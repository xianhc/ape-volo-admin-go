package utils

import (
	"fmt"
	"github.com/cloudfoundry/gosigar"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"go-apevolo/global"
	"go.uber.org/zap"
	"runtime"
	"strconv"
	"time"
)

const (
	B  = 1
	KB = 1024 * B
	MB = 1024 * KB
	GB = 1024 * MB
)

type ServerResourcesInfo struct {
	Time   string `json:"time"`
	Sys    Sys    `json:"sys"`
	Cpu    Cpu    `json:"cpu"`
	Memory Memory `json:"memory"`
	Disk   Disk   `json:"disk"`
	Swap   Swap   `json:"swap"`
}

type Sys struct {
	Os             string `json:"os"`             //服务器系统版本
	Day            string `json:"day"`            //运行时间
	RuntimeVersion string `json:"runtimeVersion"` //业务系统运行时版本
}

type Cpu struct {
	Name       string `json:"name"`       //cpu名称
	Package    string `json:"package"`    //物理cpu
	Core       string `json:"core"`       //物理核心
	CoreNumber string `json:"coreNumber"` //核心数量
	Logic      string `json:"logic"`      //逻辑cpu
	Used       string `json:"used"`       //使用
	Idle       string `json:"idle"`       //剩余
}

type Memory struct {
	Total     string `json:"total"`
	Available string `json:"available"`
	Used      string `json:"used"`
	UsageRate string `json:"usageRate"`
}

type Disk struct {
	Total     string `json:"total"`
	Available string `json:"available"`
	Used      string `json:"used"`
	UsageRate string `json:"usageRate"`
}

type Swap struct {
	Total     string `json:"total"`
	Available string `json:"available"`
	Used      string `json:"used"`
	UsageRate string `json:"usageRate"`
}

// GetServerResourcesInfo
// @description:  获取服务器资源信息
// @return: ServerResourcesInfo

func GetServerResourcesInfo() *ServerResourcesInfo {
	serverInfo := &ServerResourcesInfo{}
	currentTime := time.Now()
	formattedTime := currentTime.Format("15:04:05")
	serverInfo.Time = formattedTime
	serverInfo.Sys = getSys()
	serverInfo.Cpu = getCpu()
	serverInfo.Memory = getMemory()
	serverInfo.Disk = getDisk()
	serverInfo.Swap = getSwap()
	return serverInfo
}

func getSys() (sys Sys) {
	info, err := host.Info()
	if err != nil {
		global.Logger.Error("获取系统信息失败", zap.Error(err))
		return
	}
	// 获取系统运行时间
	timeResult := ""
	day := 0
	hour := 0
	minute := 0
	second := info.Uptime

	if second > 60 {
		minute = int(second / 60)
		second %= 60
	}

	if minute > 60 {
		hour = minute / 60
		minute %= 60
	}

	if hour <= 24 {
		timeResult = fmt.Sprintf("%d%s%d%s%d%s", hour, "小时", minute, "分钟", second, "秒")
	}
	day = hour / 24
	hour %= 24
	timeResult = fmt.Sprintf("%d%s%d%s%d%s%d%s", day, "天", hour, "小时", minute, "分钟", second, "秒")

	sys.Os = info.Platform
	sys.Day = timeResult
	sys.RuntimeVersion = runtime.Version()
	return sys
}

func getCpu() (c Cpu) {

	// 获取CPU信息
	cpuInfo, err := cpu.Info()
	if err != nil {
		global.Logger.Error("无法获取CPU信息", zap.Error(err))
		return
	}
	if len(cpuInfo) > 0 {
		c.Name = cpuInfo[0].ModelName
	}
	c.Package = fmt.Sprintf("%d个物理CPU", len(cpuInfo))

	// 获取物理CPU数量
	physicalCPUCount, err := cpu.Counts(false)
	if err != nil {
		global.Logger.Error("无法获取物理CPU数量", zap.Error(err))
		return
	}
	c.CoreNumber = strconv.Itoa(physicalCPUCount)
	c.Core = fmt.Sprintf("%s个物理核心", c.CoreNumber)

	// 获取逻辑CPU数量
	logicalCPUCount, err := cpu.Counts(true)
	if err != nil {
		global.Logger.Error("无法获取逻辑CPU数量", zap.Error(err))
		return
	}
	c.Logic = fmt.Sprintf("%s个逻辑CPU", strconv.Itoa(logicalCPUCount))

	// 获取CPU使用率
	cpuUsage, err := cpu.Percent(0, false)
	if err != nil {
		fmt.Println("无法获取CPU使用率：", err)
		return
	}

	// 计算平均使用率
	avgUsageFloat := 0.0
	for _, usage := range cpuUsage {
		avgUsageFloat += usage
	}
	avgUsageFloat /= float64(len(cpuUsage))

	// 将平均使用率转换为整数
	//avgUsage := int(avgUsageFloat)

	c.Used = fmt.Sprintf("%.2f", avgUsageFloat)
	c.Idle = fmt.Sprintf("%.2f", 100-avgUsageFloat)

	return c
}

func getMemory() (m Memory) {
	if o, err := mem.VirtualMemory(); err != nil {
		return
	} else {
		m.Total = fmt.Sprintf("%.2f", float64(o.Total)/GB) + "GiB"
		m.Available = fmt.Sprintf("%.2f", float64(o.Total-o.Used)/GB) + "GiB"
		m.Used = fmt.Sprintf("%.2f", float64(o.Used)/GB) + "GiB"
		m.UsageRate = fmt.Sprintf("%.2f", o.UsedPercent)
	}
	return m
}

func getDisk() (d Disk) {

	partitions, err := disk.Partitions(false)
	if err != nil {
		return d
	}

	var allTotal uint64
	var allUsed uint64

	for _, p := range partitions {
		d, _ := disk.Usage(p.Mountpoint)
		allTotal += d.Total
		allUsed += d.Used
	}
	d.Total = fmt.Sprintf("%.2f", float64(allTotal)/GB) + "GiB"
	d.Used = fmt.Sprintf("%.2f", float64(allUsed)/GB) + "GiB"
	d.Available = fmt.Sprintf("%.2f", float64(allTotal-allUsed)/GB) + "GiB"
	d.UsageRate = fmt.Sprintf("%.2f", ((float64(allUsed)/GB)/(float64(allTotal)/GB))*100)
	//只有一个磁盘这样写就行
	//disk.Usage("/")
	return d
}

func getSwap() (s Swap) {
	swap := sigar.Swap{}
	err := swap.Get()
	if err != nil {
		return
	}

	s.Total = fmt.Sprintf("%.2f", float64(swap.Total)/GB) + "GiB"
	s.Used = fmt.Sprintf("%.2f", float64(swap.Used)/GB) + "GiB"
	s.Available = fmt.Sprintf("%.2f", float64(swap.Free)/GB) + "GiB"
	s.UsageRate = fmt.Sprintf("%.2f", ((float64(swap.Used)/GB)/(float64(swap.Total)/GB))*100)
	return s
}
