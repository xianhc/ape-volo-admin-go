package initialize

import (
	"context"

	"github.com/xianhc/ape-volo-admin-go/internal/global"
	"github.com/xianhc/ape-volo-admin-go/internal/global/constants/cachePrefix"
	"github.com/xianhc/ape-volo-admin-go/internal/utils"
	"github.com/xianhc/ape-volo-admin-go/internal/utils/querycache"
	"gorm.io/gorm"
)

// 表名 → 需要失效的缓存前缀。
//
// 只登记「写入后会导致缓存内容过期」的表：审计日志、任务日志这类只追加、
// 不进缓存的表不登记，回调里查不到就直接返回，避免无谓的 SCAN。
var tableCachePrefixes = map[string][]string{
	"sys_user": {
		cachePrefix.UserRoleCodes,
		cachePrefix.UserInfoById,
		cachePrefix.UserAuthCodes,
		cachePrefix.UserMenuById,
		cachePrefix.UserDataScopeById,
	},
	"sys_user_role": {
		cachePrefix.UserRoleCodes,
		cachePrefix.UserInfoById,
		cachePrefix.UserAuthCodes,
		cachePrefix.UserMenuById,
		cachePrefix.UserDataScopeById,
	},
	"sys_user_position": {
		cachePrefix.UserInfoById,
	},
	"sys_role": {
		cachePrefix.UserInfoById,
		cachePrefix.UserRoleCodes,
		cachePrefix.UserAuthCodes,
		cachePrefix.UserMenuById,
		cachePrefix.UserDataScopeById,
	},
	"sys_role_menu": {
		cachePrefix.UserAuthCodes,
		cachePrefix.UserMenuById,
	},
	"sys_role_dept": {
		cachePrefix.UserAuthCodes,
		cachePrefix.UserDataScopeById,
	},
	"sys_menu": {
		cachePrefix.UserMenuById,
		cachePrefix.UserAuthCodes,
	},
	"sys_department": {
		cachePrefix.UserInfoById,
		cachePrefix.UserMenuById,
		cachePrefix.UserDataScopeById,
	},
	"sys_setting": {
		cachePrefix.LoadSettingByName,
	},
	"sys_dict": {
		cachePrefix.LoadDictByName,
	},
	"sys_dict_detail": {
		cachePrefix.LoadDictByName,
	},
}

// RegisterCacheInvalidator 在事务提交成功后按表名失效相关查询缓存。
//
// 为什么用回调而不是在每个 service 里手工删缓存：
//   - 写操作有 70+ 处（Create/Update/Delete），逐个补漏几乎必然遗漏；
//   - 关联表（sys_role_menu / sys_user_role）会被 GORM 的 many2many 直接写入，
//     service 层根本感知不到；
//   - 回调拿到的 db.Statement.Table 是最终表名，覆盖批量更新与关联写入。
//
// 同一业务事务内按前缀合并失效；未包业务事务的写入在默认事务提交后失效。
// 缓存尚未初始化时跳过，避免启动种子阶段访问缓存。
func RegisterCacheInvalidator(db *gorm.DB) {
	invalidate := func(tx *gorm.DB) {
		if tx.Error != nil || tx.Statement == nil {
			return
		}
		if global.Cache == nil {
			return
		}
		for _, prefix := range tableCachePrefixes[tx.Statement.Table] {
			utils.AfterCommit(tx.Statement.Context, "cache:"+prefix, func(ctx context.Context) {
				querycache.Invalidate(ctx, prefix)
			})
		}
	}
	_ = db.Callback().Create().After("gorm:commit_or_rollback_transaction").Register("ape:cache:invalidate", invalidate)
	_ = db.Callback().Update().After("gorm:commit_or_rollback_transaction").Register("ape:cache:invalidate", invalidate)
	_ = db.Callback().Delete().After("gorm:commit_or_rollback_transaction").Register("ape:cache:invalidate", invalidate)
}
