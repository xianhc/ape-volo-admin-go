package config

type Mssql struct {
	GeneralDB `yaml:",inline" mapstructure:",squash"`
}

func (m *Mssql) Dsn() string {
	return "sqlserver://" + m.Username + ":" + m.Password + "@" + m.Host + ":" + m.Port + "?database=" + m.Dbname + "&encrypt=disable"
}

func (m *Mssql) GetLogMode() string {
	return m.LogMode
}
