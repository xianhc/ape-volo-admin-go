package dto

import (
	"go-apevolo/model"
	"go-apevolo/model/permission"
	"go-apevolo/payloads/request"
	"go-apevolo/utils"
	"go-apevolo/utils/ext"
)

type CreateUpdateDeptDto struct {
	model.RootKey
	Name     string `json:"name"  validate:"required"`
	ParentId int64  `json:"parentId,string" `
	Sort     int    `json:"sort"  validate:"required,min=1,max=999"`
	Enabled  bool   `json:"enabled"`
	model.BaseModel
}

func (req *CreateUpdateDeptDto) Generate(model *permission.Department) {
	if req.Id != 0 {
		model.Id = req.Id
	} else {
		model.Id = int64(utils.GenerateID())
		model.CreateBy = req.CreateBy
		model.CreateTime = ext.GetCurrentTime()
	}
	model.Name = req.Name
	model.ParentId = req.ParentId
	model.Sort = req.Sort
	model.Enabled = req.Enabled
	if req.UpdateBy != nil {
		model.UpdateBy = req.UpdateBy
	}
	if req.UpdateTime != nil {
		localTime := ext.GetCurrentTime()
		model.UpdateTime = &localTime
	}
}

type DeptQueryCriteria struct {
	DeptName   string   `json:"deptName" form:"deptName"`     //部门名称
	Enabled    *bool    `json:"enabled" form:"enabled"`       //是否启用
	ParentId   *int64   `json:"parentId" form:"parentId"`     //父ID
	CreateTime []string `json:"createTime" form:"createTime"` //创建时间
	request.Pagination
}

type DeptDto struct {
	Id int64 `json:"id,string"`
}
