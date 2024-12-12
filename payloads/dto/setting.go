package dto

import (
	"go-apevolo/model"
	"go-apevolo/model/system"
	"go-apevolo/payloads/request"
	"go-apevolo/utils"
	"go-apevolo/utils/ext"
	"time"
)

type CreateUpdateSettingDto struct {
	model.RootKey
	Name        string `json:"name" form:"name" validate:"required"`
	Value       string `json:"value" form:"value" validate:"required"`
	Enabled     bool   `json:"enabled" form:"enabled"`
	Description string `json:"description" form:"description"`
	model.BaseModel
}

func (req *CreateUpdateSettingDto) Generate(model *system.Setting) {
	if req.Id != 0 {
		model.Id = req.Id
	} else {
		model.Id = int64(utils.GenerateID())
		model.CreateBy = req.CreateBy
		model.CreateTime = ext.GetCurrentTime()
	}
	model.Name = req.Name
	model.Value = req.Value
	model.Enabled = req.Enabled
	model.Description = req.Description
	if req.UpdateBy != nil {
		model.UpdateBy = req.UpdateBy
	}
	if req.UpdateTime != nil {
		localTime := ext.GetCurrentTime()
		model.UpdateTime = &localTime
	}
}

type SettingQueryCriteria struct {
	KeyWords   string      `json:"keyWords" form:"keyWords"`
	Enabled    *bool       `json:"enabled" form:"enabled"`
	CreateTime []time.Time `json:"createTime" form:"createTime"`
	request.Pagination
}
