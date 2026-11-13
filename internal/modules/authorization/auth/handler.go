package auth

import (
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/xianhc/ape-volo-admin-go/internal/payloads/request"
	"github.com/xianhc/ape-volo-admin-go/internal/payloads/response"
	"github.com/xianhc/ape-volo-admin-go/internal/utils"
	"github.com/xianhc/ape-volo-admin-go/internal/utils/ext"
)

// Handler 处理验证码、登录、认证上下文和令牌会话 HTTP 请求。
//
// Handler 只负责绑定、校验、客户端环境提取、调用 Service 和写入响应。
type Handler struct{ service Service }

// Captcha 生成图形验证码。
//
// 客户端 IP 通过统一受信代理规则解析。
// @Tags Auth
// @Summary 生成验证码
// @Success 200 {object} response.Captcha
// @Router /auth/captcha [get]
func (handler *Handler) Captcha(c *gin.Context) {
	result, err := handler.service.Captcha(c.Request.Context(), utils.GetClientIP(c))
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	c.JSON(http.StatusOK, result)
}

// Login 校验登录请求并创建服务端会话。
//
// 成功时返回脱敏用户、权限上下文与令牌对。
// @Tags Auth
// @Summary 用户登录
// @Param data body request.LoginAuthUser true "用户名, 密码, 验证码"
// @Success 200 {object} response.Response
// @Router /auth/login [post]
func (handler *Handler) Login(c *gin.Context) {
	var req request.LoginAuthUser
	if !request.BindAndValidateJSON(c, &req) {
		return
	}
	result, err := handler.service.Login(c.Request.Context(), req, ClientContext{IP: utils.GetClientIP(c), UserAgent: c.Request.UserAgent()})
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	c.JSON(http.StatusOK, result)
}

// GetInfo 查询当前用户认证上下文。
//
// 用户主键从认证中间件上下文读取。
// @Tags Auth
// @Summary 获取用户信息
// @Success 200 {object} response.Response
// @Router /auth/info [get]
func (handler *Handler) GetInfo(c *gin.Context) {
	result, err := handler.service.Info(c.Request.Context(), utils.GetId(c))
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	c.JSON(http.StatusOK, result)
}

// Logout 撤销请求体刷新令牌或 Authorization 访问令牌所属会话。
//
// 两种令牌都可以省略，空令牌或已失效令牌按幂等成功处理。
// @Tags Auth
// @Summary 用户登出
// @Param data body request.RefreshToken false "刷新凭证"
// @Success 200 {object} response.ActionResult
// @Router /auth/logout [delete]
func (handler *Handler) Logout(c *gin.Context) {
	var req request.RefreshToken
	if err := c.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		response.Error(err.Error(), nil, c)
		return
	}
	for _, token := range []string{req.Token, ext.StringReplace(utils.GetToken(c), "Bearer ", "", -1)} {
		if err := handler.service.Logout(c.Request.Context(), token); err != nil {
			response.Error(err.Error(), nil, c)
			return
		}
	}
	response.Success("", c)
}

// ResetEmail 发送邮箱验证码。
//
// email 缺失或格式无效时直接返回参数错误。
// @Tags Auth
// @Summary 获取邮箱验证码
// @Param data body request.EmailCode true "邮箱地址"
// @Success 200 {object} response.ActionResult
// @Router /auth/reset-email-code [post]
func (handler *Handler) ResetEmail(c *gin.Context) {
	var req request.EmailCode
	if err := c.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		response.Error(err.Error(), nil, c)
		return
	}
	if req.Email == "" {
		req.Email = c.Query("email")
	}
	if req.Email == "" {
		response.Error("email is null", nil, c)
		return
	}
	if err := validator.New().Var(req.Email, "email"); err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	if err := handler.service.ResetEmailCode(c.Request.Context(), req.Email); err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	response.Success("", c)
}

// RefreshToken 校验刷新令牌并轮换令牌对。
//
// 无效或重放令牌要求客户端重新登录。
// @Tags Auth
// @Summary 刷新Token，以旧换新
// @Param data body request.RefreshToken true "token值"
// @Success 200 {object} utils.JwtToken
// @Router /auth/refresh-token [post]
func (handler *Handler) RefreshToken(c *gin.Context) {
	var req request.RefreshToken
	if !request.BindAndValidateJSON(c, &req) {
		return
	}
	pair, err := handler.service.Refresh(c.Request.Context(), req.Token)
	if err != nil || pair == nil {
		response.Unauthorized("刷新凭证无效、已使用或会话已撤销，请重新登录。", nil, c)
		return
	}
	c.JSON(http.StatusOK, *pair)
}
