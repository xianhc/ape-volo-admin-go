package config

import (
	"path/filepath"
)

type Sqlite struct {
	GeneralDB `yaml:",inline" mapstructure:",squash"`
}

func (s *Sqlite) Dsn() string {
	return filepath.Join(s.Host, s.Dbname+".db")
}

func (s *Sqlite) GetLogMode() string {
	return s.LogMode
}
