package queuedemail

import (
	"context"
	"crypto/rand"
	"errors"
	"math/big"
	"strconv"
	"time"

	"github.com/xianhc/ape-volo-admin-go/internal/global"
	"github.com/xianhc/ape-volo-admin-go/internal/global/constants/cachePrefix"
	"github.com/xianhc/ape-volo-admin-go/internal/model"
	"github.com/xianhc/ape-volo-admin-go/internal/modules/message/emailaccount"
	"github.com/xianhc/ape-volo-admin-go/internal/modules/message/emailtemplate"
	"github.com/xianhc/ape-volo-admin-go/internal/utils"
	"github.com/xianhc/ape-volo-admin-go/internal/utils/cache"
	emailsend "github.com/xianhc/ape-volo-admin-go/internal/utils/email"
	"github.com/xianhc/ape-volo-admin-go/internal/utils/ext"
	"go.uber.org/zap"
)

// Service 实现队列邮件的业务规则与验证码发送流程。
//
// Service 只通过 repository 访问队列表，并通过账户和模板 Service 获取跨模块数据。
type Service struct{ repository repository }

// Create 创建队列邮件。
//
// 未提供主键时由 Service 生成主键。
func (service *Service) Create(ctx context.Context, req CreateAndUpdateQueuedEmailRequest) error {
	return service.create(ctx, req, false)
}

func (service *Service) create(ctx context.Context, req CreateAndUpdateQueuedEmailRequest, isSend bool) error {
	account, err := (&emailaccount.Service{}).GetCredentials(ctx, req.EmailAccountId)
	if err != nil {
		return err
	}
	email := newEntity(req.Id, req.To, req.ToName, req.ReplyTo, req.ReplyToName, req.Priority, req.Cc, req.Bcc, req.Subject, req.Body, req.SentTries, req.SendTime, req.EmailAccountId)
	email.From, email.FromName, email.IsSend = account.Email, account.DisplayName, isSend
	if err := service.repository.Create(ctx, email); err != nil {
		return logDatabaseError(err)
	}
	return nil
}

// Update 修改队列邮件。
//
// 方法确认目标记录存在后更新请求允许的字段。
func (service *Service) Update(ctx context.Context, req CreateAndUpdateQueuedEmailRequest) error {
	_, err := service.repository.FindByID(ctx, req.Id)
	if errors.Is(err, errNotFound) {
		return errors.New("数据不存在或您无权查看！")
	}
	if err != nil {
		return logDatabaseError(err)
	}
	account, err := (&emailaccount.Service{}).GetCredentials(ctx, req.EmailAccountId)
	if err != nil {
		return err
	}
	email := newEntity(req.Id, req.To, req.ToName, req.ReplyTo, req.ReplyToName, req.Priority, req.Cc, req.Bcc, req.Subject, req.Body, req.SentTries, req.SendTime, req.EmailAccountId)
	email.From, email.FromName = account.Email, account.DisplayName
	if err := service.repository.Update(ctx, email); err != nil {
		return logDatabaseError(err)
	}
	return nil
}

// Delete 逻辑删除一组队列邮件。
//
// 没有记录受影响时返回统一不可见错误。
func (service *Service) Delete(ctx context.Context, ids []string) error {
	rows, err := service.repository.Delete(ctx, ids)
	if err != nil {
		return logDatabaseError(err)
	}
	if rows == 0 {
		return errors.New("数据不存在或您无权查看！")
	}
	return nil
}

// Query 查询队列邮件分页列表。
//
// 返回前统一把 Entity 转换为 QueuedEmailResponse。
func (service *Service) Query(ctx context.Context, req QueryQueuedEmailRequest) ([]QueuedEmailResponse, int64, error) {
	emails, total, err := service.repository.List(ctx, listOptions(req))
	if err != nil {
		return nil, 0, logDatabaseError(err)
	}
	return fromEntities(emails), total, nil
}

