package config

import (
	"net"
	"strconv"
)

type Oracle struct {
	GeneralDB `yaml:",inline" mapstructure:",squash"`
}

func (m *Oracle) Dsn() string {
	return m.Config + " user=" + strconv.Quote(m.Username) + " password=" + strconv.Quote(m.Password) + " connectString=" + strconv.Quote(net.JoinHostPort(m.Host, m.Port)+"/"+m.Dbname)

}

func (m *Oracle) GetLogMode() string {
	return m.LogMode
}
