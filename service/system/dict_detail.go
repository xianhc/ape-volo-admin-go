package system

import (
	"errors"
	"go-apevolo/global"
	"go-apevolo/model"
	"go-apevolo/model/system"
	"go-apevolo/payloads/dto"
	"go-apevolo/utils"
	"go-apevolo/utils/ext"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type DictDetailService struct{}

// Create
// @description: 添加
// @receiver: dictDetailService
// @param: req
// @return: error
func (dictDetailService *DictDetailService) Create(req *dto.CreateUpdateDictDetailDto) error {
	dictDetail := &system.DictDetail{}
	var total int64
	err := global.Db.Model(&system.DictDetail{}).Scopes(utils.IsDeleteSoft).Where(" dict_id = ? and label = ? and value = ?", req.Dict.Id, req.Label, req.Value).Count(&total).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		return err
	}
	if total > 0 {
		return errors.New("字典详情标签=>" + req.Label + "=>已存在!")
	}
	req.Generate(dictDetail)
	err = global.Db.Create(dictDetail).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		return err
	}
	return nil
}

// Update
// @description: 修改
// @receiver: dictDetailService
// @param: req
// @return: error
func (dictDetailService *DictDetailService) Update(req *dto.CreateUpdateDictDetailDto) error {
	dictDetail := &system.DictDetail{}
	oldDictDetail := system.DictDetail{}
	err := global.Db.Scopes(utils.IsDeleteSoft).First(&oldDictDetail, req.Id).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("数据不存在或您无权查看！")
		}
		return err
	}
	req.Generate(dictDetail)
	err = global.Db.Updates(dictDetail).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		return err
	}
	return nil
}

// Delete
// @description: 删除
// @receiver: dictDetailService
// @param: id
// @param: updateBy
// @return: error
func (dictDetailService *DictDetailService) Delete(id int64, updateBy string) error {
	err := global.Db.Scopes(utils.IsDeleteSoft).First(&system.DictDetail{}, id).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("数据不存在或您无权查看！")
		}
		return err
	}
	localTime := ext.GetCurrentTime()
	err = global.Db.Scopes(utils.IsDeleteSoft).Where("id = ?", id).Updates(
		system.DictDetail{BaseModel: model.BaseModel{
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
// @receiver: dictDetailService
// @param: dictName
// @param: list
// @return: error
func (dictDetailService *DictDetailService) Query(dictName string, list *[]system.DictDetail) error {
	// 创建db
	db := global.Db.Model(&system.Dict{})
	var dict system.Dict
	err := db.Scopes(utils.IsDeleteSoft).Preload("DictDetail", func(db *gorm.DB) *gorm.DB {
		return db.Scopes(utils.IsDeleteSoft).Order("dict_sort")
	}).Where("name = ?", dictName).First(&dict).Error

	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		return err
	}
	for i, _ := range dict.DictDetail {
		dict.DictDetail[i].DictDto = system.DictDto2{Id: dict.Id}
	}
	*list = dict.DictDetail
	return nil
}
