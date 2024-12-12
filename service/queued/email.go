package queued

import (
	"context"
	"errors"
	"go-apevolo/global"
	"go-apevolo/global/constants/cachePrefix"
	"go-apevolo/model"
	"go-apevolo/model/message/email"
	"go-apevolo/model/queued"
	"go-apevolo/payloads/dto"
	"go-apevolo/payloads/request"
	"go-apevolo/utils"
	emailSend "go-apevolo/utils/email"
	"go-apevolo/utils/ext"
	"go-apevolo/utils/redis"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"math/rand"
	"strconv"
	"time"
)

type EmailQueuedService struct{}

// Create
// @description: 创建
// @receiver: emailQueuedService
// @param: req
// @return: error
func (emailQueuedService *EmailQueuedService) Create(req *dto.CreateUpdateEmailQueuedDto) error {
	emailQueued := &queued.Email{}
	req.Generate(emailQueued)
	err := global.Db.Create(&emailQueued).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		return err
	}
	return nil
}

// Update
// @description: 编辑
// @receiver: emailQueuedService
// @param: req
// @return: error
func (emailQueuedService *EmailQueuedService) Update(req *dto.CreateUpdateEmailQueuedDto) error {
	emailQueued := &queued.Email{}
	oldEmailQueued := &queued.Email{}
	err := global.Db.Scopes(utils.IsDeleteSoft).First(&oldEmailQueued, req.Id).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("数据不存在或您无权查看！")
		}
		return err
	}
	req.Generate(emailQueued)
	err = global.Db.Updates(emailQueued).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		return err
	}
	return nil
}

// Delete
// @description: 删除
// @receiver: emailQueuedService
// @param: idArray
// @param: updateBy
// @return: error
func (emailQueuedService *EmailQueuedService) Delete(idArray request.IdCollection, updateBy string) error {
	var emailQueuedList []queued.Email
	err := global.Db.Scopes(utils.IsDeleteSoft).Find(&emailQueuedList, "id in ?", idArray.IdArray).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		return err
	}
	if len(emailQueuedList) == 0 {
		return errors.New("数据不存在或您无权查看！")
	}
	localTime := ext.GetCurrentTime()
	err = global.Db.Model(&queued.Email{}).Scopes(utils.IsDeleteSoft).Where("id in (?)", idArray.IdArray).Updates(
		queued.Email{BaseModel: model.BaseModel{
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
// @receiver: emailQueuedService
// @param: info
// @param: list
// @param: count
// @return: error
func (emailQueuedService *EmailQueuedService) Query(info *dto.EmailQueuedQueryCriteria, list *[]queued.Email, count *int64) error {
	limit := info.PageSize
	offset := info.PageSize * (info.PageIndex - 1)
	// 创建db并构建查询条件
	db := buildEmailQueuedQuery(global.Db.Model(&queued.Email{}), info)
	err := db.Count(count).Limit(limit).Offset(offset).Find(list).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		return err
	}
	return nil
}

// ResetEmail
// @description: 重置邮箱验证码
// @receiver: emailQueuedService
// @param: emailAddress
// @param: messageTemplateName
// @return: error
func (emailQueuedService *EmailQueuedService) ResetEmail(emailAddress string, messageTemplateName string) error {
	var messageTemplate email.MessageTemplate
	var emailAccount email.Account
	err := global.Db.Scopes(utils.IsDeleteSoft).Where("name = ? ", messageTemplateName).First(&messageTemplate).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("数据不存在或您无权查看！")
		}
		return err
	}
	err = global.Db.Scopes(utils.IsDeleteSoft).First(&emailAccount, messageTemplate.EmailAccountId).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		return err
	}
	rand.Seed(time.Now().UnixNano())
	randomNumber := rand.Intn(900000) + 100000
	global.Redis.Del(context.Background(), cachePrefix.EmailCaptcha+utils.MD5(emailAddress))
	err = redis.Set(cachePrefix.EmailCaptcha+utils.MD5(emailAddress), randomNumber, time.Minute*time.Duration(5))
	if err != nil {
		return err
	}
	emailQueued := &dto.CreateUpdateEmailQueuedDto{}
	emailQueued.To = emailAddress
	emailQueued.Priority = 1
	emailQueued.Bcc = messageTemplate.BccEmailAddresses
	emailQueued.Subject = messageTemplate.Subject
	emailQueued.Body = ext.StringReplace(string(messageTemplate.Body), "%captcha%", strconv.Itoa(randomNumber), -1)
	emailQueued.SentTries = 1
	emailQueued.EmailAccountId = emailAccount.Id
	err = emailSend.SendEmail(emailAccount.Email, emailAccount.DisplayName, emailAccount.Password, emailAccount.Host, emailAccount.Port, emailAccount.EnableSsl, []string{emailQueued.To}, []string{}, []string{}, emailQueued.Subject, emailQueued.Body)
	if err == nil {
		localTime := ext.GetCurrentTime()
		emailQueued.SendTime = &localTime
	}
	err = emailQueuedService.Create(emailQueued)
	return err
}

// buildEmailQueuedQuery
// @description: 查询表达式
// @param: db
// @param: info
// @return: *gorm.DB
func buildEmailQueuedQuery(db *gorm.DB, info *dto.EmailQueuedQueryCriteria) *gorm.DB {
	if info.Id != nil {
		db = db.Where("id = ?", info.Id)
	}
	if info.MaxTries > 0 {
		db = db.Where("sent_tries < ?", info.MaxTries)
	}
	if info.EmailAccountId != nil {
		db = db.Where("email_account_id = ? ", info.EmailAccountId)
	}
	if info.To != "" {
		db = db.Where("to like ? or to_name like ?", "%"+info.To+"%", "%"+info.To+"%")
	}
	if info.IsSend != nil {
		if *info.IsSend == true {
			db = db.Where("send_time is not null")
		} else {
			db = db.Where("send_time is null")
		}
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
