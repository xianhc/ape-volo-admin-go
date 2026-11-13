package config

import (
	"github.com/go-sql-driver/mysql"
	"net"
	"time"
)

type Mysql struct {
	GeneralDB `yaml:",inline" mapstructure:",squash"`
}

func (m *Mysql) Dsn() string {
	return m.buildDSN(m.Dbname)
}

func (m *Mysql) DsnNotDb() string {
	return m.buildDSN("")
}

func (m *Mysql) buildDSN(database string) string {
	c := mysql.NewConfig()
	c.User, c.Passwd, c.Net, c.Addr, c.DBName = m.Username, m.Password, "tcp", net.JoinHostPort(m.Host, m.Port), database
	c.ParseTime = true
	c.Timeout, c.ReadTimeout, c.WriteTimeout = 10*time.Second, 10*time.Second, 10*time.Second
	// 使用驱动解析高级参数，凭据和数据库名仍由驱动编码。
	if m.Config != "" {
		if parsed, err := mysql.ParseDSN(c.FormatDSN() + "&" + m.Config); err == nil {
			c = parsed
		}
	}
	return c.FormatDSN()
}

func (m *Mysql) GetLogMode() string {
	return m.LogMode
}
