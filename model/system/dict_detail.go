package system

import (
	"go-apevolo/model"
)

type DictDetail struct {
	model.RootKey
	DictId   int64  `json:"dictId,string" form:"dictId" gorm:"column:dict_id;comment:关联标记;not null;"` // 字典ID 初始化数据之后可把json设置-不响应给前端
	Label    string `json:"label" gorm:"comment:字典标签"`                                                // 字典标签
	Value    string `json:"value" gorm:"comment:排序"`                                                  // 字典值
	DictSort int    `json:"dictSort"  gorm:"comment:排序"`                                              // 排序
	model.BaseModel
	model.SoftDeleted
	DictDto DictDto2 `json:"dict" gorm:"-"` // 字典ID
}

type DictDto2 struct {
	Id int64 `json:"id,string"` // 字典ID
}

func (DictDetail) TableName() string {
	return "sys_dict_detail"
}
