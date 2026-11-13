package department

import (
	"context"
	"errors"

	"github.com/xianhc/ape-volo-admin-go/internal/utils"
	"gorm.io/gorm"
)

// errNotFound 表示未找到可见的部门。
//
// repository 用该错误隔离 GORM 的记录不存在错误。
var errNotFound = errors.New("department not found")

// queryOptions 描述部门列表查询的数据库参数。
//
// ParentId 为 nil 时查询根节点。
type queryOptions struct {
	Name          string
	Enabled       *bool
	ParentId      *int64
	CreateTime    []string
	Sorting       string
	Limit, Offset int
	Paginate      bool
}

// repository 封装部门和关联使用情况的数据库访问。
//
// 零值即可使用，数据库连接从 context 中获取。
type repository struct{}

// ExistsByName 判断部门名称是否已存在。
//
// 查询继承软删除和数据权限过滤。
func (repository) ExistsByName(ctx context.Context, name string) (bool, error) {
	var total int64
	err := utils.FromContext(ctx).Model(&Department{}).Where("name = ?", name).Count(&total).Error
	return total > 0, err
}

// FindByID 按主键读取部门。
//
// 未找到记录时返回 errNotFound。
func (repository) FindByID(ctx context.Context, id int64) (Department, error) {
	var department Department
	err := utils.FromContext(ctx).Where("id = ?", id).First(&department).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Department{}, errNotFound
	}
	return department, err
}

// ParentExists 判断指定父部门是否存在。
//
// ParentId 为 nil 时调用方无需执行该检查。
func (repository) ParentExists(ctx context.Context, id int64) (bool, error) {
	var total int64
	err := utils.FromContext(ctx).Model(&Department{}).Where("id = ?", id).Count(&total).Error
	return total > 0, err
}

// Create 持久化新的部门。
//
// 创建审计字段由 ORM 回调自动填充。
func (repository) Create(ctx context.Context, department *Department) error {
	return utils.FromContext(ctx).Create(department).Error
}

// Update 更新部门允许修改的字段。
//
// 修改审计字段由 ORM 回调自动填充。
func (repository) Update(ctx context.Context, department *Department) error {
	return utils.FromContext(ctx).Select("Name", "ParentId", "Sort", "Enabled", "LastModifierId", "LastModificationTime").Updates(department).Error
}

// Delete 逻辑删除指定部门集合。
//
// 返回受影响行数和数据库错误。
func (repository) Delete(ctx context.Context, ids []int64) (int64, error) {
	result := utils.FromContext(ctx).Delete(&Department{}, ids)
	return result.RowsAffected, result.Error
}

// EnabledChildren 查询指定父部门下的启用子部门。
//
// 该查询用于递归展开“本部门及以下”结构。
func (repository) EnabledChildren(ctx context.Context, parentId int64) ([]Department, error) {
	children := make([]Department, 0)
	err := utils.FromContext(utils.WithoutDataScope(ctx)).Where("enabled = 1 and parent_id = ?", parentId).Find(&children).Error
	return children, err
}

// FindByIDs 查询指定部门集合。
//
// 该方法用于删除前确认数据存在。
func (repository) FindByIDs(ctx context.Context, ids []int64) ([]Department, error) {
	departments := make([]Department, 0)
	err := utils.FromContext(ctx).Where("enabled = 1 and id in (?)", ids).Find(&departments).Error
	return departments, err
}

// HasUsers 判断部门集合是否被用户引用。
//
// 查询直接统计 sys_user.dept_id，避免 Department 反向依赖 User Entity。
func (repository) HasUsers(ctx context.Context, ids []int64) (bool, error) {
	var total int64
	err := utils.FromContext(ctx).Table("sys_user").Where("dept_id in ?", ids).Count(&total).Error
	return total > 0, err
}

// HasRoles 判断部门集合是否被角色引用。
//
// 查询直接统计关联表，避免 Department 反向依赖 Role Entity。
func (repository) HasRoles(ctx context.Context, ids []int64) (bool, error) {
	var total int64
	err := utils.FromContext(ctx).Table("sys_role_dept").Where("dept_id in ?", ids).Count(&total).Error
	return total > 0, err
}

// ParentIDs 查询部门集合中实际存在子节点的父主键。
//
// Service 用结果批量填充 HasChildren 与 Leaf。
func (repository) ParentIDs(ctx context.Context, ids []int64) ([]int64, error) {
	parentIds := make([]int64, 0)
	err := utils.FromContext(ctx).Model(&Department{}).Where("parent_id in ?", ids).Distinct().Pluck("parent_id", &parentIds).Error
	return parentIds, err
}

// List 查询部门列表及总数。
//
// 根节点查询可分页，指定父节点查询保持既有不分页行为。
func (repository) List(ctx context.Context, options queryOptions) ([]Department, int64, error) {
	departments := make([]Department, 0)
	var total int64
	db := buildQuery(utils.FromContext(ctx).Model(&Department{}), options)
	countDB := db.Count(&total)
	if options.Paginate {
		countDB = countDB.Limit(options.Limit).Offset(options.Offset)
	}
	if err := countDB.Find(&departments).Error; err != nil {
		return nil, 0, err
	}
	return departments, total, nil
}

// FindAll 查询符合条件的全部部门。
//
// 该方法用于导出，不应用分页参数。
func (repository) FindAll(ctx context.Context, options queryOptions) ([]Department, error) {
	departments := make([]Department, 0)
	err := buildQuery(utils.FromContext(ctx).Model(&Department{}), options).Find(&departments).Error
	return departments, err
}

// All 查询全部部门并按 Sort 排序。
//
// 该方法供下拉框和结构计算使用。
func (repository) All(ctx context.Context) ([]Department, error) {
	departments := make([]Department, 0)
	err := utils.FromContext(ctx).Order("sort").Find(&departments).Error
	return departments, err
}

func buildQuery(db *gorm.DB, options queryOptions) *gorm.DB {
	if options.ParentId != nil {
		db = db.Where("parent_id = ?", options.ParentId)
	} else {
		db = db.Where("parent_id IS NULL")
	}
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
