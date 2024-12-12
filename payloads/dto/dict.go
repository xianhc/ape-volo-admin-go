package dto

import (
	"go-apevolo/model"
	"go-apevolo/model/system"
	"go-apevolo/payloads/request"
	"go-apevolo/utils"
	"go-apevolo/utils/ext"
)

type CreateUpdateDictDto struct {
	model.RootKey
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"  validate:"required"`
	model.BaseModel
}

type CreateUpdateDictDetailDto struct {
	model.RootKey
	Label    string  `json:"label" validate:"required"`
	Value    string  `json:"value"  validate:"required"`
	DictSort int     `json:"dictSort"  validate:"required,min=1,max=999"`
	Dict     DictDto `json:"dict"`
	model.BaseModel
}

type DictDto struct {
	Id int64 `json:"id,string"` // 字典ID
}

func (req *CreateUpdateDictDto) Generate(model *system.Dict) {
	if req.Id != 0 {
		model.Id = req.Id
	} else {
		model.Id = int64(utils.GenerateID())
		model.CreateBy = req.CreateBy
		model.CreateTime = ext.GetCurrentTime()
	}
	model.Name = req.Name
	model.Description = req.Description
	if req.UpdateBy != nil {
		model.UpdateBy = req.UpdateBy
	}
	if req.UpdateTime != nil {
		localTime := ext.GetCurrentTime()
		model.UpdateTime = &localTime
	}
}

func (req *CreateUpdateDictDetailDto) Generate(model *system.DictDetail) {
	if req.Id != 0 {
		model.Id = req.Id
	} else {
		model.Id = int64(utils.GenerateID())
		model.CreateBy = req.CreateBy
		model.CreateTime = ext.GetCurrentTime()
	}
	model.Label = req.Label
	model.Value = req.Value
	model.DictSort = req.DictSort
	model.DictId = req.Dict.Id
	if req.UpdateBy != nil {
		model.UpdateBy = req.UpdateBy
	}
	if req.UpdateTime != nil {
		localTime := ext.GetCurrentTime()
		model.UpdateTime = &localTime
	}
}

type DictQueryCriteria struct {
	KeyWords string `json:"keyWords" form:"keyWords"`
	request.Pagination
}

type DictDetailQueryCriteria struct {
	Label    string `json:"label" form:"label"`
	DictName string `json:"dictName" form:"dictName"`
	request.Pagination
}
