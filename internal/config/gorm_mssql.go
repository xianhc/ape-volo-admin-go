package config

type Mssql struct {
	GeneralDB `yaml:",inline" mapstructure:",squash"`
}

func (m *Mssql) Dsn() string {
	return sqlServerURL(m.GeneralDB, m.Dbname)
}

// DsnNotDb 返回用于检查及创建数据库的管理连接地址。
//
// 凭据和查询参数均使用 URL 编码。
func (m *Mssql) DsnNotDb() string { return sqlServerURL(m.GeneralDB, "master") }

func (m *Mssql) GetLogMode() string {
	return m.LogMode
}
