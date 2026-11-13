package user

import (
	"encoding/json"

	"github.com/xianhc/ape-volo-admin-go/internal/model"
	"github.com/xianhc/ape-volo-admin-go/internal/modules/identity/role"
	"github.com/xianhc/ape-volo-admin-go/internal/modules/organization/department"
	"github.com/xianhc/ape-volo-admin-go/internal/modules/organization/position"
)

// UserResponse 表示用户接口的脱敏返回模型。
//
// Password 始终为空，并额外输出 headerImg 兼容前端头像字段。
type UserResponse struct {
	Id                   int64                         `json:"id,string"`
	UserName             string                        `json:"userName"`
	NickName             string                        `json:"nickName"`
	Email                string                        `json:"email"`
	Enabled              bool                          `json:"enabled"`
	Password             string                        `json:"password,omitempty"`
	DeptId               int64                         `json:"deptId,string"`
	Phone                string                        `json:"phone"`
	AvatarPath           string                        `json:"avatarPath"`
	PasswordReSetTime    *model.Time                   `json:"passwordReSetTime"`
	GenderCode           int                           `json:"genderCode"`
	PreferencesConfig    string                        `json:"preferencesConfig"`
	Positions            []position.PositionResponse   `json:"positions"`
	Roles                []role.RoleResponse           `json:"roles"`
	Dept                 department.DepartmentResponse `json:"dept"`
	CreatorId            *int64                        `json:"creatorId,omitempty,string"`
	CreationTime         model.Time                    `json:"creationTime"`
	LastModifierId       *int64                        `json:"lastModifierId,omitempty,string"`
	LastModificationTime *model.Time                   `json:"lastModificationTime,omitempty"`
	DeleterId            *int64                        `json:"deleterId,omitempty,string"`
	DeletionTime         *model.Time                   `json:"deletionTime,omitempty"`
}

// MarshalJSON 添加 headerImg 头像字段别名。
//
// 其余字段保持 UserResponse 的标准 JSON 标签。
func (user UserResponse) MarshalJSON() ([]byte, error) {
	type alias UserResponse
	data, err := json.Marshal(alias(user))
	if err != nil {
		return nil, err
	}
	var fields map[string]json.RawMessage
	if err = json.Unmarshal(data, &fields); err != nil {
		return nil, err
	}
	fields["headerImg"], err = json.Marshal(user.AvatarPath)
	if err != nil {
		return nil, err
	}
	return json.Marshal(fields)
}

// fromEntity 将用户实体和关联转换为脱敏接口返回模型。
//
// 转换有意不复制 Password 字段。
func fromEntity(user User) UserResponse {
	return UserResponse{Id: user.Id, UserName: user.UserName, NickName: user.NickName, Email: user.Email, Enabled: user.Enabled, DeptId: user.DeptId, Phone: user.Phone, AvatarPath: user.AvatarPath, PasswordReSetTime: user.PasswordReSetTime, GenderCode: user.GenderCode, PreferencesConfig: user.PreferencesConfig, Positions: position.FromEntities(user.Positions), Roles: role.FromEntities(user.Roles), Dept: department.FromEntity(user.Dept), CreatorId: user.CreatorId, CreationTime: user.CreationTime, LastModifierId: user.LastModifierId, LastModificationTime: user.LastModificationTime, DeleterId: user.DeleterId, DeletionTime: user.DeletionTime}
}

// fromEntities 将用户实体列表转换为脱敏返回模型列表。
//
// 返回值始终是非 nil 切片。
func fromEntities(users []User) []UserResponse {
	result := make([]UserResponse, 0, len(users))
	for _, user := range users {
		result = append(result, fromEntity(user))
	}
	return result
}
