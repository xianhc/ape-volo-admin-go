package menu

import (
	"github.com/xianhc/ape-volo-admin-go/internal/model"
)

const (
	// MenuTypeCatalog 表示目录菜单。
	//
	// 目录用于组织子菜单，不要求权限码。
	MenuTypeCatalog = 1
	// MenuTypeMenu 表示可访问页面的菜单。
	//
	// 页面菜单需要权限码并可配置组件路径。
	MenuTypeMenu = 2
	// MenuTypeButton 表示页面内按钮权限。
	//
	// 按钮不进入前端路由树，只提供权限码。
	MenuTypeButton = 3
	// MenuTypeInternalLink 表示系统内链菜单。
	//
	// 内链在当前应用路由容器中打开。
	MenuTypeInternalLink = 4
	// MenuTypeExternalLink 表示外部链接菜单。
	//
	// 外链路径必须以 HTTP 或 HTTPS 协议开头。
	MenuTypeExternalLink = 5
)

// Menu 表示持久化到数据库的菜单及运行时树节点。
//
// Children、HasChildren、Leaf 和 Label 为派生字段，不写入数据库。
type Menu struct {
	model.RootKey
	Title         string `gorm:"comment:菜单标题;not null;"`
	Path          string `gorm:"comment:组件路径"`
	AuthCode      string `gorm:"column:auth_code;comment:权限标识符;"`
	Component     string `gorm:"comment:组件"`
	ComponentName string `gorm:"comment:组件名称"`
	ParentId      *int64 `gorm:"index:index_sys_menu_ParentId;comment:父级菜单ID"`
	Sort          int    `gorm:"comment:排序"`
	Icon          string `gorm:"comment:icon图标"`
	MenuType      int    `gorm:"column:menu_type;comment:类型;not null"`
	Enabled       bool   `gorm:"comment:是否启用"`
	KeepAlive     bool   `gorm:"column:keep_alive;comment:是否缓存"`
	Hidden        bool   `gorm:"comment:是否隐藏"`
	Children      []Menu `gorm:"-"`
	HasChildren   bool   `gorm:"-"`
	Leaf          bool   `gorm:"-"`
	Label         string `gorm:"-"`
	model.BaseModel
	model.SoftDeleted
}

// ParentKey 返回用于树分组的父级键。
//
// nil 根节点统一映射为 0，仅用于内存 map 分组。
func (menu Menu) ParentKey() int64 {
	if menu.ParentId == nil {
		return 0
	}
	return *menu.ParentId
}

// IsRoot 判断菜单是否为顶级菜单。
//
// ParentId 为 nil 表示根节点。
func (menu Menu) IsRoot() bool { return menu.ParentId == nil }

// IsExternalLink 判断菜单是否为外链。
//
// 外链路径必须由 Service 校验为 HTTP 或 HTTPS 地址。
func (menu Menu) IsExternalLink() bool { return menu.MenuType == MenuTypeExternalLink }

// CalculateHasChildren 根据已加载子节点更新 HasChildren。
//
// 子节点数量不持久化，树构建后按内存数据实时计算。
func (menu *Menu) CalculateHasChildren() { menu.HasChildren = len(menu.Children) > 0 }

// CalculateLeaf 根据已加载子节点更新 Leaf。
//
// Leaf 与 HasChildren 保持互斥，供前端树控件判断能否展开。
func (menu *Menu) CalculateLeaf() { menu.Leaf = len(menu.Children) == 0 }

// CalculateLabel 使用菜单标题更新树节点标签。
//
// Label 是前端树控件的派生展示字段，不写入数据库。
func (menu *Menu) CalculateLabel() { menu.Label = menu.Title }

// InitChildren 将 Children 初始化为非 nil 空切片。
//
// 该行为保证无子节点菜单序列化为既有的空数组形状。
func (menu *Menu) InitChildren() { menu.Children = []Menu{} }

// TableName 返回菜单实体对应的数据库表名。
//
// GORM 和种子初始化统一使用该名称。
func (Menu) TableName() string { return "sys_menu" }

// RoleMenu 表示角色与菜单的多对多关联。
//
// 关联实体归属 Menu 模块，Role 模块通过表名复用该关系。
type RoleMenu struct {
	RoleId int64 `gorm:"primaryKey;comment:角色ID;not null;"`
	MenuId int64 `gorm:"primaryKey;comment:菜单ID;not null;"`
}

// TableName 返回角色菜单关联表名。
//
// GORM 在授权写入和菜单查询时统一使用该名称。
func (RoleMenu) TableName() string { return "sys_role_menu" }
