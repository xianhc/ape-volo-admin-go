package logstore

import (
	"context"
	"crypto/sha256"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"
)

// Schema 指定一种日志的模型及月份索引。
//
// 模型由装配层传入，共享存储不依赖业务模块。
type Schema struct {
	Prefix  string
	Model   any
	Indexes [][]string
}

// Store 管理独立日志连接的月表元数据。
//
// 建表与表清单刷新串行化，查询不持有跨月事务。
type Store struct {
	DB        *gorm.DB
	Location  *time.Location
	allowDDL  bool
	schemas   map[string]Schema
	mu        sync.Mutex
	tables    []string
	refreshed time.Time
	ready     map[string]bool
}

// New 创建待初始化的日志存储。
//
// allowDDL 为 false 时所有维护路径都只校验。
func New(db *gorm.DB, loc *time.Location, allowDDL bool, schemas ...Schema) (*Store, error) {
	if db == nil || loc == nil {
		return nil, fmt.Errorf("日志连接和时区必须就绪")
	}
	s := &Store{DB: db, Location: loc, allowDDL: allowDDL, schemas: make(map[string]Schema), ready: make(map[string]bool)}
	for _, schema := range schemas {
		if !validPrefix(schema.Prefix) {
			return nil, fmt.Errorf("无效日志类型")
		}
		s.schemas[schema.Prefix] = schema
	}
	return s, nil
}

// Initialize 在启动时维护历史月表与当月表。
//
// 关闭 DDL 时仅发现、验证现有表。
func (s *Store) Initialize(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.refreshLocked(ctx); err != nil {
		return err
	}
	for prefix, schema := range s.schemas {
		tables := append([]string(nil), s.tables...)
		for _, table := range tables {
			if _, ok := tableMonth(prefix, table, s.Location); ok {
				if err := s.maintainLocked(ctx, table, schema); err != nil {
					return err
				}
			}
		}
		table, _ := MonthTable(prefix, time.Now(), s.Location)
		if s.allowDDL {
			if err := s.maintainLocked(ctx, table, schema); err != nil {
				return err
			}
		}
	}
	return s.refreshLocked(ctx)
}

func (s *Store) refreshLocked(ctx context.Context) error {
	tables, err := s.DB.WithContext(ctx).Migrator().GetTables()
	if err != nil {
		return fmt.Errorf("发现日志月表失败: %w", err)
	}
	s.tables = s.tables[:0]
	seen := make(map[string]bool)
	for _, table := range tables {
		if s.DB.Dialector.Name() == "oracle" {
			table = strings.ToLower(table)
		}
		for prefix := range s.schemas {
			if _, ok := tableMonth(prefix, table, s.Location); ok && !seen[table] {
				seen[table] = true
				s.tables = append(s.tables, table)
				break
			}
		}
	}
	sort.Strings(s.tables)
	s.refreshed = time.Now()
	return nil
}

// Tables 返回和时间范围相交的现存月表。
//
// 无范围时返回全部历史月份，元数据错误不会作为空结果处理。
func (s *Store) Tables(ctx context.Context, prefix string, bounds []time.Time) ([]string, error) {
	if !validPrefix(prefix) {
		return nil, fmt.Errorf("无效日志类型")
	}
	if len(bounds) != 0 && len(bounds) != 2 {
		return nil, fmt.Errorf("创建时间必须包含起止两个值")
	}
	if len(bounds) == 2 && bounds[0].After(bounds[1]) {
		return nil, fmt.Errorf("创建时间起止顺序无效")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if time.Since(s.refreshed) >= 30*time.Second {
		if err := s.refreshLocked(ctx); err != nil {
			return nil, err
		}
	}
	tables := make([]string, 0)
	for _, table := range s.tables {
		month, ok := tableMonth(prefix, table, s.Location)
		if !ok {
			continue
		}
		if len(bounds) == 2 && (month.After(bounds[1]) || !month.AddDate(0, 1, 0).After(bounds[0])) {
			continue
		}
		tables = append(tables, table)
	}
	return tables, nil
}

// EnsureMonth 在写入前确保记录所属月份可用。
//
// 按月份加锁维护，成功新建后立即刷新清单。
func (s *Store) EnsureMonth(ctx context.Context, prefix string, at time.Time) (string, error) {
	table, err := MonthTable(prefix, at, s.Location)
	if err != nil {
		return "", err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.ready[table] {
		return table, nil
	}
	schema, ok := s.schemas[prefix]
	if !ok {
		return "", fmt.Errorf("日志类型未注册")
	}
	if err := s.maintainLocked(ctx, table, schema); err != nil {
		return "", err
	}
	if err := s.refreshLocked(ctx); err != nil {
		return "", err
	}
	return table, nil
}

func (s *Store) maintainLocked(ctx context.Context, table string, schema Schema) error {
	db := s.DB.WithContext(ctx).Table(table)
	if s.allowDDL {
		if err := db.AutoMigrate(schema.Model); err != nil {
			return err
		}
	} else {
		columns, err := db.Migrator().ColumnTypes(table)
		if err != nil {
			return fmt.Errorf("校验月表 %s: %w", table, err)
		}
		have := make(map[string]bool)
		for _, c := range columns {
			have[strings.ToLower(c.Name())] = true
		}
		stmt := &gorm.Statement{DB: db}
		if err := stmt.Parse(schema.Model); err != nil {
			return err
		}
		for _, field := range stmt.Schema.Fields {
			if field.DBName != "" && !have[strings.ToLower(field.DBName)] {
				return fmt.Errorf("月表 %s 缺少列 %s，自动建表已关闭", table, field.DBName)
			}
		}
	}
	for _, columns := range schema.Indexes {
		sum := sha256.Sum256([]byte(table + ":" + strings.Join(columns, ",")))
		name := fmt.Sprintf("lx_%x", sum[:10])
		exists, err := s.indexExists(ctx, table, name)
		if err != nil {
			return err
		}
		if !exists {
			if !s.allowDDL {
				return fmt.Errorf("月表 %s 缺少索引 %s，自动建表已关闭", table, name)
			}
			quoted := make([]string, len(columns))
			for i, c := range columns {
				quoted[i] = Quote(s.DB, c)
			}
			if err := db.Exec("CREATE INDEX " + Quote(s.DB, name) + " ON " + Quote(s.DB, table) + " (" + strings.Join(quoted, ",") + ")").Error; err != nil {
				return err
			}
		}
	}
	s.ready[table] = true
	return nil
}

func (s *Store) indexExists(ctx context.Context, table, name string) (bool, error) {
	var count int64
	var sql string
	switch s.DB.Dialector.Name() {
	case "sqlite":
		sql = "SELECT count(*) FROM sqlite_master WHERE type = 'index' AND tbl_name = ? AND name = ?"
	case "mysql":
		sql = "SELECT count(*) FROM information_schema.statistics WHERE table_schema = DATABASE() AND table_name = ? AND index_name = ?"
	case "postgres":
		sql = "SELECT count(*) FROM pg_indexes WHERE schemaname = current_schema() AND tablename = ? AND indexname = ?"
	case "sqlserver":
		sql = "SELECT count(*) FROM sys.indexes WHERE object_id = OBJECT_ID(?) AND name = ?"
	case "oracle":
		sql = "SELECT count(*) FROM user_indexes WHERE UPPER(table_name) = UPPER(?) AND UPPER(index_name) = UPPER(?)"
	default:
		return false, fmt.Errorf("不支持日志索引方言")
	}
	err := s.DB.WithContext(ctx).Raw(sql, table, name).Scan(&count).Error
	return count > 0, err
}
