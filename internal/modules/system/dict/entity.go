package dict

import (
	"github.com/xianhc/ape-volo-admin-go/internal/model"
)

// Dict 表示持久化到数据库的字典主表。
//
// 字典详情通过 DictDetails 关联加载，实体不直接作为 HTTP 响应返回。
type Dict struct {
	model.RootKey
	DictType    int    `gorm:"comment:字典类型"`
	Name        string `gorm:"size:255;uniqueIndex:unique_sys_dict_Name;comment:字典名称"`
	Description string `gorm:"comment:描述"`
	model.BaseModel
	model.SoftDeleted
	DictDetails []DictDetail `gorm:"foreignKey:DictId;references:Id;comment:字典明细"`
}

// TableName 返回字典主表对应的数据库表名。
//
// GORM 和种子初始化统一使用该名称。
func (Dict) TableName() string { return "sys_dict" }

// DictDetail 表示持久化到数据库的字典明细。
//
// DictId 保存所属字典主键，接口中的嵌套字典引用由 Response 组装。
type DictDetail struct {
	model.RootKey
	DictId   int64  `gorm:"column:dict_id;index:index_sys_dict_detail_DictId_DictSort,priority:1;comment:关联标记;not null;"`
	Label    string `gorm:"comment:字典标签"`
	Value    string `gorm:"comment:排序"`
	DictSort int    `gorm:"index:index_sys_dict_detail_DictId_DictSort,priority:2;comment:排序"`
	model.BaseModel
	model.SoftDeleted
}

// TableName 返回字典明细对应的数据库表名。
//
// GORM 和种子初始化统一使用该名称。
func (DictDetail) TableName() string { return "sys_dict_detail" }
