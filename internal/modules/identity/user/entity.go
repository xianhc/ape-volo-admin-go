package user

import (
	"github.com/xianhc/ape-volo-admin-go/internal/model"
	"github.com/xianhc/ape-volo-admin-go/internal/modules/identity/role"
	"github.com/xianhc/ape-volo-admin-go/internal/modules/organization/department"
	"github.com/xianhc/ape-volo-admin-go/internal/modules/organization/position"
)

const (
	// GenderCodeMan 表示男性。
	//
	// 数值保持现有接口与数据库契约。
	GenderCodeMan = 1
	// GenderCodeGirl 表示女性。
	//
	// 数值保持现有接口与数据库契约。
	GenderCodeGirl = 2
)

// User 表示持久化到数据库的系统用户。
//
// Password 只用于认证与写入，Entity 不直接作为 HTTP 响应返回。
type User struct {
	model.RootKey
	UserName          string                `gorm:"column:user_name;size:255;uniqueIndex:unique_sys_user_UserName;comment:用户名;not null;"`
	NickName          string                `gorm:"comment:用户昵称"`
	Email             string                `gorm:"size:255;uniqueIndex:unique_sys_user_Email;comment:用户邮箱"`
	Enabled           bool                  `gorm:"comment:是否激活"`
	Password          string                `gorm:"comment:登录密码;not null;"`
	DeptId            int64                 `gorm:"index:index_sys_user_DeptId;comment:部门ID;not null"`
	Phone             string                `gorm:"index:index_sys_user_Phone;comment:用户手机号"`
	AvatarPath        string                `gorm:"comment:头像路径"`
	PasswordReSetTime *model.Time           `gorm:"comment:密码最后修改时间"`
	GenderCode        int                   `gorm:"column:gender_code;comment:性别"`
	PreferencesConfig string                `gorm:"column:preferences_config;comment:偏好配置"`
	Positions         []position.Position   `gorm:"many2many:sys_user_position;"`
	Roles             []role.Role           `gorm:"many2many:sys_user_role;"`
	Dept              department.Department `gorm:"foreignKey:DeptId;references:Id;comment:部门"`
	model.BaseModel
	model.SoftDeleted
	model.DataScoped
}

// TableName 返回用户实体对应的数据库表名。
//
// GORM 和种子初始化统一使用该名称。
func (User) TableName() string { return "sys_user" }

// UserRole 表示用户与角色的多对多关联。
//
// 关联实体归属 User 模块。
type UserRole struct {
	UserId int64 `gorm:"primaryKey;comment:用户ID;not null;"`
	RoleId int64 `gorm:"primaryKey;index:index_sys_user_role_RoleId;comment:角色ID;not null;"`
}

// TableName 返回用户角色关联表名。
//
// GORM 在用户授权写入时统一使用该名称。
func (UserRole) TableName() string { return "sys_user_role" }
