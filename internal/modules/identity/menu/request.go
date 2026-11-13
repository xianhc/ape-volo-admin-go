package menu

import (
	"time"

	"github.com/xianhc/ape-volo-admin-go/internal/payloads/request"
)

// CreateAndUpdateMenuRequest 表示创建或修改菜单的请求参数。
//
// 菜单类型、权限码和外链路径的组合约束由 Service 校验。
type CreateAndUpdateMenuRequest struct {
	Id            int64  `json:"id,string"`
	Title         string `json:"title" validate:"required"`
	Path          string `json:"path"`
	AuthCode      string `json:"authCode"`
	Component     string `json:"component"`
	ComponentName string `json:"componentName"`
	ParentId      *int64 `json:"parentId,string"`
	Sort          int    `json:"sort" validate:"required,min=1,max=999"`
	Icon          string `json:"icon"`
	MenuType      int    `json:"menuType" validate:"required,min=1,max=5"`
	Enabled       bool   `json:"enabled"`
	KeepAlive     bool   `json:"keepAlive"`
	Hidden        bool   `json:"hidden"`
}

// QueryMenuRequest 表示菜单列表与导出的筛选参数。
//
// ParentId 为 nil 时查询根菜单，非 nil 时查询直接子菜单。
type QueryMenuRequest struct {
	request.CreationTimeRange[time.Time]
	Title    string `json:"title" form:"title"`
	ParentId *int64 `json:"parentId" form:"parentId"`
}
