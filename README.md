<h1 align="center">Ape-Volo-Admin-Go</h1>

<p align="center">基于 Go、Gin 与 GORM 的后台管理系统后端</p>
<p align="center">集成认证授权、组织管理与常用业务模块，为中后台业务开发提供基础。</p>

<p align="center">
	<img src="https://img.shields.io/badge/Go-1.27.3-00ADD8?style=flat-square" alt="Go 1.27.3" />
	<img src="https://img.shields.io/badge/Gin-1.12-008ECF?style=flat-square" alt="Gin 1.12" />
	<img src="https://img.shields.io/badge/GORM-1.31-00A29A?style=flat-square" alt="GORM 1.31" />
	<a href="LICENSE"><img src="https://img.shields.io/badge/License-Apache--2.0-blue?style=flat-square" alt="Apache License 2.0" /></a>
</p>

<p align="center">
	<a href="#版权与授权">版权与授权</a> ·
	<a href="https://www.apevolo.com">官网与在线文档</a> ·
	<a href="https://vip.apevolo.com">在线体验</a> ·
	<a href="#快速开始">快速开始</a> ·
	<a href="#系统架构">系统架构</a> ·
	<a href="https://gitee.com/xianhc/ape-volo-admin-go">Gitee</a>
</p>

## 项目文档

