package cachePrefix

const (
	// OnlineKey 在线
	OnlineKey = "online:"

	// LoginCaptcha 验证码
	LoginCaptcha = "login:captcha:"

	// EmailCaptcha 邮箱验证码
	EmailCaptcha = "email:captcha:"

	// UserInfoById 用户信息
	UserInfoById = "user:info:id:"

	// UserMenuById 用户菜单
	UserMenuById = "user:menu:id:"

	// UserPermissionRoles 用户权限标识
	UserPermissionRoles = "user:permissionRole:id:"

	// UserPermissionUrls 用户权限Url
	UserPermissionUrls = "user:permissionUrl:id:"

	// LoadMenusByPId 加载菜单根据PID
	LoadMenusByPId = "menu:pid:"

	// LoadMenusById 加载菜单根据ID
	LoadMenusById = "menu:id:"

	// LoadSettingByName 加载设置信息
	LoadSettingByName = "setting:name:"

	// Threshold 登录失败阈值
	Threshold = "login:threshold:"

	// Attempts 登录失败次数
	Attempts = "login:attempts:"
)
