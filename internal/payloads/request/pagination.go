package request

// LimitedResultRequest 限制一次分页查询返回的数量。
//
// 列表校验要求页大小为 1～1000；仅绑定参数时允许省略，供导出使用。
type LimitedResultRequest struct {
	PageSize int `json:"pageSize" form:"pageSize" validate:"required,min=1,max=1000" binding:"omitempty,min=1,max=1000"`
}

// PagedResultRequest 在数量限制上增加从 1 开始的页码。
//
// 匿名嵌入使调用方可以直接访问 PageSize 与 PageIndex。
type PagedResultRequest struct {
	LimitedResultRequest
	PageIndex int `json:"pageIndex" form:"pageIndex" validate:"required,min=1,max=1000000" binding:"omitempty,min=1,max=1000000"`
}

// SortedResultRequest 承载多字段排序条件。
//
// 字段名及方向由查询入口校验，不作为原始 SQL 直接使用。
type SortedResultRequest struct {
	Sorting string `json:"sorting" form:"sorting"`
}

// Pagination 组合分页和排序请求。
//
// JSON 和表单参数保持扁平结构，不增加 pagination 包装层。
type Pagination struct {
	PagedResultRequest
	SortedResultRequest
}

// OptionalPagination 承载可省略的分页及排序参数。
//
// 树查询和会话列表可不分页或由 Service 应用默认值；传入的非零分页值仍需满足范围约束。
type OptionalPagination struct {
	PageIndex int `json:"pageIndex" form:"pageIndex" validate:"omitempty,min=1,max=1000000" binding:"omitempty,min=1,max=1000000"`
	PageSize  int `json:"pageSize" form:"pageSize" validate:"omitempty,min=1,max=1000" binding:"omitempty,min=1,max=1000"`
	SortedResultRequest
}

// CreationTimeRange 承载可选的创建时间范围。
//
// T 保留各接口现有的时间值类型；空范围表示不过滤，非空范围必须包含起止两个值。
type CreationTimeRange[T any] struct {
	CreateTime []T `json:"createTime" form:"createTime" binding:"omitempty,len=0|len=2" validate:"omitempty,len=0|len=2"`
}

// DateRangePagination 在分页和排序请求上增加创建时间范围。
//
// 业务查询匿名嵌入后可直接访问全部公共字段，无需重复声明或逐层转发。
type DateRangePagination[T any] struct {
	Pagination
	CreationTimeRange[T]
}

// NewPagination 创建第一页、每页十条的分页参数。
//
// 匿名嵌入不会自动调用构造函数，需要默认值的内部调用显式使用本函数。
func NewPagination() *Pagination {
	p := &Pagination{}
	p.PageIndex, p.PageSize = 1, 10
	return p
}
