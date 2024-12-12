package email

import (
	"errors"
	"github.com/tealeg/xlsx"
	"go-apevolo/global"
	"go-apevolo/model"
	"go-apevolo/model/message/email"
	"go-apevolo/payloads/dto"
	"go-apevolo/payloads/request"
	"go-apevolo/utils"
	"go-apevolo/utils/ext"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"strconv"
)

type AccountService struct{}

// Create
// @description: 创建
// @receiver: accountService
// @param: req
// @return: error
func (accountService *AccountService) Create(req *dto.CreateUpdateEmailAccountDto) error {
	emailAccount := &email.Account{}
	var count int64
	err := global.Db.Model(&email.Account{}).Scopes(utils.IsDeleteSoft).Where("email = ?", req.Email).Count(&count).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		return err
	}
	if count > 0 {
		return errors.New("邮箱账户=>" + req.Email + "=>已存在!")
	}
	req.Generate(emailAccount)
	err = global.Db.Create(emailAccount).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		return err
	}
	return nil
}

// Update
// @description: 编辑
// @receiver: accountService
// @param: req
// @return: error
func (accountService *AccountService) Update(req *dto.CreateUpdateEmailAccountDto) error {
	oldEmailAccount := &email.Account{}
	emailAccount := &email.Account{}
	err := global.Db.Scopes(utils.IsDeleteSoft).First(oldEmailAccount, req.Id).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("数据不存在或您无权查看！")
		}
		return err
	}
	if oldEmailAccount.Email != req.Email {
		var count int64
		err = global.Db.Model(&email.Account{}).Scopes(utils.IsDeleteSoft).Where("email = ?", req.Email).Count(&count).Error
		if err != nil {
			global.Logger.Error("db error: ", zap.Error(err))
			return err
		}
		if count > 0 {
			return errors.New("邮箱账户=>" + req.Email + "=>已存在!")
		}
	}
	req.Generate(emailAccount)
	err = global.Db.Updates(emailAccount).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		return err
	}
	return nil
}

// Delete
// @description: 删除
// @receiver: accountService
// @param: idArray
// @param: updateBy
// @return: error
func (accountService *AccountService) Delete(idArray request.IdCollection, updateBy string) error {
	var emailAccounts []email.Account
	err := global.Db.Scopes(utils.IsDeleteSoft).Find(&emailAccounts, "id in ?", idArray.IdArray).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		return err
	}
	if len(emailAccounts) == 0 {
		return errors.New("数据不存在或您无权查看！")
	}
	localTime := ext.GetCurrentTime()
	err = global.Db.Model(&email.Account{}).Scopes(utils.IsDeleteSoft).Where("id in (?)", idArray.IdArray).Updates(
		email.Account{BaseModel: model.BaseModel{
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
// @receiver: accountService
// @param: info
// @param: list
// @param: count
// @return: error
func (accountService *AccountService) Query(info *dto.EmailAccountQueryCriteria, list *[]email.Account, count *int64) error {
	limit := info.PageSize
	offset := info.PageSize * (info.PageIndex - 1)
	// 创建db并构建查询条件
	db := buildEmailAccountQuery(global.Db.Model(&email.Account{}), info)

	err := db.Count(count).Limit(limit).Offset(offset).Find(list).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		return err
	}
	return nil
}

// Download
// @description: 导出
// @receiver: accountService
// @param: info
// @return: filePath
// @return: fileName
// @return: err
func (accountService *AccountService) Download(info *dto.EmailAccountQueryCriteria) (filePath string, fileName string, err error) {
	var emailAccounts []email.Account
	// 创建db并构建查询条件
	err = buildEmailAccountQuery(global.Db.Model(&email.Account{}), info).Find(&emailAccounts).Error

	if err != nil {
		return
	}
	// 创建一个新的 Excel 文件
	file := xlsx.NewFile()
	sheet, err := file.AddSheet("EmailAccounts")
	if err != nil {
		return
	}
	style := xlsx.NewStyle()
	style.Fill = *xlsx.NewFill("solid", "1e90ff", "1e90ff")
	style.ApplyFill = true
	// 表头数据
	header := []string{"ID", "邮箱", "显示名称", "主机", "端口", "账户名称", "密码", "是否SSL", "发送默认系统凭据", "创建时间"}
	// 将表头添加到 Excel 工作表
	row := sheet.AddRow()
	for _, cellValue := range header {
		cell := row.AddCell()
		cell.Value = cellValue
		cell.SetStyle(style)
	}

	// 将数据添加到 Excel 工作表
	for _, emailAccount := range emailAccounts {
		row := sheet.AddRow()
		row.WriteSlice(&[]interface{}{
			strconv.FormatInt(emailAccount.Id, 10),
			emailAccount.Email,
			emailAccount.DisplayName,
			emailAccount.Host,
			emailAccount.Port,
			emailAccount.Username,
			emailAccount.Password,
			func() string {
				if emailAccount.EnableSsl {
					return "是"
				}
				return "否"
			}(),
			func() string {
				if emailAccount.UseDefaultCredentials {
					return "是"
				}
				return "否"
			}(),
			emailAccount.CreateTime.Format("2006-01-02 15:04:05"),
		}, -1)
	}
	// 保存 Excel 文件到本地
	fileName = "EmailAccounts_" + utils.GenerateID().String() + ".xlsx"
	filePath = global.Config.Excel.Dir + fileName
	err = file.Save(filePath)
	return filePath, fileName, err
}

// buildEmailAccountQuery
// @description: 条件表达式
// @param: db
// @param: info
// @return: *gorm.DB
func buildEmailAccountQuery(db *gorm.DB, info *dto.EmailAccountQueryCriteria) *gorm.DB {
	if info.Email != "" {
		db = db.Where("email LIKE ?", "%"+info.Email+"%")
	}
	if info.DisplayName != "" {
		db = db.Where("display_name like ?", "%"+info.DisplayName+"%")
	}
	if info.Username != "" {
		db = db.Where("username like ?", "%"+info.Username+"%")
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
