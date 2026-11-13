package cachePrefix

const (
	// LoginCaptcha 验证码
	LoginCaptcha = "login:captcha:"

	// EmailCaptcha 邮箱验证码
	EmailCaptcha = "email:captcha:"

	// UserInfoById 用户信息
	UserInfoById = "identity:user:info:id:"

	// UserMenuById 用户菜单
	UserMenuById = "identity:user:menu:id:"

	// UserAuthCodes 用户权限标识
	UserAuthCodes = "identity:user:authCode:id:"

	// UserRoleCodes 用户角色码。
	//
	// 按用户保存去重后的非空角色码，角色或用户关联写入后失效。
	UserRoleCodes = "identity:user:roleCode:id:"

	// UserDataScopeById 用户数据权限快照
	UserDataScopeById = "identity:user:dataScope:id:"

	// LoadSettingByName 加载设置信息
	LoadSettingByName = "setting:name:"

	// LoadDictByName 按名称加载字典及明细
	LoadDictByName = "dict:name:"

	// TokenRevoked 已撤销的令牌
	TokenRevoked = "login:revoked:"

	// Threshold 登录失败阈值
	Threshold = "login:threshold:"

	// Attempts 登录失败次数
	Attempts = "login:attempts:"
)
