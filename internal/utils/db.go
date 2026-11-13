package utils

import (
	"context"

	"github.com/xianhc/ape-volo-admin-go/internal/global"
	"gorm.io/gorm"
)

// FromContext 返回绑定了请求上下文的数据库句柄。
//
// 为什么不用 global.Db：
//   - 客户端断开或超时后，绑定了 context 的 SQL 会被中断，不会白跑；
//   - ORM 写回调能通过 Statement.Context 读到操作人，审计字段不必在入口手工写；
//   - 请求处于事务中时直接复用该事务，多个 service 的写入落在同一事务里。
//
// ctx 为 nil 时退化为 global.Db，避免调用方漏传导致 panic。
func FromContext(ctx context.Context) *gorm.DB {
	if ctx == nil {
		return global.Db
	}
	if !TxSkipped(ctx) {
		if tx, ok := TxFromContext(ctx).(*gorm.DB); ok && tx != nil {
			return tx.Session(&gorm.Session{NewDB: true, Context: ctx})
		}
	}
	return global.Db.WithContext(ctx)
}

// LogDB 始终使用独立日志连接，不复用业务事务。
//
// 查询保留请求取消，后台写入使用队列创建的独立 context。
func LogDB(ctx context.Context) *gorm.DB {
	if ctx == nil {
		ctx = context.Background()
	}
	return global.LogDb.WithContext(ctx)
}
