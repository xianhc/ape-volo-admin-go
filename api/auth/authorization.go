package auth

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/mojocn/base64Captcha"
	"github.com/mssola/user_agent"
	"github.com/redis/go-redis/v9"
	"go-apevolo/global"
	"go-apevolo/global/constants/cachePrefix"
	"go-apevolo/middleware/aop"
	"go-apevolo/model/permission"
	"go-apevolo/payloads/auth"
	"go-apevolo/payloads/request"
	"go-apevolo/payloads/response"
	"go-apevolo/utils"
	"go-apevolo/utils/ext"
	_redis "go-apevolo/utils/redis"
	"go-apevolo/utils/redis/captcha"
	"go.uber.org/zap"
	"net/http"
	"time"
)

// 当开启多服务器部署时，替换下面的配置，使用redis共享存储验证码
var store = captcha.NewDefaultRedisStore()

//var store = base64Captcha.DefaultMemStore

type AuthorizationApi struct{}

// Captcha
// @Tags      Auth
// @Summary   生成验证码
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Success   200  {object}  response.Captcha  "生成验证码,返回包括随机数id,base64"
// @Router    /auth/captcha [get]
func (b *AuthorizationApi) Captcha(c *gin.Context) {
	showCaptcha := true
	threshold := global.Config.Captcha.Threshold
	if threshold > 0 {
		timeOut := global.Config.Captcha.TimeOut
		key := cachePrefix.Threshold + c.ClientIP()
		failedThreshold, err := global.Redis.Get(context.Background(), key).Int()
		if errors.Is(err, redis.Nil) {
			_ = _redis.Set(key, 1, time.Second*time.Duration(timeOut))
		}

		if threshold >= failedThreshold {
			showCaptcha = false
		}
	}

	// 生成默认数字的driver
	driver := base64Captcha.NewDriverDigit(global.Config.Captcha.ImgHeight, global.Config.Captcha.ImgWidth, global.Config.Captcha.KeyLength, 0.3, 30)
	cp := base64Captcha.NewCaptcha(driver, store.UseWithCtx(c))
	id, b64s, err := cp.Generate()
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	c.JSON(http.StatusOK, response.Captcha{
		CaptchaId:   id,
		Img:         b64s,
		ShowCaptcha: showCaptcha,
	})
}

