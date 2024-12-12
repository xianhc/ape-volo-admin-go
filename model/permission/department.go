package permission

import (
	"go-apevolo/model"
)

type Department struct {
	model.RootKey
	Name        string       `json:"name" gorm:"comment:岗位名称"`                        // 岗位名称
	ParentId    int64        `json:"parentId,omitempty,string" gorm:"comment:父级部门ID"` // 父级部门ID
	Sort        int          `json:"sort" gorm:"comment:排序"`                          // 排序
	Enabled     bool         `json:"enabled" gorm:"comment:是否启用"`                     // 是否启用
	SubCount    int          `json:"subCount" gorm:"comment:子节点个数"`                   // 子节点个数
	Children    []Department `json:"children,omitempty"  gorm:"-"`                    // 子部门集合
	HasChildren bool         `json:"hasChildren"  gorm:"-"`                           // 是否有子级
	Leaf        bool         `json:"leaf"  gorm:"-"`                                  // 是否能展开
	Label       string       `json:"label"  gorm:"-"`                                 // 标题
	Roles       []Role       `json:"roles"  gorm:"many2many:sys_role_dept;"`          // 角色列表
	Users       []User       `json:"users" gorm:"foreignKey:DeptId;"`                 //用户列表
	model.BaseModel
	model.SoftDeleted
}

func (d *Department) CalculateHasChildren() {
	d.HasChildren = d.SubCount > 0
}

func (d *Department) CalculateLeaf() {
	d.Leaf = d.SubCount == 0
}

func (d *Department) CalculateLabel() {
	d.Label = d.Name
}

func (Department) TableName() string {
	return "sys_department"
}
