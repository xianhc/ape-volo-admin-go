package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/mssola/user_agent"
	"go-apevolo/global/constants/cachePrefix"
	"go-apevolo/middleware/aop"
	"go-apevolo/model/permission"
	"go-apevolo/model/system"
	"go-apevolo/payloads/auth"
	"go-apevolo/payloads/response"
	"go-apevolo/payloads/vo"
	"go-apevolo/service"
	"go-apevolo/utils"
	"go-apevolo/utils/ext"
	"go-apevolo/utils/redis"
	"net/http"
	"strconv"
	"strings"
)

var rolePermissionService = service.ServiceGroupApp.PermissionServiceGroup.RolePermissionService
var settingService = service.ServiceGroupApp.SystemServiceGroup.SettingService
var userService = service.ServiceGroupApp.PermissionServiceGroup.UserService
var tokenBlacklistService = service.ServiceGroupApp.SystemServiceGroup.TokenBlacklistService

// PermissionHandler 权限拦截处理器
func PermissionHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		message, code := permissionHandler(c)
		if code == http.StatusOK {
			c.Next()
		} else {
			if code == http.StatusUnauthorized {
				response.Unauthorized(message, nil, c)
			} else if code == http.StatusForbidden {
				response.Forbidden(message, nil, c)
			} else if code == http.StatusBadRequest {
				response.Error(message, nil, c)
			}
			c.Abort()
		}
	}
}

func permissionHandler(c *gin.Context) (message string, httpCode int) {
	path := strings.ToLower(c.Request.URL.Path)
	method := strings.ToLower(c.Request.Method)
	token := utils.GetToken(c)
	userId := utils.GetId(c)
	token = ext.StringReplace(token, "Bearer ", "", -1)
	var loginUserInfo *auth.LoginUserInfo
	_ = redis.Get(cachePrefix.OnlineKey+utils.MD5(token), &loginUserInfo)
	if loginUserInfo == nil {
		//return "抱歉，您无权访问该接口", http.StatusUnauthorized
		isExist, err := tokenBlacklistService.DoesItExist(utils.MD5(token))
		if err != nil {
			return err.Error(), http.StatusBadRequest
		}
		if isExist {
			return "抱歉，您无权访问该接口", http.StatusUnauthorized
		}
		var user permission.User
		err = userService.QueryById(userId, &user)
		if err != nil {
			return err.Error(), http.StatusBadRequest
		}
		// 创建在线用户
		ip := utils.GetClientIP(c)
		ipAddress, _ := utils.SearchIpAddress(ip)
		ua := user_agent.New(c.Request.UserAgent())
		browserName, browserVersion := ua.Browser()
		deviceType := utils.GetDeviceType(ua.Platform(), ua.OS(), ua.Mobile())
		loginUserInfoNew := auth.LoginUserInfo{UserId: user.Id, Account: user.Username, NickName: user.NickName, DeptId: user.DeptId, IsAdmin: user.IsAdmin,
			DeptName: user.Dept.Name, Ip: ip, Address: ipAddress, LoginTime: ext.GetCurrentTime(), AccessToken: token, OperatingSystem: ua.Platform(), DeviceType: deviceType, BrowserName: browserName, Version: browserVersion}
		err = redis.Set(cachePrefix.OnlineKey+utils.MD5(token), loginUserInfoNew, ext.GetTimeDuration(3, ext.Hour))
		if err != nil {
			return err.Error(), http.StatusBadRequest
		}

	}
	newCacheConfig := aop.NewCacheConfig(cachePrefix.LoadSettingByName, ext.GetTimeDuration(30, ext.Minute), nil)

	setting := &system.Setting{}
	cacheSetting := aop.CacheAop(newCacheConfig, settingService.FindSettingByName, setting, "IsAdminNotAuthentication", setting)
	settingErr := cacheSetting()
	if settingErr == nil {
		isTrue, err1 := strconv.ParseBool(setting.Value)
		if err1 == nil {
			if isTrue {
				newCacheConfig := aop.NewCacheConfig(cachePrefix.UserInfoById, ext.GetTimeDuration(2, ext.Hour), nil)
				user := &permission.User{}
				cacheUser := aop.CacheAop(newCacheConfig, userService.QueryById, user, userId, user)
				userResult := cacheUser()
				if userResult == nil && user.IsAdmin {
					return "", http.StatusOK
				}
			}
		} else {
			return err1.(error).Error(), http.StatusBadRequest
		}
	} else {
		return settingErr.(error).Error(), http.StatusBadRequest
	}

	newCacheConfig = aop.NewCacheConfig(cachePrefix.UserPermissionUrls, ext.GetTimeDuration(2, ext.Hour), nil)
	urlAccessControlList := &[]vo.UrlAccessControl{}
	cachePermissionVos := aop.CacheAop(newCacheConfig, rolePermissionService.GetUrlAccessControl, urlAccessControlList, userId, urlAccessControlList)
	permissionVosResult := cachePermissionVos()
	if permissionVosResult != nil {
		return permissionVosResult.Error(), http.StatusBadRequest
	} else {
		for _, item := range *urlAccessControlList {
			if strings.ToLower(item.Url) == path && strings.ToLower(item.Method) == method {
				return "", http.StatusOK
			}
		}
	}
	return "抱歉，您访问权限等级不够", http.StatusForbidden
}
