package utils

import "context"

// contextKey 是本包私有的上下文键类型，避免与其他包的键冲突。
type contextKey string

const operatorIDKey contextKey = "apevolo.operatorId"

// roleCodesKey 保存当前用户的角色码快照。
const roleCodesKey contextKey = "apevolo.roleCodes"

// dataScopeSkipKey 标记本次查询不做数据权限过滤。
const dataScopeSkipKey contextKey = "apevolo.skipDataScope"

// txKey 保存请求级事务句柄。
const txKey contextKey = "apevolo.tx"

// noTxKey 标记本次读写不使用请求级事务。
const noTxKey contextKey = "apevolo.noTx"

// WithOperatorID 把当前操作人写入请求上下文。
//
// 之所以要进 context 而不是只放在 gin.Context：ORM 的写回调只能拿到
// 数据库连接的 context，拿不到 gin 的键值对，审计字段要在那里自动填充。
func WithOperatorID(ctx context.Context, id int64) context.Context {
	return context.WithValue(ctx, operatorIDKey, id)
}

// OperatorID 从上下文取当前操作人。
//
// 第二个返回值为 false 表示没有登录用户（定时任务、启动初始化等），
// 此时审计字段不应写入用户 ID。
func OperatorID(ctx context.Context) (int64, bool) {
	if ctx == nil {
		return 0, false
	}
	id, ok := ctx.Value(operatorIDKey).(int64)
	if !ok || id <= 0 {
		return 0, false
	}
	return id, true
}

// WithRoleCodes 把当前用户的角色码快照写入请求上下文。
//
// 方法复制输入切片，避免后续复用缓存结果时修改当前请求的授权判断。
func WithRoleCodes(ctx context.Context, codes []string) context.Context {
	return context.WithValue(ctx, roleCodesKey, append([]string(nil), codes...))
}

// HasRole 判断请求上下文中的当前用户是否具有指定角色码。
//
// 上下文没有角色快照时返回 false，敏感业务方法因此保持 fail closed。
func HasRole(ctx context.Context, code string) bool {
	if ctx == nil || code == "" {
		return false
	}
	codes, _ := ctx.Value(roleCodesKey).([]string)
	for _, current := range codes {
		if current == code {
			return true
		}
	}
	return false
}

// WithoutDataScope 标记本次查询不做数据权限过滤。
//
// 用于两类场景：解析数据权限本身（否则会递归），以及必须读全量数据的内部查询
// （登录校验、权限码与菜单树计算）。业务查询不要用它绕过权限。
func WithoutDataScope(ctx context.Context) context.Context {
	return context.WithValue(ctx, dataScopeSkipKey, true)
}

// WithTx 把请求级事务写入上下文。
//
// 由事务管理入口在进入业务处理前写入，业务代码不直接调用。
func WithTx(ctx context.Context, tx any) context.Context {
	ctx = context.WithValue(ctx, txKey, tx)
	ctx = context.WithValue(ctx, noTxKey, false)
	return context.WithValue(ctx, txEffectsKey, &txEffects{keys: make(map[string]struct{})})
}

// WithoutTx 返回不带请求级事务的上下文。
//
// 用于必须独立于业务事务落库的记录（操作日志、异常日志）：
// 业务回滚时这些记录仍要保留，否则「失败的操作」反而查不到。
func WithoutTx(ctx context.Context) context.Context {
	return context.WithValue(ctx, noTxKey, true)
}

// TxSkipped 判断本次读写是否已被标记为不使用请求级事务。
func TxSkipped(ctx context.Context) bool {
	if ctx == nil {
		return false
	}
	skip, _ := ctx.Value(noTxKey).(bool)
	return skip
}

// TxFromContext 取请求级事务，没有则返回 nil。
func TxFromContext(ctx context.Context) any {
	if ctx == nil {
		return nil
	}
	return ctx.Value(txKey)
}

// DataScopeSkipped 判断本次查询是否已被标记为不做数据权限过滤。
func DataScopeSkipped(ctx context.Context) bool {
	if ctx == nil {
		return false
	}
	skip, _ := ctx.Value(dataScopeSkipKey).(bool)
	return skip
}
