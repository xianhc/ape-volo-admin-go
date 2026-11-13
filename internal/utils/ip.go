package utils

import (
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/xianhc/ape-volo-admin-go/internal/utils/ipregion"
)

var ipData struct {
	once    sync.Once
	service *ipregion.Service
	err     error
}

// GetClientIP 返回请求的真实客户端 IP。
//
// 是否采信代理头由 Gin 的可信代理配置决定。
func GetClientIP(c *gin.Context) string {
	if ip := c.ClientIP(); ip != "" {
		return ip
	}
	if ip := c.RemoteIP(); ip != "" {
		return ip
	}
	return "127.0.0.1"
}

// InitIpData 加载 IPv4 和 IPv6 归属地数据库。
//
// 进程内只加载一次，初始化失败时由启动入口处理错误。
func InitIpData() error {
	ipData.once.Do(func() {
		ipData.service, ipData.err = ipregion.New("resource/ip/ip2region_v4.xdb", "resource/ip/ip2region_v6.xdb")
	})
	return ipData.err
}

// SearchIpAddress 返回客户端归属地文本。
//
// 保留国家、省份、城市、ISP、国家代码的竖线分隔格式，供日志和会话记录使用。
func SearchIpAddress(ip string) (string, error) {
	if err := InitIpData(); err != nil {
		return "", err
	}
	region, err := ipData.service.Lookup(ip)
	if err != nil {
		return "", err
	}
	return region.String(), nil
}
