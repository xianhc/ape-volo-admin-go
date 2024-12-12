package monitor

import (
	"go-apevolo/global"
	"go-apevolo/model/monitor"
	"go-apevolo/payloads/dto"
	"go-apevolo/payloads/request"
	"go-apevolo/utils"
	"go.uber.org/zap"
	"time"
)

type AuditLogService struct{}

// Create
// @description: 创建
// @receiver: auditLogService
// @param: auditLog
// @return: error
func (auditLogService *AuditLogService) Create(auditLog monitor.AuditLog) error {
	auditLog.CreateTime = time.Now()
	err := global.Db.Create(&auditLog).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		return err
	}
	return nil
}

// Query
// @description: 查询
// @receiver: auditLogService
// @param: info
// @param: list
// @param: count
// @return: error
func (auditLogService *AuditLogService) Query(info *dto.LogQueryCriteria, list *[]monitor.AuditLog, count *int64) error {
	limit := info.PageSize
	offset := info.PageSize * (info.PageIndex - 1)
	// 创建db
	db := global.Db.Model(&monitor.AuditLog{})
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

// QueryByCurrent
// @description: 查询当前用户
// @receiver: auditLogService
// @param: info
// @param: account
// @param: list
// @param: count
// @return: error
func (auditLogService *AuditLogService) QueryByCurrent(info request.Pagination, account string, list *[]monitor.AuditLog, count *int64) error {
	limit := info.PageSize
	offset := info.PageSize * (info.PageIndex - 1)
	// 创建db
	db := global.Db.Model(&monitor.AuditLog{})

	db = db.Where("create_by = ? ", account)

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
