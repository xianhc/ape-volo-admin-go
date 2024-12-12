package monitor

import (
	"go-apevolo/global"
	"go-apevolo/model/monitor"
	"go-apevolo/payloads/dto"
	"go-apevolo/utils"
	"go.uber.org/zap"
	"time"
)

type ExceptionLogService struct{}

// Create
// @description: 创建
// @receiver: exceptionLogService
// @param: exceptionLog
// @return: error
func (exceptionLogService *ExceptionLogService) Create(exceptionLog monitor.ExceptionLog) error {
	exceptionLog.CreateTime = time.Now()
	err := global.Db.Create(&exceptionLog).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		return err
	}
	return nil
}

// Query
// @description: 查询
// @receiver: exceptionLogService
// @param: info
// @param: list
// @param: count
// @return: error
func (exceptionLogService *ExceptionLogService) Query(info *dto.LogQueryCriteria, list *[]monitor.ExceptionLog, count *int64) error {
	limit := info.PageSize
	offset := info.PageSize * (info.PageIndex - 1)
	// 创建db
	db := global.Db.Model(&monitor.ExceptionLog{})
	if info.KeyWords != "" {
		db = db.Where("description LIKE ? ", "%"+info.KeyWords+"%")
	}

	if len(info.CreateTime) > 0 {
		db = db.Where("create_time >= ? and create_time <= ? ", info.CreateTime[0], info.CreateTime[1])
	}
	if len(info.SortFields) > 0 {
		for _, sort := range info.SortFields {
			db = db.Order(sort)
		}
	}
	err := db.Scopes(utils.IsDeleteSoft).Count(count).Limit(limit).Offset(offset).Find(list).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		return err
	}
	return nil
}
