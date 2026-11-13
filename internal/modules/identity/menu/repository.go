package menu

import (
	"context"
	"errors"
	"time"

	"github.com/xianhc/ape-volo-admin-go/internal/utils"
	"gorm.io/gorm"
)

// errNotFound 表示未找到可见的菜单。
//
// repository 用该错误隔离 GORM 的记录不存在错误。
var errNotFound = errors.New("menu not found")

// repository 封装菜单、用户角色和角色菜单关联的数据库访问。
//
// 菜单按项目约定不接入数据权限，但仍继承软删除过滤。
type repository struct{}

// CountByField 统计指定菜单字段和值的记录数。
//
// Service 用于标题、权限码和组件名称唯一性校验。
func (repository) CountByField(ctx context.Context, field string, value any) (int64, error) {
	var total int64
	err := utils.FromContext(ctx).Model(&Menu{}).Where(field+" = ?", value).Count(&total).Error
	return total, err
}

// FindByID 按主键读取菜单。
//
// 未找到记录时返回 errNotFound。
func (repository) FindByID(ctx context.Context, id int64) (Menu, error) {
	var menu Menu
	err := utils.FromContext(ctx).First(&menu, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Menu{}, errNotFound
	}
	return menu, err
}

// Create 持久化新的菜单。
//
// 创建审计字段由 ORM 回调自动填充。
func (repository) Create(ctx context.Context, menu *Menu) error {
	return utils.FromContext(ctx).Create(menu).Error
}

// Update 按主键更新菜单。
//
// 修改审计字段由 ORM 回调自动填充。
func (repository) Update(ctx context.Context, menu *Menu) error {
	return utils.FromContext(ctx).Select("Title", "Path", "AuthCode", "Component", "ComponentName", "ParentId", "Sort", "Icon", "MenuType", "Enabled", "KeepAlive", "Hidden", "LastModifierId", "LastModificationTime").Updates(menu).Error
}

// Delete 逻辑删除指定菜单集合。
//
// 返回受影响行数和数据库错误。
func (repository) Delete(ctx context.Context, ids []int64) (int64, error) {
	result := utils.FromContext(ctx).Delete(&Menu{}, ids)
	return result.RowsAffected, result.Error
}

// Children 查询指定父菜单的直接子菜单。
//
// 结果保持数据库默认顺序，Service 按具体场景补派生字段。
func (repository) Children(ctx context.Context, parentId int64) ([]Menu, error) {
	menus := make([]Menu, 0)
	err := utils.FromContext(ctx).Where("parent_id = ?", parentId).Find(&menus).Error
	return menus, err
}

// ParentIDs 查询菜单集合中实际存在子节点的父主键。
//
// Service 用结果批量填充 HasChildren 与 Leaf。
func (repository) ParentIDs(ctx context.Context, ids []int64) ([]int64, error) {
	parentIds := make([]int64, 0)
	err := utils.FromContext(ctx).Model(&Menu{}).Where("parent_id in ?", ids).Distinct().Pluck("parent_id", &parentIds).Error
	return parentIds, err
}

// Query 查询符合筛选条件的菜单。
//
// ParentId 为 nil 时查询根菜单，结果按 Sort 升序。
func (repository) Query(ctx context.Context, title string, parentId *int64, createTime []time.Time) ([]Menu, error) {
	menus := make([]Menu, 0)
	db := utils.FromContext(ctx).Model(&Menu{})
	if title != "" {
		db = db.Where("title LIKE ?", "%"+title+"%")
	}
	if parentId == nil {
		db = db.Where("parent_id IS NULL")
	} else {
		db = db.Where("parent_id = ?", parentId)
	}
	if len(createTime) > 0 {
		db = db.Where("creation_time >= ? and creation_time <= ?", createTime[0], createTime[1])
	}
	err := db.Order("sort").Find(&menus).Error
	return menus, err
}

// All 查询全部菜单并按 Sort 排序。
//
// 该方法供角色授权树构建使用。
func (repository) All(ctx context.Context) ([]Menu, error) {
	menus := make([]Menu, 0)
	err := utils.FromContext(ctx).Order("sort").Find(&menus).Error
	return menus, err
}

// MenusForUser 查询用户角色关联的非按钮菜单。
//
// 查询通过关联表获取菜单主键，并按 Sort 排序。
func (repository) MenusForUser(ctx context.Context, userId int64) ([]Menu, error) {
	roleIds := make([]int64, 0)
	if err := utils.FromContext(ctx).Table("sys_user_role").Where("user_id = ?", userId).Pluck("role_id", &roleIds).Error; err != nil {
		return nil, err
	}
	menuIds := make([]int64, 0)
	if err := utils.FromContext(ctx).Model(&RoleMenu{}).Where("role_id in ?", roleIds).Pluck("menu_id", &menuIds).Error; err != nil {
		return nil, err
	}
	menus := make([]Menu, 0)
	err := utils.FromContext(ctx).Where("id in ? and menu_type <> ?", menuIds, MenuTypeButton).Order("sort").Find(&menus).Error
	return menus, err
}
