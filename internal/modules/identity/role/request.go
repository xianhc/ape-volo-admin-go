package role

import (
	"time"

	"github.com/xianhc/ape-volo-admin-go/internal/payloads/request"
)

// IDReference 表示角色请求中的关联资源主键。
//
// Departments 和 Menus 共用该轻量请求类型。
type IDReference struct {
	Id int64 `json:"id,string"`
}

// CreateAndUpdateRoleRequest 表示创建或修改角色的请求参数。
//
// DataScopeType 保留 0 到 5 的现有校验范围。
type CreateAndUpdateRoleRequest struct {
	Id            int64         `json:"id,string"`
	Name          string        `json:"name" validate:"required"`
	Description   string        `json:"description" validate:"required"`
	DataScopeType int           `json:"dataScopeType" validate:"omitempty,min=0,max=5"`
	AuthCode      string        `json:"authCode" validate:"required"`
	Departments   []IDReference `json:"depts"`
	Menus         []IDReference `json:"menus"`
}

// QueryRoleRequest 表示角色列表与导出的筛选参数。
//
// 支持角色名称和创建时间范围筛选。
type QueryRoleRequest struct {
	request.DateRangePagination[time.Time]
	RoleName string `json:"roleName" form:"roleName"`
}