// Login
// @Tags     Auth
// @Summary  用户登录
// @Produce   application/json
// @Param    data  body  request.LoginAuthUser true  "用户名, 密码, 验证码 随机数"
// @Success  200   {object}  response.Login  "返回包括用户信息,token,过期时间"
// @Router   /auth/login [post]
func (b *AuthorizationApi) Login(c *gin.Context) {
	var l request.LoginAuthUser
	err := c.ShouldBindJSON(&l)

	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	err = utils.VerifyData(l)
	if err != nil {
		response.Error("", utils.GetVerifyErr(err), c)
		return
	}

	loginFailedLimitEnabled := global.Config.LoginFailedLimit.Enabled
	var attempsCacheKey = cachePrefix.Attempts + c.ClientIP() + l.Username

	var loginAttempt *auth.LoginAttempt
	if loginFailedLimitEnabled {
		err = _redis.Get(attempsCacheKey, &loginAttempt)
		if err != nil {
			if errors.Is(err, _redis.Nil) {
				loginAttempt = &auth.LoginAttempt{}
				loginAttempt.Count = 1
				loginAttempt.IsLocked = false
				loginAttempt.LockUntil = time.Date(1, time.January, 1, 0, 0, 0, 0, time.Local)
				_ = _redis.Set(attempsCacheKey, loginAttempt, time.Second*time.Duration(global.Config.LoginFailedLimit.Lockout))
			} else {
				response.Error(err.Error(), nil, c)
				return
			}
		}
		currentTime := time.Now()
		if loginAttempt.IsLocked && currentTime.Before(loginAttempt.LockUntil) {
			// 可以实施账户锁定时，通过邮件或短信通知用户。
			// 可以实施账户锁定后要求管理员手动解锁
			response.Error("账户已锁定，请稍后重试。解锁时间："+loginAttempt.LockUntil.Format("2006-01-02 15:04:05"), nil, c)
			return
		}
	}

	showCaptcha := true
	thresholdKey := cachePrefix.Threshold + c.ClientIP()
	threshold := global.Config.Captcha.Threshold
	if threshold > 0 {

		timeOut := global.Config.Captcha.TimeOut
		failedThreshold, err := global.Redis.Get(context.Background(), thresholdKey).Int()
		if err != nil {
			if errors.Is(err, redis.Nil) {
				_ = _redis.Set(thresholdKey, 1, time.Second*time.Duration(timeOut))
			}
		}
		if threshold >= failedThreshold {
			showCaptcha = false
		}
	}

	if !global.Config.System.IsQuickDebug && showCaptcha {
		if !store.Verify(l.CaptchaId, l.Captcha, true) {
			if threshold > 0 {
				_ = global.Redis.Incr(context.Background(), thresholdKey)
			}
			response.Error("验证码输入错误!", nil, c)
			return
		}
	}

	var user permission.User
	err = userService.QueryByName(l.Username, &user)
	if err != nil {
		if threshold > 0 {
			_ = global.Redis.Incr(context.Background(), thresholdKey)
		}
		response.Error(err.Error(), nil, c)
		return
	}
	password, err := utils.Decrypt(l.Password)
	if err != nil {
		if threshold > 0 {
			_ = global.Redis.Incr(context.Background(), thresholdKey)
		}
		response.Error("密码解密失败:"+err.Error(), nil, c)
		return
	}
	if ok := utils.BcryptCheck(password, user.Password); !ok {
		if threshold > 0 {
			_ = global.Redis.Incr(context.Background(), thresholdKey)
		}

		if loginFailedLimitEnabled && loginAttempt != nil {
			loginAttempt.Count++
			if loginAttempt.Count >= global.Config.LoginFailedLimit.MaxAttempts {
				loginAttempt.IsLocked = true
				currentTime := time.Now()
				loginAttempt.LockUntil = currentTime.Add(time.Duration(global.Config.LoginFailedLimit.Lockout) * time.Second)
			}
			_ = _redis.Set(attempsCacheKey, loginAttempt, time.Second*time.Duration(global.Config.LoginFailedLimit.Lockout))
		}
		response.Error("密码错误", nil, c)
		return
	}
	if !user.Enabled {
		if threshold > 0 {
			_ = global.Redis.Incr(context.Background(), thresholdKey)
		}
		response.Error("用户未激活", nil, c)
		return
	}

	userNet := &permission.User{}
	newCacheConfig := aop.NewCacheConfig(cachePrefix.UserInfoById, ext.GetTimeDuration(2, ext.Hour), nil)
	cacheUserById := aop.CacheAop(newCacheConfig, userService.QueryById, userNet, user.Id, userNet)
	err = cacheUserById()
	if err != nil {
		if threshold > 0 {
			_ = global.Redis.Incr(context.Background(), thresholdKey)
		}
		response.Error(err.Error(), nil, c)
		return
	}

	var roleIds []int64

	for i := range userNet.Roles {
		roleIds = append(roleIds, userNet.Roles[i].Id)
	}
	permissionIdentifierList := &[]string{}
	newCacheConfig = aop.NewCacheConfig(cachePrefix.UserPermissionRoles, ext.GetTimeDuration(2, ext.Hour), nil)
	cachePermissionIdentifierList := aop.CacheAop(newCacheConfig, rolePermissionService.GetPermissionIdentifier, permissionIdentifierList, userNet.Id, roleIds, permissionIdentifierList)
	err = cachePermissionIdentifierList()
	if err != nil {
		if threshold > 0 {
			_ = global.Redis.Incr(context.Background(), thresholdKey)
		}
		response.Error(err.Error(), nil, c)
		return
	}

	ip := utils.GetClientIP(c)
	jwt := &utils.JwtAuthOptions{SigningKey: []byte(global.Config.JwtAuthOptions.SecurityKey)}
	claims := request.Claims{
		Jti:  user.Id,
		Name: user.Username,
		//Iat:  time.Now().Format("2006-01-02 15:04:05"),
		Iat: time.Now().UnixNano() / int64(time.Millisecond),
		Ip:  ip,
	}
	token, err := jwt.IssuedToken(claims)
	if err != nil {
		if threshold > 0 {
			_ = global.Redis.Incr(context.Background(), thresholdKey)
		}
		response.Error("创建token失败", nil, c)
		return
	}

	// 创建在线用户
	ipAddress, _ := utils.SearchIpAddress(ip)
	ua := user_agent.New(c.Request.UserAgent())
	browserName, browserVersion := ua.Browser()
	deviceType := utils.GetDeviceType(ua.Platform(), ua.OS(), ua.Mobile())
	loginUserInfo := auth.LoginUserInfo{UserId: userNet.Id, Account: userNet.Username, NickName: userNet.NickName, DeptId: userNet.DeptId,
		DeptName: userNet.Dept.Name, Ip: ip, Address: ipAddress, LoginTime: ext.GetCurrentTime(), AccessToken: token, OperatingSystem: ua.Platform(), DeviceType: deviceType, BrowserName: browserName, Version: browserVersion}
	loginUserInfo.IsAdmin = userNet.IsAdmin
	err = _redis.Set(cachePrefix.OnlineKey+utils.MD5(token), loginUserInfo, time.Hour*time.Duration(3))
	if err != nil {
		if threshold > 0 {
			_ = global.Redis.Incr(context.Background(), thresholdKey)
		}
		global.Logger.Error(err.Error(), zap.Error(err))
		response.Error(err.Error(), nil, c)
		return
	}
	c.JSON(http.StatusOK, response.Login{
		User:  response.JwtUser{User: *userNet, Role: *permissionIdentifierList, DataScope: []string{}},
		Token: response.JwtToken{AccessToken: token, Expires: global.Config.JwtAuthOptions.Expires * 3600, TokenType: "Bearer", RefreshToken: "", RefreshTokenExpires: global.Config.JwtAuthOptions.RefreshTokenExpires * 3600},
	})
}

