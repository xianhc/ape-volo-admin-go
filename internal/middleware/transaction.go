package middleware

import (
	"context"
	"errors"
	"net/http"
	"path"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/xianhc/ape-volo-admin-go/internal/global"
	"github.com/xianhc/ape-volo-admin-go/internal/modules/identity/datascope"
	"github.com/xianhc/ape-volo-admin-go/internal/payloads/response"
	"github.com/xianhc/ape-volo-admin-go/internal/utils"
	"go.uber.org/zap"
)

// Transaction 为写请求开启请求级事务。
//
// 本中间件统一开启请求事务，service 的 InTransaction 调用复用已有事务，
// 整个请求要么全成功要么全回滚。
//
// # 只包写请求
//
// GET、查询/导出类 POST 以及 job-manager 不开启。作业管理由 Service 提交配置后
// 同步运行时，才能向客户端区分“配置已保存”和“调度同步失败”。
//
// # 回滚判定
//
// 响应状态 >= 400，或事务上已标记错误，即回滚。业务错误都会走到这两个条件之一。
// 写请求的响应暂存到事务提交成功后再发送，提交失败返回 HTTP 500。
func Transaction() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request = c.Request.WithContext(datascope.WithSnapshotCache(c.Request.Context()))
		// 关闭事务切面时退化为各 service 自己开事务
		if !global.Config.Aop.Transactions || !needsTransaction(c.Request.Method, c.FullPath()) {
			c.Next()
			return
		}

		defer runAfterTx(c)
		original := c.Writer
		buffered := newTransactionWriter(original)
		defer func() {
			c.Writer = original
			if r := recover(); r != nil {
				writeTransactionError(c, "请求处理失败")
				panic(r)
			}
		}()
		err := utils.InTransaction(c.Request.Context(), func(ctx context.Context) error {
			c.Request = c.Request.WithContext(ctx)
			c.Writer = buffered
			defer func() { c.Writer = original }()
			c.Next()
			if buffered.Status() >= http.StatusBadRequest {
				return errResponseRollback
			}
			return nil
		})
		if err != nil && !errors.Is(err, errResponseRollback) {
			global.Logger.Error("事务执行失败", zap.String("path", c.Request.URL.Path), zap.Error(err))
			writeTransactionError(c, "事务执行失败")
			c.Abort()
			return
		}
		if err := buffered.send(); err != nil {
			global.Logger.Warn("发送响应失败", zap.Error(err))
		}
	}
}

var errResponseRollback = errors.New("response requires rollback")

// afterTxKey 保存「事务结束后执行」的回调列表。
const afterTxKey = "apevolo.afterTx"

// DeferAfterTx 把回调推迟到请求级事务结束（提交或回滚）之后执行。
//
// 操作日志、异常日志不能写在业务事务内部：一来业务回滚会把日志一起回滚，
// 失败的操作反而查不到；二来在 SQLite 这类库上，事务未结束时的独立写入
// 会直接撞 SQLITE_BUSY。推迟到事务结束后写，两个问题一起解决。
func DeferAfterTx(c *gin.Context, fn func()) {
	existing, _ := c.Get(afterTxKey)
	fns, _ := existing.([]func())
	c.Set(afterTxKey, append(fns, fn))
}

// runAfterTx 执行推迟到事务结束后的回调。
func runAfterTx(c *gin.Context) {
	c.Request = c.Request.WithContext(utils.WithoutTx(c.Request.Context()))
	value, ok := c.Get(afterTxKey)
	if !ok {
		return
	}
	fns, _ := value.([]func())
	for _, fn := range fns {
		fn()
	}
}

// needsTransaction 判断请求方法是否会产生写入。
func needsTransaction(method, route string) bool {
	if strings.Contains(route, "/job-manager/") {
		return false
	}
	switch method {
	case http.MethodPost:
		action := path.Base(route)
		return !strings.HasPrefix(action, "get-") && action != "download" && action != "build-tree"
	case http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	default:
		return false
	}
}

func writeTransactionError(c *gin.Context, message string) {
	response.JsonContent(response.ActionResult{Status: http.StatusInternalServerError, Message: message}, c)
}
