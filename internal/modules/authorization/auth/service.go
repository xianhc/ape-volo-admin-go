package auth

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"time"

	"github.com/xianhc/ape-volo-admin-go/internal/global"
	"github.com/xianhc/ape-volo-admin-go/internal/global/constants/cachePrefix"
	sessionmodule "github.com/xianhc/ape-volo-admin-go/internal/modules/authorization/session"
	rolemodule "github.com/xianhc/ape-volo-admin-go/internal/modules/identity/role"
	usermodule "github.com/xianhc/ape-volo-admin-go/internal/modules/identity/user"
	"github.com/xianhc/ape-volo-admin-go/internal/modules/message/queuedemail"
	"github.com/xianhc/ape-volo-admin-go/internal/payloads/auth"
	"github.com/xianhc/ape-volo-admin-go/internal/payloads/request"
	"github.com/xianhc/ape-volo-admin-go/internal/payloads/response"
	"github.com/xianhc/ape-volo-admin-go/internal/utils"
	"github.com/xianhc/ape-volo-admin-go/internal/utils/cache"
	captchastore "github.com/xianhc/ape-volo-admin-go/internal/utils/cache/captcha"
	"github.com/xianhc/ape-volo-admin-go/internal/utils/ext"
	"github.com/xianhc/ape-volo-admin-go/internal/utils/querycache"
	"go.uber.org/zap"
)

var store = captchastore.NewStore()

// ClientContext 表示认证业务需要的客户端环境。
//
// IP 必须由 Handler 通过受信代理规则解析，UserAgent 为原始浏览器标识。
type ClientContext struct {
	IP        string
	UserAgent string
}

// Service 实现验证码、登录、认证上下文和令牌会话业务。
//
// Service 不依赖 Gin，通过用户、角色、队列和会话模块协作。
type Service struct {
	users    usermodule.Service
	roles    rolemodule.Service
	queue    queuedemail.Service
	sessions sessionmodule.Service
}

// Captcha 生成并保存图形验证码。
//
// ShowCaptcha 根据客户端登录失败计数决定，验证码存储跟随缓存后端。
func (service *Service) Captcha(ctx context.Context, clientIP string) (response.Captcha, error) {
	cfg := global.Config.Captcha
	show := true
	if cfg.FailureThreshold > 0 {
		key := cachePrefix.Threshold + clientIP
		var failed int
		err := global.Cache.Get(ctx, key, &failed)
		if errors.Is(err, cache.ErrNotFound) {
			failed = 1
			_ = global.Cache.Set(ctx, key, 1, time.Duration(cfg.CounterTtlSeconds)*time.Second)
		}
		show = failed > cfg.FailureThreshold
	}
	code := captchastore.RandomCode(cfg.CodeLength)
	png, err := captchastore.RenderDigits(code, cfg.ImageWidth, cfg.ImageHeight, cfg.FontSize)
	if err != nil {
		return response.Captcha{}, errors.New("生成验证码失败")
	}
	id := captchastore.NewID()
	if err = store.Set(id, code); err != nil {
		return response.Captcha{}, errors.New("生成验证码失败")
	}
	return response.Captcha{CaptchaId: id, Img: "data:image/png;base64," + base64.StdEncoding.EncodeToString(png), ShowCaptcha: show}, nil
}

