package onlineuser

import (
	"github.com/xianhc/ape-volo-admin-go/internal/payloads/request"
)

// QueryRequest 表示在线用户列表的分页参数。
//
// 非正页码与页大小由 Service 分别回落到 1 和 10。
type QueryRequest struct {
	request.OptionalPagination
}
