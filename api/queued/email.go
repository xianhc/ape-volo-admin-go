package queued

import (
	"github.com/gin-gonic/gin"
	"go-apevolo/model/queued"
	"go-apevolo/payloads/dto"
	"go-apevolo/payloads/request"
	"go-apevolo/payloads/response"
	"go-apevolo/utils"
)

type EmailQueuedApi struct{}

// Create
// @Tags   QueuedEmail
// @Summary 创建
// @Accept json
// @Produce json
// @Param request body dto.CreateUpdateEmailQueuedDto true "CreateUpdateEmailQueuedDto"
// @Success 201 {object} response.ActionResult "创建成功"
// @Router /queued/email/create [post]
func (e *EmailQueuedApi) Create(c *gin.Context) {
	var reqInfo dto.CreateUpdateEmailQueuedDto
	err := c.ShouldBindJSON(&reqInfo)
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	err = utils.VerifyData(reqInfo)
	if err != nil {
		response.Error("", utils.GetVerifyErr(err), c)
		return
	}
	reqInfo.SetCreateBy(utils.GetAccount(c))
	err = emailQueuedService.Create(&reqInfo)
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	response.Create("", c)
}

// Update
// @Tags   QueuedEmail
// @Summary 编辑
// @Accept json
// @Produce json
// @Param request body dto.CreateUpdateEmailQueuedDto true "CreateUpdateEmailQueuedDto"
// @Success 204 {object} response.ActionResult "编辑成功"
// @Router /queued/email/edit [put]
func (e *EmailQueuedApi) Update(c *gin.Context) {
	var reqInfo dto.CreateUpdateEmailQueuedDto
	err := c.ShouldBindJSON(&reqInfo)
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	err = utils.VerifyData(reqInfo)
	if err != nil {
		response.Error("", utils.GetVerifyErr(err), c)
		return
	}
	reqInfo.SetUpdateBy(utils.GetAccount(c))
	err = emailQueuedService.Update(&reqInfo)
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	response.NoContent(c)
}

// Delete
// @Tags   QueuedEmail
// @Summary 删除
// @Accept json
// @Produce json
// @Param request body request.IdCollection true "ID数组"
// @Success 200 {object} response.ActionResult "删除成功"
// @Router /queued/email/delete [delete]
func (e *EmailQueuedApi) Delete(c *gin.Context) {
	var idArray request.IdCollection
	err := c.ShouldBindJSON(&idArray)
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}

	err = emailQueuedService.Delete(idArray, utils.GetAccount(c))
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	response.Success("", c)
}

// Query
// @Tags QueuedEmail
// @Summary 查询
// @Accept json
// @Produce json
// @Param request body dto.EmailQueuedQueryCriteria true "request object"
// @Success 200 {object} response.ActionResultPage "查询成功"
// @Router /queued/email/query [get]
func (e *EmailQueuedApi) Query(c *gin.Context) {
	var reqInfo dto.EmailQueuedQueryCriteria
	err := c.ShouldBindQuery(&reqInfo)
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	err = utils.VerifyData(reqInfo.Pagination)
	if err != nil {
		response.Error("", utils.GetVerifyErr(err), c)
		return
	}
	list := make([]queued.Email, 0)
	var count int64
	err = emailQueuedService.Query(&reqInfo, &list, &count)
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	response.ResultPage(response.ActionResultPage{
		Content:       list,
		TotalElements: count,
	}, c)
}
