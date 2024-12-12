package dto

import (
	"go-apevolo/model"
	"go-apevolo/model/permission"
	"go-apevolo/payloads/request"
	"go-apevolo/utils"
	"go-apevolo/utils/ext"
	"time"
)

type CreateUpdateRoleDto struct {
	model.RootKey
	Name        string      `json:"name" validate:"required"`
	Level       int         `json:"level" validate:"required,min=1,max=999"`
	Description string      `json:"description" validate:"required"`
	DataScope   string      `json:"dataScope"`
	Permission  string      `json:"permission" validate:"required"`
	Departments []DeptDto   `json:"depts"`
	Menus       []MenuIdDto `json:"menus"`
	Apis        []ApisIdDto `json:"apis"`
	model.BaseModel
}

func (req *CreateUpdateRoleDto) Generate(model *permission.Role) {
	if req.Id != 0 {
		model.Id = req.Id
	} else {
		model.Id = int64(utils.GenerateID())
		model.CreateBy = req.CreateBy
		model.CreateTime = ext.GetCurrentTime()
	}
	model.Name = req.Name
	model.Level = req.Level
	model.Description = req.Description
	model.DataScope = req.DataScope
	model.Permission = req.Permission
	if req.UpdateBy != nil {
		model.UpdateBy = req.UpdateBy
	}
	if req.UpdateTime != nil {
		localTime := ext.GetCurrentTime()
		model.UpdateTime = &localTime
	}
}

type RoleQueryCriteria struct {
	RoleName   string      `json:"roleName" form:"roleName"`
	CreateTime []time.Time `json:"createTime" form:"createTime"`
	request.Pagination
}