// Login 校验用户凭据并创建服务端会话。
//
// 登录成功返回脱敏用户、角色码、权限码和令牌对。
func (service *Service) Login(ctx context.Context, req request.LoginAuthUser, client ClientContext) (Login, error) {
	cfg := global.Config.Captcha
	thresholdKey := cachePrefix.Threshold + client.IP
	show := true
	if cfg.FailureThreshold > 0 {
		var failed int
		err := global.Cache.Get(ctx, thresholdKey, &failed)
		if errors.Is(err, cache.ErrNotFound) {
			failed = 1
			_ = global.Cache.Set(ctx, thresholdKey, 1, time.Duration(cfg.CounterTtlSeconds)*time.Second)
		}
		show = failed > cfg.FailureThreshold
	}
	attemptKey := cachePrefix.Attempts + client.IP + req.UserName
	var attempt *auth.LoginAttempt
	if global.Config.LoginFailedLimit.Enabled {
		if err := global.Cache.Get(ctx, attemptKey, &attempt); err != nil {
			if !errors.Is(err, cache.ErrNotFound) {
				return Login{}, err
			}
			attempt = &auth.LoginAttempt{Count: 1, LockUntil: time.Date(1, time.January, 1, 0, 0, 0, 0, time.Local)}
			_ = global.Cache.Set(ctx, attemptKey, attempt, time.Duration(global.Config.LoginFailedLimit.Lockout)*time.Second)
		}
		if attempt.IsLocked && time.Now().Before(attempt.LockUntil) {
			return Login{}, errors.New("账户已锁定，请稍后重试。解锁时间：" + attempt.LockUntil.Format("2006-01-02 15:04:05"))
		}
	}
	if !global.Config.System.IsQuickDebug && show && !store.Verify(req.CaptchaId, req.Captcha, true) {
		service.incrementThreshold(ctx, thresholdKey)
		return Login{}, errors.New("验证码输入错误!")
	}
	publicUser, passwordHash, err := service.users.AuthByName(ctx, req.UserName)
	if err != nil {
		service.incrementThreshold(ctx, thresholdKey)
		return Login{}, err
	}
	password, err := utils.Decrypt(req.Password)
	if err != nil {
		service.incrementThreshold(ctx, thresholdKey)
		return Login{}, errors.New("密码解密失败:" + err.Error())
	}
	if !utils.BcryptCheck(password, passwordHash) {
		service.incrementThreshold(ctx, thresholdKey)
		if attempt != nil {
			attempt.Count++
			if attempt.Count >= global.Config.LoginFailedLimit.MaxAttempts {
				attempt.IsLocked = true
				attempt.LockUntil = time.Now().Add(time.Duration(global.Config.LoginFailedLimit.Lockout) * time.Second)
			}
			_ = global.Cache.Set(ctx, attemptKey, attempt, time.Duration(global.Config.LoginFailedLimit.Lockout)*time.Second)
		}
		return Login{}, errors.New("密码错误")
	}
	if !publicUser.Enabled {
		service.incrementThreshold(ctx, thresholdKey)
		return Login{}, errors.New("用户未激活")
	}
	user, err := service.cachedUser(ctx, publicUser.Id)
	if err != nil {
		return Login{}, err
	}
	roleIds := roleIDs(user)
	codes, err := service.cachedRoleCodes(ctx, user.Id)
	if err != nil {
		return Login{}, err
	}
	authCodes := make([]string, 0)
	err = querycache.GetOrLoad(ctx, querycache.Key(cachePrefix.UserAuthCodes, user.Id), &authCodes, time.Hour, func() error { return service.roles.GetPermissionIdentifier(ctx, user.Id, roleIds, &authCodes) })
	if err != nil {
		return Login{}, err
	}
	ua := utils.ParseUserAgent(client.UserAgent)
	address, _ := utils.SearchIpAddress(client.IP)
	info := auth.LoginUserInfo{UserId: user.Id, Account: user.UserName, NickName: user.NickName, DeptId: user.DeptId, DeptName: user.Dept.Name, Ip: client.IP, Address: address, LoginTime: ext.GetCurrentTime(), OperatingSystem: ua.OperatingSystem, DeviceType: ua.DeviceType, BrowserName: ua.BrowserName, Version: ua.BrowserVersion}
	if cfg.FailureThreshold > 0 {
		_ = global.Cache.Del(ctx, thresholdKey)
	}
	token, err := service.issueSession(ctx, &info)
	if err != nil {
		return Login{}, err
	}
	return Login{User: user, RoleCodes: codes, AuthCodes: authCodes, Token: *token}, nil
}

// Info 返回当前用户的脱敏认证上下文。
//
// 用户与权限码均使用现有查询缓存。
func (service *Service) Info(ctx context.Context, userId int64) (AuthContext, error) {
	user, err := service.cachedUser(ctx, userId)
	if err != nil {
		return AuthContext{}, err
	}
	roleIds := roleIDs(user)
	codes, err := service.cachedRoleCodes(ctx, user.Id)
	if err != nil {
		return AuthContext{}, err
	}
	authCodes := make([]string, 0)
	err = querycache.GetOrLoad(ctx, querycache.Key(cachePrefix.UserAuthCodes, user.Id), &authCodes, time.Hour, func() error { return service.roles.GetPermissionIdentifier(ctx, user.Id, roleIds, &authCodes) })
	if err != nil {
		return AuthContext{}, err
	}
	return AuthContext{User: user, RoleCodes: codes, AuthCodes: authCodes}, nil
}

