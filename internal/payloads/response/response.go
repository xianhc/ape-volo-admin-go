package response

import (
	"net/http"
	"sort"

	"github.com/gin-gonic/gin"
	"github.com/xianhc/ape-volo-admin-go/internal/utils/ext"
)

type Response struct {
	Code int         `json:"code"`
	Data interface{} `json:"data"`
	Msg  string      `json:"msg"`
}

// ResultPage 分页
func ResultPage(actionResultPage ActionResultPage, c *gin.Context) {
	// 开始时间
	c.JSON(http.StatusOK, actionResultPage)
}

// Success 请求成功
func Success(message string, c *gin.Context) {
	if message == "" {
		message = "请求成功"
	}
	actionResult := ActionResult{IsSuccess: true, Message: message, Status: http.StatusOK}
	JsonContent(actionResult, c)
}

// Create 创建
func Create(message string, c *gin.Context) {
	if message == "" {
		message = "创建成功"
	}
	actionResult := ActionResult{IsSuccess: true, Message: message, Status: http.StatusCreated}
	JsonContent(actionResult, c)
}

// NoContent 编辑
func NoContent(c *gin.Context) {
	c.JSON(http.StatusNoContent, nil)
}

// Error 错误
func Error(message string, error *ActionError, c *gin.Context) {
	actionResult := ActionResult{IsSuccess: false, Message: message, Status: http.StatusBadRequest, ActionError: error}
	JsonContent(actionResult, c)
}

// Unauthorized 无权限
func Unauthorized(message string, error *ActionError, c *gin.Context) {
	actionResult := ActionResult{IsSuccess: false, Message: message, Status: http.StatusUnauthorized, ActionError: error}
	JsonContent(actionResult, c)
}

// Forbidden 权限不够
func Forbidden(message string, error *ActionError, c *gin.Context) {
	actionResult := ActionResult{IsSuccess: false, Message: message, Status: http.StatusForbidden, ActionError: error}
	JsonContent(actionResult, c)
}

func JsonContent(actionResult ActionResult, c *gin.Context) {
	// isSuccess 以状态码为准，避免调用方漏设
	actionResult.IsSuccess = actionResult.Status < http.StatusBadRequest
	actionResult.Path = c.Request.URL.Path
	actionResult.Timestamp = ext.GetTimestamp(ext.GetCurrentTime())
	if actionResult.ActionError == nil {
		emptyActionError := ActionError{}
		actionResult.ActionError = &emptyActionError
	} else if len(actionResult.ActionError.Errors) > 0 {
		// 按字段名排序后取第一条。
		// 直接 range map 会因 Go 的随机遍历顺序导致同一请求多次提交
		// 返回不同的校验提示，问题无法复现、前端提示也会跳变。
		keys := make([]string, 0, len(actionResult.ActionError.Errors))
		for k := range actionResult.ActionError.Errors {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		actionResult.Message = actionResult.ActionError.Errors[keys[0]]
	}
	c.JSON(actionResult.Status, actionResult)
}
