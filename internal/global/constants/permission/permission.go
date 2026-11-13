// Package permission 定义系统权限码常量。
//
// 权限码格式为 `模块:资源:动作`，例如 `sys:user:list`。
//
// 路由通过 routeauth.Group 注册时声明权限，权限码与菜单种子的 AuthCode 一致。
package permission

const (
	// SysUserList 用户-查询
	SysUserList = "sys:user:list"
	// SysUserAdd 用户-新增
	SysUserAdd = "sys:user:add"
	// SysUserEdit 用户-编辑
	SysUserEdit = "sys:user:edit"
	// SysUserDelete 用户-删除
	SysUserDelete = "sys:user:del"
	// SysUserDownload 用户-导出
	SysUserDownload = "sys:user:download"

	// SysDeptList 部门-查询
	SysDeptList = "sys:dept:list"
	// SysDeptAdd 部门-新增
	SysDeptAdd = "sys:dept:add"
	// SysDeptEdit 部门-编辑
	SysDeptEdit = "sys:dept:edit"
	// SysDeptDelete 部门-删除
	SysDeptDelete = "sys:dept:del"
	// SysDeptDownload 部门-导出
	SysDeptDownload = "sys:dept:download"

	// SysPositionList 岗位-查询
	SysPositionList = "sys:position:list"
	// SysPositionAdd 岗位-新增
	SysPositionAdd = "sys:position:add"
	// SysPositionEdit 岗位-编辑
	SysPositionEdit = "sys:position:edit"
	// SysPositionDelete 岗位-删除
	SysPositionDelete = "sys:position:del"
	// SysPositionDownload 岗位-导出
	SysPositionDownload = "sys:position:download"

	// SysMenuList 菜单-查询
	SysMenuList = "sys:menu:list"
	// SysMenuAdd 菜单-新增
	SysMenuAdd = "sys:menu:add"
	// SysMenuEdit 菜单-编辑
	SysMenuEdit = "sys:menu:edit"
	// SysMenuDelete 菜单-删除
	SysMenuDelete = "sys:menu:del"
	// SysMenuDownload 菜单-导出
	SysMenuDownload = "sys:menu:download"

	// SysRoleList 角色-查询
	SysRoleList = "sys:role:list"
	// SysRoleAdd 角色-新增
	SysRoleAdd = "sys:role:add"
	// SysRoleEdit 角色-编辑（含菜单授权）
	SysRoleEdit = "sys:role:edit"
	// SysRoleDelete 角色-删除
	SysRoleDelete = "sys:role:del"
	// SysRoleDownload 角色-导出
	SysRoleDownload = "sys:role:download"

	// SysDictList 字典-查询（含字典详情）
	SysDictList = "sys:dict:list"
	// SysDictAdd 字典-新增
	SysDictAdd = "sys:dict:add"
	// SysDictEdit 字典-编辑
	SysDictEdit = "sys:dict:edit"
	// SysDictDelete 字典-删除
	SysDictDelete = "sys:dict:del"
	// SysDictDownload 字典-导出
	SysDictDownload = "sys:dict:download"

	// SysSettingList 参数-查询
	SysSettingList = "sys:setting:list"
	// SysSettingAdd 参数-新增
	SysSettingAdd = "sys:setting:add"
	// SysSettingEdit 参数-编辑
	SysSettingEdit = "sys:setting:edit"
	// SysSettingDelete 参数-删除
	SysSettingDelete = "sys:setting:del"
	// SysSettingDownload 参数-导出
	SysSettingDownload = "sys:setting:download"

	// SysAppSecretList 应用密钥-查询
	SysAppSecretList = "sys:appSecret:list"
	// SysAppSecretAdd 应用密钥-新增
	SysAppSecretAdd = "sys:appSecret:add"
	// SysAppSecretEdit 应用密钥-编辑
	SysAppSecretEdit = "sys:appSecret:edit"
	// SysAppSecretDelete 应用密钥-删除
	SysAppSecretDelete = "sys:appSecret:del"
	// SysAppSecretDownload 应用密钥-导出
	SysAppSecretDownload = "sys:appSecret:download"

	// SysStorageList 文件记录-查询
	SysStorageList = "sys:storage:list"
	// SysStorageAdd 文件记录-上传
	SysStorageAdd = "sys:storage:add"
	// SysStorageEdit 文件记录-编辑
	SysStorageEdit = "sys:storage:edit"
	// SysStorageDelete 文件记录-删除
	SysStorageDelete = "sys:storage:del"
	// SysStorageDownload 文件记录-下载
	SysStorageDownload = "sys:storage:download"

	// SysJobList 作业-查询
	//
	// 分页列表和全部作业查询共用此权限。
	SysJobList = "sys:job:list"
	// SysJobAdd 作业-新增
	//
	// 创建配置通过作业管理入口同步运行时。
	SysJobAdd = "sys:job:add"
	// SysJobEdit 作业-编辑（含执行/暂停/恢复）
	//
	// 配置修改与运行时控制共用此权限。
	SysJobEdit = "sys:job:edit"
	// SysJobDelete 作业-删除
	//
	// 删除动作沿用菜单种子的 del 权限码。
	SysJobDelete = "sys:job:del"
	// SysJobDownload 作业-导出
	//
	// 导出只通过 POST download 入口提供。
	SysJobDownload = "sys:job:download"

	// SysScheduledJobLogList 任务执行日志-查询
	SysScheduledJobLogList = "log:job:list"

	// LogOperateList 操作日志-查询
	LogOperateList = "log:operate:list"
	// LogExceptionList 异常日志-查询
	LogExceptionList = "log:exception:list"

	// MonitorOnlineList 在线用户-查询（含强制下线）
	MonitorOnlineList = "sys:online:list"
	// MonitorServerList 服务器监控-查询
	MonitorServerList = "sys:server:list"

	// MsgAccountList 邮箱账户-查询
	MsgAccountList = "sys:emailAccount:list"
	// MsgAccountAdd 邮箱账户-新增
	MsgAccountAdd = "sys:emailAccount:add"
	// MsgAccountEdit 邮箱账户-编辑
	MsgAccountEdit = "sys:emailAccount:edit"
	// MsgAccountDelete 邮箱账户-删除
	MsgAccountDelete = "sys:emailAccount:del"
	// MsgAccountDownload 邮箱账户-导出
	MsgAccountDownload = "sys:emailAccount:download"

	// MsgTemplateList 邮件模板-查询
	MsgTemplateList = "sys:emailTemplate:list"
	// MsgTemplateAdd 邮件模板-新增
	MsgTemplateAdd = "sys:emailTemplate:add"
	// MsgTemplateEdit 邮件模板-编辑
	MsgTemplateEdit = "sys:emailTemplate:edit"
	// MsgTemplateDelete 邮件模板-删除
	MsgTemplateDelete = "sys:emailTemplate:del"

	// MsgQueuedList 邮件队列-查询
	MsgQueuedList = "sys:queuedEmail:list"
	// MsgQueuedAdd 邮件队列-新增
	MsgQueuedAdd = "sys:queuedEmail:add"
	// MsgQueuedEdit 邮件队列-编辑
	MsgQueuedEdit = "sys:queuedEmail:edit"
	// MsgQueuedDelete 邮件队列-删除
	MsgQueuedDelete = "sys:queuedEmail:del"
)
