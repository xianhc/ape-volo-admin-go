package onlineuser

import (
	"time"

	"github.com/xianhc/ape-volo-admin-go/internal/model"
	"github.com/xianhc/ape-volo-admin-go/internal/payloads/auth"
)

// OnlineUserResponse 表示在线用户接口的返回模型。
//
// AccessToken 字段沿用现有契约，实际承载可撤销的会话标识而非认证令牌。
type OnlineUserResponse struct {
	UserId          int64      `json:"userId,string"`
	Account         string     `json:"account"`
	NickName        string     `json:"nickName"`
	DeptId          int64      `json:"deptId,string"`
	DeptName        string     `json:"deptName"`
	Ip              string     `json:"ip"`
	Address         string     `json:"address"`
	OperatingSystem string     `json:"operatingSystem"`
	DeviceType      string     `json:"deviceType"`
	BrowserName     string     `json:"browserName"`
	Version         string     `json:"version"`
	AccessToken     string     `json:"accessToken"`
	LoginTime       model.Time `json:"loginTime"`
}

// fromLoginInfo 将会话登录信息转换为在线用户响应。
//
// sessionId 写入 AccessToken 字段以保持既有强制下线请求契约。
// 会话创建时间转换到服务器本地时区，再使用统一的日期时间格式输出。
func fromLoginInfo(info auth.LoginUserInfo, sessionId string, createdAt time.Time) OnlineUserResponse {
	return OnlineUserResponse{UserId: info.UserId, Account: info.Account, NickName: info.NickName, DeptId: info.DeptId, DeptName: info.DeptName, Ip: info.Ip, Address: info.Address, OperatingSystem: info.OperatingSystem, DeviceType: info.DeviceType, BrowserName: info.BrowserName, Version: info.Version, AccessToken: sessionId, LoginTime: model.Time{Time: createdAt.Local()}}
}
