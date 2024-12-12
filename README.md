#### 📚系统说明

- 基于 Gin、GORM、Vue 2.X、RBAC、前后端分离的开箱则用的企业级中后台**权限管理系统**
- 无业务逻辑代码入侵，适用于任何 Go 应用程序。
- 预览体验：  [https://www.apevolo.com](https://apevolo.com)
- 开发文档：  [http://doc.apevolo.com](http://doc.apevolo.com)
- 账号密码： `apevolo / 123456`

#### 💒代码仓库(api)
- go 版本(Github) <a href="https://github.com/xianhc/ape-volo-admin-go" target="_blank">https://github.com/xianhc/ape-volo-admin-go</a>
- go 版本(Gitee) <a href="https://gitee.com/xianhc/ape-volo-admin-go" target="_blank">https://gitee.com/xianhc/ape-volo-admin-go</a>
<br><br>
- net 版本(Github) <a href="https://github.com/xianhc/ape-volo-admin" target="_blank">https://github.com/xianhc/ape-volo-admin</a>
- net 版本(Gitee) <a href="https://gitee.com/xianhc/ape-volo-admin" target="_blank">https://gitee.com/xianhc/ape-volo-admin</a>

#### 💒代码仓库(web)
- vue2.x 版本(Github) <a href="https://github.com/xianhc/ape-volo-web" target="_blank">https://github.com/xianhc/ape-volo-web</a>
- vue2.x 版本(Gitee) <a href="https://gitee.com/xianhc/ape-volo-web" target="_blank">https://gitee.com/xianhc/ape-volo-web</a>

#### ⚙️模块说明

| # |  项目文件                    | 说明|
|---| -------------------------------|-------------------------------|
| 1 | api | 接口交互 |
| 2 | config | 配置文件 |
| 3 | core | 核心功能 |
| 4 | docs | 接口文档 |
| 5 | global | 全局对象 |
| 6 | initialize | 初始化 |
| 7 | job | 任务调度作业|
| 8 | middleware | 中间件 |
| 9 | model | 实体对象 |
| 10 | payloads | 请求、响应结构体 |
| 11 | resource | 资源文件 |
| 12 | router | 路由 |
| 13 | service | 业务实现 |
| 14 | uploads | 文件上传路径 |
| 15 | utils | 工具包 |

#### 🚀系统特性
- 使用 Gin 搭建基础restful风格API
- 使用 GORM 简化与数据库的交互
- 使用 Swagger UI 自动生成 WebAPI 说明文档
- 使用 Zap 日志组件
- 使用 Cron 封装任务调度中心功能
- 封装异常过滤器  实现统一记录系统异常日志
- 封装审计过滤器  实现统一记录接口请求日志
- 封装缓存拦截器  实现对业务方法结果缓存处理
- 封装事务拦截器  实现对业务方法操作数据库事务处理
- 封装系统config.yaml配置Configs类
- 自定义权限拦截处理器实现鉴权
- 支持多种主流数据库(MySql、SqlServer、Sqlite、Oracle、postgresql)；
- 支持 CORS 跨域配置
- 支持数据字典、自定义设置处理
- 支持接口限流 避免恶意请求攻击
- <span style="color:red;">[X]</span>支持多租户 ID隔离 、 库隔离
- <span style="color:red;">[X]</span>支持数据权限 (全部、本人、本部门、本部门及以下、自定义)


#### ⚡快速开始

##### 环境
推荐使用 `JetBrains GoLand`、`WebStorm`<br/>
GoLand版本 >= v1.23

##### 运行

1. 下载项目，安装go依赖包。然后启动项目
2. 系统便会自动创建数据库表并初始化相关基础数据
3. 系统默认使用`Sqlite`数据库
4. 🚨🚨🚨系统当前高强度依赖Redis缓存，必须启动Redis服务并确保连接正确，项目才能正常运行

#### ⭐️支持作者
如果觉得框架不错，或者已经在使用了，希望你可以去 <a target="_blank" href="https://github.com/xianhc/ape-volo-admin-go">Github</a> 或者
<a target="_blank" href="https://gitee.com/xianhc/ape-volo-admin-go">Gitee</a> 帮我点个 ⭐ Star，这将是对我极大的鼓励与支持。

#### 🙋反馈交流
##### QQ群：1015661568
| QQ 群 |
|  :---:  |
| <img width="150" src="https://www.apevolo.com/uploads/file/wechat/20230723172504.jpg">

##### 微信群
| 微信 |
|  :---:  |
| <img width="150" src="https://www.apevolo.com/uploads/file/wechat/20230723172451.jpg">

添加微信，备注"go"

#### 🤟捐赠
如果你觉得这个项目对你有帮助，你可以请作者喝饮料 :tropical_drink: [点我](http://doc.apevolo.com/donate/)

#### 🤝致谢
![JetBrains Logo (Main) logo](https://resources.jetbrains.com/storage/products/company/brand/logos/jb_beam.svg)


#### 💡其他
<a target="_blank" href="https://github.com/xianhc/ape-volo-admin">ape-volo-admin</a> 是一个主要基于 .NET Core 的开源框架<br>
同时我也使用 Go 语言开发了一个功能完全复刻的版本 <a target="_blank" href="https://github.com/xianhc/ape-volo-admin-go">ape-volo-admin-go</a><br>
尽管当前 Go 版本的功能尚未完全与 .NET Core 版本对等，但我正在逐步完善<br>
目标是确保 Go 版本具备与 .NET Core 版本完全一致的功能