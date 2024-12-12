package response

import (
	"github.com/gin-gonic/gin"
	"go-apevolo/utils/ext"
	"net/http"
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
	actionResult := ActionResult{Message: message, Status: http.StatusOK}
	JsonContent(actionResult, c)
}

// Create 创建
func Create(message string, c *gin.Context) {
	if message == "" {
		message = "创建成功"
	}
	actionResult := ActionResult{Message: message, Status: http.StatusCreated}
	JsonContent(actionResult, c)
}

// NoContent 编辑
func NoContent(c *gin.Context) {
	c.JSON(http.StatusNoContent, nil)
}

// Error 错误
func Error(message string, error *ActionError, c *gin.Context) {
	actionResult := ActionResult{Message: message, Status: http.StatusBadRequest, ActionError: error}
	JsonContent(actionResult, c)
}

// Unauthorized 无权限
func Unauthorized(message string, error *ActionError, c *gin.Context) {
	actionResult := ActionResult{Message: message, Status: http.StatusUnauthorized, ActionError: error}
	JsonContent(actionResult, c)
}

// Forbidden 权限不够
func Forbidden(message string, error *ActionError, c *gin.Context) {
	actionResult := ActionResult{Message: message, Status: http.StatusForbidden, ActionError: error}
	JsonContent(actionResult, c)
}

func JsonContent(actionResult ActionResult, c *gin.Context) {
	actionResult.Path = c.Request.URL.Path
	actionResult.Timestamp = ext.GetTimestamp(ext.GetCurrentTime())
	if actionResult.ActionError == nil {
		emptyActionError := ActionError{}
		actionResult.ActionError = &emptyActionError
	} else {
		//randomIndex := rand.Intn(len(actionResult.ActionError.Errors))
		//i := 0
		for _, value := range actionResult.ActionError.Errors {
			//if i == randomIndex {
			actionResult.Message = value
			break
			//}
			//i++
		}
	}
	c.JSON(actionResult.Status, actionResult)
}
