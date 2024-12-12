package middleware

import (
	"bytes"
	"encoding/json"
	"github.com/mssola/user_agent"
	"go-apevolo/model"
	"go-apevolo/model/monitor"
	"go-apevolo/payloads/response"
	"go-apevolo/utils"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"go-apevolo/global"
	"go-apevolo/service"
	"go.uber.org/zap"
)

var auditLogService = service.ServiceGroupApp.MonitorServiceGroup.AuditLogService
var exceptionLogService = service.ServiceGroupApp.MonitorServiceGroup.ExceptionLogService

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
		ip := utils.GetClientIP(c)
		audit := monitor.AuditLog{
			RootKey:           model.RootKey{Id: int64(utils.GenerateID())},
			Method:            c.Request.Method,
			RequestUrl:        c.Request.URL.Path,
			RequestParameters: utils.CustomFieldText(body),
			RequestIp:         ip,
			//IpAddress:         ipAddress,
			//OperatingSystem:   ua.Platform(),
			//DeviceType:        deviceType,
			//BrowserName:       browserName,
			//Version:           browserVersion,
			BaseModel: model.BaseModel{CreateBy: utils.GetAccount(c)},
		}
		// 上传文件时候 中间件日志进行裁断操作
		if strings.Contains(c.GetHeader("Content-Type"), "multipart/form-data") {
			if len(audit.RequestParameters) > 1024 {
				// 截断
				newBody := respPool.Get().([]byte)
				copy(newBody, audit.RequestParameters)
				audit.RequestParameters = utils.CustomFieldText(newBody)
				defer respPool.Put(newBody[:0])
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
		ua := user_agent.New(c.Request.UserAgent())
		ipAddress, _ := utils.SearchIpAddress(ip)
		browserName, browserVersion := ua.Browser()
		deviceType := utils.GetDeviceType(ua.Platform(), ua.OS(), ua.Mobile())
		if c.Writer.Status() == http.StatusOK || c.Writer.Status() == http.StatusCreated || c.Writer.Status() == http.StatusNoContent {
			audit.ResponseData = utils.CustomFieldText(writer.body.String())
			audit.ExecutionDuration = duration.Milliseconds()
			audit.IpAddress = ipAddress
			audit.OperatingSystem = ua.Platform()
			audit.DeviceType = deviceType
			audit.BrowserName = browserName
			audit.Version = browserVersion

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
					copy(newBody, audit.ResponseData)
					audit.ResponseData = utils.CustomFieldText(newBody)
					defer respPool.Put(newBody[:0])
				}
			}
			_ = auditLogService.Create(audit)
		} else {
			errMsgFull := c.Errors.ByType(gin.ErrorTypePrivate).String()
			errMsg := ""
			vm := writer.body.String()
			var resultAction response.ActionResult
			if err := json.Unmarshal([]byte(vm), &resultAction); err == nil {
				errMsg = resultAction.Message
			}
			exceptionLog := monitor.ExceptionLog{
				RootKey:              model.RootKey{Id: int64(utils.GenerateID())},
				Method:               c.Request.Method,
				RequestUrl:           c.Request.URL.Path,
				RequestParameters:    utils.CustomFieldText(body),
				RequestIp:            ip,
				IpAddress:            ipAddress,
				OperatingSystem:      ua.Platform(),
				DeviceType:           deviceType,
				BrowserName:          browserName,
				Version:              browserVersion,
				BaseModel:            model.BaseModel{CreateBy: utils.GetAccount(c)},
				LogLevel:             c.Writer.Status(),
				ExceptionMessage:     errMsg,
				ExceptionMessageFull: utils.CustomFieldText(errMsgFull),
				//ExceptionStack:       stackTrace,
			}
			_ = exceptionLogService.Create(exceptionLog)
		}
	}
}

type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (r responseBodyWriter) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}