- **系统官网 / 在线文档**：[https://www.apevolo.com](https://www.apevolo.com)
- **代码与开发入口**：[系统架构](#系统架构) · [源码导航](#源码导航) · [协作约定](AGENTS.md)
- **软件著作权**：[查看证书](https://www.apevolo.com/copyright.pdf)

## 在线演示

- **系统预览**：[https://vip.apevolo.com](https://vip.apevolo.com)
- **体验账号**：`apevolo`
- **体验密码**：`123456`

## 项目介绍

Ape-Volo-Admin 是面向管理后台与企业内部业务系统的开源项目系列，覆盖 **C#、Go、Java 三种后端语言**，并提供基于 **Vue 3 + Element Plus** 的 **TypeScript 与 JavaScript 双前端版本**。团队可以结合现有技术栈与开发习惯，选择适合自己的版本。

项目内置用户、角色、菜单、部门和岗位管理，集成认证授权、数据权限、文件管理、任务调度、邮件与日志等常用能力。

**本仓库为 Go 后端**，使用 Gin 提供 HTTP 接口，使用 GORM 管理持久化，按业务模块组织代码。管理界面可使用主仓库中的 TypeScript 前端或独立 JavaScript 前端，详见[项目系列](#项目系列)。

### 项目特点

- **模块化分层**：业务模块独立组织 Handler、Service、Repository、Entity 及请求、响应模型。
- **权限控制**：基于角色、菜单和按钮权限授权，支持全部、本人、本部门、本部门及以下、自定义五种数据范围。
- **会话认证**：JWT、刷新令牌轮换、验证码与登录失败限制，会话状态统一存储在缓存中。
- **缓存与限流**：支持进程内存和 Redis 两种模式，查询缓存通过统一接口管理。
- **数据一致性**：提供事务与提交后回调，统一处理审计字段、软删除、数据范围和缓存失效。
- **日志与调度**：操作、异常与任务执行日志写入独立日志库月表，支持定时任务与邮件队列。
- **接口文档**：集成 Swagger UI，便于接口调试与前后端联调。

## 功能模块

| 模块       | 已实现功能                                           |
| ---------- | ---------------------------------------------------- |
| 权限与组织 | 用户、角色、菜单、部门、岗位、菜单授权、数据范围配置 |
| 系统管理   | 字典与字典项、系统参数、应用密钥、文件上传与文件记录 |
| 运维管理   | 定时任务配置、启停与执行、服务器监控                 |
| 日志管理   | 操作日志、异常日志、任务执行日志                     |
| 邮件管理   | 邮件账户、邮件模板、发送队列                         |
| 会话管理   | 在线用户查询、导出与强制下线，需要启用 Redis         |
| 个人中心   | 个人资料、头像、密码、邮箱与界面偏好                 |

## 技术栈

| 层次       | 主要技术                                           |
| ---------- | -------------------------------------------------- |
| 后端框架   | Go、Gin                                            |
| 数据访问   | GORM、SQLite / MySQL / SQL Server / PostgreSQL     |
| 缓存与限流 | 进程内缓存、go-redis、ulule/limiter                |
| 认证与安全 | JWT、BCrypt、RSA、base64Captcha                    |
| 配置与校验 | Viper、validator                                   |
| 日志与调度 | Zap、robfig/cron                                   |
| 接口与工具 | Swaggo、Excelize、go-mail、gopsutil、ip2region     |
| 配套前端   | Vue 3、TypeScript / JavaScript、Vite、Element Plus |

后端依赖及 Go 版本要求维护在 [go.mod](go.mod)。Oracle 驱动依赖已包含，启用时需要 64 位 CGO 编译器、Oracle Instant Client，并使用 `oracle` 构建标签；常规构建不包含 Oracle 驱动。

## 快速开始

安装 Go 1.27.3 或更高版本。系统默认使用 **SQLite + 内置内存缓存**，首次启动自动初始化数据库；使用内置缓存时，确认 [config.yaml](config.yaml) 中的 `system.use-redis` 为 `false`。

在仓库根目录执行即可启动：

```bash
go run ./cmd/server
```

- **后端地址**：[http://localhost:9002](http://localhost:9002)
- **Swagger UI**：[http://localhost:9002/swagger/api/index.html](http://localhost:9002/swagger/api/index.html)

前端使用[主仓库的 TypeScript 前端](https://github.com/xianhc/ape-volo-admin)或[独立 JavaScript 前端](https://github.com/xianhc/ape-volo-web-js)，按对应仓库说明启动，后端代理地址设为 `http://localhost:9002`。

## Swagger 文档

项目已包含生成的接口文档。修改接口注解后，在仓库根目录执行：

```bash
go run github.com/swaggo/swag/cmd/swag@v1.16.6 init -g cmd/server/main.go -o internal/docs --parseInternal --parseDependency
```

文档生成到 [internal/docs](internal/docs)，重启后端后即可在 Swagger UI 查看更新；[config.yaml](config.yaml) 中的 `swagger.enabled` 需为 `true`。

## 系统架构

### 请求与分层

请求由 Gin 路由进入，经中间件处理认证、限流、审计等横切逻辑，再调用业务模块。私有路由统一通过 `routeauth.Group` 注册，并显式声明登录要求、权限码及必要角色。

```text
HTTP 请求
	-> Router / RouteAuth / Middleware
	-> Handler     参数绑定、校验与 HTTP 响应
	-> Service     业务规则、事务编排与响应映射
	-> Repository  数据查询与持久化
	-> Entity      数据库实体
```

业务代码位于 `internal/modules/<group>/<module>/`。各模块按职责拆分文件，请求、响应与实体分别定义；Handler 不直接访问数据库，Service 不依赖 Gin，Repository 不使用 HTTP 请求或响应模型。

缓存统一通过 `global.Cache` 访问，查询缓存由 `querycache` 管理；多步写入使用事务，缓存失效等副作用在提交后执行。审计字段、软删除和数据权限由 GORM 回调统一处理。业务库与日志库分别管理，操作、异常和任务执行日志按月存储。

操作日志和异常日志共用路由元数据：模块通过 `middleware.LogModule` 声明 `area`、`controller`，接口通过 `middleware.LogAction` 声明中文 `description`，`action` 取路由动作。元数据在请求处理后统一写入日志；未声明的接口使用路由信息作为缺省值，历史日志不自动回填。

### 目录结构

```text
ape-volo-admin-go/
├── cmd/server/
│   └── main.go                 程序入口、启动装配与退出处理
├── internal/
│   ├── routeauth/              路由权限策略与授权检查
│   ├── config/                 配置结构定义
│   ├── core/                   配置加载、日志与 HTTP 服务生命周期
│   ├── docs/                   Swagger 接口文档与元数据
│   ├── global/                 全局基础设施对象与常量
│   ├── initialize/             数据库、种子、缓存、路由与任务初始化
│   ├── job/                    定时任务定义与执行逻辑
│   ├── logstore/               日志月表、查询与写入
│   ├── middleware/             认证、限流、事务、审计等中间件
│   ├── model/                  公共实体字段与数据范围模型
│   ├── modules/                按领域组织的业务模块
│   │   ├── authorization/      登录认证与会话
│   │   ├── identity/           用户、角色、菜单与在线用户
│   │   ├── organization/       部门与岗位
│   │   ├── system/             字典、设置、文件、应用密钥与任务
│   │   ├── message/            邮件账户、模板与发送队列
│   │   └── log/                操作、异常与任务日志查询
│   ├── payloads/               公共请求、响应与认证载荷
│   ├── router/                 按模块注册路由与权限策略
│   └── utils/                  事务、缓存、限流、上传与通用工具
├── resource/
│   ├── db/                     初始化种子数据
│   ├── ip/                     IP 归属地离线资源
│   └── keys/                   RSA 密钥资源
├── data/                       本地数据库目录
├── logs/                       应用日志输出
├── uploads/                    上传文件目录
├── bin/                        本地构建输出
├── config.yaml                 服务配置
├── go.mod                      Go 版本与依赖声明
├── AGENTS.md                   协作与验证约定
├── README.md                   项目说明
└── LICENSE                     许可证
```

## 开发检查

在仓库根目录执行：

```bash
gofmt -l .
go build ./...
go vet ./...
go test ./...
```

`gofmt -l .` 应无输出。修改依赖后执行 `go mod tidy` 并检查依赖文件变化；缓存、限流、会话与鉴权改动还需分别验证 Redis 和内存模式，在线用户功能仅在 Redis 模式验收。具体要求见 [AGENTS.md](AGENTS.md)。

## 项目系列

主仓库提供 .NET 后端与 TypeScript 前端，其他版本可按团队技术栈选择。

| 项目                                                                 | 定位                 | 核心技术                                               |
| -------------------------------------------------------------------- | -------------------- | ------------------------------------------------------ |
| [ape-volo-admin](https://github.com/xianhc/ape-volo-admin)           | 主仓库 · 完整前后端  | .NET 10 + SqlSugar / Vue 3 + TypeScript + Element Plus |
| **[ape-volo-admin-go](https://github.com/xianhc/ape-volo-admin-go)** | **Go 后端 · 本仓库** | Go + Gin + GORM                                        |
| [ape-volo-admin-java](https://github.com/xianhc/ape-volo-admin-java) | Java 后端            | Java 25 + Spring Boot 4 + MyBatis-Plus                 |
| [ape-volo-web-js](https://github.com/xianhc/ape-volo-web-js)         | JavaScript 前端      | Vue 3 + JavaScript + Vite + Element Plus               |

## 贡献指南

欢迎提交问题、功能建议、文档改进和代码修复。

### Issue 规范

- 提交前先搜索已有 Issue，避免重复反馈。
- Bug 请提供运行环境、复现步骤、预期与实际结果，以及必要的错误信息。
- 功能建议请说明使用场景、要解决的问题和预期行为。
- 日志、截图和示例配置请移除密码、令牌、密钥等敏感信息。

### Pull Request 规范

- Fork 仓库后在自己的分支修改，一个 PR 聚焦一个问题或功能。
- 提交信息使用中文，格式为 `类型: 简述`；PR 描述包含改动目的、影响范围和验证结果。
- 遵循现有分层与命名约定；修改接口、配置或目录职责时，同步更新说明。
- 按改动执行[开发检查](#开发检查)，说明未验证的范围，不提交构建物、日志、本地数据库或临时配置。

## 交流与反馈

[官方网站](https://www.apevolo.com) · [GitHub](https://github.com/xianhc/ape-volo-admin-go) · [Gitee](https://gitee.com/xianhc/ape-volo-admin-go)

QQ 群：`839263566` · 微信：`apevolo`（备注“加群”）

如果项目对你有帮助，欢迎通过 Star 或[捐赠](https://www.apevolo.com/donate)支持持续维护。

## 版权与授权

**版权所有 © 2022–2026 xianhc。**

本项目采用 [Apache License 2.0](LICENSE)，使用、修改与分发请遵循许可证要求。

Apache License 2.0 允许在遵守许可证条款的前提下使用、修改、商业使用与再分发本项目。

再分发时须提供许可证副本；以源码形式分发衍生作品时，须保留与作品相关的原始版权及署名通知，修改过的文件须显著注明已作修改。完整义务以 [LICENSE](LICENSE) 为准。

商业合作咨询：[apevolo@gmail.com](mailto:apevolo@gmail.com)。