// GetInfo
// @Tags     Auth
// @Summary  获取用户信息
// @Produce   application/json
// @Success  200   {object}  response.JwtUser  "返回包括用户信息,token,过期时间"
// @Router   /auth/info [post]
func (b *AuthorizationApi) GetInfo(c *gin.Context) {
	userNet := &permission.User{}
	newCacheConfig := aop.NewCacheConfig(cachePrefix.UserInfoById, ext.GetTimeDuration(2, ext.Hour), nil)
	cacheUserById := aop.CacheAop(newCacheConfig, userService.QueryById, userNet, utils.GetId(c), userNet)
	err := cacheUserById()
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}

	var roleIds []int64

	for i := range userNet.Roles {
		roleIds = append(roleIds, userNet.Roles[i].Id)
	}

	permissionIdentifierList := &[]string{}
	newCacheConfig = aop.NewCacheConfig(cachePrefix.UserPermissionRoles, ext.GetTimeDuration(2, ext.Hour), nil)
	cachePermissionIdentifierList := aop.CacheAop(newCacheConfig, rolePermissionService.GetPermissionIdentifier, permissionIdentifierList, userNet.Id, roleIds, permissionIdentifierList)
	err = cachePermissionIdentifierList()
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}

	jwtUser := response.JwtUser{User: *userNet, Role: *permissionIdentifierList, DataScope: []string{}}
	c.JSON(http.StatusOK, jwtUser)
}

// Logout
// @Tags     Auth
// @Summary  用户登出
// @Produce   application/json
// @Success 200 {object} response.ActionResult "登出成功"
// @Router   /auth/logout [delete]
func (b *AuthorizationApi) Logout(c *gin.Context) {
	token := utils.GetToken(c)
	if token != "" {
		_, err := global.Redis.Del(context.Background(), utils.MD5(token)).Result()
		if err != nil {
			response.Success(err.Error(), c)
		}
	}
	response.Success("", c)
}

