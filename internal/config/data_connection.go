package config

import (
	"fmt"
	"github.com/go-sql-driver/mysql"
	"net"
	"net/url"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

// DataConnection 保存具名业务连接，日志使用单独的固定连接。
//
// 映射键是可供环境变量覆盖的稳定连接 ID。
type DataConnection struct {
	Connections map[string]Connection `mapstructure:"connections" json:"connections" yaml:"connections"`
}

// Connection 定义一个可独立启用的数据库连接。
//
// 驱动由 DbType 决定，不能从连接 ID 推断。
type Connection struct {
	GeneralDB `mapstructure:",squash" yaml:",inline"`
	Enabled   bool   `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
	DbType    string `mapstructure:"db-type" json:"db-type" yaml:"db-type"`
}

// Validate 检查启用连接必需的参数。
//
// 不建立网络连接，不把凭据包含在错误中。
func (c Connection) Validate() error {
	switch c.DbType {
	case "sqlite", "mysql", "mssql", "pgsql", "oracle":
	default:
		return fmt.Errorf("不支持的 db-type %q", c.DbType)
	}
	if strings.TrimSpace(c.Dbname) == "" {
		return fmt.Errorf("db-name 不能为空")
	}
	if c.MaxIdleConns < 0 || c.MaxOpenConns < 1 || c.MaxIdleConns > c.MaxOpenConns {
		return fmt.Errorf("连接池须满足 0 <= max-idle-conns <= max-open-conns 且 max-open-conns > 0")
	}
	if c.DbType != "sqlite" {
		port, err := strconv.Atoi(c.Port)
		if strings.TrimSpace(c.Host) == "" || strings.TrimSpace(c.Username) == "" || err != nil || port < 1 || port > 65535 {
			return fmt.Errorf("host、username 和有效 port 必须配置")
		}
	}
	if c.DbType == "mysql" && c.Engine != "" && c.Engine != "InnoDB" && c.Engine != "MyISAM" {
		return fmt.Errorf("不支持的 MySQL engine")
	}
	if c.DbType == "mysql" {
		if _, err := mysql.ParseDSN("tcp(localhost:3306)/?" + c.Config); err != nil {
			return fmt.Errorf("mysql.config 无效")
		}
	}
	return nil
}

// Target 返回规范化数据库目标，供主库与日志库冲突检查使用。
//
// 网络和文件系统别名仍需部署核验。
func (c Connection) Target() (string, error) {
	if c.DbType == "sqlite" {
		p, err := filepath.Abs((&Sqlite{GeneralDB: c.GeneralDB}).Dsn())
		if err != nil {
			return "", err
		}
		if runtime.GOOS == "windows" {
			p = strings.ToLower(p)
		}
		return "sqlite:" + filepath.Clean(p), nil
	}
	host := strings.ToLower(strings.TrimSuffix(c.Host, "."))
	if host == "localhost" || host == "::1" {
		host = "127.0.0.1"
	}
	// 同库不同 schema 也拒绝，避免连接级默认 schema 变化导致串写。
	return c.DbType + ":" + net.JoinHostPort(host, c.Port) + "/" + strings.ToLower(c.Dbname), nil
}

// ValidateConnections 检查主库选择及独立日志连接。
//
// 关闭的业务连接不检查凭据和连接池，但保留 ID 仍不可占用。
func (s Server) ValidateConnections() error {
	if _, ok := s.DataConnection.Connections["log"]; ok {
		return fmt.Errorf("log 是保留连接 ID")
	}
	main, ok := s.DataConnection.Connections[s.System.MainDatabase]
	if !ok || !main.Enabled {
		return fmt.Errorf("主库 %q 不存在或未启用", s.System.MainDatabase)
	}
	if !s.LogDataConnection.Enabled {
		return fmt.Errorf("日志连接必须启用")
	}
	for id, c := range s.DataConnection.Connections {
		if strings.TrimSpace(id) == "" || strings.ContainsAny(id, ". ") {
			return fmt.Errorf("连接 ID 无效")
		}
		if c.Enabled {
			if err := c.Validate(); err != nil {
				return fmt.Errorf("连接 %s: %w", id, err)
			}
		}
	}
	if err := s.LogDataConnection.Validate(); err != nil {
		return fmt.Errorf("连接 log: %w", err)
	}
	a, err := main.Target()
	if err != nil {
		return err
	}
	b, err := s.LogDataConnection.Target()
	if err != nil {
		return err
	}
	if a == b {
		return fmt.Errorf("主库与日志库不能指向同一数据库目标")
	}
	return nil
}

func sqlServerURL(c GeneralDB, database string) string {
	u := url.URL{Scheme: "sqlserver", User: url.UserPassword(c.Username, c.Password), Host: net.JoinHostPort(c.Host, c.Port)}
	q, _ := url.ParseQuery(c.Config)
	if q.Get("encrypt") == "" {
		q.Set("encrypt", "disable")
	}
	q.Set("database", database)
	if q.Get("connection timeout") == "" {
		q.Set("connection timeout", "10")
	}
	u.RawQuery = q.Encode()
	return u.String()
}
