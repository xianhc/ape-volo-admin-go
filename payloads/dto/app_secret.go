package dto

import (
	"go-apevolo/model"
	"go-apevolo/model/system"
	"go-apevolo/payloads/request"
	"go-apevolo/utils"
	"go-apevolo/utils/ext"
	"strconv"
	"time"
)

type CreateUpdateAppSecretDto struct {
	model.RootKey
	AppName string `json:"appName" form:"appName" validate:"required"`
	Remark  string `json:"remark" form:"remark"`
	model.BaseModel
}

func (req *CreateUpdateAppSecretDto) Generate(model *system.AppSecret) {
	if req.Id != 0 {
		model.Id = req.Id
	} else {
		model.Id = int64(utils.GenerateID())
		model.CreateBy = req.CreateBy
		model.CreateTime = ext.GetCurrentTime()
		currentTime := time.Now()
		dateString := currentTime.Format("20060102")
		var strId = strconv.FormatInt(model.Id, 10)
		model.AppId = dateString + strId[len(strId)-8:]
		secretKey := utils.BcryptHash(model.AppId + strId)
		model.AppSecretKey = secretKey
	}
	model.AppName = req.AppName
	model.Remark = req.Remark
	if req.UpdateBy != nil {
		model.UpdateBy = req.UpdateBy
	}
	if req.UpdateTime != nil {
		localTime := ext.GetCurrentTime()
		model.UpdateTime = &localTime
	}
}

type AppSecretQueryCriteria struct {
	KeyWords   string   `json:"keyWords" form:"keyWords"`
	CreateTime []string `json:"createTime" form:"createTime"`
	request.Pagination
}
