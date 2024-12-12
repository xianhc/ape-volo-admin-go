package monitor

import (
	"go-apevolo/model"
	"go-apevolo/utils"
)

type AuditLog struct {
	model.RootKey
	Area              string                `json:"area" gorm:"comment:区域"`                     // 区域
	Controller        string                `json:"controller" gorm:"comment:控制器"`              // 控制器
	Action            string                `json:"action"  gorm:"comment:方法"`                  // 方法
	Method            string                `json:"method" gorm:"comment:请求方式"`                 // 请求方式
	Description       string                `json:"description" gorm:"comment:描述"`              // 描述
	RequestUrl        string                `json:"requestUrl"  gorm:"comment:请求url"`           // 请求url
	RequestParameters utils.CustomFieldText `json:"requestParameters" gorm:"comment:请求参数"`      // 请求参数
	ResponseData      utils.CustomFieldText `json:"responseData" gorm:"comment:响应数据"`           // 响应数据
	ExecutionDuration int64                 `json:"executionDuration"  gorm:"comment:执行耗时(毫秒)"` // 执行耗时(毫秒)
	RequestIp         string                `json:"requestIp" gorm:"comment:请求IP"`              // 请求IP
	IpAddress         string                `json:"ipAddress" gorm:"comment:IP所属真实地址"`          // IP所属真实地址
	OperatingSystem   string                `json:"operatingSystem"  gorm:"comment:操作系统"`       // 操作系统
	DeviceType        string                `json:"deviceType" gorm:"comment:设备类型"`             // 设备类型
	BrowserName       string                `json:"browserName" gorm:"comment:浏览器名称"`           // 浏览器名称
	Version           string                `json:"version"  gorm:"comment:浏览器版本"`              // 浏览器版本
	model.BaseModel
	model.SoftDeleted
}

func (AuditLog) TableName() string {
	return "log_audit"
}
