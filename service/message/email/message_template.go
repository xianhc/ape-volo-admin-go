package email

import (
	"errors"
	"go-apevolo/global"
	"go-apevolo/model"
	"go-apevolo/model/message/email"
	"go-apevolo/payloads/dto"
	"go-apevolo/payloads/request"
	"go-apevolo/utils"
	"go-apevolo/utils/ext"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type MessageTemplateService struct{}

// Create
// @description: 创建
// @receiver: messageTemplateService
// @param: req
// @return: error
func (messageTemplateService *MessageTemplateService) Create(req *dto.CreateUpdateEmailMessageTemplateDto) error {
	messageTemplate := &email.MessageTemplate{}
	var count int64
	err := global.Db.Model(&email.MessageTemplate{}).Scopes(utils.IsDeleteSoft).Where("name = ?", req.Name).Count(&count).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		return err
	}
	if count > 0 {
		return errors.New("模板名称=>" + req.Name + "=>已存在!")
	}
	req.Generate(messageTemplate)
	err = global.Db.Create(messageTemplate).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		return err
	}
	return nil
}

// Update
// @description: 编辑
// @receiver: messageTemplateService
// @param: req
// @return: error
func (messageTemplateService *MessageTemplateService) Update(req *dto.CreateUpdateEmailMessageTemplateDto) error {
	oldEmailMessageTemplate := &email.MessageTemplate{}
	emailMessageTemplate := &email.MessageTemplate{}
	err := global.Db.Scopes(utils.IsDeleteSoft).First(oldEmailMessageTemplate, req.Id).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("数据不存在或您无权查看！")
		}
		return err
	}
	if oldEmailMessageTemplate.Name != req.Name {
		var count int64
		err = global.Db.Model(&email.MessageTemplate{}).Scopes(utils.IsDeleteSoft).Where("name = ?", req.Name).Count(&count).Error
		if err != nil {
			global.Logger.Error("db error: ", zap.Error(err))
			return err
		}
		if count > 0 {
			return errors.New("模板名称=>" + req.Name + "=>已存在!")
		}
	}
	req.Generate(emailMessageTemplate)
	err = global.Db.Updates(emailMessageTemplate).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		return err
	}
	return nil
}

// Delete
// @description: 删除
// @receiver: messageTemplateService
// @param: idArray
// @param: updateBy
// @return: error
func (messageTemplateService *MessageTemplateService) Delete(idArray request.IdCollection, updateBy string) error {
	var emailMessageTemplates []email.MessageTemplate
	err := global.Db.Scopes(utils.IsDeleteSoft).Find(&emailMessageTemplates, "id in ?", idArray.IdArray).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		return err
	}
	if len(emailMessageTemplates) == 0 {
		return errors.New("数据不存在或您无权查看！")
	}
	localTime := ext.GetCurrentTime()
	err = global.Db.Model(&email.MessageTemplate{}).Scopes(utils.IsDeleteSoft).Where("id in (?)", idArray.IdArray).Updates(
		email.MessageTemplate{BaseModel: model.BaseModel{
			UpdateBy:   &updateBy,
			UpdateTime: &localTime,
		}, SoftDeleted: model.SoftDeleted{IsDeleted: true}},
	).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		return err
	}
	return nil
}

// Query
// @description: 查询
// @receiver: messageTemplateService
// @param: info
// @param: list
// @param: count
// @return: error
func (messageTemplateService *MessageTemplateService) Query(info *dto.EmailMessageTemplateQueryCriteria, list *[]email.MessageTemplate, count *int64) error {
	limit := info.PageSize
	offset := info.PageSize * (info.PageIndex - 1)
	// 创建db并构建查询条件
	db := buildEmailMessageTemplateQuery(global.Db.Model(&email.MessageTemplate{}), info)

	err := db.Count(count).Limit(limit).Offset(offset).Find(list).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		return err
	}
	return nil
}

// buildEmailMessageTemplateQuery
// @description: 条件表达式
// @param: db
// @param: info
// @return: *gorm.DB
func buildEmailMessageTemplateQuery(db *gorm.DB, info *dto.EmailMessageTemplateQueryCriteria) *gorm.DB {
	if info.Name != "" {
		db = db.Where("name LIKE ?", "%"+info.Name+"%")
	}
	if info.IsActive != nil {
		db = db.Where("is_active = ?", info.IsActive)
	}
	if len(info.CreateTime) > 0 {
		db = db.Where("create_time >= ? and create_time <= ? ", info.CreateTime[0], info.CreateTime[1])
	}
	if len(info.SortFields) > 0 {
		for _, sort := range info.SortFields {
			db = db.Order(sort)
		}
	}
	db = db.Scopes(utils.IsDeleteSoft)

	return db
}
