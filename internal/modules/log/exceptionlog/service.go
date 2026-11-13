package exceptionlog

import (
	"context"

	"github.com/xianhc/ape-volo-admin-go/internal/global"
	"github.com/xianhc/ape-volo-admin-go/internal/logstore"
	"github.com/xianhc/ape-volo-admin-go/internal/model"
	"github.com/xianhc/ape-volo-admin-go/internal/utils"
	"github.com/xianhc/ape-volo-admin-go/internal/utils/logqueue"
	"go.uber.org/zap"
)

// Service 实现异常日志的写入与查询业务。
//
// Service 只通过 Repository 访问数据库，并把查询实体转换为独立 Response。
type Service struct{ repository repository }

// Create 在最终结果确定后构建快照并非阻塞入队。
//
// 返回值只表示是否接收，不能据此推定已经落库。
func (service *Service) Create(ctx context.Context, log ExceptionLog) error {
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

// Query 查询异常日志分页列表。
//
// 查询支持描述关键字和创建时间范围。
func (service *Service) Query(ctx context.Context, req QueryExceptionLogRequest) ([]ExceptionLogResponse, int64, error) {
	logs, total, err := service.repository.List(ctx, queryOptions{KeyWords: req.KeyWords, CreateTime: req.CreateTime, Sorting: req.Sorting, Limit: req.PageSize, Offset: req.PageSize * (req.PageIndex - 1)})
	if err != nil {
		return nil, 0, logDatabaseError(err)
	}
	return fromEntities(logs), total, nil
}

func logDatabaseError(err error) error { global.Logger.Error("db error: ", zap.Error(err)); return err }
