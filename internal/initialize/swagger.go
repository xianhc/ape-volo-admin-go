package initialize

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/xianhc/ape-volo-admin-go/internal/docs"
	"github.com/xianhc/ape-volo-admin-go/internal/global"
)

func registerSwagger(router *gin.Engine) {
	cfg := global.Config.Swagger
	if !cfg.Enabled {
		return
	}

	// 复制生成文档再覆盖元数据，重新生成接口定义不会丢失配置能力。
	spec := *docs.SwaggerInfo
	var document map[string]any
	if err := json.Unmarshal([]byte(spec.ReadDoc()), &document); err != nil {
		panic(fmt.Errorf("解析接口文档失败: %w", err))
	}
	info := document["info"].(map[string]any)
	info["title"], info["version"] = cfg.Title, cfg.Version
	info["contact"] = map[string]string{
		"name": cfg.ContactName, "email": cfg.ContactEmail, "url": cfg.ContactURL,
	}
	document["basePath"] = "/" + strings.Trim(global.Config.System.RouterPrefix, "/")
	body, err := json.Marshal(document)
	if err != nil {
		panic(fmt.Errorf("序列化接口文档失败: %w", err))
	}

	documentPath := "/" + cfg.Name + "/doc.json"
	ui := ginSwagger.WrapHandler(swaggerFiles.Handler,
		ginSwagger.URL("../"+cfg.Name+"/doc.json"),
		func(c *ginSwagger.Config) { c.Title = cfg.Title },
	)
	router.GET("/swagger/*any", func(c *gin.Context) {
		path := c.Param("any")
		// 保留既有页面使用的 JSON 地址。
		if path == documentPath || path == "/api/doc.json" {
			c.Data(http.StatusOK, "application/json; charset=utf-8", body)
			return
		}
		if strings.HasSuffix(path, "/doc.json") {
			c.Status(http.StatusNotFound)
			return
		}
		ui(c)
	})
}
