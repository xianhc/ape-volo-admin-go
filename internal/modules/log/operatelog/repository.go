package operatelog

import (
	"context"
	"errors"
	"time"

	"github.com/xianhc/ape-volo-admin-go/internal/global"
	"github.com/xianhc/ape-volo-admin-go/internal/logstore"
	"github.com/xianhc/ape-volo-admin-go/internal/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// queryOptions 描述操作日志列表查询的数据库参数。
//
// CreatorId 非 nil 时只查询指定用户创建的日志。
type queryOptions struct {
	KeyWords   string
	CreateTime []time.Time
	CreatorId  *int64
	Sorting    string
	Limit      int
	Offset     int
}

// DailyCount 表示某一天的操作日志数量。
//
// Day 使用 `yyyy-MM-dd`，Total 为该日期的记录数。
type DailyCount struct {
	Day   string
	Total int
}

// repository 封装操作日志的数据库访问。
//
// 零值即可使用，查询只使用独立日志连接。
type repository struct{}

// CreateBatch 由独立消费者调用，按生产月份写入日志连接。
//
// 返回每个月份子批次的最终处理分类。
func (repository) CreateBatch(ctx context.Context, entries []OperateLog) logstore.Outcome {
	return logstore.WriteBatch(ctx, logstore.Default(), logstore.Operate, entries, time.Duration(global.Config.LogQueue.WriteTimeoutSeconds)*time.Second, global.Logger)
}

// List 查询一页操作日志及分页前总数。
//
// 查询支持描述关键字、创建时间和创建者筛选。
func (repository) List(ctx context.Context, options queryOptions) ([]OperateLog, int64, error) {
	return logstore.Page[OperateLog](ctx, logstore.Default(), logstore.Operate, logstore.Query{DB: utils.LogDB(ctx), Bounds: options.CreateTime, Sorting: options.Sorting, Limit: options.Limit, Offset: options.Offset, Filter: func(db *gorm.DB) *gorm.DB { return buildQuery(db, options) }})
}

// CountByDay 统计指定起始时间后的每日操作数。
//
// 返回数据只包含实际有记录的日期，缺失日期由 Service 补零。
func (repository) CountByDay(ctx context.Context, start time.Time) ([]DailyCount, error) {
	s := logstore.Default()
	if s == nil {
		return nil, errors.New("日志存储尚未初始化")
	}
	tables, err := s.Tables(ctx, logstore.Operate, []time.Time{start, time.Now().In(s.Location)})
	if err != nil {
		return nil, err
	}
	db := utils.LogDB(ctx)
	day, err := logstore.DayExpression(db, s.Location.String())
	if err != nil {
		return nil, err
	}
	counts := make(map[string]int)
	for _, table := range tables {
		var rows []DailyCount
		if err := db.Model(&OperateLog{}).Table(table).Select(day + " AS " + logstore.Quote(db, "day") + ", count(*) AS " + logstore.Quote(db, "total")).Where(clause.Gte{Column: clause.Column{Name: "creation_time"}, Value: start}).Group(day).Find(&rows).Error; err != nil {
			return nil, err
		}
		for _, row := range rows {
			counts[row.Day] += row.Total
		}
	}
	rows := make([]DailyCount, 0, len(counts))
	for day, total := range counts {
		rows = append(rows, DailyCount{Day: day, Total: total})
	}
	return rows, nil
}

func buildQuery(db *gorm.DB, options queryOptions) *gorm.DB {
	if options.KeyWords != "" {
		db = db.Where(clause.Like{Column: clause.Column{Name: "description"}, Value: "%" + options.KeyWords + "%"})
	}
	if len(options.CreateTime) != 0 && len(options.CreateTime) != 2 {
		_ = db.AddError(errors.New("创建时间必须包含起止两个值"))
		return db
	}
	if len(options.CreateTime) == 2 {
		db = db.Where(clause.Gte{Column: clause.Column{Name: "creation_time"}, Value: options.CreateTime[0].In(logstore.Location())}).Where(clause.Lte{Column: clause.Column{Name: "creation_time"}, Value: options.CreateTime[1].In(logstore.Location())})
	}
	if options.CreatorId != nil {
		db = db.Where(clause.Eq{Column: clause.Column{Name: "creator_id"}, Value: *options.CreatorId})
	}
	return db
}
