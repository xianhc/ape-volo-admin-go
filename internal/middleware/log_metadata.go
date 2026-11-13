package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
)

const logModuleKey = "apevolo.log.module"
const logDescriptionKey = "apevolo.log.description"

type logMetadata struct {
	Area        string
	Controller  string
	Action      string
	Description string
}

func LogModule(area, controller string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(logModuleKey, logMetadata{Area: area, Controller: controller})
	}
}

func LogAction(description string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(logDescriptionKey, description)
	}
}

func requestLogMetadata(c *gin.Context) logMetadata {
	metadata := logMetadata{}
	if value, exists := c.Get(logModuleKey); exists {
		metadata, _ = value.(logMetadata)
	}
	route := c.FullPath()
	if route == "" {
		route = c.Request.URL.Path
	}
	segments := strings.Split(strings.Trim(route, "/"), "/")
	if len(segments) > 0 {
		metadata.Action = segments[len(segments)-1]
	}
	if metadata.Controller == "" && len(segments) > 1 {
		metadata.Controller = segments[len(segments)-2]
	}
	if metadata.Area == "" && len(segments) > 2 {
		metadata.Area = strings.Join(segments[:len(segments)-2], "/")
	}
	metadata.Description = c.GetString(logDescriptionKey)
	if metadata.Description == "" {
		metadata.Description = c.Request.Method + " " + route
	}
	return metadata
}
