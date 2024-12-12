package dto

import (
	"go-apevolo/model"
	"go-apevolo/model/permission"
	"go-apevolo/payloads/request"
	"go-apevolo/utils"
	"go-apevolo/utils/ext"
)

type ApisIdDto struct {
	Id int64 `json:"id,string"`
}

type CreateUpdateApisDto struct {
	model.RootKey
	Group       string `json:"group" validate:"required"`
	Url         string `json:"url" validate:"required"`
	Description string `json:"description" validate:"required"`
	Method      string `json:"method" validate:"required"`
	model.BaseModel
}

func (req *CreateUpdateApisDto) Generate(model *permission.Apis) {
	if req.Id != 0 {
		model.Id = req.Id
	} else {
		model.Id = int64(utils.GenerateID())
		model.CreateBy = req.CreateBy
		model.CreateTime = ext.GetCurrentTime()
	}
	model.Group = req.Group
	model.Url = req.Url
	model.Description = req.Description
	model.Method = req.Method
	if req.UpdateBy != nil {
		model.UpdateBy = req.UpdateBy
	}
	if req.UpdateTime != nil {
		localTime := ext.GetCurrentTime()
		model.UpdateTime = &localTime
	}
}

type ApisQueryCriteria struct {
	Group       string `json:"group" form:"group"`
	Description string `json:"description" form:"description"`
	Method      string `json:"method" form:"method"`
	request.Pagination
}
