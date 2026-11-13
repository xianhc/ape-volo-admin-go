package model

// DataScoped 标记实体参与数据权限过滤。
//
// 内嵌本结构体的实体，查询时由 ORM 回调自动追加「可见创建者」条件，
// 见 initialize/data_scope_filter.go。新增实体只要内嵌这里就自动生效，
// 不需要再往查询里传用户 ID。
//
// 不内嵌的实体不受数据权限约束：菜单（按父子关系组装，过滤会导致父节点缺失、
// 导航断裂）与字典（全局配置，不随创建者变化）。
type DataScoped struct{}
