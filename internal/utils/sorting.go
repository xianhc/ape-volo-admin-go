package utils

import (
	"errors"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ApplySorting 把客户端排序转换为带标识符转义的排序子句。
//
// 只接受当前模型的数据库列名或字段名，以及可选的 asc / desc 方向；
// 不允许表达式、注释或任意 SQL，非法排序通过查询的 Error 返回。
func ApplySorting(db *gorm.DB, sorting string) *gorm.DB {
	if strings.TrimSpace(sorting) == "" || db.Error != nil {
		return db
	}
	stmt := &gorm.Statement{DB: db}
	if err := stmt.Parse(db.Statement.Model); err != nil {
		_ = db.AddError(errors.New("无法确定排序字段"))
		return db
	}
	columns := make(map[string]string)
	for _, field := range stmt.Schema.Fields {
		if field.DBName != "" && field.Readable {
			columns[strings.ToLower(field.DBName)] = field.DBName
			columns[strings.ToLower(field.Name)] = field.DBName
		}
	}
	parts := strings.Split(sorting, ",")
	if len(parts) > 8 {
		_ = db.AddError(errors.New("排序字段不能超过8个"))
		return db
	}
	orders := make([]clause.OrderByColumn, 0, len(parts))
	for _, part := range parts {
		fields := strings.Fields(part)
		if len(fields) < 1 || len(fields) > 2 {
			_ = db.AddError(errors.New("排序格式不正确"))
			return db
		}
		column, ok := columns[strings.ToLower(fields[0])]
		if !ok {
			_ = db.AddError(errors.New("排序字段不允许"))
			return db
		}
		desc := false
		if len(fields) == 2 {
			switch strings.ToLower(fields[1]) {
			case "asc":
			case "desc":
				desc = true
			default:
				_ = db.AddError(errors.New("排序方向只允许asc或desc"))
				return db
			}
		}
		orders = append(orders, clause.OrderByColumn{Column: clause.Column{Name: column}, Desc: desc})
	}
	return db.Clauses(clause.OrderBy{Columns: orders})
}
