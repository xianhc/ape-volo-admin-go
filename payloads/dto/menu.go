package dto

import (
	"go-apevolo/model"
	"go-apevolo/model/permission"
	"go-apevolo/utils"
	"go-apevolo/utils/ext"
	"time"
)

type MenuIdDto struct {
	Id int64 `json:"id,string"`
}

type CreateUpdateMenuDto struct {
	model.RootKey
	Title         string `json:"title" validate:"required"`
	Path          string `json:"path" `
	Permission    string `json:"permission" `
	IFrame        bool   `json:"iFrame" `
	Component     string `json:"component" `
	ComponentName string `json:"componentName" `
	ParentId      int64  `json:"parentId,string" `
	Sort          int    `json:"sort" validate:"required,min=1,max=999"`
	Icon          string `json:"icon" `
	Type          int    `json:"type" validate:"required,min=1,max=3"`
	Cache         bool   `json:"cache" `
	Hidden        bool   `json:"hidden" `
	SubCount      int    `json:"subCount"`
	model.BaseModel
}

func (req *CreateUpdateMenuDto) Generate(model *permission.Menu) {
	if req.Id != 0 {
		model.Id = req.Id
	} else {
		model.Id = int64(utils.GenerateID())
		model.CreateBy = req.CreateBy
		model.CreateTime = ext.GetCurrentTime()
	}
	model.Title = req.Title
	model.Path = req.Path
	model.Permission = req.Permission
	model.IFrame = req.IFrame
	model.Component = req.Component
	model.ComponentName = req.ComponentName
	model.ParentId = req.ParentId
	model.Sort = req.Sort
	model.Icon = req.Icon
	model.Type = req.Type
	model.Cache = req.Cache
	model.Hidden = req.Hidden
	model.SubCount = req.SubCount
	if req.UpdateBy != nil {
		model.UpdateBy = req.UpdateBy
	}
	if req.UpdateTime != nil {
		localTime := ext.GetCurrentTime()
		model.UpdateTime = &localTime
	}
}

type MenuQueryCriteria struct {
	Title      string      `json:"title" form:"title"`
	ParentId   *int64      `json:"parentId" form:"parentId"`
	CreateTime []time.Time `json:"createTime" form:"createTime"`
	// request.Pagination
}
