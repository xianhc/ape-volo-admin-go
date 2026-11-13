package queuedemail

import (
	"context"
	"errors"
	"time"

	"github.com/xianhc/ape-volo-admin-go/internal/model"
	"github.com/xianhc/ape-volo-admin-go/internal/utils"
	"gorm.io/gorm"
)

// errNotFound 表示未找到可见的队列邮件。
//
// repository 用该错误隔离 GORM 的记录不存在错误。
var errNotFound = errors.New("queued email not found")

// queryOptions 描述队列邮件列表查询的数据库参数。
//
// Service 负责从 HTTP 请求转换该结构。
type queryOptions struct {
	Id             *int64
	MaxTries       int32
	EmailAccountId *int64
	To             string
	IsSend         *bool
	CreateTime     []time.Time
	Sorting        string
	Limit, Offset  int
}

// repository 封装队列邮件的数据库访问。
//
// 零值即可使用，数据库连接从 context 中获取。
type repository struct{}

// FindByID 按主键读取队列邮件。
//
// 未找到记录时返回 errNotFound。
func (repository) FindByID(ctx context.Context, id int64) (QueuedEmail, error) {
	var email QueuedEmail
	err := utils.FromContext(ctx).First(&email, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return QueuedEmail{}, errNotFound
	}
	return email, err
}

// Create 持久化新的队列邮件。
//
// 创建审计字段由 ORM 回调自动填充。
func (repository) Create(ctx context.Context, email *QueuedEmail) error {
	return utils.FromContext(ctx).Create(email).Error
}

// Update 按主键更新队列邮件。
//
// 修改审计字段由 ORM 回调自动填充。
func (repository) Update(ctx context.Context, email *QueuedEmail) error {
	return utils.FromContext(ctx).Select("From", "FromName", "To", "ToName", "ReplyTo", "ReplyToName", "Priority", "Cc", "Bcc", "Subject", "Body", "SentTries", "IsSend", "SendTime", "EmailAccountId", "LastModifierId", "LastModificationTime").Updates(email).Error
}

// UpdateDelivery 更新发送状态、发送时间和重试次数。
//
// 该方法供定时发送任务使用，不修改邮件正文等业务字段。
func (repository) UpdateDelivery(ctx context.Context, id int64, isSend bool, sendTime *model.Time, sentTries int32) error {
	return utils.FromContext(ctx).Model(&QueuedEmail{}).Where("id = ?", id).Updates(map[string]any{"is_send": isSend, "send_time": sendTime, "sent_tries": sentTries}).Error
}

// Delete 逻辑删除指定队列邮件。
//
// 返回受影响行数，Service 据此判断数据是否存在。
func (repository) Delete(ctx context.Context, ids []string) (int64, error) {
	result := utils.FromContext(ctx).Delete(&QueuedEmail{}, ids)
	return result.RowsAffected, result.Error
}

// List 查询一页队列邮件及分页前总数。
//
// 查询支持发送状态和重试次数等队列筛选条件。
func (repository) List(ctx context.Context, options queryOptions) ([]QueuedEmail, int64, error) {
	emails := make([]QueuedEmail, 0)
	var total int64
	db := buildQuery(utils.FromContext(ctx).Model(&QueuedEmail{}), options)
	if err := db.Count(&total).Limit(options.Limit).Offset(options.Offset).Find(&emails).Error; err != nil {
		return nil, 0, err
	}
	return emails, total, nil
}

func buildQuery(db *gorm.DB, options queryOptions) *gorm.DB {
	if options.Id != nil {
		db = db.Where("id = ?", options.Id)
	}
	if options.MaxTries > 0 {
		db = db.Where("sent_tries < ?", options.MaxTries)
	}
	if options.EmailAccountId != nil {
		db = db.Where("email_account_id = ? ", options.EmailAccountId)
	}
	if options.To != "" {
		db = db.Where("to like ? or to_name like ?", "%"+options.To+"%", "%"+options.To+"%")
	}
	if options.IsSend != nil {
		db = db.Where("is_send = ?", *options.IsSend)
	}
	if len(options.CreateTime) != 0 && len(options.CreateTime) != 2 {
		_ = db.AddError(errors.New("创建时间必须包含起止两个值"))
		return db
	}
	if len(options.CreateTime) == 2 {
		db = db.Where("creation_time >= ? and creation_time <= ? ", options.CreateTime[0], options.CreateTime[1])
	}
	if options.Sorting != "" {
		db = utils.ApplySorting(db, options.Sorting)
	}
	return db
}
