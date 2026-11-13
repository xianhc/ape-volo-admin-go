package exceptionlog

import (
	"time"

	"github.com/xianhc/ape-volo-admin-go/internal/model"
)

// ExceptionLog 表示持久化到数据库的异常日志。
//
// 实体记录请求环境、异常消息和堆栈信息，不直接作为 HTTP 响应返回。
type ExceptionLog struct {
	model.RootKey
	Area                 string                `gorm:"comment:区域"`
	Controller           string                `gorm:"comment:控制器"`
	Action               string                `gorm:"comment:方法"`
	Method               string                `gorm:"comment:请求方式"`
	Description          string                `gorm:"comment:描述"`
	RequestUrl           string                `gorm:"comment:请求url"`
	RequestParameters    model.CustomFieldText `gorm:"comment:请求参数"`
	ExceptionMessage     string                `gorm:"comment:异常短信息"`
	ExceptionMessageFull model.CustomFieldText `gorm:"comment:异常完整信息"`
	ExceptionStack       model.CustomFieldText `gorm:"comment:异常堆栈信息"`
	LogLevel             int                   `gorm:"comment:等级"`
	RequestIp            string                `gorm:"comment:请求IP"`
	IpAddress            string                `gorm:"comment:IP所属真实地址"`
	UserAgent            string                `gorm:"comment:用户代理信息"`
	OperatingSystem      string                `gorm:"comment:操作系统"`
	DeviceType           string                `gorm:"comment:设备类型"`
	BrowserName          string                `gorm:"comment:浏览器名称"`
	Version              string                `gorm:"comment:浏览器版本"`
	CreatorName          string                `gorm:"comment:创建者账号"`
	model.BaseModel
	model.SoftDeleted
}

// TableName 返回异常日志实体的月表基名。
//
// 实际读写由仓储显式选择月份。
func (ExceptionLog) TableName() string { return "log_exception" }

// LogID 返回生产端已经确定的记录 ID。
//
// 用于写入错误摘要，不输出异常正文。
func (l ExceptionLog) LogID() int64 { return l.Id }

// LogTime 返回生产端已经确定的创建时刻。
//
// 消费者据此选月，不按消费时间改写。
func (l ExceptionLog) LogTime() time.Time { return l.CreationTime.Time }
