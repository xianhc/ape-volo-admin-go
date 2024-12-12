package permission

import (
	"github.com/gin-gonic/gin"
	"go-apevolo/global"
	"go-apevolo/model/permission"
	"go-apevolo/payloads/dto"
	"go-apevolo/payloads/request"
	"go-apevolo/payloads/response"
	"go-apevolo/utils"
	"go-apevolo/utils/ext"
	"go.uber.org/zap"
	"strconv"
)

type DeptApi struct{}

// Create
// @Tags   Dept
// @Summary 创建
// @Accept json
// @Produce json
// @Param request body dto.CreateUpdateDeptDto true "CreateUpdateDeptDto object"
// @Success 204 {object} response.ActionResult "编辑成功"
// @Router /dept/create [post]
func (d *DeptApi) Create(c *gin.Context) {
	var reqInfo dto.CreateUpdateDeptDto
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
	err = deptService.Create(&reqInfo)
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	response.Create("", c)
}

// Update
// @Tags   Dept
// @Summary 编辑
// @Accept json
// @Produce json
// @Param request body dto.CreateUpdateDeptDto true "CreateUpdateDeptDto object"
// @Success 204 {object} response.ActionResult "编辑成功"
// @Router /dept/edit [put]
func (d *DeptApi) Update(c *gin.Context) {
	var reqInfo dto.CreateUpdateDeptDto
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
	err = deptService.Update(&reqInfo)
	if err != nil {
		global.Logger.Error("修改失败!", zap.Error(err))
		response.Error(err.Error(), nil, c)
		return
	}
	response.NoContent(c)
}

// Delete
// @Tags   Dept
// @Summary 删除
// @Accept json
// @Produce json
// @Param request body request.IdCollection true "ID数组"
// @Success 200 {object} response.ActionResult "删除成功"
// @Router /dept/delete [delete]
func (d *DeptApi) Delete(c *gin.Context) {
	var idArray request.IdCollection
	err := c.ShouldBindJSON(&idArray)
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	// 创建一个新的 int64 切片
	var intSlice []int64

	// 遍历字符串切片，将每个字符串转换为 int64 并添加到新切片中
	for _, str := range idArray.IdArray {
		// 使用 strconv.ParseInt 将字符串解析为 int64
		num, err := strconv.ParseInt(str, 10, 64)
		if err == nil {
			intSlice = append(intSlice, num)
		}
	}
	err = deptService.Delete(intSlice, utils.GetAccount(c))
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	response.Success("", c)
}

// Superior
// @Tags   Dept
// @Summary 查询
// @Accept json
// @Produce json
// @Param  id query string true "ID"
// @Success 200 {object} response.ActionResultPage "查询成功"
// @Router /dept/superior [get]
func (d *DeptApi) Superior(c *gin.Context) {
	var id = c.Query("id")
	if id == "" {
		response.Error("id is null", nil, c)
		return
	}
	list, total, err := deptService.Superior(ext.StringToInt64(id))
	if err != nil {
		global.Logger.Error("获取失败!", zap.Error(err))
		response.Error("获取失败", nil, c)
		return
	}
	response.ResultPage(response.ActionResultPage{
		Content:       list,
		TotalElements: total,
	}, c)
}

// Query
// @Tags   Dept
// @Summary 查询
// @Accept json
// @Produce json
// @Param request body dto.DeptQueryCriteria true "Dept request object"
// @Success 200 {object} response.ActionResultPage "查询成功"
// @Router /dept/query [get]
func (d *DeptApi) Query(c *gin.Context) {
	var pageInfo dto.DeptQueryCriteria
	err := c.ShouldBindQuery(&pageInfo)
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	list := make([]permission.Department, 0)
	var count int64
	err = deptService.Query(&pageInfo, &list, &count)
	if err != nil {
		response.Error("获取失败:"+err.Error(), nil, c)
		return
	} else {
		for i := range list {
			list[i].CalculateHasChildren()
			list[i].CalculateLeaf()
			list[i].CalculateLabel()
		}
	}
	response.ResultPage(response.ActionResultPage{
		Content:       list,
		TotalElements: count,
	}, c)
}

// Download
// @Tags   Dept
// @Summary 查询
// @Accept json
// @Produce json
// @Param request body dto.DeptQueryCriteria true "Dept request object"
// @Success 200 {object} response.ActionResultPage "查询成功"
// @Router /dept/query [get]
func (d *DeptApi) Download(c *gin.Context) {
	var pageInfo dto.DeptQueryCriteria
	err := c.ShouldBindQuery(&pageInfo)
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	filePath, fileName, err := deptService.Download(&pageInfo)
	if err != nil {
		global.Logger.Error("导出失败!", zap.Error(err))
		response.Error("导出失败", nil, c)
		return
	}

	// 设置响应头，告诉浏览器将文件作为附件下载
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", "attachment; filename="+fileName)
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Expires", "0")
	c.Header("Cache-Control", "must-revalidate")
	c.Header("Pragma", "public")
	c.File(filePath)
}
