package role

import (
	"context"
	"errors"
	"time"

	"github.com/xianhc/ape-volo-admin-go/internal/modules/identity/menu"
	"github.com/xianhc/ape-volo-admin-go/internal/utils"
	"gorm.io/gorm"
)

// errNotFound 表示未找到可见的角色。
//
// repository 用该错误隔离 GORM 的记录不存在错误。
var errNotFound = errors.New("role not found")

// queryOptions 描述角色列表查询的数据库参数。
//
// Service 负责从 HTTP 请求转换该结构。
type queryOptions struct {
	Name          string
	CreateTime    []time.Time
	Sorting       string
	Limit, Offset int
}

// repository 封装角色、角色部门、角色菜单和用户角色关联访问。
//
// 零值即可使用，写操作复用请求事务。
type repository struct{}

// CountByField 统计角色字段值的记录数。
//
// Service 用于名称和角色码唯一性校验。
func (repository) CountByField(ctx context.Context, field string, value any) (int64, error) {
	var total int64
	err := utils.FromContext(ctx).Model(&Role{}).Where(field+" = ?", value).Count(&total).Error
	return total, err
}

// FindByID 按主键读取角色和可选关联。
//
// details 为 true 时预加载菜单与部门。
func (repository) FindByID(ctx context.Context, id int64, details bool) (Role, error) {
	var role Role
	db := utils.FromContext(ctx)
	if details {
		db = db.Preload("Menus").Preload("Departments")
	}
	err := db.First(&role, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Role{}, errNotFound
	}
	return role, err
}

// Create 持久化角色及部门关联。
//
// 调用方事务保证角色和关联表原子写入。
func (repository) Create(ctx context.Context, role *Role, deptIds []int64) error {
	db := utils.FromContext(ctx)
	if err := db.Create(role).Error; err != nil {
		return err
	}
	return replaceDepartments(db, role.Id, deptIds)
}

// Update 更新角色并替换部门关联。
//
// 调用方事务保证角色和关联表原子写入。
func (repository) Update(ctx context.Context, role *Role, deptIds []int64) error {
	db := utils.FromContext(ctx)
	if err := db.Select("Name", "Description", "DataScopeType", "AuthCode", "LastModifierId", "LastModificationTime").Updates(role).Error; err != nil {
		return err
	}
	return replaceDepartments(db, role.Id, deptIds)
}

// Delete 逻辑删除指定角色。
//
// 返回受影响行数和数据库错误。
func (repository) Delete(ctx context.Context, ids []string) (int64, error) {
	result := utils.FromContext(ctx).Delete(&Role{}, ids)
	return result.RowsAffected, result.Error
}

// HasUsers 判断角色是否被用户引用。
//
// 直接统计关联表，避免 Role 反向依赖 User Entity。
func (repository) HasUsers(ctx context.Context, ids []string) (bool, error) {
	var total int64
	err := utils.FromContext(ctx).Table("sys_user_role").Where("role_id in ?", ids).Count(&total).Error
	return total > 0, err
}

// List 查询一页角色及菜单关联。
//
// 返回分页前总数。
func (repository) List(ctx context.Context, o queryOptions) ([]Role, int64, error) {
	roles := make([]Role, 0)
	var total int64
	db := buildQuery(utils.FromContext(ctx).Model(&Role{}), o)
	if err := db.Count(&total).Limit(o.Limit).Offset(o.Offset).Preload("Menus").Find(&roles).Error; err != nil {
		return nil, 0, err
	}
	return roles, total, nil
}

// All 查询全部角色及完整关联。
//
// 该方法供 get-all 与详情复用。
func (repository) All(ctx context.Context) ([]Role, error) {
	roles := make([]Role, 0)
	err := utils.FromContext(ctx).Preload("Menus").Preload("Departments").Find(&roles).Error
	return roles, err
}

// FindForExport 查询符合条件的角色及部门关联。
//
// 该方法用于导出，不预加载菜单。
func (repository) FindForExport(ctx context.Context, o queryOptions) ([]Role, error) {
	roles := make([]Role, 0)
	err := buildQuery(utils.FromContext(ctx).Model(&Role{}), o).Preload("Departments").Find(&roles).Error
	return roles, err
}

// ReplaceMenus 替换角色菜单关联。
//
// menus 为空时保持旧关联，延续现有接口行为。
func (repository) ReplaceMenus(ctx context.Context, roleId int64, menuIds []int64) error {
	if len(menuIds) == 0 {
		return nil
	}
	db := utils.FromContext(ctx)
	if err := db.Where("role_id = ?", roleId).Delete(&menu.RoleMenu{}).Error; err != nil {
		return err
	}
	rows := make([]menu.RoleMenu, 0, len(menuIds))
	for _, id := range menuIds {
		rows = append(rows, menu.RoleMenu{RoleId: roleId, MenuId: id})
	}
	return db.Create(&rows).Error
}

// PermissionCodes 查询用户角色对应的菜单权限码。
//
// 结果只包含非空权限码并去重。
func (repository) PermissionCodes(ctx context.Context, userId int64, roleIds []int64) ([]string, error) {
	if len(roleIds) == 0 {
		if err := utils.FromContext(ctx).Table("sys_user_role").Where("user_id = ?", userId).Pluck("role_id", &roleIds).Error; err != nil {
			return nil, err
		}
	}
	menuIds := make([]int64, 0)
	if err := utils.FromContext(ctx).Model(&menu.RoleMenu{}).Where("role_id in ?", roleIds).Pluck("menu_id", &menuIds).Error; err != nil {
		return nil, err
	}
	codes := make([]string, 0)
	err := utils.FromContext(ctx).Model(&menu.Menu{}).Where("id in ? and auth_code <> ''", menuIds).Distinct().Pluck("auth_code", &codes).Error
	return codes, err
}

// RoleCodes 查询启用且未删除用户关联的角色码。
//
// 使用子查询读取关联主键，角色自身的软删除由模型回调过滤。
func (repository) RoleCodes(ctx context.Context, userId int64) ([]string, error) {
	db := utils.FromContext(ctx)
	roleIDs := db.Table("sys_user_role AS ur").Select("ur.role_id").
		Joins("JOIN sys_user AS u ON u.id = ur.user_id").
		Where("ur.user_id = ? AND u.enabled = ? AND u.is_deleted = ?", userId, true, false)
	codes := make([]string, 0)
	err := db.Model(&Role{}).Where("id IN (?)", roleIDs).Distinct().Order("auth_code").Pluck("auth_code", &codes).Error
	return codes, err
}

func replaceDepartments(db *gorm.DB, roleId int64, ids []int64) error {
	if err := db.Where("role_id = ?", roleId).Delete(&RoleDepartment{}).Error; err != nil {
		return err
	}
	if len(ids) == 0 {
		return nil
	}
	rows := make([]RoleDepartment, 0, len(ids))
	for _, id := range ids {
		rows = append(rows, RoleDepartment{RoleId: roleId, DeptId: id})
	}
	return db.Create(&rows).Error
}
func buildQuery(db *gorm.DB, o queryOptions) *gorm.DB {
	if o.Name != "" {
		db = db.Where("name LIKE ?", "%"+o.Name+"%")
	}
	if len(o.CreateTime) != 0 && len(o.CreateTime) != 2 {
		_ = db.AddError(errors.New("创建时间必须包含起止两个值"))
		return db
	}
	if len(o.CreateTime) == 2 {
		db = db.Where("creation_time >= ? and creation_time <= ?", o.CreateTime[0], o.CreateTime[1])
	}
	if o.Sorting != "" {
		db = utils.ApplySorting(db, o.Sorting)
	}
	return db
}
