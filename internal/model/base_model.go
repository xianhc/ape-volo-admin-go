package model

// RootKey 主键
type RootKey struct {
	Id int64 `json:"id,string" gorm:"primaryKey;comment:主键编码;not null;"` //主键编码
}

// BaseModel 审计字段。
//
// 创建者与修改者记录的是用户 ID 而不是账号名：账号允许改名，
// 用 ID 才能保证历史记录始终反查到唯一的人。
//
// 时间字段用 model.Time 而不是 time.Time：前者序列化为 `yyyy-MM-dd HH:mm:ss`，
// 后者是 RFC3339，前端表格直接渲染会显示成 ISO 串。
type BaseModel struct {
	CreatorId            *int64 `json:"creatorId,omitempty,string" gorm:"column:creator_id;index;comment:创建者ID"`
	CreationTime         Time   `json:"creationTime" gorm:"column:creation_time;comment:创建时间;not null"`
	LastModifierId       *int64 `json:"lastModifierId,omitempty,string" gorm:"column:last_modifier_id;comment:最后修改者ID"`
	LastModificationTime *Time  `json:"lastModificationTime,omitempty" gorm:"column:last_modification_time;comment:最后修改时间"`
}

// SoftDeleted 逻辑删除标记。
//
// 除标记位外还记录删除人与删除时间：排查「数据为什么没了」时，
// 只有标记位无法回答是谁、什么时候删的。
//
// 这些字段由 ORM 回调自动维护（见 initialize/audit_filler.go 与
// initialize/soft_delete_filter.go），业务代码不要手工赋值。
type SoftDeleted struct {
	IsDeleted    bool   `json:"-" gorm:"default:0;comment:是否删除(逻辑性);not null;"`
	DeleterId    *int64 `json:"deleterId,omitempty,string" gorm:"column:deleter_id;comment:删除者ID"`
	DeletionTime *Time  `json:"deletionTime,omitempty" gorm:"column:deletion_time;comment:删除时间"`
}