// Logout 根据访问令牌或刷新令牌撤销服务端会话。
//
// 空令牌或无法解析令牌按幂等登出处理。
func (service *Service) Logout(ctx context.Context, token string) error {
	if token == "" {
		return nil
	}
	jwt := utils.NewJwt()
	for _, use := range []string{utils.TokenUseAccess, utils.TokenUseRefresh} {
		claims, err := jwt.ParseToken(token, use, false)
		if err == nil && claims.Sid != "" {
			return service.sessions.RevokeSession(ctx, claims.Sid)
		}
	}
	return nil
}

// ResetEmailCode 发送邮箱验证码。
//
// 验证码邮件使用固定模板 EmailVerificationCode。
func (service *Service) ResetEmailCode(ctx context.Context, email string) error {
	return service.queue.ResetEmail(ctx, email, "EmailVerificationCode")
}

// Refresh 校验刷新令牌并原子轮换令牌对。
//
// 无效、过期或重放令牌返回 nil，调用方要求重新登录。
func (service *Service) Refresh(ctx context.Context, token string) (*utils.JwtToken, error) {
	if token == "" {
		return nil, nil
	}
	jwt := utils.NewJwt()
	claims, err := jwt.ParseToken(token, utils.TokenUseRefresh, true)
	if err != nil {
		return nil, err
	}
	session, err := service.sessions.FindActive(ctx, claims.Sid)
	if err != nil || session == nil {
		return nil, err
	}
	if session.UserId != utils.UserID(claims) {
		return nil, nil
	}
	previous := hashToken(token)
	if session.RefreshTokenHash != previous {
		_ = service.sessions.RevokeSession(ctx, session.Id)
		return nil, nil
	}
	info, err := session.LoginUserInfo()
	if err != nil {
		return nil, err
	}
	accessID := sessionmodule.NewID()
	pair, err := jwt.IssueTokenPair(session.Id, session.UserId, info.Account, info.DeptId, accessID, session.ExpiresAt)
	if err != nil {
		return nil, err
	}
	ok, err := service.sessions.Rotate(ctx, session.Id, previous, hashToken(pair.RefreshToken), accessID)
	if err != nil {
		return nil, err
	}
	if !ok {
		_ = service.sessions.RevokeSession(ctx, session.Id)
		return nil, nil
	}
	return pair, nil
}

func (service *Service) cachedUser(ctx context.Context, userId int64) (usermodule.UserResponse, error) {
	var user usermodule.UserResponse
	err := querycache.GetOrLoad(ctx, querycache.Key(cachePrefix.UserInfoById, userId), &user, time.Hour, func() error {
		loaded, _, loadErr := service.users.AuthByID(ctx, userId)
		user = loaded
		return loadErr
	})
	return user, err
}
func (service *Service) issueSession(ctx context.Context, info *auth.LoginUserInfo) (*utils.JwtToken, error) {
	now := time.Now()
	session := &sessionmodule.Session{Id: sessionmodule.NewID(), UserId: info.UserId, AccessTokenId: sessionmodule.NewID(), CreatedAt: now, ExpiresAt: now.Add(time.Duration(global.Config.JwtAuthOptions.RefreshTokenLifetimeMs) * time.Millisecond)}
	info.AccessToken = session.Id
	payload, err := json.Marshal(info)
	if err != nil {
		return nil, err
	}
	session.LoginInfo = string(payload)
	pair, err := utils.NewJwt().IssueTokenPair(session.Id, info.UserId, info.Account, info.DeptId, session.AccessTokenId, session.ExpiresAt)
	if err != nil {
		return nil, err
	}
	session.RefreshTokenHash = hashToken(pair.RefreshToken)
	ok, err := service.sessions.Create(ctx, session)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.New("创建会话失败")
	}
	return pair, nil
}
func (service *Service) incrementThreshold(ctx context.Context, key string) {
	if global.Config.Captcha.FailureThreshold > 0 {
		_, _ = global.Cache.Incr(ctx, key, 1)
	}
}
func (service *Service) cachedRoleCodes(ctx context.Context, userId int64) ([]string, error) {
	codes := make([]string, 0)
	err := querycache.GetOrLoad(ctx, querycache.Key(cachePrefix.UserRoleCodes, userId), &codes, time.Hour, func() error {
		var err error
		codes, err = service.roles.RoleCodes(ctx, userId)
		return err
	})
	return codes, err
}

func roleIDs(user usermodule.UserResponse) []int64 {
	ids := make([]int64, 0, len(user.Roles))
	for _, role := range user.Roles {
		ids = append(ids, role.Id)
	}
	return ids
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

var _ = zap.Error
