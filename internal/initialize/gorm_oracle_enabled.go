//go:build oracle

package initialize

import (
	"database/sql"
	"strings"

	"github.com/oracle-samples/gorm-oracle/oracle"
	"github.com/xianhc/ape-volo-admin-go/internal/config"
	"github.com/xianhc/ape-volo-admin-go/internal/initialize/internal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// GormOracle 的可选构建需要 CGO 与运行期 Instant Client。
//
// 数据库服务及用户由部署预先准备，不自动建库。
func GormOracle(c config.GeneralDB) (*gorm.DB, error) {
	o := config.Oracle{GeneralDB: c}
	cfg := internal.Gorm.Config(c)
	cfg.CreateBatchSize = 1
	return gorm.Open(oracleDialect{Dialector: &oracle.Dialector{Config: &oracle.Config{DataSourceName: o.Dsn(), SkipQuoteIdentifiers: true}}}, cfg)
}

type oracleDialect struct{ *oracle.Dialector }

func (dialect oracleDialect) QuoteTo(writer clause.Writer, name string) {
	parts := strings.Split(name, ".")
	for index, part := range parts {
		if index > 0 {
			writer.WriteByte('.')
		}
		if part == "*" {
			writer.WriteByte('*')
			continue
		}
		writer.WriteByte('"')
		writer.WriteString(strings.ReplaceAll(strings.ToUpper(strings.Trim(part, "\"")), "\"", "\"\""))
		writer.WriteByte('"')
	}
}

func (dialect oracleDialect) Migrator(db *gorm.DB) gorm.Migrator {
	return oracleMigrator{Migrator: dialect.Dialector.Migrator(db).(oracle.Migrator), db: db}
}

type oracleMigrator struct {
	oracle.Migrator
	db *gorm.DB
}

type oracleColumnMetadata interface{ gorm.ColumnType }

type oracleColumn struct {
	oracleColumnMetadata
	defaultValue sql.NullString
	charLength   int64
}

func (column oracleColumn) Name() string {
	return strings.ToLower(column.oracleColumnMetadata.Name())
}

func (column oracleColumn) DefaultValue() (string, bool) {
	value := strings.TrimSpace(column.defaultValue.String)
	return value, column.defaultValue.Valid && value != "" && !strings.EqualFold(value, "NULL")
}

func (column oracleColumn) Length() (int64, bool) {
	if column.charLength > 0 {
		return column.charLength, true
	}
	return column.oracleColumnMetadata.Length()
}

func (migration oracleMigrator) ColumnTypes(value interface{}) ([]gorm.ColumnType, error) {
	columns, err := migration.Migrator.ColumnTypes(value)
	if err != nil {
		return nil, err
	}
	statement := &gorm.Statement{DB: migration.db, Table: migration.db.Statement.Table}
	if table, ok := value.(string); ok {
		statement.Table = table
	} else {
		if err := statement.Parse(value); err != nil {
			return nil, err
		}
	}
	var defaults []struct {
		Name         string
		DefaultValue sql.NullString
		CharLength   int64
	}
	if err := migration.db.Raw(`SELECT column_name AS "name", data_default_vc AS "default_value", char_length AS "char_length" FROM user_tab_columns WHERE table_name = ?`, strings.ToUpper(statement.Table)).Scan(&defaults).Error; err != nil {
		return nil, err
	}
	byName := make(map[string]oracleColumn, len(defaults))
	for _, column := range defaults {
		byName[strings.ToUpper(column.Name)] = oracleColumn{defaultValue: column.DefaultValue, charLength: column.CharLength}
	}
	for index, column := range columns {
		metadata := byName[strings.ToUpper(column.Name())]
		metadata.oracleColumnMetadata = column
		columns[index] = metadata
	}
	return columns, nil
}

func (migration oracleMigrator) HasTable(value interface{}) bool {
	statement := &gorm.Statement{DB: migration.db, Table: migration.db.Statement.Table}
	if table, ok := value.(string); ok {
		statement.Table = table
	} else if err := statement.Parse(value); err != nil {
		return false
	}
	var count int64
	err := migration.db.Raw("SELECT COUNT(*) FROM user_tables WHERE table_name = ?", strings.ToUpper(statement.Table)).Row().Scan(&count)
	return err == nil && count > 0
}

func (migration oracleMigrator) HasIndex(value interface{}, name string) bool {
	statement := &gorm.Statement{DB: migration.db, Table: migration.db.Statement.Table}
	if table, ok := value.(string); ok {
		statement.Table = table
	} else if err := statement.Parse(value); err != nil {
		return false
	} else if index := statement.Schema.LookIndex(name); index != nil {
		name = index.Name
	}
	var count int64
	err := migration.db.Raw("SELECT COUNT(*) FROM user_indexes WHERE table_name = ? AND index_name = ?", strings.ToUpper(statement.Table), strings.ToUpper(name)).Row().Scan(&count)
	return err == nil && count > 0
}
