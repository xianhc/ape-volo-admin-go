package logstore

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/xianhc/ape-volo-admin-go/internal/model"
	"github.com/xianhc/ape-volo-admin-go/internal/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

// Query 描述跨月筛选、排序和分页。
//
// DB 应为日志连接，Filter 为每张表构建相同条件。
type Query struct {
	DB            *gorm.DB
	Bounds        []time.Time
	Sorting       string
	Limit, Offset int
	Filter        func(*gorm.DB) *gorm.DB
}

type pageKey struct {
	Id       int64
	LogMonth string
}

func filtered[T any](ctx context.Context, s *Store, table string, q Query) *gorm.DB {
	db := q.DB
	if db == nil {
		db = s.DB.WithContext(ctx)
	}
	db = db.Model(new(T)).Table(table)
	if q.Filter != nil {
		db = q.Filter(db)
	}
	return db
}

// Page 只联合排序键，再按月表分批读取正文。
//
// 大文本不参与集合运算；任一月份或正文回读失败都返回错误。
func Page[T any](ctx context.Context, s *Store, prefix string, q Query) ([]T, int64, error) {
	items := make([]T, 0)
	if s == nil {
		return nil, 0, fmt.Errorf("日志存储尚未初始化")
	}
	if q.Limit < 1 || q.Limit > 100000 || q.Offset < 0 {
		return nil, 0, fmt.Errorf("日志分页范围无效")
	}
	db := q.DB
	if db == nil {
		db = s.DB.WithContext(ctx)
	}
	probe := db.Model(new(T))
	sorting := q.Sorting
	if strings.TrimSpace(sorting) == "" {
		sorting = "creation_time desc, id desc"
	}
	probe = utils.ApplySorting(probe, sorting)
	if probe.Error != nil {
		return nil, 0, probe.Error
	}
	order, ok := probe.Statement.Clauses["ORDER BY"].Expression.(clause.OrderBy)
	if !ok {
		return nil, 0, fmt.Errorf("无法确定日志排序")
	}
	stmt := &gorm.Statement{DB: s.DB}
	if err := stmt.Parse(new(T)); err != nil {
		return nil, 0, err
	}
	cols := []string{Quote(s.DB, "id")}
	seen := map[string]bool{"id": true}
	hasID := false
	for _, o := range order.Columns {
		field := stmt.Schema.LookUpField(o.Column.Name)
		if field == nil || field.FieldType == stmtTextType || field.DBName == "run_params" || field.DBName == "exception_detail" {
			return nil, 0, fmt.Errorf("日志排序不允许大文本字段")
		}
		if o.Column.Name == "id" {
			hasID = true
		}
		if !seen[o.Column.Name] {
			seen[o.Column.Name] = true
			cols = append(cols, Quote(s.DB, o.Column.Name))
		}
	}
	if !hasID {
		order.Columns = append(order.Columns, clause.OrderByColumn{Column: clause.Column{Name: "id"}, Desc: true})
	}
	tables, err := s.Tables(ctx, prefix, q.Bounds)
	if err != nil {
		return nil, 0, err
	}
	if len(tables) == 0 {
		return items, 0, nil
	}
	var total int64
	args := make([]any, 0, len(tables))
	parts := make([]string, 0, len(tables))
	for _, table := range tables {
		var count int64
		if err := filtered[T](ctx, s, table, q).Count(&count).Error; err != nil {
			return nil, 0, err
		}
		total += count
		sub := filtered[T](ctx, s, table, q).Select(strings.Join(cols, ",")+", ? AS "+Quote(s.DB, "log_month"), table).Session(&gorm.Session{DryRun: true}).Find(&[]T{})
		if sub.Error != nil {
			return nil, 0, sub.Error
		}
		args = append(args, sub)
		parts = append(parts, "?")
	}
	for offset := 0; offset < q.Limit; {
		size := min(500, q.Limit-offset)
		keys := make([]pageKey, 0, size)
		outer := db.Table("("+strings.Join(parts, " UNION ALL ")+") "+Quote(s.DB, "log_keys"), args...).Clauses(order).Limit(size).Offset(q.Offset + offset)
		if err := outer.Select(Quote(s.DB, "id") + ", " + Quote(s.DB, "log_month")).Find(&keys).Error; err != nil {
			return nil, 0, err
		}
		if len(keys) == 0 {
			break
		}
		groups := make(map[string][]int64)
		for _, key := range keys {
			groups[key.LogMonth] = append(groups[key.LogMonth], key.Id)
		}
		rows := make(map[pageKey]T, len(keys))
		for table, ids := range groups {
			for start := 0; start < len(ids); start += 200 {
				var batch []T
				if err := filtered[T](ctx, s, table, q).Where(clause.IN{Column: clause.Column{Name: "id"}, Values: intValues(ids[start:min(start+200, len(ids))])}).Find(&batch).Error; err != nil {
					return nil, 0, err
				}
				for _, row := range batch {
					if err := normalizeTimes(ctx, &row, stmt.Schema, s.Location); err != nil {
						return nil, 0, err
					}
					id, _ := stmt.Schema.PrioritizedPrimaryField.ValueOf(ctx, reflect.ValueOf(row))
					rows[pageKey{Id: id.(int64), LogMonth: table}] = row
				}
			}
		}
		for _, key := range keys {
			row, ok := rows[key]
			if !ok {
				return nil, 0, fmt.Errorf("日志分页记录在回读前变化，请重试")
			}
			items = append(items, row)
		}
		offset += len(keys)
		if len(keys) < size {
			break
		}
	}
	return items, total, nil
}

func intValues(ids []int64) []any {
	values := make([]any, len(ids))
	for i, id := range ids {
		values[i] = id
	}
	return values
}

var stmtTextType = reflect.TypeOf(model.CustomFieldText(""))

func normalizeTimes[T any](ctx context.Context, row *T, parsed *schema.Schema, loc *time.Location) error {
	rv := reflect.ValueOf(row).Elem()
	for _, field := range parsed.Fields {
		value, zero := field.ValueOf(ctx, rv)
		if zero {
			continue
		}
		switch v := value.(type) {
		case model.Time:
			if err := field.Set(ctx, rv, model.Time{Time: v.In(loc)}); err != nil {
				return err
			}
		case *model.Time:
			if v != nil {
				at := model.Time{Time: v.In(loc)}
				if err := field.Set(ctx, rv, &at); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
