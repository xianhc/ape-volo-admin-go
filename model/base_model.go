package model

import (
	"go-apevolo/utils/ext"
	"time"
)

// RootKey 主键
type RootKey struct {
	Id int64 `json:"id,string" gorm:"primaryKey;comment:主键编码;not null;"` //主键编码
}

type BaseModel struct {
	CreateBy   string     `json:"createBy" gorm:"comment:创建者;not null;"`    //创建者
	CreateTime time.Time  `json:"createTime" gorm:"comment:创建时间;not null"`  //创建时间
	UpdateBy   *string    `json:"updateBy,omitempty" gorm:"comment:更新者"`    //更新者
	UpdateTime *time.Time `json:"updateTime,omitempty" gorm:"comment:更新时间"` //更新时间
}

// SetCreateBy 设置创建人id
func (e *BaseModel) SetCreateBy(createBy string) {
	e.CreateTime = ext.GetCurrentTime()
	e.CreateBy = createBy
	e.UpdateBy = nil
	e.UpdateTime = nil
}

// SetUpdateBy 设置修改人id
func (e *BaseModel) SetUpdateBy(updateBy string) {
	localTime := ext.GetCurrentTime()
	e.UpdateTime = &localTime
	e.UpdateBy = &updateBy
}

// SoftDeleted 删除
type SoftDeleted struct {
	IsDeleted bool `json:"-" gorm:"default:0;comment:是否删除(逻辑性);not null;"` //用户是否被冻结 1正常 2冻结
}
