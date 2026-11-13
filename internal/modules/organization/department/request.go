package department

import (
	"github.com/xianhc/ape-volo-admin-go/internal/payloads/request"
)

// CreateAndUpdateDepartmentRequest 表示创建或修改部门的请求参数。
//
// ParentId 为 nil 时创建顶级部门。
type CreateAndUpdateDepartmentRequest struct {
	Id       int64  `json:"id,string"`
	Name     string `json:"name" validate:"required"`
	ParentId *int64 `json:"parentId,string"`
	Sort     int    `json:"sort" validate:"required,min=1,max=999"`
	Enabled  bool   `json:"enabled"`
}

// QueryDepartmentRequest 表示部门列表与导出的筛选参数。
//
// ParentId 为 nil 时查询根部门，非 nil 时查询指定父节点的直接子部门。
type QueryDepartmentRequest struct {
	request.OptionalPagination
	request.CreationTimeRange[string]
	Name     string `json:"name" form:"name"`
	DeptName string `json:"deptName" form:"deptName"`
	Enabled  *bool  `json:"enabled" form:"enabled"`
	ParentId *int64 `json:"parentId" form:"parentId"`
}
