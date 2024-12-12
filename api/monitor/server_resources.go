package monitor

import (
	"github.com/gin-gonic/gin"
	"go-apevolo/utils"
	"net/http"
)

type ServerResourcesApi struct{}

// Query
// @Tags   ServerResourcesInfo
// @Summary 查询
// @Accept application/json
// @Produce application/json
// @Success 200 {object} utils.ServerResourcesInfo "查询成功"
// @Router /api/service/resources/info [get]
func (s *ServerResourcesApi) Query(c *gin.Context) {

	serverInfo := utils.GetServerResourcesInfo()

	c.JSON(http.StatusOK, serverInfo)
}
