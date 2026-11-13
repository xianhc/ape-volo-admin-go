package utils

import (
	"context"
	"database/sql"
	"errors"
	"sync"
	"time"

	"github.com/xianhc/ape-volo-admin-go/internal/global"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const txEffectsKey contextKey = "apevolo.txEffects"

type txEffects struct {
	mu               sync.Mutex
	keys             map[string]struct{}
	commit, rollback []func(context.Context)
	finished         bool
}

// InTransaction 在同一事务中执行多步业务操作，已有事务时复用。
//
// 提交、回滚及其副作用均由最外层调用负责。副作用在连接释放后执行，
// 避免日志或补偿操作等待仍被业务事务占用的连接。
func InTransaction(ctx context.Context, run func(context.Context) error) (err error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if tx, ok := TxFromContext(ctx).(*gorm.DB); ok && tx != nil && !TxSkipped(ctx) {
		err = run(ctx)
		if err != nil {
			tx.AddError(err)
		}
		return err
	}
	var txCtx context.Context
	committed := false
	defer func() {
		if txCtx != nil {
			finishEffects(txCtx, committed)
		}
	}()
	return global.Db.WithContext(ctx).Connection(func(connection *gorm.DB) (err error) {
		tx := connection.Begin()
		if tx.Error != nil {
			return tx.Error
		}
		txCtx = WithTx(ctx, tx)
		defer func() {
			if !committed {
				rollbackTransaction(tx)
			}
		}()
		err = run(txCtx)
		if err == nil {
			err = tx.Error
		}
		if err != nil {
			return err
		}
		if err = tx.Commit().Error; err != nil {
			// SQLite 的延迟约束可能在 COMMIT 时失败，需要在原连接清理。
			if connection.Dialector.Name() == "sqlite" {
				if cleanupErr := connection.WithContext(context.WithoutCancel(ctx)).Exec("ROLLBACK").Error; cleanupErr != nil {
					global.Logger.Warn("清理失败事务的连接失败", zap.Error(cleanupErr))
				}
			}
			return err
		}
		committed = true
		return nil
	})
}

func rollbackTransaction(tx *gorm.DB) {
	if err := tx.Rollback().Error; err != nil && !errors.Is(err, context.Canceled) && !errors.Is(err, sql.ErrTxDone) {
		global.Logger.Warn("事务回滚失败", zap.Error(err))
	}
}

// AfterCommit 在提交成功后执行；没有业务事务时立即执行。
// 非空 key 用于合并同一事务内重复登记的副作用。
func AfterCommit(ctx context.Context, key string, fn func(context.Context)) {
	if effects := effectsFromContext(ctx); effects != nil {
		effects.mu.Lock()
		defer effects.mu.Unlock()
		if effects.finished {
			return
		}
		if key != "" {
			if _, exists := effects.keys[key]; exists {
				return
			}
			effects.keys[key] = struct{}{}
		}
		effects.commit = append(effects.commit, fn)
		return
	}
	runEffect(ctx, fn)
}

// AfterRollback 登记失败补偿，没有业务事务时由调用方处理当前操作错误。
func AfterRollback(ctx context.Context, fn func(context.Context)) {
	if effects := effectsFromContext(ctx); effects != nil {
		effects.mu.Lock()
		defer effects.mu.Unlock()
		if !effects.finished {
			effects.rollback = append(effects.rollback, fn)
		}
	}
}

func effectsFromContext(ctx context.Context) *txEffects {
	if ctx == nil || TxSkipped(ctx) {
		return nil
	}
	effects, _ := ctx.Value(txEffectsKey).(*txEffects)
	return effects
}

func finishEffects(ctx context.Context, committed bool) {
	effects := effectsFromContext(ctx)
	if effects == nil {
		return
	}
	effects.mu.Lock()
	if effects.finished {
		effects.mu.Unlock()
		return
	}
	effects.finished = true
	fns := effects.rollback
	if committed {
		fns = effects.commit
	}
	effects.mu.Unlock()
	for _, fn := range fns {
		runEffect(ctx, fn)
	}
}

func runEffect(ctx context.Context, fn func(context.Context)) {
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithTimeout(WithoutTx(context.WithoutCancel(ctx)), 5*time.Second)
	defer cancel()
	defer func() {
		if r := recover(); r != nil {
			global.Logger.Error("事务后处理失败", zap.Any("panic", r))
		}
	}()
	fn(ctx)
}
