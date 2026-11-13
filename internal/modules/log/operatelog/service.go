package operatelog

import (
	"context"
	"time"

	"github.com/xianhc/ape-volo-admin-go/internal/global"
	"github.com/xianhc/ape-volo-admin-go/internal/logstore"
	"github.com/xianhc/ape-volo-admin-go/internal/model"
	"github.com/xianhc/ape-volo-admin-go/internal/utils"
	"github.com/xianhc/ape-volo-admin-go/internal/utils/logqueue"
	"go.uber.org/zap"
)

// Service 实现操作日志的写入、查询和统计业务。
//
// Service 只通过 Repository 访问数据库，并把查询实体转换为独立 Response。
type Service struct{ repository repository }

// Create 在最终结果确定后构建快照并非阻塞入队。
//
// 返回值只表示是否接收，不能据此推定已经落库。
func (service *Service) Create(ctx context.Context, log OperateLog) error {
	if log.Id == 0 {
		log.Id = int64(utils.GenerateID())
	}
	log.CreationTime = model.Time{Time: logstore.Now()}
	if id, ok := utils.OperatorID(ctx); ok && log.CreatorId == nil {
		log.CreatorId = &id
	}
	log.LastModifierId = logqueue.ClonePointer(log.LastModifierId)
	log.LastModificationTime = logqueue.ClonePointer(log.LastModificationTime)
	log.DeleterId = logqueue.ClonePointer(log.DeleterId)
	log.DeletionTime = logqueue.ClonePointer(log.DeletionTime)
	if !writer.TryEnqueue(log) {
		return logqueue.ErrRejected
	}
	return nil
}

// Query 查询操作日志分页列表。
//
// 查询支持描述关键字和创建时间范围。
func (service *Service) Query(ctx context.Context, req QueryOperateLogRequest) ([]OperateLogResponse, int64, error) {
	logs, total, err := service.repository.List(ctx, queryOptions{KeyWords: req.KeyWords, CreateTime: req.CreateTime, Sorting: req.Sorting, Limit: req.PageSize, Offset: req.PageSize * (req.PageIndex - 1)})
	if err != nil {
		return nil, 0, logDatabaseError(err)
	}
	return fromEntities(logs), total, nil
}

// QueryByCurrent 查询当前用户的操作日志分页列表。
//
// creatorId 由认证上下文提供，不接受客户端指定。
func (service *Service) QueryByCurrent(ctx context.Context, req QueryCurrentRequest, creatorId int64) ([]OperateLogResponse, int64, error) {
	logs, total, err := service.repository.List(ctx, queryOptions{CreatorId: &creatorId, Sorting: req.Sorting, Limit: req.PageSize, Offset: req.PageSize * (req.PageIndex - 1)})
	if err != nil {
		return nil, 0, logDatabaseError(err)
	}
	return fromEntities(logs), total, nil
}

// GetOperationNumber 统计最近 days 天每天的操作数。
//
// 缺失日期补零，返回切片长度始终等于 days；非正数使用 7 天。
func (service *Service) GetOperationNumber(ctx context.Context, days int) ([]int, error) {
	if days <= 0 {
		days = 7
	}
	now := time.Now().In(logstore.Location())
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, -(days - 1))
	rows, err := service.repository.CountByDay(ctx, start)
	if err != nil {
		return nil, logDatabaseError(err)
	}
	counts := make(map[string]int, len(rows))
	for _, row := range rows {
		counts[row.Day] = row.Total
	}
	numbers := make([]int, 0, days)
	for index := 0; index < days; index++ {
		numbers = append(numbers, counts[start.AddDate(0, 0, index).Format("2006-01-02")])
	}
	return numbers, nil
}

func logDatabaseError(err error) error { global.Logger.Error("db error: ", zap.Error(err)); return err }
