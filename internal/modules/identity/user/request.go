package user

import (
	"time"

	"github.com/xianhc/ape-volo-admin-go/internal/payloads/request"
)

// IDReference 表示用户请求中的关联资源主键。
//
// 部门、角色与岗位请求字段复用该轻量模型。
type IDReference struct {
	Id       int64  `json:"id,string"`
	Name     string `json:"name,omitempty"`
	AuthCode string `json:"authCode,omitempty"`
}

// CreateAndUpdateUserRequest 表示创建或修改用户的请求参数。
//
// 角色和岗位至少各选择一项，初始密码由服务端配置生成。
type CreateAndUpdateUserRequest struct {
	Id         int64         `json:"id,string"`
	UserName   string        `json:"userName" validate:"required"`
	NickName   string        `json:"nickName" validate:"required"`
	Email      string        `json:"email" validate:"required,email"`
	Enabled    bool          `json:"enabled"`
	Phone      string        `json:"phone" validate:"required"`
	GenderCode int           `json:"genderCode" validate:"required,min=1,max=2"`
	Dept       IDReference   `json:"dept" validate:"required"`
	Roles      []IDReference `json:"roles" validate:"required,gt=0"`
	Positions  []IDReference `json:"positions" validate:"required,gt=0"`
}

// QueryUserRequest 表示用户列表与导出的筛选参数。
//
// DepartmentIdArray 精确匹配逗号分隔的部门主键；DeptId 保留展开子部门的兼容行为。
type QueryUserRequest struct {
	request.DateRangePagination[time.Time]
	Id                *int64 `json:"id,string" form:"id"`
	UserName          string `json:"userName" form:"userName"`
	NickName          string `json:"nickName" form:"nickName"`
	DepartmentIdArray string `json:"departmentIdArray" form:"departmentIdArray"`
	KeyWords          string `json:"keyWords" form:"keyWords"`
	Enabled           *bool  `json:"enabled" form:"enabled"`
	DeptId            *int64 `json:"deptId" form:"deptId"`
}

// UpdateCenterRequest 表示个人中心基础信息请求。
//
// 用户主键从认证上下文读取。
type UpdateCenterRequest struct {
	NickName   string `json:"nickName" validate:"required"`
	GenderCode int    `json:"genderCode" validate:"required,min=1,max=2"`
	Phone      string `json:"phone"`
}

// UpdatePasswordRequest 表示个人中心修改密码请求。
//
// 三个密码字段均为前端 RSA 加密文本。
type UpdatePasswordRequest struct {
	OldPassword     string `json:"oldPassword" validate:"required"`
	NewPassword     string `json:"newPassword" validate:"required"`
	ConfirmPassword string `json:"confirmPassword" validate:"required"`
}

// UpdateEmailRequest 表示个人中心修改邮箱请求。
//
// 请求同时校验登录密码和一次性邮箱验证码。
type UpdateEmailRequest struct {
	Password string `json:"password" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Code     string `json:"code" validate:"required"`
}

// UpdatePositionsRequest 表示替换用户岗位关联请求。
//
// PositionIdArray 至少包含一个岗位主键。
type UpdatePositionsRequest struct {
	Id              int64           `json:"id,string" validate:"required,gt=0"`
	PositionIdArray request.IDArray `json:"positionIdArray" validate:"required,gt=0"`
}

// UpdateRolesRequest 表示替换用户角色关联请求。
//
// RoleIdArray 至少包含一个角色主键。
type UpdateRolesRequest struct {
	Id          int64           `json:"id,string" validate:"required,gt=0"`
	RoleIdArray request.IDArray `json:"roleIdArray" validate:"required,gt=0"`
}

// UpdatePreferencesRequest 表示修改用户偏好配置请求。
//
// PreferencesConfig 保存前端布局和主题等 JSON 字符串。
type UpdatePreferencesRequest struct {
	PreferencesConfig string `json:"preferencesConfig" validate:"required"`
}
