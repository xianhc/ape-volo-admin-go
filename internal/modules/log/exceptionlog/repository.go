package exceptionlog

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

// queryOptions 描述异常日志列表查询的数据库参数。
//
// Service 负责从 HTTP 请求转换该结构。
type queryOptions struct {
	KeyWords      string
	CreateTime    []time.Time
	Sorting       string
	Limit, Offset int
}

// repository 封装异常日志的数据库访问。
//
// 零值即可使用，查询只使用独立日志连接。
type repository struct{}

// CreateBatch 按生产月份批量写入独立日志连接。
//
// 返回每个月份子批次的最终处理分类。
func (repository) CreateBatch(ctx context.Context, entries []ExceptionLog) logstore.Outcome {
	return logstore.WriteBatch(ctx, logstore.Default(), logstore.Exception, entries, time.Duration(global.Config.LogQueue.WriteTimeoutSeconds)*time.Second, global.Logger)
}

// List 查询一页异常日志及分页前总数。
//
// 查询支持描述关键字和创建时间范围。
func (repository) List(ctx context.Context, options queryOptions) ([]ExceptionLog, int64, error) {
	return logstore.Page[ExceptionLog](ctx, logstore.Default(), logstore.Exception, logstore.Query{DB: utils.LogDB(ctx), Bounds: options.CreateTime, Sorting: options.Sorting, Limit: options.Limit, Offset: options.Offset, Filter: func(db *gorm.DB) *gorm.DB { return buildQuery(db, options) }})
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
	return db
}
