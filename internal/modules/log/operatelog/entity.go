package operatelog

import (
	"time"

	"github.com/xianhc/ape-volo-admin-go/internal/model"
)

// OperateLog 表示持久化到数据库的操作审计日志。
//
// 实体记录请求环境、请求参数、响应数据与执行耗时，不直接作为 HTTP 响应返回。
type OperateLog struct {
	model.RootKey
	Area              string                `gorm:"comment:区域"`
	Controller        string                `gorm:"comment:控制器"`
	Action            string                `gorm:"comment:方法"`
	Method            string                `gorm:"comment:请求方式"`
	Description       string                `gorm:"comment:描述"`
	RequestUrl        string                `gorm:"comment:请求url"`
	RequestParameters model.CustomFieldText `gorm:"comment:请求参数"`
	ResponseData      model.CustomFieldText `gorm:"comment:响应数据"`
	ExecutionDuration int64                 `gorm:"comment:执行耗时(毫秒)"`
	RequestIp         string                `gorm:"comment:请求IP"`
	IpAddress         string                `gorm:"comment:IP所属真实地址"`
	UserAgent         string                `gorm:"comment:用户代理信息"`
	OperatingSystem   string                `gorm:"comment:操作系统"`
	DeviceType        string                `gorm:"comment:设备类型"`
	BrowserName       string                `gorm:"comment:浏览器名称"`
	Version           string                `gorm:"comment:浏览器版本"`
	CreatorName       string                `gorm:"comment:创建者账号"`
	model.BaseModel
	model.SoftDeleted
}

// TableName 返回基名，持久化由仓储显式选择月表。
//
// 不在实体中固定当前月份。
func (OperateLog) TableName() string { return "log_operate" }

// LogID 返回生产端已经确定的记录 ID。
//
// 用于写入错误摘要，不输出请求正文。
func (l OperateLog) LogID() int64 { return l.Id }

// LogTime 返回生产端已经确定的创建时刻。
//
// 消费者据此选月，不按消费时间改写。
func (l OperateLog) LogTime() time.Time { return l.CreationTime.Time }
