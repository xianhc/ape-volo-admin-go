package request

type Pagination struct {
	PageIndex  int      `json:"pageIndex" form:"pageIndex" validate:"required,numeric"` // 页码
	PageSize   int      `json:"pageSize" form:"pageSize" validate:"required,numeric"`   // 每页大小
	SortFields []string `json:"sortFields" form:"sortFields"`                           //排序
}

func NewPagination() *Pagination {
	return &Pagination{
		PageIndex:  1,          // 默认页码为 1
		PageSize:   10,         // 默认每页大小为 10
		SortFields: []string{}, // 默认排序字段为空
	}
}

type IdCollection struct {
	IdArray []string `json:"idArray" form:"idArray"` //id集合
}
