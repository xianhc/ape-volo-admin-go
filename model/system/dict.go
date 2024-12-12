package system

import (
	"go-apevolo/model"
)

type Dict struct {
	model.RootKey
	DictType    int    `json:"dictType" gorm:"comment:字典类型"`  // 字典类型 系统OR业务
	Name        string `json:"name" gorm:"comment:字典名称"`      // 字典名称
	Description string `json:"description" gorm:"comment:描述"` // 描述
	model.BaseModel
	model.SoftDeleted
	DictDetail []DictDetail `json:"dictDetail"  gorm:"foreignKey:DictId;references:Id;comment:字典明细"` // 字典明细
}

func (Dict) TableName() string {
	return "sys_dict"
}
