package position

import (
	"context"
	"errors"
	"time"

	"github.com/xianhc/ape-volo-admin-go/internal/utils"
	"gorm.io/gorm"
)

// errNotFound 表示未找到可见的岗位。
//
// repository 用该错误隔离 GORM 的记录不存在错误。
var errNotFound = errors.New("position not found")

// queryOptions 描述岗位列表查询的数据库参数。
//
// Service 负责从 HTTP 请求转换该结构。
type queryOptions struct {
	Name          string
	Enabled       *bool
	CreateTime    []time.Time
	Sorting       string
	Limit, Offset int
}

// repository 封装岗位及用户岗位关联的数据库访问。
//
// 零值即可使用，数据库连接从 context 中获取。
type repository struct{}

// ExistsByName 判断岗位名称是否已存在。
//
// 查询继承软删除和数据权限过滤。
func (repository) ExistsByName(ctx context.Context, name string) (bool, error) {
	var total int64
	err := utils.FromContext(ctx).Model(&Position{}).Where("name = ?", name).Count(&total).Error
	return total > 0, err
}

// FindByID 按主键读取岗位。
//
// 未找到记录时返回 errNotFound。
func (repository) FindByID(ctx context.Context, id int64) (Position, error) {
	var position Position
	err := utils.FromContext(ctx).First(&position, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Position{}, errNotFound
	}
	return position, err
}

// FindByIDs 查询一组岗位。
//
// 返回结果供删除前关联使用校验。
func (repository) FindByIDs(ctx context.Context, ids []string) ([]Position, error) {
	positions := make([]Position, 0)
	err := utils.FromContext(ctx).Find(&positions, "id in ?", ids).Error
	return positions, err
}

// HasUsers 判断岗位集合是否被用户引用。
//
// 任一用户岗位关联存在时返回 true。
func (repository) HasUsers(ctx context.Context, ids []int64) (bool, error) {
	var total int64
	err := utils.FromContext(ctx).Model(&UserPosition{}).Where("position_id in ?", ids).Count(&total).Error
	return total > 0, err
}

// Create 持久化新的岗位。
//
// 创建审计字段由 ORM 回调自动填充。
func (repository) Create(ctx context.Context, position *Position) error {
	return utils.FromContext(ctx).Create(position).Error
}

// Update 按主键更新岗位。
//
// 修改审计字段由 ORM 回调自动填充。
func (repository) Update(ctx context.Context, position *Position) error {
	return utils.FromContext(ctx).Select("Name", "Sort", "Enabled", "LastModifierId", "LastModificationTime").Updates(position).Error
}

// Delete 逻辑删除指定岗位。
//
// 返回受影响行数和数据库错误。
func (repository) Delete(ctx context.Context, ids []string) (int64, error) {
	result := utils.FromContext(ctx).Delete(&Position{}, ids)
	return result.RowsAffected, result.Error
}

// List 查询一页岗位及分页前总数。
//
// 查询支持名称、启用状态和创建时间范围筛选。
func (repository) List(ctx context.Context, options queryOptions) ([]Position, int64, error) {
	positions := make([]Position, 0)
	var total int64
	db := buildQuery(utils.FromContext(ctx).Model(&Position{}), options)
	if err := db.Count(&total).Limit(options.Limit).Offset(options.Offset).Find(&positions).Error; err != nil {
		return nil, 0, err
	}
	return positions, total, nil
}

// FindAll 查询符合条件的全部岗位。
//
// 该方法用于导出，不应用分页参数。
func (repository) FindAll(ctx context.Context, options queryOptions) ([]Position, error) {
	positions := make([]Position, 0)
	err := buildQuery(utils.FromContext(ctx).Model(&Position{}), options).Find(&positions).Error
	return positions, err
}

// AllEnabled 查询全部启用岗位。
//
// 返回结果按 Sort 升序排列。
func (repository) AllEnabled(ctx context.Context) ([]Position, error) {
	positions := make([]Position, 0)
	err := utils.FromContext(ctx).Where("enabled = 1").Order("sort asc").Find(&positions).Error
	return positions, err
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
