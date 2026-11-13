package scheduledjoblog

import (
	"context"
	"strconv"
	"time"

	"github.com/xianhc/ape-volo-admin-go/internal/global"
	"github.com/xianhc/ape-volo-admin-go/internal/logstore"
	"github.com/xianhc/ape-volo-admin-go/internal/model"
	"github.com/xianhc/ape-volo-admin-go/internal/utils"
	"github.com/xianhc/ape-volo-admin-go/internal/utils/excel"
	"go.uber.org/zap"
)

const maxExceptionDetail = 4000

// Service 实现调度任务日志的记录、查询和导出业务。
//
// Service 只通过 Repository 访问数据库，写入失败不会中断任务执行。
type Service struct{ repository repository }

// Record 记录一次调度任务执行结果。
//
// 异常详情最多保留 4000 个字节，写入失败只记录错误日志。
func (service *Service) Record(ctx context.Context, task Execution, duration time.Duration, runErr error) {
	entry := ScheduledJobLog{RootKey: model.RootKey{Id: int64(utils.GenerateID())}, BaseModel: model.BaseModel{CreationTime: model.Time{Time: logstore.Now()}}, JobId: task.JobId, JobName: task.JobName, JobGroup: task.JobGroup, AssemblyName: task.AssemblyName, ClassName: task.ClassName, Cron: task.Cron, RunParams: model.CustomFieldText(task.RunParams), ExecutionDuration: duration.Milliseconds(), IsSuccess: runErr == nil, CreatorName: "ScheduledJob"}
	if id, ok := utils.OperatorID(ctx); ok {
		entry.CreatorId = &id
	}
	if runErr != nil {
		detail := runErr.Error()
		if len(detail) > maxExceptionDetail {
			detail = detail[:maxExceptionDetail]
		}
		entry.ExceptionDetail = model.CustomFieldText(detail)
	}
	writer.TryEnqueue(entry)
}

// Query 查询调度任务日志分页列表。
//
// 返回前统一把 Entity 转换为 ScheduledJobLogResponse。
func (service *Service) Query(ctx context.Context, req QueryScheduledJobLogRequest) ([]ScheduledJobLogResponse, int64, error) {
	logs, total, err := service.repository.List(ctx, listOptions(req))
	if err != nil {
		return nil, 0, logDatabaseError(err)
	}
	return fromEntities(logs), total, nil
}

// Download 导出符合条件的调度任务日志。
//
// 导出最多读取十万条记录，保持既有行为。
func (service *Service) Download(ctx context.Context, req QueryScheduledJobLogRequest) (filePath, fileName string, err error) {
	options := listOptions(req)
	options.Limit = 100000
	options.Offset = 0
	logs, _, err := service.repository.List(ctx, options)
	if err != nil {
		return "", "", logDatabaseError(err)
	}
	file := excel.NewFile("ScheduledJobLogs")
	defer func() { _ = file.Close() }()
	file.SetHeader([]string{"ID", "任务名称", "任务分组", "Cron表达式", "程序集", "执行类", "执行传参", "执行耗时(ms)", "是否成功", "异常详情", "创建时间"})
	for _, log := range logs {
		success := "否"
		if log.IsSuccess {
			success = "是"
		}
		file.AddRow(strconv.FormatInt(log.Id, 10), log.JobName, log.JobGroup, log.Cron, log.AssemblyName, log.ClassName, string(log.RunParams), log.ExecutionDuration, success, string(log.ExceptionDetail), log.CreationTime.Format("2006-01-02 15:04:05"))
	}
	fileName = "ScheduledJobLogs_" + utils.GenerateID().String() + ".xlsx"
	filePath = global.Config.Excel.Dir + fileName
	return filePath, fileName, file.Save(filePath)
}

func listOptions(req QueryScheduledJobLogRequest) queryOptions {
	return queryOptions{JobName: req.JobName, JobGroup: req.JobGroup, IsSuccess: req.IsSuccess, Sorting: req.Sorting, Limit: req.PageSize, Offset: req.PageSize * (req.PageIndex - 1)}
}
func logDatabaseError(err error) error { global.Logger.Error("db error: ", zap.Error(err)); return err }
