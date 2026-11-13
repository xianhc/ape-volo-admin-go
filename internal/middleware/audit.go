package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xianhc/ape-volo-admin-go/internal/global"
	"github.com/xianhc/ape-volo-admin-go/internal/model"
	"github.com/xianhc/ape-volo-admin-go/internal/modules/log/exceptionlog"
	"github.com/xianhc/ape-volo-admin-go/internal/modules/log/operatelog"
	"github.com/xianhc/ape-volo-admin-go/internal/modules/system/setting"
	"github.com/xianhc/ape-volo-admin-go/internal/payloads/response"
	"github.com/xianhc/ape-volo-admin-go/internal/utils"
	"go.uber.org/zap"
)

var auditLogService = operatelog.Service{}
var exceptionLogService = exceptionlog.Service{}

var respPool sync.Pool

func init() {
	respPool.New = func() interface{} {
		return make([]byte, 1024)
	}
}

func OperationRecord() gin.HandlerFunc {
	return func(c *gin.Context) {
		var body []byte
		if c.Request.Method != http.MethodGet {
			var err error
			body, err = io.ReadAll(c.Request.Body)
			if err != nil {
				global.Logger.Error("获取body参数数据错误:", zap.Error(err))
			} else {
				c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
			}
		} else {
			query := c.Request.URL.RawQuery
			query, _ = url.QueryUnescape(query)
			split := strings.Split(query, "&")
			m := make(map[string]string)
			for _, v := range split {
				kv := strings.Split(v, "=")
				if len(kv) == 2 {
					m[kv[0]] = kv[1]
				}
			}
			body, _ = json.Marshal(&m)
		}
		// 客户端环境（IP / 归属地 / 操作系统 / 设备类型 / 浏览器 / 版本）只解析一次，
		// 审计日志与异常日志共用，避免同一请求内重复解析 UA 与查询 IP 归属地。
		env := utils.ParseClientEnvironment(c)
		var creatorID *int64
		if id := utils.GetId(c); id != 0 {
			creatorID = &id
		}
		audit := operatelog.OperateLog{
			RootKey:           model.RootKey{Id: int64(utils.GenerateID())},
			Method:            c.Request.Method,
			RequestUrl:        c.Request.URL.Path,
			RequestParameters: model.CustomFieldText(body),
			RequestIp:         env.IP,
			IpAddress:         env.Location,
			UserAgent:         c.Request.UserAgent(),
			OperatingSystem:   env.OperatingSystem,
			DeviceType:        env.DeviceType,
			BrowserName:       env.BrowserName,
			Version:           env.BrowserVersion,
			CreatorName:       utils.GetAccount(c),
			BaseModel:         model.BaseModel{CreatorId: creatorID},
		}
		// 上传文件时候 中间件日志进行裁断操作
		if strings.Contains(c.GetHeader("Content-Type"), "multipart/form-data") {
			if len(audit.RequestParameters) > 1024 {
				// 截断
				newBody := respPool.Get().([]byte)
				newBody = newBody[:1024]
				copy(newBody, audit.RequestParameters)
				audit.RequestParameters = model.CustomFieldText(newBody)
				defer respPool.Put(newBody)
			}
		}

		writer := responseBodyWriter{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
		}
		c.Writer = writer
		now := time.Now()
		c.Next()
		duration := time.Since(now)
		metadata := requestLogMetadata(c)
		audit.Area = metadata.Area
		audit.Controller = metadata.Controller
		audit.Action = metadata.Action
		audit.Description = metadata.Description
		handlerStatus := c.Writer.Status()
		record := func() {
			if c.Writer.Status() == http.StatusOK || c.Writer.Status() == http.StatusCreated || c.Writer.Status() == http.StatusNoContent {
				audit.ResponseData = model.CustomFieldText(writer.body.String())
				audit.ExecutionDuration = duration.Milliseconds()
				if strings.Contains(c.Writer.Header().Get("Pragma"), "public") ||
					strings.Contains(c.Writer.Header().Get("Expires"), "0") ||
					strings.Contains(c.Writer.Header().Get("Cache-Control"), "must-revalidate, post-check=0, pre-check=0") ||
					strings.Contains(c.Writer.Header().Get("Content-Type"), "application/force-download") ||
					strings.Contains(c.Writer.Header().Get("Content-Type"), "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet") ||
					strings.Contains(c.Writer.Header().Get("Content-Type"), "application/octet-stream") ||
					strings.Contains(c.Writer.Header().Get("Content-Type"), "application/vnd.ms-excel") ||
					strings.Contains(c.Writer.Header().Get("Content-Type"), "application/download") ||
					strings.Contains(c.Writer.Header().Get("Content-Disposition"), "attachment") ||
					strings.Contains(c.Writer.Header().Get("Content-Transfer-Encoding"), "binary") {
					if len(audit.ResponseData) > 1024 {
						// 截断
						newBody := respPool.Get().([]byte)
						newBody = newBody[:1024]
						copy(newBody, audit.ResponseData)
						audit.ResponseData = model.CustomFieldText(newBody)
						defer respPool.Put(newBody)
					}
				}
				// 日志必须独立于业务事务：事务结束后再写，回滚时这条记录也要留下
				writeAuditLog(c, audit)
			} else {
				errMsgFull := c.Errors.ByType(gin.ErrorTypePrivate).String()
				errMsg := ""
				vm := writer.body.String()
				if c.Writer.Status() != handlerStatus {
					vm = ""
				}
				var resultAction response.ActionResult
				if err := json.Unmarshal([]byte(vm), &resultAction); err == nil {
					errMsg = resultAction.Message
				}
				exceptionLog := exceptionlog.ExceptionLog{
					RootKey:              model.RootKey{Id: int64(utils.GenerateID())},
					Area:                 metadata.Area,
					Controller:           metadata.Controller,
					Action:               metadata.Action,
					Description:          metadata.Description,
					Method:               c.Request.Method,
					RequestUrl:           c.Request.URL.Path,
					RequestParameters:    model.CustomFieldText(body),
					RequestIp:            env.IP,
					IpAddress:            env.Location,
					UserAgent:            c.Request.UserAgent(),
					OperatingSystem:      env.OperatingSystem,
					DeviceType:           env.DeviceType,
					BrowserName:          env.BrowserName,
					Version:              env.BrowserVersion,
					CreatorName:          utils.GetAccount(c),
					BaseModel:            model.BaseModel{CreatorId: creatorID},
					LogLevel:             c.Writer.Status(),
					ExceptionMessage:     errMsg,
					ExceptionMessageFull: model.CustomFieldText(errMsgFull),
					//ExceptionStack:       stackTrace,
				}
				writeExceptionLog(c, exceptionLog)
			}
		}
		if utils.TxFromContext(c.Request.Context()) != nil && !utils.TxSkipped(c.Request.Context()) {
			DeferAfterTx(c, record)
		} else {
			record()
		}
	}
}

