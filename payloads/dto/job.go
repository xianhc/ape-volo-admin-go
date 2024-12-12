package dto

import (
	"go-apevolo/model"
	"go-apevolo/model/permission"
	"go-apevolo/payloads/request"
	"go-apevolo/utils"
	"go-apevolo/utils/ext"
	"time"
)

type CreateUpdateJobDto struct {
	model.RootKey
	Name    string `json:"name" validate:"required"`
	Sort    int    `json:"sort" validate:"required,min=1,max=999"`
	Enabled bool   `json:"enabled"`
	model.BaseModel
}

func (req *CreateUpdateJobDto) Generate(model *permission.Job) {
	if req.Id != 0 {
		model.Id = req.Id
	} else {
		model.Id = int64(utils.GenerateID())
		model.CreateBy = req.CreateBy
		model.CreateTime = ext.GetCurrentTime()
	}
	model.Name = req.Name
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

type JobQueryCriteria struct {
	JobName    string      `json:"jobName" form:"jobName"`
	Enabled    *bool       `json:"enabled" form:"enabled"`
	CreateTime []time.Time `json:"createTime" form:"createTime"`
	request.Pagination
}
