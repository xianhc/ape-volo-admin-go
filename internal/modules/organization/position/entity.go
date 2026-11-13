package position

import (
	"github.com/xianhc/ape-volo-admin-go/internal/model"
)

// Position 表示持久化到数据库的岗位。
//
// 实体只描述岗位字段和审计能力，不直接作为 HTTP 响应返回。
type Position struct {
	model.RootKey
	Name    string `gorm:"size:255;uniqueIndex:unique_sys_position_Name;comment:岗位名称"`
	Sort    int    `gorm:"comment:排序"`
	Enabled bool   `gorm:"comment:是否启用"`
	model.BaseModel
	model.SoftDeleted
	model.DataScoped
}

// TableName 返回岗位实体对应的数据库表名。
//
// GORM 和种子初始化统一使用该名称。
func (Position) TableName() string { return "sys_position" }

// UserPosition 表示用户与岗位的多对多关联。
//
// 关联实体归属 Position 模块，User Entity 通过表名复用该关系。
type UserPosition struct {
	UserId     int64 `gorm:"primaryKey;comment:用户ID;not null;"`
	PositionId int64 `gorm:"primaryKey;index:index_sys_user_position_PositionId;comment:岗位ID;not null;"`
}

// TableName 返回用户岗位关联表名。
//
// GORM 在关联写入和删除校验时统一使用该名称。
func (UserPosition) TableName() string { return "sys_user_position" }
