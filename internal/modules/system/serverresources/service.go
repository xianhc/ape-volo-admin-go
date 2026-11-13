package serverresources

import (
	"github.com/xianhc/ape-volo-admin-go/internal/utils"
)

// Service 提供服务器资源查询业务。
//
// 该模块不访问数据库，直接调用基础设施采集能力并转换为独立 Response。
type Service struct{}

// Query 获取当前服务器资源快照。
//
// 每次调用都会重新采集系统、CPU、内存、磁盘和交换分区信息。
func (Service) Query() ServerResourcesResponse {
	return fromInfo(utils.GetServerResourcesInfo())
}
