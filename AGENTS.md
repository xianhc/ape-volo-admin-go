# Agent 工作约定

本仓库是基于 Go、Gin 与 GORM 的后台管理系统，module 为
`github.com/xianhc/ape-volo-admin-go`。

## 开始修改

先读 [项目说明](README.md)，并检查相关源码与已有测试。

文档说明的是当前实现；发现文档与代码冲突时，以本项目当前源码和可执行验证为准，并在同一改动中修正文档。

## 修改范围

- 只处理当前任务，不顺手重构或还原用户改动。
- 文档、注释和说明只描述本项目当前职责、行为及必要约束，不保留其他项目或语言实现的对比、对齐状态和迁移过程记录。
- 不读取或提交 `bin/`、`logs/`、`uploads/`、`*.db*` 及临时探针。
- 不保留方案草稿、调试日志、执行快照或已完成迁移记录；长期文档只写当前行为和必要决策。
- 修改接口、配置、数据库或目录职责时，同步更新对应文档和示例。
- 不新增外部依赖，除非标准库和现有依赖均不能满足；新增后执行 `go mod tidy`。

## 分层

- 入口 `cmd/server/main.go` 只装配和启动，保留顶部版权与 SPDX 说明。
- 业务模块放在 `internal/modules/<group>/<module>/`，单包内按 `entity`、`request`、`response`、
  `repository`、`service`、`handler` 分文件；没有职责时不创建空文件。
- 依赖方向为 `handler -> service -> repository -> entity`。Handler 不访问数据库，Service 不依赖 Gin，
  Repository 不使用 HTTP 请求或响应模型。
- Entity、Request、Response 不复用同一 struct；响应必须显式映射，密码和密钥不得进入响应。
- 数据库入口只在 Repository；业务库走 `utils.FromContext(ctx)`，日志查询走 `utils.LogDB(ctx)`。

## 接口与授权

- 当前接口面是 105 个端点。不要增加 GET download、兼容查询路由或 `/api/health`。
- 查询动作以 `get-` 开头；模块使用短横线。导出使用 POST `download`。
- 私有接口必须通过 `routeauth.Group` 注册，并在同一行使用 `routeauth.LoginOnly()` 或
  `routeauth.Require(permissionCode, roleCodes...)` 声明策略；禁止直接调用 Gin 注册私有路由。
- 新增权限码时同步权限常量、菜单种子和角色菜单种子，不维护独立路由权限表。
- 权限码使用 `internal/global/constants/permission`，角色码使用
  `internal/global/constants/rolecode`，禁止散落字符串。
- 未登记私有路由 fail closed。敏感用户、角色、菜单操作还需 admin 角色，在路由注册处显式声明。
- 用户岗位/角色更新操作请求中的目标 `id`，不得擅自替换为当前登录用户。
- IP 只用 `utils.GetClientIP(c)`，JWT 用户只用 `utils.UserID(claims)`。

## 数据库

- 领域命名使用 `Position`、`Department`、`AuthCode`、`MenuType`、`DataScopeType`、`GenderCode`、
  `UserName`、`PreferencesConfig`。
- 树根 `ParentId` 为 `*int64` 且数据库为 `NULL`，使用 `IsRoot()` / `ParentKey()`。
- 审计字段由 ORM 回调填充。`creator_id` 可空；后台任务不得写 0。请求不接收审计字段。
- 软删除和数据权限由 GORM 回调统一施加；空数据范围必须返回空，不得退化为不过滤。
- 四张关联表使用复合主键。新增唯一性或高频筛选字段时同步实体索引与数据库文档。
- 主库和日志库目标必须不同；日志只写独立日志库的 `YYYYMM01` 月表。

## 缓存与事务

- 缓存只通过 `global.Cache`；键前缀只从 `cachePrefix` 取。
- 查询缓存使用 `querycache.GetOrLoad`；用户菜单使用 `GetOrLoadAlways`，固定 120 分钟。
- 事务内不读写共享查询缓存；写入失效在提交后执行，回滚不得清缓存。
- ORM 按表前缀失效是有意的保守实现，覆盖关联写和非 Service 写，不改成分散的随手删除。
- 会话只存在缓存，缺失即未登录；刷新轮换使用 `CompareExchange`。
- 多步写使用 `utils.InTransaction`，提交后副作用使用 `utils.AfterCommit`。

## 验证

| 改动范围               | 必须执行                                                        |
| ---------------------- | --------------------------------------------------------------- |
| 仅文档                 | 核对链接、源码入口、命令和事实                                  |
| 任一 Go 代码           | `gofmt -l .`、`go build ./...`、`go vet ./...`、`go test ./...` |
| 依赖                   | 追加 `go mod tidy`，确认 `go.mod` / `go.sum` 无意外变化         |
| 缓存、限流、会话、鉴权 | Redis 与内存模式各启动一次；在线用户只验 Redis 模式             |
| 数据库结构             | SQLite 临时库检查实际列/索引；其他受支持数据库按改动风险验收    |
| 接口契约               | 核对 105 端点集合、权限路由、菜单种子和实际 HTTP 状态           |

测试方法使用英文。完成时报告实际执行的命令、结果和未验证范围，不把静态检查通过描述为业务全量验收。

## 提交

提交信息使用中文，格式 `类型: 简述`，类型为 `feat` / `fix` / `refactor` / `docs` / `chore` / `test`。
一次提交只处理一个主题，不提交构建物、日志、本地数据库和临时配置。
