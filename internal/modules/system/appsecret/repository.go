package appsecret

import (
	"context"
	"errors"

	"github.com/xianhc/ape-volo-admin-go/internal/utils"
	"gorm.io/gorm"
)

// errNotFound 表示按主键未找到可见的应用密钥。
//
// repository 将 GORM 的记录不存在错误转换为该稳定错误，避免 Service 依赖 ORM 实现。
var errNotFound = errors.New("app secret not found")

// queryOptions 描述应用密钥列表查询的数据库参数。
//
// Service 负责把 HTTP 查询请求转换为该结构，repository 不依赖请求模型。
type queryOptions struct {
	KeyWords   string
	CreateTime []string
	Sorting    string
	Limit      int
	Offset     int
}

// repository 封装应用密钥的数据库访问。
//
// 零值即可使用，数据库连接和事务从 context 中获取。
type repository struct{}

// ExistsByName 判断指定应用名称是否已存在。
//
// 查询自动继承软删除与数据权限回调，只统计当前请求可见的数据。
func (repository) ExistsByName(ctx context.Context, appName string) (bool, error) {
	var total int64
	err := utils.FromContext(ctx).Model(&AppSecret{}).Where("app_name = ?", appName).Count(&total).Error
	return total > 0, err
}

// FindByID 按主键读取单个应用密钥实体。
//
// 未找到记录时返回 errNotFound，其余数据库错误原样返回。
func (repository) FindByID(ctx context.Context, id int64) (AppSecret, error) {
	var appSecret AppSecret
	err := utils.FromContext(ctx).Where("id = ?", id).First(&appSecret).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return AppSecret{}, errNotFound
	}
	return appSecret, err
}

// Create 持久化新的应用密钥实体。
//
// 创建人与创建时间由 ORM 审计回调根据 context 自动填充。
func (repository) Create(ctx context.Context, appSecret *AppSecret) error {
	return utils.FromContext(ctx).Create(appSecret).Error
}

// Update 按实体主键更新应用密钥。
//
// 修改人与修改时间由 ORM 审计回调根据 context 自动填充。
func (repository) Update(ctx context.Context, appSecret *AppSecret) error {
	return utils.FromContext(ctx).Select("AppName", "Remark", "LastModifierId", "LastModificationTime").Updates(appSecret).Error
}

// Delete 逻辑删除指定主键集合对应的应用密钥。
//
// 返回受影响行数，调用方据此区分删除成功与数据不存在。
func (repository) Delete(ctx context.Context, ids []string) (int64, error) {
	result := utils.FromContext(ctx).Delete(&AppSecret{}, ids)
	return result.RowsAffected, result.Error
}

// List 查询一页应用密钥及分页前总数。
//
// 筛选、排序和分页在同一个数据库查询构造器上执行。
func (repository) List(ctx context.Context, options queryOptions) ([]AppSecret, int64, error) {
	appSecrets := make([]AppSecret, 0)
	var total int64
	db := buildQuery(utils.FromContext(ctx).Model(&AppSecret{}), options)
	if err := db.Count(&total).Limit(options.Limit).Offset(options.Offset).Find(&appSecrets).Error; err != nil {
		return nil, 0, err
	}
	return appSecrets, total, nil
}

// FindAll 查询符合条件的全部应用密钥。
//
// 该方法用于导出，不应用分页参数，但保留筛选与排序条件。
func (repository) FindAll(ctx context.Context, options queryOptions) ([]AppSecret, error) {
	appSecrets := make([]AppSecret, 0)
	err := buildQuery(utils.FromContext(ctx).Model(&AppSecret{}), options).Find(&appSecrets).Error
	return appSecrets, err
}

func buildQuery(db *gorm.DB, options queryOptions) *gorm.DB {
	if options.KeyWords != "" {
		db = db.Where("app_id LIKE ? or app_name LIKE ? or remark LIKE ?", "%"+options.KeyWords+"%", "%"+options.KeyWords+"%", "%"+options.KeyWords+"%")
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
