package utils

import (
	"github.com/gin-gonic/gin"
	"github.com/lionsoul2014/ip2region/binding/golang/xdb"
	"go-apevolo/global"
	"go.uber.org/zap"
	"os"
	"strings"
	"sync"
)

// 全局变量，用于存储文件内容
var dbFileContent []byte
var mu sync.Mutex // 用于确保并发安全

func GetClientIP(c *gin.Context) string {
	ClientIP := c.ClientIP()
	RemoteIP := c.RemoteIP()
	ip := c.Request.Header.Get("X-Forwarded-For")
	if strings.Contains(ip, "127.0.0.1") || ip == "" {
		ip = c.Request.Header.Get("X-real-ip")
	}
	if ip == "" {
		ip = "127.0.0.1"
	}
	if RemoteIP != "127.0.0.1" {
		ip = RemoteIP
	}
	if ClientIP != "127.0.0.1" {
		ip = ClientIP
	}
	return ip
}

func InitIpData() {
	var dbPath = "./resource/ip/ip2region.xdb"
	// 1、从 dbPath 加载整个 xdb 到内存
	cBuff, err := xdb.LoadContentFromFile(dbPath)
	if err != nil {
		global.Logger.Error("failed to load content from ", zap.Error(err))
		os.Exit(0)
	}

	// 使用互斥锁确保并发安全
	mu.Lock()
	defer mu.Unlock()

	dbFileContent = cBuff
}

// 获取文件内容的函数
func getFileContent() []byte {
	// 使用互斥锁确保并发安全
	mu.Lock()
	defer mu.Unlock()

	return dbFileContent
}

func SearchIpAddress(ip string) (ipAddress string, err error) {
	content := getFileContent()
	// 2、用全局的 cBuff 创建完全基于内存的查询对象。
	searcher, err := xdb.NewWithBuffer(content)
	if err != nil {
		global.Logger.Error("failed to create searcher with content ", zap.Error(err))
		return
	}
	region, err := searcher.SearchByStr(ip)

	return region, err
}
