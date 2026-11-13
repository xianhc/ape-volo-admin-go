package department

import (
	"github.com/xianhc/ape-volo-admin-go/internal/model"
)

// DepartmentResponse 表示部门接口的返回模型。
//
// Children 递归使用相同 Response，数据库实体不会进入 HTTP 输出。
type DepartmentResponse struct {
	Id                   int64                `json:"id,string"`
	Name                 string               `json:"name"`
	ParentId             *int64               `json:"parentId,omitempty,string"`
	Sort                 int                  `json:"sort"`
	Enabled              bool                 `json:"enabled"`
	Children             []DepartmentResponse `json:"children,omitempty"`
	HasChildren          bool                 `json:"hasChildren"`
	Leaf                 bool                 `json:"leaf"`
	Label                string               `json:"label"`
	CreatorId            *int64               `json:"creatorId,omitempty,string"`
	CreationTime         model.Time           `json:"creationTime"`
	LastModifierId       *int64               `json:"lastModifierId,omitempty,string"`
	LastModificationTime *model.Time          `json:"lastModificationTime,omitempty"`
	DeleterId            *int64               `json:"deleterId,omitempty,string"`
	DeletionTime         *model.Time          `json:"deletionTime,omitempty"`
}

// FromEntity 将部门实体及子树转换为接口返回模型。
//
// 转换保持树控件所需的派生字段。
func FromEntity(department Department) DepartmentResponse {
	children := make([]DepartmentResponse, 0, len(department.Children))
	for _, child := range department.Children {
		children = append(children, FromEntity(child))
	}
	return DepartmentResponse{Id: department.Id, Name: department.Name, ParentId: department.ParentId, Sort: department.Sort, Enabled: department.Enabled, Children: children, HasChildren: department.HasChildren, Leaf: department.Leaf, Label: department.Label, CreatorId: department.CreatorId, CreationTime: department.CreationTime, LastModifierId: department.LastModifierId, LastModificationTime: department.LastModificationTime, DeleterId: department.DeleterId, DeletionTime: department.DeletionTime}
}

// FromEntities 将部门实体列表转换为接口返回模型列表。
//
// 返回值始终是非 nil 切片。
func FromEntities(departments []Department) []DepartmentResponse {
	result := make([]DepartmentResponse, 0, len(departments))
	for _, department := range departments {
		result = append(result, FromEntity(department))
	}
	return result
}
