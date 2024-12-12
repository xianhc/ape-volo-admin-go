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

type AppSecretService struct{}

// Create
// @description: 创建
// @receiver: appSecretService
// @param: req
// @return: error
func (appSecretService *AppSecretService) Create(req *dto.CreateUpdateAppSecretDto) error {
	appSecret := &system.AppSecret{}
	var total int64
	err := global.Db.Model(&system.AppSecret{}).Scopes(utils.IsDeleteSoft).Where("app_name = ?", req.AppName).Count(&total).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		return err
	}
	if total > 0 {
		return errors.New("应用=>" + req.AppName + "=>已存在!")
	}
	req.Generate(appSecret)
	err = global.Db.Create(appSecret).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		return err
	}
	return nil
}

// Update
// @description: 编辑
// @receiver: appSecretService
// @param: req
// @return: error
func (appSecretService *AppSecretService) Update(req *dto.CreateUpdateAppSecretDto) error {
	oldAppSecret := &system.AppSecret{}
	appSecret := &system.AppSecret{}
	err := global.Db.Scopes(utils.IsDeleteSoft).Where("id = ?", req.Id).First(oldAppSecret).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("数据不存在或您无权查看！")
		}
		return err
	}
	if oldAppSecret.AppName != req.AppName {
		var total int64
		err = global.Db.Model(&system.AppSecret{}).Scopes(utils.IsDeleteSoft).Where("app_name = ?", req.AppName).Count(&total).Error
		if err != nil {
			global.Logger.Error("db error: ", zap.Error(err))
			return err
		}
		if total > 0 {

			return errors.New("应用名称=>" + req.AppName + "=>已存在!")
		}
	}
	req.Generate(appSecret)
	err = global.Db.Updates(appSecret).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		return err
	}
	return nil
}

// Delete
// @description: 删除
// @receiver: appSecretService
// @param: idArray
// @param: updateBy
// @return: error
func (appSecretService *AppSecretService) Delete(idArray request.IdCollection, updateBy string) error {
	appSecrets := &[]system.AppSecret{}
	err := global.Db.Scopes(utils.IsDeleteSoft).Find(appSecrets, "id in ?", idArray.IdArray).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		return err
	}
	if len(*appSecrets) <= 0 {
		return errors.New("数据不存在")
	}

	localTime := ext.GetCurrentTime()
	err = global.Db.Model(&system.AppSecret{}).Scopes(utils.IsDeleteSoft).Where("id in (?)", idArray.IdArray).Updates(
		system.AppSecret{BaseModel: model.BaseModel{
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
// @receiver: appSecretService
// @param: info
// @param: list
// @param: total
// @return: error
func (appSecretService *AppSecretService) Query(info *dto.AppSecretQueryCriteria, list *[]system.AppSecret, total *int64) error {
	limit := info.PageSize
	offset := info.PageSize * (info.PageIndex - 1)
	// 创建db
	db := buildAppSecretQuery(global.Db.Model(&system.AppSecret{}), info)

	err := db.Count(total).Limit(limit).Offset(offset).Find(list).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		return err
	}
	return nil
}

// Download
// @description: 导出
// @receiver: appSecretService
// @param: info
// @return: filePath
// @return: fileName
// @return: err
func (appSecretService *AppSecretService) Download(info *dto.AppSecretQueryCriteria) (filePath string, fileName string, err error) {
	var appSecrets []system.AppSecret
	// 创建db并构建查询条件
	err = buildAppSecretQuery(global.Db.Model(&system.AppSecret{}), info).Find(&appSecrets).Error

	if err != nil {
		return
	}
	// 创建一个新的 Excel 文件
	file := xlsx.NewFile()
	sheet, err := file.AddSheet("AppSecret")
	if err != nil {
		return
	}
	style := xlsx.NewStyle()
	style.Fill = *xlsx.NewFill("solid", "1e90ff", "1e90ff")
	style.ApplyFill = true
	// 表头数据
	header := []string{"ID", "应用密钥", "应用名称", "备注", "创建时间"}
	// 将表头添加到 Excel 工作表
	row := sheet.AddRow()
	for _, cellValue := range header {
		cell := row.AddCell()
		cell.Value = cellValue
		cell.SetStyle(style)
	}

	// 将数据添加到 Excel 工作表
	for _, appSecret := range appSecrets {
		row := sheet.AddRow()
		row.WriteSlice(&[]interface{}{
			strconv.FormatInt(appSecret.Id, 10),
			appSecret.AppSecretKey,
			appSecret.AppName,
			appSecret.Remark,
			appSecret.CreateTime.Format("2006-01-02 15:04:05"),
		}, -1)
	}
	// 保存 Excel 文件到本地
	fileName = "AppSecret_" + utils.GenerateID().String() + ".xlsx"
	filePath = global.Config.Excel.Dir + fileName
	err = file.Save(filePath)
	return filePath, fileName, err
}

// buildAppSecretQuery
// @description: 条件表达式
// @param: db
// @param: info
// @return: *gorm.DB
func buildAppSecretQuery(db *gorm.DB, info *dto.AppSecretQueryCriteria) *gorm.DB {
	if info.KeyWords != "" {
		db = db.Where("app_id LIKE ? or app_name LIKE ? or remark LIKE ?", "%"+info.KeyWords+"%", "%"+info.KeyWords+"%", "%"+info.KeyWords+"%")
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
