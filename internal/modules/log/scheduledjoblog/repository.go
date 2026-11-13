package scheduledjoblog

import (
	"context"
	"time"

	"github.com/xianhc/ape-volo-admin-go/internal/global"
	"github.com/xianhc/ape-volo-admin-go/internal/logstore"
	"github.com/xianhc/ape-volo-admin-go/internal/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// queryOptions 描述调度任务日志列表查询的数据库参数。
//
// Service 负责从 HTTP 请求转换该结构。
type queryOptions struct {
	JobName, JobGroup string
	IsSuccess         *bool
	Sorting           string
	Limit, Offset     int
}

// repository 封装调度任务日志的数据库访问。
//
// 零值即可使用，查询只使用独立日志连接。
type repository struct{}

// CreateBatch 按生产月份批量写入独立日志连接。
//
// 返回每个月份子批次的最终处理分类。
func (repository) CreateBatch(ctx context.Context, entries []ScheduledJobLog) logstore.Outcome {
	return logstore.WriteBatch(ctx, logstore.Default(), logstore.Job, entries, time.Duration(global.Config.LogQueue.WriteTimeoutSeconds)*time.Second, global.Logger)
}

// List 查询一页任务执行日志及分页前总数。
//
// 未指定排序时默认按创建时间倒序。
func (repository) List(ctx context.Context, options queryOptions) ([]ScheduledJobLog, int64, error) {
	return logstore.Page[ScheduledJobLog](ctx, logstore.Default(), logstore.Job, logstore.Query{DB: utils.LogDB(ctx), Sorting: options.Sorting, Limit: options.Limit, Offset: options.Offset, Filter: func(db *gorm.DB) *gorm.DB { return buildQuery(db, options) }})
}

func buildQuery(db *gorm.DB, options queryOptions) *gorm.DB {
	if options.JobName != "" {
		db = db.Where(clause.Like{Column: clause.Column{Name: "job_name"}, Value: "%" + options.JobName + "%"})
	}
	if options.JobGroup != "" {
		db = db.Where(clause.Eq{Column: clause.Column{Name: "job_group"}, Value: options.JobGroup})
	}
	if options.IsSuccess != nil {
		db = db.Where(clause.Eq{Column: clause.Column{Name: "is_success"}, Value: *options.IsSuccess})
	}
	return db
}
