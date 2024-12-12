package permission

import (
	"go-apevolo/model"
	"time"
)

type User struct {
	model.RootKey
	Username          string     `json:"username" gorm:"index;comment:用户名;not null;"` // 用户名
	NickName          string     `json:"nickName" gorm:"comment:用户昵称"`                // 用户昵称
	Email             string     `json:"email"  gorm:"comment:用户邮箱"`                  // 用户邮箱
	IsAdmin           bool       `json:"isAdmin"  gorm:"comment:是否管理员"`               // 是否管理员
	Enabled           bool       `json:"enabled"  gorm:"comment:是否激活"`                // 是否激活
	Password          string     `json:"password"  gorm:"comment:登录密码;not null;"`     // 登录密码 初始化数据之后可把json设置-不响应给前端
	DeptId            int64      `json:"deptId,string" gorm:"comment:部门ID;not null"`  // 部门ID
	Phone             string     `json:"phone"  gorm:"comment:用户手机号"`                 // 用户手机号
	AvatarPath        string     `json:"avatarPath" gorm:"comment:头像路径"`              // 头像路径
	PasswordReSetTime *time.Time `json:"passwordReSetTime" gorm:"comment:密码最后修改时间"`   // 密码最后修改时间
	Sex               bool       `json:"sex" gorm:"comment:性别"`                       // 性别
	Gender            string     `json:"gender" gorm:"comment:性别"`                    // 性别
	Jobs              []Job      `json:"jobs" gorm:"many2many:sys_user_job;"`
	Roles             []Role     `json:"roles" gorm:"many2many:sys_user_role;"`
	Dept              Department `json:"dept" gorm:"foreignKey:DeptId;references:Id;comment:部门"`
	model.BaseModel
	model.SoftDeleted
}

func (User) TableName() string {
	return "sys_user"
}
