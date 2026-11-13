package logqueue

import (
	"context"
	"time"
)

func (q *Queue[T]) consume(ctx context.Context, items <-chan T, done chan struct{}, options Options, write func(context.Context, []T) Result) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	defer close(done)
	defer q.cancel()
	last := time.Now()
	report := func() {
		if options.Summary != nil {
			options.Summary(q.Stats())
		}
		last = time.Now()
	}
	defer report()
	batch := make([]T, 0, options.BatchSize)
	for {
		if ctx.Err() != nil {
			q.remaining.Add(int64(len(items)))
			return
		}
		select {
		case <-ctx.Done():
			q.remaining.Add(int64(len(items)))
			return
		case <-ticker.C:
			report()
			continue
		case value, ok := <-items:
			if !ok {
				return
			}
			batch = append(batch[:0], value)
		}
	fill:
		for len(batch) < options.BatchSize {
			select {
			case value, ok := <-items:
				if !ok {
					break fill
				}
				batch = append(batch, value)
			default:
				break fill
			}
		}
		q.inFlight.Store(int64(len(batch)))
		result := write(ctx, batch)
		q.persisted.Add(int64(result.Persisted))
		q.failed.Add(int64(result.Failed))
		q.unknown.Add(int64(result.Unknown))
		q.remaining.Add(int64(result.Remaining))
		q.inFlight.Store(0)
		clear(batch)
		if time.Since(last) >= 10*time.Second {
			report()
		}
	}
}
