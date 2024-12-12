package dto

import (
	"go-apevolo/model"
	"go-apevolo/model/permission"
	"go-apevolo/payloads/request"
	"go-apevolo/utils"
	"go-apevolo/utils/ext"
	"time"
)

type CreateUpdateUserReq struct {
	model.RootKey
	Username string    `json:"username" validate:"required"`
	NickName string    `json:"nickName" validate:"required"`
	Email    string    `json:"email" validate:"required,email"`
	Enabled  bool      `json:"enabled"`
	Phone    string    `json:"phone" validate:"required"`
	Gender   string    `json:"gender" validate:"required"`
	Dept     DeptDto   `json:"dept" validate:"required"`
	Roles    []RoleDto `json:"roles" validate:"required,gt=0"`
	Jobs     []JobDto  `json:"jobs" validate:"required,gt=0"`
	model.BaseModel
}

type RoleDto struct {
	Id         int64  `json:"id,string"`
	Name       string `json:"name"`
	Permission string `json:"permission"`
}

type JobDto struct {
	Id   int64  `json:"id,string"`
	Name string `json:"name"`
}

func (req *CreateUpdateUserReq) Generate(model *permission.User) {
	if req.Id != 0 {
		model.Id = req.Id
	} else {
		model.Id = int64(utils.GenerateID())
		model.CreateBy = req.CreateBy
		model.CreateTime = ext.GetCurrentTime()
	}
	model.Username = req.Username
	model.NickName = req.NickName
	model.Email = req.Email
	model.Enabled = req.Enabled
	model.Phone = req.Phone
	model.Email = req.Email
	model.Gender = req.Gender
	model.DeptId = req.Dept.Id
	if req.UpdateBy != nil {
		model.UpdateBy = req.UpdateBy
	}
	if req.UpdateTime != nil {
		localTime := ext.GetCurrentTime()
		model.UpdateTime = &localTime
	}
}

type UserQueryCriteria struct {
	KeyWords   string      `json:"keyWords" form:"keyWords"`
	Enabled    *bool       `json:"enabled" form:"enabled"`
	DeptId     *int64      `json:"deptId" form:"deptId"`
	CreateTime []time.Time `json:"createTime" form:"createTime"`
	request.Pagination
}
