package config

import "strings"

type Pgsql struct {
	GeneralDB `yaml:",inline" mapstructure:",squash"`
}

// Dsn 基于配置文件获取 dsn
func (p *Pgsql) Dsn() string {
	return p.LinkDsn(p.Dbname)
}

// LinkDsn 根据 dbname 生成 dsn
func (p *Pgsql) LinkDsn(dbname string) string {
	quote := func(s string) string { return "'" + strings.NewReplacer("\\", "\\\\", "'", "\\'").Replace(s) + "'" }
	return "connect_timeout=10 " + p.Config + " host=" + quote(p.Host) + " user=" + quote(p.Username) + " password=" + quote(p.Password) + " dbname=" + quote(dbname) + " port=" + quote(p.Port)
}

func (m *Pgsql) GetLogMode() string {
	return m.LogMode
}
