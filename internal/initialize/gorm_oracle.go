//go:build !oracle

package initialize

import (
	"github.com/xianhc/ape-volo-admin-go/internal/config"
	"gorm.io/gorm"
)

// GormOracle 在普通构建中明确拒绝 Oracle 配置。
//
// 防止把不具备的驱动能力降级为其他数据库。
func GormOracle(config.GeneralDB) (*gorm.DB, error) {
	return nil, errOracleDriverMissing
}
