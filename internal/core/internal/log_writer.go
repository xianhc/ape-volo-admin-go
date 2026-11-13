package internal

import (
	"errors"
	"os"
	"sync"

	"github.com/xianhc/ape-volo-admin-go/internal/global"
	"go.uber.org/zap/zapcore"
)

var LogWriter = new(logWriter)

type logWriter struct {
	mu      sync.Mutex
	writers []*DailyWriter
}

// GetWriteSyncer 获取 zapcore.WriteSyncer。
// 日志按天写入 <director>/<YYYY-MM-DD>/<level>.log，单文件超过 zap.max-size 时切分序号。
func (l *logWriter) GetWriteSyncer(level string) (zapcore.WriteSyncer, error) {
	daily := NewDailyWriter(
		global.Config.Zap.Director,
		level,
		global.Config.Zap.MaxAge,
		global.Config.Zap.MaxSize,
	)
	l.mu.Lock()
	l.writers = append(l.writers, daily)
	l.mu.Unlock()
	syncer := zapcore.AddSync(daily)

	if global.Config.Zap.LogInConsole {
		return zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), syncer), nil
	}
	return syncer, nil
}

func (l *logWriter) Close() error {
	l.mu.Lock()
	writers := l.writers
	l.writers = nil
	l.mu.Unlock()
	var failures []error
	for _, writer := range writers {
		if err := writer.Close(); err != nil {
			failures = append(failures, err)
		}
	}
	return errors.Join(failures...)
}
