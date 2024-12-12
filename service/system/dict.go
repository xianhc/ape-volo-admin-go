package system

import (
	"errors"
	"github.com/tealeg/xlsx"
	"go-apevolo/global"
	"go-apevolo/model"
	"go-apevolo/model/system"
	"go-apevolo/payloads/dto"
	"go-apevolo/payloads/request"
	"go-apevolo/utils"
	"go-apevolo/utils/ext"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"strconv"
)

type DictService struct{}

// Create
// @description: 创建
// @receiver: dictService
// @param: req
// @return: error
func (dictService *DictService) Create(req *dto.CreateUpdateDictDto) error {
	dict := &system.Dict{}
	var total int64
	err := global.Db.Model(&system.Dict{}).Scopes(utils.IsDeleteSoft).Where("name = ?", req.Name).Count(&total).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		return err
	}
	if total > 0 {
		return errors.New("字典名称=>" + req.Name + "=>已存在!")
	}
	req.Generate(dict)
	err = global.Db.Create(dict).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		return err
	}
	return nil
}

// Update
// @description: 编辑
// @receiver: dictService
// @param: req
// @return: error
func (dictService *DictService) Update(req *dto.CreateUpdateDictDto) error {
	dict := &system.Dict{}
	oldDict := &system.Dict{}
	err := global.Db.Scopes(utils.IsDeleteSoft).Where("id = ?", req.Id).First(oldDict).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("数据不存在或您无权查看！")
		}
		return err
	}
	if oldDict.Name != req.Name {
		var total int64
		err = global.Db.Model(&system.AppSecret{}).Scopes(utils.IsDeleteSoft).Where("name = ?", req.Name).Count(&total).Error
		if err != nil {
			global.Logger.Error("db error: ", zap.Error(err))
			return err
		}
		if total > 0 {

			return errors.New("字典名称=>" + req.Name + "=>已存在!")
		}
	}
	req.Generate(dict)
	err = global.Db.Updates(dict).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		return err
	}
	return nil
}

// Delete
// @description: 删除
// @receiver: dictService
// @param: collection
// @param: updateBy
// @return: error
func (dictService *DictService) Delete(collection request.IdCollection, updateBy string) error {
	dicts := &[]system.Dict{}
	err := global.Db.Scopes(utils.IsDeleteSoft).Find(dicts, "id in ?", collection.IdArray).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		return err
	}
	if len(*dicts) <= 0 {
		return errors.New("数据不存在或您无权查看！")
	}
	localTime := ext.GetCurrentTime()
	err = global.Db.Model(&system.Dict{}).Scopes(utils.IsDeleteSoft).Where("id in (?)", collection.IdArray).Updates(
		system.Dict{BaseModel: model.BaseModel{
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
// @receiver: dictService
// @param: info
// @param: list
// @param: total
// @return: error
func (dictService *DictService) Query(info *dto.DictQueryCriteria, list *[]system.Dict, total *int64) error {
	limit := info.PageSize
	offset := info.PageSize * (info.PageIndex - 1)
	// 创建db
	db := buildDictQuery(global.Db.Model(&system.Dict{}), info)
	err := db.Count(total).Limit(limit).Offset(offset).Preload("DictDetail").Find(list).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		return err
	}
	for _, dict := range *list {
		for i, _ := range dict.DictDetail {
			dict.DictDetail[i].DictDto = system.DictDto2{Id: dict.Id}
		}
	}
	return err
}

// Download
// @description: 导出
// @receiver: dictService
// @param: info
// @return: filePath
// @return: fileName
// @return: err
func (dictService *DictService) Download(info *dto.DictQueryCriteria) (filePath string, fileName string, err error) {
	var dicts []system.Dict
	// 创建db并构建查询条件
	err = buildDictQuery(global.Db.Model(&system.Dict{}), info).Preload("DictDetail").Find(&dicts).Error

	if err != nil {
		return
	}
	// 创建一个新的 Excel 文件
	file := xlsx.NewFile()
	sheet, err := file.AddSheet("Dicts")
	if err != nil {
		return
	}
	style := xlsx.NewStyle()
	style.Fill = *xlsx.NewFill("solid", "1e90ff", "1e90ff") // 使用红色作为背景色
	style.ApplyFill = true

	// 表头数据
	header := []string{"ID", "字典名称", "字典描述", "字典标签", "字典值", "创建时间"}
	// 将表头添加到 Excel 工作表
	row := sheet.AddRow()
	for _, cellValue := range header {
		cell := row.AddCell()
		cell.Value = cellValue
		cell.SetStyle(style)
	}

	// 将数据添加到 Excel 工作表
	for _, dict := range dicts {
		for _, detail := range dict.DictDetail {
			row := sheet.AddRow()
			row.WriteSlice(&[]interface{}{
				strconv.FormatInt(dict.Id, 10),
				dict.Name,
				dict.Description,
				detail.Label,
				detail.Value,
				detail.CreateTime.Format("2006-01-02 15:04:05"),
			}, -1)
		}
	}
	// 保存 Excel 文件到本地
	fileName = "Dicts_" + utils.GenerateID().String() + ".xlsx"
	filePath = global.Config.Excel.Dir + fileName
	err = file.Save(filePath)
	return filePath, fileName, err
}

// buildDictQuery
// @description: 条件表达式
// @param: db
// @param: info
// @return: *gorm.DB
func buildDictQuery(db *gorm.DB, info *dto.DictQueryCriteria) *gorm.DB {
	if info.KeyWords != "" {
		db = db.Where("name LIKE ?", "%"+info.KeyWords+"%").Or("description LIKE ?", "%"+info.KeyWords+"%")
	}
	if len(info.SortFields) > 0 {
		for _, sort := range info.SortFields {
			db = db.Order(sort)
		}
	}
	db = db.Scopes(utils.IsDeleteSoft)

	return db
}
