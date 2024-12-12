package permission

import (
	"go-apevolo/model"
)

type Menu struct {
	model.RootKey
	Title         string `json:"title" gorm:"comment:菜单标题;not null;"`            // 菜单标题
	Path          string `json:"path" gorm:"comment:组件路径"`                       // 组件路径
	Permission    string `json:"permission" gorm:"comment:权限标识符;"`               // 权限标识符
	IFrame        bool   `json:"iFrame" gorm:"comment:是否iframe"`                 // 是否iframe
	Component     string `json:"component,omitempty" gorm:"comment:组件"`          // 组件
	ComponentName string `json:"componentName,omitempty" gorm:"comment:组件名称"`    // 组件名称
	ParentId      int64  `json:"parentId,string" gorm:"comment:父级菜单ID;not null"` // 父级菜单ID
	Sort          int    `json:"sort" gorm:"comment:排序"`                         // 排序
	Icon          string `json:"icon" gorm:"comment:icon图标"`                     // icon图标
	Type          int    `json:"type" gorm:"comment:类型 1.目录 2.菜单 3.按钮;not null"` // 类型 1.目录 2.菜单 3.按钮
	Cache         bool   `json:"cache" gorm:"comment:是否缓存"`                      // 是否缓存
	Hidden        bool   `json:"hidden" gorm:"comment:是否隐藏"`                     // 是否隐藏
	SubCount      int    `json:"subCount" gorm:"comment:子节点个数"`                  // 子节点个数
	Children      []Menu `json:"children,omitempty"  gorm:"-"`                   // 子菜单集合
	HasChildren   bool   `json:"hasChildren"  gorm:"-"`                          // 是否有子级
	Leaf          bool   `json:"leaf"  gorm:"-"`                                 // 是否能展开
	Label         string `json:"label"  gorm:"-"`                                // 标题
	model.BaseModel
	model.SoftDeleted
}

func (m *Menu) CalculateHasChildren() {
	m.HasChildren = m.SubCount > 0
}

func (m *Menu) CalculateLeaf() {
	m.Leaf = m.SubCount == 0
}

func (m *Menu) InitChildren() {
	m.Children = []Menu{}
}

func (m *Menu) CalculateLabel() {
	m.Label = m.Title
}

func (Menu) TableName() string {
	return "sys_menu"
}
