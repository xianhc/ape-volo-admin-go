package emailaccount

import (
	"context"
	"errors"
	"time"

	"github.com/xianhc/ape-volo-admin-go/internal/utils"
	"gorm.io/gorm"
)

// errNotFound 表示未找到可见的邮件账户。
//
// repository 用该错误隔离 GORM 的记录不存在错误。
var errNotFound = errors.New("email account not found")

// queryOptions 描述邮件账户列表查询的数据库参数。
//
// Service 负责从 HTTP 请求转换该结构。
type queryOptions struct {
	Email, DisplayName, UserName string
	CreateTime                   []time.Time
	Sorting                      string
	Limit, Offset                int
}

// repository 封装邮件账户的数据库访问。
//
// 零值即可使用，数据库连接从 context 中获取。
type repository struct{}

// ExistsByEmail 判断邮箱地址是否已存在。
//
// 查询继承软删除和数据权限过滤。
func (repository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var total int64
	err := utils.FromContext(ctx).Model(&Account{}).Where("email = ?", email).Count(&total).Error
	return total > 0, err
}

// FindByID 按主键读取邮件账户。
//
// 未找到记录时返回 errNotFound。
func (repository) FindByID(ctx context.Context, id int64) (Account, error) {
	var account Account
	err := utils.FromContext(ctx).First(&account, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Account{}, errNotFound
	}
	return account, err
}

// Create 持久化新的邮件账户。
//
// 创建审计字段由 ORM 回调自动填充。
func (repository) Create(ctx context.Context, account *Account) error {
	return utils.FromContext(ctx).Create(account).Error
}

// Update 按主键更新邮件账户。
//
// 修改审计字段由 ORM 回调自动填充。
func (repository) Update(ctx context.Context, account *Account) error {
	return utils.FromContext(ctx).Select("Email", "DisplayName", "Host", "Port", "UserName", "Password", "EnableSsl", "UseDefaultCredentials", "LastModifierId", "LastModificationTime").Updates(account).Error
}

// Delete 逻辑删除指定邮件账户。
//
// 返回受影响行数，Service 据此判断数据是否存在。
func (repository) Delete(ctx context.Context, ids []string) (int64, error) {
	result := utils.FromContext(ctx).Delete(&Account{}, ids)
	return result.RowsAffected, result.Error
}

// List 查询一页邮件账户及分页前总数。
//
// 查询支持邮箱、显示名称、用户名和创建时间范围筛选。
func (repository) List(ctx context.Context, options queryOptions) ([]Account, int64, error) {
	accounts := make([]Account, 0)
	var total int64
	db := buildQuery(utils.FromContext(ctx).Model(&Account{}), options)
	if err := db.Count(&total).Limit(options.Limit).Offset(options.Offset).Find(&accounts).Error; err != nil {
		return nil, 0, err
	}
	return accounts, total, nil
}

// All 按创建时间升序查询全部邮件账户。
//
// 该方法供前端下拉框使用。
func (repository) All(ctx context.Context) ([]Account, error) {
	accounts := make([]Account, 0)
	err := utils.FromContext(ctx).Order("creation_time asc").Find(&accounts).Error
	return accounts, err
}

func buildQuery(db *gorm.DB, options queryOptions) *gorm.DB {
	if options.Email != "" {
		db = db.Where("email LIKE ?", "%"+options.Email+"%")
	}
	if options.DisplayName != "" {
		db = db.Where("display_name like ?", "%"+options.DisplayName+"%")
	}
	if options.UserName != "" {
		db = db.Where("user_name like ?", "%"+options.UserName+"%")
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
