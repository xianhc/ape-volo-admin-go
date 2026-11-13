package serverresources

import (
	"github.com/xianhc/ape-volo-admin-go/internal/utils"
)

// ServerResourcesResponse 表示服务器资源接口的返回模型。
//
// 它汇总系统、CPU、内存、磁盘和交换分区的当前快照。
type ServerResourcesResponse struct {
	Time   string         `json:"time"`
	Sys    SystemResponse `json:"sys"`
	CPU    CPUResponse    `json:"cpu"`
	Memory MemoryResponse `json:"memory"`
	Disk   DiskResponse   `json:"disk"`
	Swap   SwapResponse   `json:"swap"`
}

// SystemResponse 表示操作系统和运行时信息。
//
// Day 使用可读文本描述当前系统运行时长。
type SystemResponse struct {
	OS             string `json:"os"`
	Day            string `json:"day"`
	RuntimeVersion string `json:"runtimeVersion"`
}

// CPUResponse 表示处理器资源信息。
//
// 使用率字段保持既有字符串格式，避免改变前端展示契约。
type CPUResponse struct {
	Name       string `json:"name"`
	Package    string `json:"package"`
	Core       string `json:"core"`
	CoreNumber string `json:"coreNumber"`
	Logic      string `json:"logic"`
	Used       string `json:"used"`
	Idle       string `json:"idle"`
}

// MemoryResponse 表示物理内存资源信息。
//
// 容量与使用率保持既有带单位字符串格式。
type MemoryResponse struct {
	Total     string `json:"total"`
	Available string `json:"available"`
	Used      string `json:"used"`
	UsageRate string `json:"usageRate"`
}

// DiskResponse 表示磁盘资源信息。
//
// 容量与使用率保持既有带单位字符串格式。
type DiskResponse struct {
	Total     string `json:"total"`
	Available string `json:"available"`
	Used      string `json:"used"`
	UsageRate string `json:"usageRate"`
}

// SwapResponse 表示交换分区资源信息。
//
// 未启用交换分区时各字段返回零值文本。
type SwapResponse struct {
	Total     string `json:"total"`
	Available string `json:"available"`
	Used      string `json:"used"`
	UsageRate string `json:"usageRate"`
}

// fromInfo 将基础设施采集结果转换为接口返回模型。
//
// 转换保持所有现有 JSON 字段和值格式不变。
func fromInfo(info *utils.ServerResourcesInfo) ServerResourcesResponse {
	return ServerResourcesResponse{
		Time:   info.Time,
		Sys:    SystemResponse{OS: info.Sys.Os, Day: info.Sys.Day, RuntimeVersion: info.Sys.RuntimeVersion},
		CPU:    CPUResponse{Name: info.Cpu.Name, Package: info.Cpu.Package, Core: info.Cpu.Core, CoreNumber: info.Cpu.CoreNumber, Logic: info.Cpu.Logic, Used: info.Cpu.Used, Idle: info.Cpu.Idle},
		Memory: MemoryResponse{Total: info.Memory.Total, Available: info.Memory.Available, Used: info.Memory.Used, UsageRate: info.Memory.UsageRate},
		Disk:   DiskResponse{Total: info.Disk.Total, Available: info.Disk.Available, Used: info.Disk.Used, UsageRate: info.Disk.UsageRate},
		Swap:   SwapResponse{Total: info.Swap.Total, Available: info.Swap.Available, Used: info.Swap.Used, UsageRate: info.Swap.UsageRate},
	}
}
