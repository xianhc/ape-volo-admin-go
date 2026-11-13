package initialize

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"time"

	"github.com/xianhc/ape-volo-admin-go/internal/config"
	"github.com/xianhc/ape-volo-admin-go/internal/global"
	"gorm.io/gorm"
)

var errOracleDriverMissing = errors.New("当前构建未包含 Oracle 驱动，请使用 oracle 构建标签与 Instant Client")

// Gorm 打开所有启用连接，成功后才发布到全局；任一步失败都关闭此前连接。
//
// 日志连接只挂载审计和软删除回调，不参与业务缓存与数据权限。
func Gorm() error { return openDatabases(global.Config, openConnection) }

func openDatabases(cfg config.Server, open func(config.Connection) (*gorm.DB, error)) error {
	if err := cfg.ValidateConnections(); err != nil {
		return err
	}
	ids := make([]string, 0, len(cfg.DataConnection.Connections))
	for id, c := range cfg.DataConnection.Connections {
		if c.Enabled {
			ids = append(ids, id)
		}
	}
	sort.Strings(ids)
	ids = append(ids, "log")
	opened := make(map[string]*gorm.DB)
	fail := func(id string, c config.Connection, err error) error {
		for _, db := range opened {
			sqlDB, e := db.DB()
			if e == nil {
				_ = sqlDB.Close()
			}
		}
		if errors.Is(err, errOracleDriverMissing) {
			return fmt.Errorf("连接 %s (%s): %w", id, c.DbType, err)
		}
		// 驱动错误可能包含完整 DSN，启动错误只报告连接身份和错误类型。
		return fmt.Errorf("连接 %s (%s) 初始化失败 (%T)，请检查目标、凭据与网络", id, c.DbType, err)
	}
	for _, id := range ids {
		c := cfg.LogDataConnection
		if id != "log" {
			c = cfg.DataConnection.Connections[id]
		} else if cfg.LogStorage.TimeZone != "" {
			switch c.DbType {
			case "mysql":
				q, _ := url.ParseQuery(c.Config)
				q.Set("loc", cfg.LogStorage.TimeZone)
				c.Config = q.Encode()
			case "pgsql":
				c.Config += " TimeZone='" + cfg.LogStorage.TimeZone + "'"
			}
		}
		db, err := open(c)
		if err != nil {
			if db != nil {
				if sqlDB, e := db.DB(); e == nil {
					_ = sqlDB.Close()
				}
			}
			return fail(id, c, err)
		}
		opened[id] = db
		sqlDB, err := db.DB()
		if err != nil {
			return fail(id, c, err)
		}
		sqlDB.SetMaxIdleConns(c.MaxIdleConns)
		sqlDB.SetMaxOpenConns(c.MaxOpenConns)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		err = sqlDB.PingContext(ctx)
		cancel()
		if err != nil {
			return fail(id, c, err)
		}
		RegisterAuditFiller(db)
		RegisterSoftDeleteFilter(db)
		if id != "log" {
			RegisterDataScopeFilter(db)
			RegisterCacheInvalidator(db)
		}
	}
	global.Databases = opened
	global.Db = opened[cfg.System.MainDatabase]
	global.LogDb = opened["log"]
	return nil
}

func openConnection(c config.Connection) (*gorm.DB, error) {
	switch c.DbType {
	case "sqlite":
		return GormSqlite(c.GeneralDB)
	case "mysql":
		return GormMysql(c.GeneralDB)
	case "mssql":
		return GormMssql(c.GeneralDB)
	case "pgsql":
		return GormPgSql(c.GeneralDB)
	case "oracle":
		return GormOracle(c.GeneralDB)
	default:
		return nil, fmt.Errorf("不支持的数据库类型 %q", c.DbType)
	}
}

// CloseDatabases 在生产者和日志消费者结束后关闭全部连接。
//
// 返回所有关闭错误，并清除已发布连接引用。
func CloseDatabases() error {
	var result error
	for id, db := range global.Databases {
		sqlDB, err := db.DB()
		if err == nil {
			err = sqlDB.Close()
		}
		if err != nil {
			result = errors.Join(result, fmt.Errorf("关闭连接 %s: %w", id, err))
		}
	}
	global.Db, global.LogDb, global.Databases = nil, nil, nil
	return result
}
