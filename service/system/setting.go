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

type SettingService struct{}

// Create
// @description: 创建
// @receiver: settingService
// @param: req
// @return: error
func (settingService *SettingService) Create(req *dto.CreateUpdateSettingDto) error {
	setting := &system.Setting{}
	var total int64
	err := global.Db.Model(&system.Setting{}).Scopes(utils.IsDeleteSoft).Where("name = ?", req.Name).Count(&total).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		return err
	}
	if total > 0 {
		return errors.New("设置键=>" + setting.Name + "=>已存在!")
	}
	req.Generate(setting)
	err = global.Db.Create(setting).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		return err
	}
	return nil
}

// Update
// @description: 编辑
// @receiver: settingService
// @param: req
// @return: error
func (settingService *SettingService) Update(req *dto.CreateUpdateSettingDto) error {
	setting := &system.Setting{}
	oldSetting := &system.Setting{}
	err := global.Db.Scopes(utils.IsDeleteSoft).First(oldSetting, req.Id).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("数据不存在或您无权查看！")
		}
		return err
	}
	if oldSetting.Name != req.Name {
		var total int64
		err = global.Db.Scopes(utils.IsDeleteSoft).Where("name = ?", req.Name).Count(&total).Error
		if err != nil {
			global.Logger.Error("db error: ", zap.Error(err))
			return err
		}
		if total > 0 {
			return errors.New("设置键=>" + setting.Name + "=>已存在!")
		}
	}
	req.Generate(setting)
	err = global.Db.Updates(setting).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		return err
	}
	return nil
}

// Delete
// @description: 删除
// @receiver: settingService
// @param: idArray
// @param: updateBy
// @return: error
func (settingService *SettingService) Delete(idArray request.IdCollection, updateBy string) error {
	var settings []system.Setting
	err := global.Db.Scopes(utils.IsDeleteSoft).Find(settings, "id in ?", idArray.IdArray).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		return err
	}
	if len(settings) <= 0 {
		return errors.New("数据不存在")
	}

	localTime := ext.GetCurrentTime()
	err = global.Db.Model(&system.Setting{}).Scopes(utils.IsDeleteSoft).Updates(
		system.Setting{BaseModel: model.BaseModel{
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
// @receiver: settingService
// @param: info
// @param: list
// @param: count
// @return: error
func (settingService *SettingService) Query(info *dto.SettingQueryCriteria, list *[]system.Setting, count *int64) error {
	limit := info.PageSize
	offset := info.PageSize * (info.PageIndex - 1)
	// 创建db
	db := buildSettingQuery(global.Db.Model(&system.Setting{}), info)

	err := db.Count(count).Limit(limit).Offset(offset).Find(&list).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		return err
	}
	return nil
}

// Download
// @description: 导出
// @receiver: settingService
// @param: info
// @return: filePath
// @return: fileName
// @return: err
func (settingService *SettingService) Download(info *dto.SettingQueryCriteria) (filePath string, fileName string, err error) {
	var settings []system.Setting
	// 创建db并构建查询条件
	err = buildSettingQuery(global.Db.Model(&system.Setting{}), info).Find(&settings).Error

	if err != nil {
		return
	}
	// 创建一个新的 Excel 文件
	file := xlsx.NewFile()
	sheet, err := file.AddSheet("Settings")
	if err != nil {
		return
	}
	style := xlsx.NewStyle()
	style.Fill = *xlsx.NewFill("solid", "1e90ff", "1e90ff")
	style.ApplyFill = true
	// 表头数据
	header := []string{"ID", "键", "值", "是否启用", "描述", "创建时间"}
	// 将表头添加到 Excel 工作表
	row := sheet.AddRow()
	for _, cellValue := range header {
		cell := row.AddCell()
		cell.Value = cellValue
		cell.SetStyle(style)
	}

	// 将数据添加到 Excel 工作表
	for _, setting := range settings {
		row := sheet.AddRow()
		row.WriteSlice(&[]interface{}{
			strconv.FormatInt(setting.Id, 10),
			setting.Name,
			setting.Value,
			func() string {
				if setting.Enabled {
					return "启用"
				}
				return "禁用"
			}(),
			setting.Description,
			setting.CreateTime.Format("2006-01-02 15:04:05"),
		}, -1)
	}
	// 保存 Excel 文件到本地
	fileName = "Settings_" + utils.GenerateID().String() + ".xlsx"
	filePath = global.Config.Excel.Dir + fileName
	err = file.Save(filePath)
	return filePath, fileName, err
}

// FindSettingByName
// @description: 查询
// @receiver: settingService
// @param: settingName
// @return: setting
// @return: err
func (settingService *SettingService) FindSettingByName(settingName string, setting *system.Setting) error {

	err := global.Db.Model(&system.Setting{}).Scopes(utils.IsDeleteSoft).Where("name = ? ", settingName).First(setting).Error

	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		return err
	}
	return nil
}

// @description: 条件表达式
// @param: db
// @param: info
// @return: *gorm.DB

func buildSettingQuery(db *gorm.DB, info *dto.SettingQueryCriteria) *gorm.DB {
	if info.KeyWords != "" {
		db = db.Where("name LIKE ? or value LIKE ? or description LIKE ?", "%"+info.KeyWords+"%", "%"+info.KeyWords+"%", "%"+info.KeyWords+"%")
	}
	if info.Enabled != nil {
		db = db.Where("enabled = ?", info.Enabled)
	}
	if len(info.SortFields) > 0 {
		for _, sort := range info.SortFields {
			db = db.Order(sort)
		}
	}
	db = db.Scopes(utils.IsDeleteSoft)

	return db
}
