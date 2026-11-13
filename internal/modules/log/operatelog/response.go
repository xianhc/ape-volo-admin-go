package operatelog

import (
	"github.com/xianhc/ape-volo-admin-go/internal/model"
)

// OperateLogResponse 表示操作日志接口的返回模型。
//
// 它显式列出审计字段，数据库实体不会进入 HTTP 输出。
type OperateLogResponse struct {
	Id                   int64                 `json:"id,string"`
	Area                 string                `json:"area"`
	Controller           string                `json:"controller"`
	Action               string                `json:"action"`
	Method               string                `json:"method"`
	Description          string                `json:"description"`
	RequestUrl           string                `json:"requestUrl"`
	RequestParameters    model.CustomFieldText `json:"requestParameters"`
	ResponseData         model.CustomFieldText `json:"responseData"`
	ExecutionDuration    int64                 `json:"executionDuration"`
	RequestIp            string                `json:"requestIp"`
	IpAddress            string                `json:"ipAddress"`
	UserAgent            string                `json:"userAgent"`
	OperatingSystem      string                `json:"operatingSystem"`
	DeviceType           string                `json:"deviceType"`
	BrowserName          string                `json:"browserName"`
	Version              string                `json:"version"`
	CreatorName          string                `json:"creatorName"`
	CreatorId            *int64                `json:"creatorId,omitempty,string"`
	CreationTime         model.Time            `json:"creationTime"`
	LastModifierId       *int64                `json:"lastModifierId,omitempty,string"`
	LastModificationTime *model.Time           `json:"lastModificationTime,omitempty"`
	DeleterId            *int64                `json:"deleterId,omitempty,string"`
	DeletionTime         *model.Time           `json:"deletionTime,omitempty"`
}

// fromEntity 将操作日志实体转换为接口返回模型。
//
// 转换保持既有 JSON 字段并隐藏软删除标记。
func fromEntity(log OperateLog) OperateLogResponse {
	return OperateLogResponse{Id: log.Id, Area: log.Area, Controller: log.Controller, Action: log.Action, Method: log.Method, Description: log.Description, RequestUrl: log.RequestUrl, RequestParameters: log.RequestParameters, ResponseData: log.ResponseData, ExecutionDuration: log.ExecutionDuration, RequestIp: log.RequestIp, IpAddress: log.IpAddress, UserAgent: log.UserAgent, OperatingSystem: log.OperatingSystem, DeviceType: log.DeviceType, BrowserName: log.BrowserName, Version: log.Version, CreatorName: log.CreatorName, CreatorId: log.CreatorId, CreationTime: log.CreationTime, LastModifierId: log.LastModifierId, LastModificationTime: log.LastModificationTime, DeleterId: log.DeleterId, DeletionTime: log.DeletionTime}
}

// fromEntities 将操作日志实体列表转换为接口返回模型列表。
//
// 返回值始终是非 nil 切片。
func fromEntities(logs []OperateLog) []OperateLogResponse {
	result := make([]OperateLogResponse, 0, len(logs))
	for _, log := range logs {
		result = append(result, fromEntity(log))
	}
	return result
}
