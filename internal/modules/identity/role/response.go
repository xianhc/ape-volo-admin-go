package role

import (
	"github.com/xianhc/ape-volo-admin-go/internal/model"
	"github.com/xianhc/ape-volo-admin-go/internal/modules/identity/menu"
	"github.com/xianhc/ape-volo-admin-go/internal/modules/organization/department"
)

// RoleResponse 表示角色接口的返回模型。
//
// 菜单与部门使用各自独立 Response，数据库实体不会进入 HTTP 输出。
type RoleResponse struct {
	Id                   int64                           `json:"id,string"`
	Name                 string                          `json:"name"`
	Description          string                          `json:"description"`
	DataScopeType        int                             `json:"dataScopeType"`
	AuthCode             string                          `json:"authCode"`
	Menus                []menu.MenuResponse             `json:"menus"`
	Departments          []department.DepartmentResponse `json:"depts"`
	CreatorId            *int64                          `json:"creatorId,omitempty,string"`
	CreationTime         model.Time                      `json:"creationTime"`
	LastModifierId       *int64                          `json:"lastModifierId,omitempty,string"`
	LastModificationTime *model.Time                     `json:"lastModificationTime,omitempty"`
	DeleterId            *int64                          `json:"deleterId,omitempty,string"`
	DeletionTime         *model.Time                     `json:"deletionTime,omitempty"`
}

// fromEntity 将角色实体及关联转换为接口返回模型。
//
// 转换保持现有嵌套菜单和部门字段。
func fromEntity(role Role) RoleResponse {
	return RoleResponse{Id: role.Id, Name: role.Name, Description: role.Description, DataScopeType: role.DataScopeType, AuthCode: role.AuthCode, Menus: menu.FromEntities(role.Menus), Departments: department.FromEntities(role.Departments), CreatorId: role.CreatorId, CreationTime: role.CreationTime, LastModifierId: role.LastModifierId, LastModificationTime: role.LastModificationTime, DeleterId: role.DeleterId, DeletionTime: role.DeletionTime}
}

// FromEntities 将角色实体列表转换为接口返回模型列表。
//
// 返回值始终是非 nil 切片。
func FromEntities(roles []Role) []RoleResponse {
	result := make([]RoleResponse, 0, len(roles))
	for _, role := range roles {
		result = append(result, fromEntity(role))
	}
	return result
}
