package permission

import (
	"errors"
	"go-apevolo/global"
	"go-apevolo/model"
	"go-apevolo/model/permission"
	"go-apevolo/payloads/dto"
	"go-apevolo/payloads/request"
	"go-apevolo/utils"
	"go-apevolo/utils/ext"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type ApisService struct{}

// Create
// @description: 创建
// @receiver: apisService
// @param: req
// @return: error
func (apisService *ApisService) Create(req *dto.CreateUpdateApisDto) error {
	api := &permission.Apis{}
	var total int64
	err := global.Db.Model(&permission.Apis{}).Scopes(utils.IsDeleteSoft).Where("url = ? and method = ?", req.Url, req.Method).Count(&total).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		return err
	}
	if total > 0 {
		return errors.New("Url=>" + req.Url + "Method=>" + req.Method + "已存在!")
	}
	req.Generate(api)
	err = global.Db.Create(&api).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		return err
	}
	return nil
}

// Update
// @description: 编辑
// @receiver: apisService
// @param: req
// @return: error
func (apisService *ApisService) Update(req *dto.CreateUpdateApisDto) error {
	var oldApis permission.Apis
	var apis permission.Apis
	err := global.Db.Scopes(utils.IsDeleteSoft).First(&oldApis, req.Id).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("数据不存在")
		}
		return err
	}
	if oldApis.Url != req.Url && oldApis.Method != req.Method {
		var total int64
		err = global.Db.Model(&permission.Apis{}).Scopes(utils.IsDeleteSoft).Where("name = ? and method = ?", req.Url, req.Method).Count(&total).Error
		if err != nil {
			global.Logger.Error("db error: ", zap.Error(err))
			return err
		}
		if total > 0 {
			return errors.New("Url=>" + req.Url + "=>已存在!")
		}
	}
	req.Generate(&apis)
	err = global.Db.Updates(&apis).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		return err
	}
	return nil
}

// Delete
// @description: 删除
// @receiver: apisService
// @param: idArray
// @param: updateBy
// @return: error
func (apisService *ApisService) Delete(idArray request.IdCollection, updateBy string) error {
	var apis []permission.Apis
	err := global.Db.Scopes(utils.IsDeleteSoft).Find(&apis, "id in ?", idArray.IdArray).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		return err
	}
	if len(apis) <= 0 {
		return errors.New("数据不存在")
	}

	localTime := ext.GetCurrentTime()
	err = global.Db.Model(&permission.Apis{}).Scopes(utils.IsDeleteSoft).Where("id in (?)", idArray.IdArray).Updates(
		permission.Apis{BaseModel: model.BaseModel{
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
// @receiver: apisService
// @param: info
// @param: list
// @param: count
// @return: error
func (apisService *ApisService) Query(info dto.ApisQueryCriteria, list *[]permission.Apis, count *int64) error {
	limit := info.PageSize
	offset := info.PageSize * (info.PageIndex - 1)
	// 创建db
	db := global.Db.Model(&permission.Apis{})
	if info.Group != "" {
		db = db.Where("group LIKE ?", "%"+info.Group+"%")
	}
	if info.Description != "" {
		db = db.Where("description LIKE ?", "%"+info.Description+"%")
	}
	if info.Method != "" {
		db = db.Where("method LIKE ?", "%"+info.Method+"%")
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
	return err
}
