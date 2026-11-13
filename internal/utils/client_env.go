package utils

import "github.com/gin-gonic/gin"

// ClientEnvironment 保存一次请求的客户端环境信息。
//
// IP 取自连接地址（是否采信代理头由 system.trusted-proxies 决定），
// 归属地由 ip2region 解析，操作系统 / 设备类型 / 浏览器 / 版本取自 User-Agent。
type ClientEnvironment struct {
	// IP 客户端真实 IP
	IP string
	// Location IP 归属地，解析失败时为空字符串
	Location string
	// UserAgentInfo 操作系统、设备类型、浏览器与版本
	UserAgentInfo
}

// ParseClientEnvironment 解析请求的客户端环境。
//
// 同一次请求只调用一次：UA 解析与 IP 归属地查询都有开销，
// 审计日志、异常日志与在线用户共用同一份结果，避免重复解析、重复查库。
func ParseClientEnvironment(c *gin.Context) ClientEnvironment {
	env := ClientEnvironment{
		IP:            GetClientIP(c),
		UserAgentInfo: ParseUserAgent(c.Request.UserAgent()),
	}
	if location, err := SearchIpAddress(env.IP); err == nil {
		env.Location = location
	}
	return env
}
