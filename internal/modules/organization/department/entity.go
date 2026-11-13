package department

import (
	"github.com/xianhc/ape-volo-admin-go/internal/model"
)

// Department 表示持久化到数据库的部门及运行时树节点。
//
// Children、HasChildren、Leaf 和 Label 为派生字段，不写入数据库。
type Department struct {
	model.RootKey
	Name        string       `gorm:"comment:部门名称"`
	ParentId    *int64       `gorm:"index:index_sys_department_ParentId;comment:父级部门ID"`
	Sort        int          `gorm:"comment:排序"`
	Enabled     bool         `gorm:"comment:是否启用"`
	Children    []Department `gorm:"-"`
	HasChildren bool         `gorm:"-"`
	Leaf        bool         `gorm:"-"`
	Label       string       `gorm:"-"`
	model.BaseModel
	model.SoftDeleted
	model.DataScoped
}

// ParentKey 返回用于树分组的父级键。
//
// nil 根节点统一映射为 0，仅用于内存 map 分组。
func (department Department) ParentKey() int64 {
	if department.ParentId == nil {
		return 0
	}
	return *department.ParentId
}

// IsRoot 判断部门是否为顶级部门。
//
// ParentId 为 nil 表示根节点。
func (department Department) IsRoot() bool { return department.ParentId == nil }

// TableName 返回部门实体对应的数据库表名。
//
// GORM 和种子初始化统一使用该名称。
func (Department) TableName() string { return "sys_department" }
