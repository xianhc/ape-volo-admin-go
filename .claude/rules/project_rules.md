# 项目规则

本项目的协作约定统一写在仓库根目录的 [AGENTS.md](../../AGENTS.md)。

改代码前请先读：

1. `AGENTS.md`：修改范围、关键约束、验证矩阵
2. `README.md`：项目介绍、启动方式和系统架构

文档、注释和说明只描述本项目当前实现，不保留其他项目或语言实现的对比及对齐记录。

要点摘录（以 `AGENTS.md` 为准）：

- 只处理当前任务，不顺手重构、不批量还原工作区
- 私有路由用 `routeauth.Group` 注册并在同一行声明权限；新增权限码同步菜单种子与角色授权
- 缓存走 `global.Cache`，限流走 `utils/ratelimit`，IP 取 `utils.GetClientIP(c)`
- 审计字段由 ORM 回调填充，`creator_id` 可空，业务代码不手写
- 敏感用户、角色、菜单操作还需 `admin` 角色，缺少角色上下文时 fail closed
- 当前接口面为 105 条，不增加 GET download、兼容查询路由或 `/api/health`
- 提交信息用中文，格式 `类型: 简述`