// ResetEmail
// @Tags      Auth
// @Summary   获取邮箱验证码
// @accept    application/json
// @Produce   application/json
// @Success   200 {object} response.ActionResult "发送成功"
// @Router    /auth/code/reset/email [post]
func (b *AuthorizationApi) ResetEmail(c *gin.Context) {
	email := c.Query("email")
	if email == "" {
		response.Error("email is null", nil, c)
		return
	}
	validate := validator.New()
	err := validate.Var(email, "email")
	if err != nil {
		response.Error(err.Error(), nil, c)
	}
	err = queuedService.ResetEmail(email, "EmailVerificationCode")
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	response.Success("", c)
}

// RefreshToken
// @Tags      Auth
// @Summary   刷新Token，以旧换新
// @accept    application/json
// @Produce   application/json
// @Param data body request.RefreshToken true "token值"
// @Success  200   {object}  response.JwtToken  "返回包括用户信息,token,过期时间"
// @Router    /auth/refreshToken [post]
func (b *AuthorizationApi) RefreshToken(c *gin.Context) {
	var refreshToken request.RefreshToken
	err := c.ShouldBindJSON(&refreshToken)

	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	err = utils.VerifyData(refreshToken)
	if err != nil {
		response.Error("", utils.GetVerifyErr(err), c)
		return
	}
	if refreshToken.Token != "" {
		isExist, err := tokenBlacklistService.DoesItExist(utils.MD5(refreshToken.Token))
		if err != nil {
			response.Error(err.Error(), nil, c)
			return
		}
		if !isExist {
			j := utils.NewJwt()
			claims, err := j.ReadJwtToken(refreshToken.Token)
			if err != nil {
				response.Error(err.Error(), nil, c)
				return
			}
			currentTime := time.Now()
			loginTime := time.Unix(claims.Iat/1000, (claims.Iat%1000)*int64(time.Millisecond))
			refreshTime := loginTime.Add(10000 * time.Second)
			if currentTime.Before(refreshTime) {
				var userNet permission.User
				err := userService.QueryById(claims.Jti, &userNet)
				if err != nil {
					response.Error("", utils.GetVerifyErr(err), c)
					return
				}
				if userNet.UpdateTime == nil || userNet.UpdateTime.After(loginTime) {
					ip := utils.GetClientIP(c)
					jwt := &utils.JwtAuthOptions{SigningKey: []byte(global.Config.JwtAuthOptions.SecurityKey)}
					claims := request.Claims{
						Jti:  userNet.Id,
						Name: userNet.Username,
						//Iat:  time.Now().Format("2006-01-02 15:04:05"),
						Iat: time.Now().UnixNano() / int64(time.Millisecond),
						Ip:  ip,
					}
					token, err := jwt.IssuedToken(claims)
					if err != nil {
						response.Error("创建token失败", nil, c)
						return
					}

					// 创建在线用户
					ipAddress, _ := utils.SearchIpAddress(ip)
					ua := user_agent.New(c.Request.UserAgent())
					browserName, browserVersion := ua.Browser()
					deviceType := utils.GetDeviceType(ua.Platform(), ua.OS(), ua.Mobile())
					loginUserInfo := auth.LoginUserInfo{UserId: userNet.Id, Account: userNet.Username, NickName: userNet.NickName, DeptId: userNet.DeptId,
						DeptName: userNet.Dept.Name, Ip: ip, Address: ipAddress, LoginTime: ext.GetCurrentTime(), AccessToken: token, OperatingSystem: ua.Platform(), DeviceType: deviceType, BrowserName: browserName, Version: browserVersion}
					loginUserInfo.IsAdmin = userNet.IsAdmin
					err = _redis.Set(cachePrefix.OnlineKey+utils.MD5(token), loginUserInfo, time.Hour*time.Duration(3))
					if err != nil {
						global.Logger.Error(err.Error(), zap.Error(err))
						response.Error(err.Error(), nil, c)
						return
					}
					jwtToken := response.JwtToken{RefreshToken: token, TokenType: "Bearer", Expires: global.Config.JwtAuthOptions.Expires * 3600}
					c.JSON(http.StatusOK, jwtToken)
					return
				}
			}
		}
	}
	response.Error("token验证失败，请重新登录！", nil, c)
}
