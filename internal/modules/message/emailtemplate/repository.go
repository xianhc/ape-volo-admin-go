package emailtemplate

import (
	"context"
	"errors"
	"time"

	"github.com/xianhc/ape-volo-admin-go/internal/utils"
	"gorm.io/gorm"
)

// errNotFound 表示未找到可见的邮件模板。
//
// repository 用该错误隔离 GORM 的记录不存在错误。
var errNotFound = errors.New("email template not found")

// queryOptions 描述邮件模板列表查询的数据库参数。
//
// Service 负责从 HTTP 请求转换该结构。
type queryOptions struct {
	Name          string
	Enabled       *bool
	CreateTime    []time.Time
	Sorting       string
	Limit, Offset int
}

// repository 封装邮件模板的数据库访问。
//
// 零值即可使用，数据库连接从 context 中获取。
type repository struct{}

// ExistsByName 判断模板名称是否已存在。
//
// 查询继承软删除和数据权限过滤。
func (repository) ExistsByName(ctx context.Context, name string) (bool, error) {
	var total int64
	err := utils.FromContext(ctx).Model(&MessageTemplate{}).Where("name = ?", name).Count(&total).Error
	return total > 0, err
}

// FindByID 按主键读取邮件模板。
//
// 未找到记录时返回 errNotFound。
func (repository) FindByID(ctx context.Context, id int64) (MessageTemplate, error) {
	var template MessageTemplate
	err := utils.FromContext(ctx).First(&template, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return MessageTemplate{}, errNotFound
	}
	return template, err
}

// FindByName 按名称读取邮件模板。
//
// 未找到记录时返回 errNotFound。
func (repository) FindByName(ctx context.Context, name string) (MessageTemplate, error) {
	var template MessageTemplate
	err := utils.FromContext(ctx).Where("name = ?", name).First(&template).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return MessageTemplate{}, errNotFound
	}
	return template, err
}

// Create 持久化新的邮件模板。
//
// 创建审计字段由 ORM 回调自动填充。
func (repository) Create(ctx context.Context, template *MessageTemplate) error {
	return utils.FromContext(ctx).Create(template).Error
}

// Update 按主键更新邮件模板。
//
// 修改审计字段由 ORM 回调自动填充。
func (repository) Update(ctx context.Context, template *MessageTemplate) error {
	return utils.FromContext(ctx).Select("Name", "BccEmailAddresses", "Subject", "Body", "Enabled", "EmailAccountId", "LastModifierId", "LastModificationTime").Updates(template).Error
}

// Delete 逻辑删除指定邮件模板。
//
// 返回受影响行数，Service 据此判断数据是否存在。
func (repository) Delete(ctx context.Context, ids []string) (int64, error) {
	result := utils.FromContext(ctx).Delete(&MessageTemplate{}, ids)
	return result.RowsAffected, result.Error
}

// List 查询一页邮件模板及分页前总数。
//
// 查询支持名称、激活状态和创建时间范围筛选。
func (repository) List(ctx context.Context, options queryOptions) ([]MessageTemplate, int64, error) {
	templates := make([]MessageTemplate, 0)
	var total int64
	db := buildQuery(utils.FromContext(ctx).Model(&MessageTemplate{}), options)
	if err := db.Count(&total).Limit(options.Limit).Offset(options.Offset).Find(&templates).Error; err != nil {
		return nil, 0, err
	}
	return templates, total, nil
}

func buildQuery(db *gorm.DB, options queryOptions) *gorm.DB {
	if options.Name != "" {
		db = db.Where("name LIKE ?", "%"+options.Name+"%")
	}
	if options.Enabled != nil {
		db = db.Where("enabled = ?", options.Enabled)
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
