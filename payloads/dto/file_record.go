package dto

import (
	"go-apevolo/model"
	"go-apevolo/model/system"
	"go-apevolo/payloads/request"
	"go-apevolo/utils"
	"go-apevolo/utils/ext"
)

type CreateUpdateFileRecordDto struct {
	model.RootKey
	Description string `json:"description"  validate:"required"`
	model.BaseModel
}

func (req *CreateUpdateFileRecordDto) Generate(model *system.FileRecord) {
	if req.Id != 0 {
		model.Id = req.Id
	} else {
		model.Id = int64(utils.GenerateID())
		model.CreateBy = req.CreateBy
		model.CreateTime = ext.GetCurrentTime()
	}
	model.Description = req.Description
	if req.UpdateBy != nil {
		model.UpdateBy = req.UpdateBy
	}
	if req.UpdateTime != nil {
		localTime := ext.GetCurrentTime()
		model.UpdateTime = &localTime
	}
}

type FileRecordQueryCriteria struct {
	KeyWords   string   `json:"keyWords" form:"keyWords"`
	CreateTime []string `json:"createTime" form:"createTime"`
	request.Pagination
}
