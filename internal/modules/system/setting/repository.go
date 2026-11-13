package setting

import (
	"context"
	"errors"

	"github.com/xianhc/ape-volo-admin-go/internal/utils"
	"gorm.io/gorm"
)

// errNotFound 表示按条件未找到可见的系统设置。
//
// repository 用该错误隔离 GORM 的记录不存在错误。
var errNotFound = errors.New("setting not found")

// queryOptions 描述系统设置列表查询的数据库参数。
//
// Service 负责从 HTTP 请求转换该结构，repository 不依赖请求模型。
type queryOptions struct {
	KeyWords string
	Enabled  *bool
	Sorting  string
	Limit    int
	Offset   int
}

// repository 封装系统设置的数据库访问。
//
// 零值即可使用，数据库连接和事务从 context 中获取。
type repository struct{}

// ExistsByName 判断指定设置键是否已存在。
//
// 查询继承当前请求的数据权限与软删除过滤。
func (repository) ExistsByName(ctx context.Context, name string) (bool, error) {
	var total int64
	err := utils.FromContext(ctx).Model(&Setting{}).Where("name = ?", name).Count(&total).Error
	return total > 0, err
}

// FindByID 按主键读取系统设置。
//
// 未找到记录时返回 errNotFound，其余数据库错误原样返回。
func (repository) FindByID(ctx context.Context, id int64) (Setting, error) {
	var setting Setting
	err := utils.FromContext(ctx).First(&setting, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Setting{}, errNotFound
	}
	return setting, err
}

// FindByName 按设置键读取系统设置。
//
// 未找到记录时返回 errNotFound，其余数据库错误原样返回。
func (repository) FindByName(ctx context.Context, name string) (Setting, error) {
	var setting Setting
	err := utils.FromContext(ctx).Where("name = ?", name).First(&setting).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Setting{}, errNotFound
	}
	return setting, err
}

// Create 持久化新的系统设置。
//
// 创建审计字段由 ORM 回调根据 context 自动填充。
func (repository) Create(ctx context.Context, setting *Setting) error {
	return utils.FromContext(ctx).Create(setting).Error
}

// Update 按实体主键更新系统设置。
//
// 修改审计字段由 ORM 回调根据 context 自动填充。
func (repository) Update(ctx context.Context, setting *Setting) error {
	return utils.FromContext(ctx).Select("Name", "Value", "Enabled", "Description", "LastModifierId", "LastModificationTime").Updates(setting).Error
}

// Delete 逻辑删除指定主键集合对应的系统设置。
//
// 返回受影响行数，Service 据此判断数据是否存在。
func (repository) Delete(ctx context.Context, ids []string) (int64, error) {
	result := utils.FromContext(ctx).Delete(&Setting{}, ids)
	return result.RowsAffected, result.Error
}

// List 查询一页系统设置及分页前总数。
//
// 筛选、排序和分页在同一个查询构造器上执行。
func (repository) List(ctx context.Context, options queryOptions) ([]Setting, int64, error) {
	settings := make([]Setting, 0)
	var total int64
	db := buildQuery(utils.FromContext(ctx).Model(&Setting{}), options)
	if err := db.Count(&total).Limit(options.Limit).Offset(options.Offset).Find(&settings).Error; err != nil {
		return nil, 0, err
	}
	return settings, total, nil
}

// FindAll 查询符合条件的全部系统设置。
//
// 该方法用于导出，不应用分页参数。
func (repository) FindAll(ctx context.Context, options queryOptions) ([]Setting, error) {
	settings := make([]Setting, 0)
	err := buildQuery(utils.FromContext(ctx).Model(&Setting{}), options).Find(&settings).Error
	return settings, err
}

func buildQuery(db *gorm.DB, options queryOptions) *gorm.DB {
	if options.KeyWords != "" {
		db = db.Where("name LIKE ? or value LIKE ? or description LIKE ?", "%"+options.KeyWords+"%", "%"+options.KeyWords+"%", "%"+options.KeyWords+"%")
	}
	if options.Enabled != nil {
		db = db.Where("enabled = ?", options.Enabled)
	}
	if options.Sorting != "" {
		db = utils.ApplySorting(db, options.Sorting)
	}
	return db
}