// writeAuditLog 写操作日志。
//
// 请求处于事务中时推迟到事务结束后写：否则业务回滚会把日志一起回滚，
// 且事务未释放时的独立写入会撞锁。
// 保存开关以独立超时读取，避免客户端断开导致已完成请求的日志丢失。
func writeAuditLog(c *gin.Context, audit operatelog.OperateLog) {
	write := func() {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(utils.WithoutTx(c.Request.Context())), 5*time.Second)
		defer cancel()
		enabled, err := (&setting.Service{}).GetBool(ctx, setting.OperateLogSaveToDatabase)
		if err != nil {
			global.Logger.Warn("读取操作日志保存开关失败", zap.Error(err))
			return
		}
		if enabled {
			_ = auditLogService.Create(ctx, audit)
		}
	}
	if utils.TxFromContext(c.Request.Context()) != nil && !utils.TxSkipped(c.Request.Context()) {
		DeferAfterTx(c, write)
		return
	}
	write()
}

// writeExceptionLog 写异常日志，推迟规则同 writeAuditLog。
func writeExceptionLog(c *gin.Context, log exceptionlog.ExceptionLog) {
	write := func() {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(utils.WithoutTx(c.Request.Context())), 5*time.Second)
		defer cancel()
		enabled, err := (&setting.Service{}).GetBool(ctx, setting.ExceptionLogSaveToDatabase)
		if err != nil {
			global.Logger.Warn("读取异常日志保存开关失败", zap.Error(err))
			return
		}
		if enabled {
			_ = exceptionLogService.Create(ctx, log)
		}
	}
	if utils.TxFromContext(c.Request.Context()) != nil && !utils.TxSkipped(c.Request.Context()) {
		DeferAfterTx(c, write)
		return
	}
	write()
}

type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (r responseBodyWriter) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}