// Pending 查询定时任务需要处理的待发送邮件。
//
// 返回内部 Delivery DTO，不向任务暴露数据库 Entity。
func (service *Service) Pending(ctx context.Context, req QueryQueuedEmailRequest) ([]Delivery, int64, error) {
	emails, total, err := service.repository.List(ctx, listOptions(req))
	if err != nil {
		return nil, 0, logDatabaseError(err)
	}
	result := make([]Delivery, 0, len(emails))
	for _, email := range emails {
		result = append(result, Delivery{Id: email.Id, From: email.From, FromName: email.FromName, To: email.To, Subject: email.Subject, Body: email.Body, Priority: email.Priority, SentTries: email.SentTries, IsSend: email.IsSend, SendTime: email.SendTime, EmailAccountId: email.EmailAccountId})
	}
	return result, total, nil
}

// UpdateDelivery 更新邮件发送结果。
//
// 定时任务只更新发送时间和重试次数，其他邮件内容保持不变。
func (service *Service) UpdateDelivery(ctx context.Context, id int64, isSend bool, sendTime *model.Time, sentTries int32) error {
	if err := service.repository.UpdateDelivery(ctx, id, isSend, sendTime, sentTries); err != nil {
		return logDatabaseError(err)
	}
	return nil
}

// ResetEmail 发送邮箱验证码并记录队列邮件。
//
// 验证码使用加密随机数生成，缓存五分钟；无论发送结果如何都记录队列邮件。
func (service *Service) ResetEmail(ctx context.Context, emailAddress, templateName string) error {
	cooldownKey := cachePrefix.EmailCaptcha + "cooldown:" + emailAddress
	var cooldown int
	if err := global.Cache.Get(ctx, cooldownKey, &cooldown); err == nil && cooldown > 0 {
		return errors.New("验证码发送过于频繁，请稍后再试")
	} else if err != nil && !errors.Is(err, cache.ErrNotFound) {
		return err
	}
	template, err := (&emailtemplate.Service{}).GetByName(ctx, templateName)
	if err != nil {
		return err
	}
	account, err := (&emailaccount.Service{}).GetCredentials(ctx, template.EmailAccountId)
	if err != nil {
		return err
	}
	code, err := generateVerifyCode()
	if err != nil {
		return err
	}
	if err := global.Cache.Set(ctx, cachePrefix.EmailCaptcha+emailAddress, code, 5*time.Minute); err != nil {
		return err
	}
	req := CreateAndUpdateQueuedEmailRequest{To: emailAddress, Priority: 1, Bcc: template.BccEmailAddresses, Subject: template.Subject, Body: ext.StringReplace(template.Body, "%captcha%", strconv.Itoa(code), -1), SentTries: 1, EmailAccountId: account.Id}
	sendErr := emailsend.SendEmail(ctx, account.Email, account.DisplayName, account.Password, account.Host, account.Port, account.EnableSSL, []string{req.To}, []string{}, []string{}, req.Subject, req.Body)
	if sendErr == nil {
		now := model.Now()
		req.SendTime = &now
	}
	if err := service.create(ctx, req, sendErr == nil); err != nil {
		return err
	}
	if sendErr != nil {
		return sendErr
	}
	return global.Cache.Set(ctx, cooldownKey, 1, time.Minute)
}

func generateVerifyCode() (int, error) {
	value, err := rand.Int(rand.Reader, big.NewInt(900000))
	if err != nil {
		return 0, err
	}
	return int(value.Int64()) + 100000, nil
}
func newEntity(id int64, to, toName, replyTo, replyToName string, priority int32, cc, bcc, subject, body string, tries int32, sendTime *model.Time, accountId int64) *QueuedEmail {
	email := &QueuedEmail{To: to, ToName: toName, ReplyTo: replyTo, ReplyToName: replyToName, Priority: priority, Cc: cc, Bcc: bcc, Subject: subject, Body: body, SentTries: tries, SendTime: sendTime, EmailAccountId: accountId}
	if id == 0 {
		id = int64(utils.GenerateID())
	}
	email.Id = id
	return email
}
func listOptions(req QueryQueuedEmailRequest) queryOptions {
	return queryOptions{Id: req.Id, MaxTries: req.MaxTries, EmailAccountId: req.EmailAccountId, To: req.To, IsSend: req.IsSend, CreateTime: req.CreateTime, Sorting: req.Sorting, Limit: req.PageSize, Offset: req.PageSize * (req.PageIndex - 1)}
}
func logDatabaseError(err error) error { global.Logger.Error("db error: ", zap.Error(err)); return err }
