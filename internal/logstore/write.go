package logstore

import (
	"context"
	"fmt"
	"sort"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Record 暴露写入月份和错误摘要所需的轻量快照字段。
//
// 生产端必须在入队前确定这两个值。
type Record interface {
	LogID() int64
	LogTime() time.Time
}

// Outcome 区分一个消费批次内各月份的最终处理数量。
//
// 提交结果未知与未开始的停机遗留分开计数。
type Outcome struct{ Persisted, Failed, Unknown, Remaining int }

// WriteBatch 按生产月份独立提交日志批次。
//
// 某个月失败不阻止其他月份，错误摘要不含正文和驱动原始错误。
func WriteBatch[T Record](ctx context.Context, s *Store, prefix string, entries []T, timeout time.Duration, logger *zap.Logger) Outcome {
	var result Outcome
	if s == nil {
		result.Failed = len(entries)
		return result
	}
	groups := make(map[string][]T)
	for _, entry := range entries {
		table, err := MonthTable(prefix, entry.LogTime(), s.Location)
		if err != nil {
			result.Failed++
			continue
		}
		groups[table] = append(groups[table], entry)
	}
	tables := make([]string, 0, len(groups))
	for table := range groups {
		tables = append(tables, table)
	}
	sort.Strings(tables)
	for _, table := range tables {
		batch := groups[table]
		if ctx.Err() != nil {
			result.Remaining += len(batch)
			continue
		}
		writeCtx, cancel := context.WithTimeout(ctx, timeout)
		unknown := false
		_, err := s.EnsureMonth(writeCtx, prefix, batch[0].LogTime())
		if err == nil {
			stmt := &gorm.Statement{DB: s.DB}
			err = stmt.Parse(new(T))
			if err == nil {
				size := max(1, 900/len(stmt.Schema.DBNames))
				if s.DB.CreateBatchSize > 0 {
					size = min(size, s.DB.CreateBatchSize)
				}
				err, unknown = insertBatch(writeCtx, s.DB, table, batch, size)
			}
		}
		if err == nil {
			result.Persisted += len(batch)
		} else {
			if unknown {
				result.Unknown += len(batch)
			} else {
				result.Failed += len(batch)
			}
			if logger != nil {
				logger.Error("日志月份批次写入失败", zap.String("type", prefix), zap.String("month", table), zap.Int("count", len(batch)), zap.Int64("first_id", batch[0].LogID()), zap.Int64("last_id", batch[len(batch)-1].LogID()), zap.String("error_type", fmt.Sprintf("%T", err)), zap.Bool("commit_unknown", unknown))
			}
		}
		cancel()
	}
	return result
}

func insertBatch[T any](ctx context.Context, db *gorm.DB, table string, batch []T, size int) (error, bool) {
	// 驱动错误可能包含整行内容，写入失败仅由上层输出脱敏摘要。
	tx := db.Session(&gorm.Session{Logger: logger.Discard}).WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error, false
	}
	defer tx.Rollback()
	if err := tx.Table(table).CreateInBatches(&batch, size).Error; err != nil {
		return err, false
	}
	err := tx.Commit().Error
	return err, err != nil
}
