package permission

import (
	"github.com/gin-gonic/gin"
	"go-apevolo/global"
	"go-apevolo/global/constants/cachePrefix"
	"go-apevolo/middleware/aop"
	"go-apevolo/model/permission"
	"go-apevolo/payloads/dto"
	"go-apevolo/payloads/request"
	"go-apevolo/payloads/response"
	"go-apevolo/utils"
	"go-apevolo/utils/ext"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

type MenuApi struct{}

// Create
// @Tags   Menu
// @Summary 创建
// @Accept json
// @Produce json
// @Param request body dto.CreateUpdateMenuDto true "CreateUpdateMenuDto object"
// @Success 200 {object} response.ActionResult "创建成功"
// @Router /menu/create [post]
func (m *MenuApi) Create(c *gin.Context) {
	var reqInfo dto.CreateUpdateMenuDto
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
	err = menuService.Create(&reqInfo)
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	response.Create("", c)
}

// Update
// @Tags   Menu
// @Summary 编辑
// @Accept json
// @Produce json
// @Param request body dto.CreateUpdateMenuDto true "CreateUpdateMenuDto object"
// @Success 204 {object} response.ActionResult "编辑成功"
// @Router /menu/edit [put]
func (m *MenuApi) Update(c *gin.Context) {
	var reqInfo dto.CreateUpdateMenuDto
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
	err = menuService.Update(&reqInfo)
	if err != nil {
		global.Logger.Error("修改失败!", zap.Error(err))
		response.Error(err.Error(), nil, c)
		return
	}
	response.NoContent(c)
}

// Delete
// @Tags   Menu
// @Summary 删除
// @Accept json
// @Produce json
// @Param request body request.IdCollection true "ID数组"
// @Success 200 {object} response.ActionResult "删除成功"
// @Router /menu/delete [delete]
func (m *MenuApi) Delete(c *gin.Context) {
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
	err = menuService.Delete(intSlice, utils.GetAccount(c))
	if err != nil {
		global.Logger.Error("删除失败!", zap.Error(err))
		response.Error(err.Error(), nil, c)
		return
	}
	response.Success("", c)
}

// Query
// @Tags   Menu
// @Summary 查询
// @Accept application/json
// @Produce application/json
// @Param request body dto.MenuQueryCriteria true "Menu request object"
// @Success 200 {object} response.ActionResultPage "查询成功"
// @Router /menu/query [get]
func (m *MenuApi) Query(c *gin.Context) {
	var pageInfo dto.MenuQueryCriteria
	err := c.ShouldBindQuery(&pageInfo)
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	//err = utils.Verify(pageInfo.Pagination, utils.PaginationVerify)
	//if err != nil {
	//	response.Error(err.Error(), nil, c)
	//	return
	//}
	list, total, err := menuService.Query(pageInfo)
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

// Download
// @Tags   Menu
// @Summary 查询
// @Accept application/json
// @Produce application/json
// @Param request body dto.RoleQueryCriteria true "Role request object"
// @Success 200 {object} response.ActionResultPage "查询成功"
// @Router /menu/download [get]
func (m *MenuApi) Download(c *gin.Context) {
	var pageInfo dto.MenuQueryCriteria
	err := c.ShouldBindQuery(&pageInfo)
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	filePath, fileName, err := menuService.Download(pageInfo)
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

// GetSuperior
// @Tags      Menu
// @Summary   获取同级、父级菜单
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Success   200   {object}  permission.Menu  "获取同级、父级菜单"
// @Router    /api/menu/superior [get]
func (m *MenuApi) GetSuperior(c *gin.Context) {
	var id = c.Query("id")
	if id == "" {
		response.Error("id is null", nil, c)
		return
	}
	newCacheConfig := aop.NewCacheConfig(cachePrefix.LoadMenusById, ext.GetTimeDuration(1, ext.Hour), nil)
	menuList := &[]permission.Menu{}
	cacheMenus := aop.CacheAop(newCacheConfig, menuService.GetSuperior, menuList, ext.StringToInt64(id), menuList)
	menuResult := cacheMenus()
	if menuResult != nil {
		response.Error(menuResult.(error).Error(), nil, c)
	}
	c.JSON(http.StatusOK, menuList)
}

// GetChild
// @Tags      Menu
// @Summary   获取所有子级菜单ID
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Success   200   {object}  []int64  "获取所有子级菜单ID"
// @Router    /api/menu/child [get]
func (m *MenuApi) GetChild(c *gin.Context) {
	pid := c.Query("id")
	if pid == "" {
		response.Error("id is null", nil, c)
		return
	}
	id, err := strconv.ParseInt(pid, 10, 64)
	allIds, err := menuService.GetChildrenIds([]int64{id}, []int64{})
	if err != nil {
		global.Logger.Error("获取失败", zap.Error(err))
		response.Error(err.Error(), nil, c)
	}
	intSlice := make([]string, 0)
	for _, id := range allIds {
		intSlice = utils.AppendIfNotExists(intSlice, strconv.FormatInt(id, 10))
	}
	c.JSON(http.StatusOK, intSlice)
}

// Lazy
// @Tags      Menu
// @Summary   获取子菜单
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Success   200   {object}  permission.Menu  "获取子菜单"
// @Router    /api/menu/lazy [get]
func (m *MenuApi) Lazy(c *gin.Context) {
	pid := c.Query("pid")
	if pid == "" {
		response.Error("pid is null", nil, c)
		return
	}

	newCacheConfig := aop.NewCacheConfig(cachePrefix.LoadMenusByPId, ext.GetTimeDuration(1, ext.Hour), nil)
	menuList := &[]permission.Menu{}
	cacheMenus := aop.CacheAop(newCacheConfig, menuService.GetMenuByParentId, menuList, ext.StringToInt64(pid), menuList)
	menuResult := cacheMenus()
	if menuResult != nil {
		response.Error(menuResult.(error).Error(), nil, c)
	}
	c.JSON(http.StatusOK, menuList)
}

// Build
// @Tags      Menu
// @Summary   构建前端路由菜单
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Success   200   {object}  response.Response{data=response.TreeMenuVo}  "构建前端路由菜单"
// @Router    /api/menu/build [get]
func (m *MenuApi) Build(c *gin.Context) {

	newCacheConfig := aop.NewCacheConfig(cachePrefix.UserMenuById, ext.GetTimeDuration(2, ext.Hour), nil)
	treeMenuVoList := &[]response.TreeMenuVo{}
	cacheTreeMenu := aop.CacheAop(newCacheConfig, menuService.GetMenuTree, treeMenuVoList, utils.GetId(c), treeMenuVoList)
	treeResult := cacheTreeMenu()
	if treeResult != nil {
		response.Error(treeResult.(error).Error(), nil, c)
	}
	c.JSON(http.StatusOK, treeMenuVoList)
}
