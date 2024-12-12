package permission

import (
	"errors"
	"github.com/tealeg/xlsx"
	"go-apevolo/global"
	"go-apevolo/model"
	"go-apevolo/model/permission"
	"go-apevolo/payloads/dto"
	"go-apevolo/payloads/response"
	"go-apevolo/utils"
	"go-apevolo/utils/ext"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"strconv"
	"strings"
)

type MenuService struct{}

// Create
// @description: 创建
// @receiver: menuService
// @param: req
// @return: error
func (menuService *MenuService) Create(req *dto.CreateUpdateMenuDto) error {
	menu := &permission.Menu{}
	var total int64
	err := global.Db.Model(&permission.Menu{}).Scopes(utils.IsDeleteSoft).Where("title = ?", req.Title).Count(&total).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		return err
	}
	if total > 0 {
		return errors.New("菜单标题=>" + req.Title + "=>已存在!")
	}
	if req.Type != 1 {
		err = global.Db.Model(&permission.Menu{}).Scopes(utils.IsDeleteSoft).Where("permission = ?", req.Permission).Count(&total).Error
		if err != nil {
			global.Logger.Error("db error: ", zap.Error(err))
			return err
		}
		if total > 0 {
			return errors.New("权限标识=>" + req.Permission + "=>已存在!")
		}
	}
	if req.ComponentName != "" {
		err = global.Db.Model(&permission.Menu{}).Scopes(utils.IsDeleteSoft).Where("component_name = ?", req.ComponentName).Count(&total).Error
		if err != nil {
			global.Logger.Error("db error: ", zap.Error(err))
			return err
		}
		if total > 0 {
			return errors.New("组件名称=>" + req.ComponentName + "=>已存在!")
		}
	}

	if req.Type != 1 {
		if req.Permission == "" {
			return errors.New("权限标识为必填")
		}
	}
	if req.IFrame {
		if !strings.HasPrefix(strings.ToLower(req.Path), "http") && !strings.HasPrefix(strings.ToLower(req.Path), "http") {
			return errors.New("外链菜单必须以http://或者https://开头")
		}
	}
	db := global.Db.Begin()
	defer func() {
		if err != nil {
			global.Logger.Error("tran error: ", zap.Error(err))
			db.Rollback()
		} else {
			db.Commit()
		}
	}()
	req.Generate(menu)
	err = db.Create(menu).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		return err
	}
	if menu.ParentId > 0 {
		//重新计算子部门个数
		parentMenu := &permission.Menu{}
		err = db.Scopes(utils.IsDeleteSoft).Where("id = ?", req.ParentId).First(parentMenu).Error
		if err != nil {
			global.Logger.Error("db error: ", zap.Error(err))
			return err
		}
		err = db.Model(&permission.Menu{}).Scopes(utils.IsDeleteSoft).Where("parent_id = ?", parentMenu.Id).Count(&total).Error
		if err != nil {
			global.Logger.Error("db error: ", zap.Error(err))
			return err
		}
		parentMenu.SubCount = int(total)
		err = db.Updates(parentMenu).Error
		if err != nil {
			global.Logger.Error("db error: ", zap.Error(err))
			return err
		}
	}
	return nil
}

// Update
// @description: 编辑
// @receiver: menuService
// @param: req
// @return: err
func (menuService *MenuService) Update(req *dto.CreateUpdateMenuDto) (err error) {
	oldMenu := &permission.Menu{}
	menu := &permission.Menu{}
	err = global.Db.Scopes(utils.IsDeleteSoft).First(oldMenu, req.Id).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("数据不存在或您无权查看！")
		}
		return err
	}
	var total int64
	if oldMenu.Title != req.Title {
		err = global.Db.Model(&permission.Menu{}).Scopes(utils.IsDeleteSoft).Where("title = ?", req.Title).Count(&total).Error
		if err != nil {
			global.Logger.Error("db error: ", zap.Error(err))
			return err
		}
		if total > 0 {
			return errors.New("菜单标题=>" + req.Title + "=>已存在!")
		}
	}
	if req.Type != 1 && oldMenu.Permission != req.Permission {
		err = global.Db.Model(&permission.Menu{}).Scopes(utils.IsDeleteSoft).Where("permission = ?", req.Permission).Count(&total).Error
		if err != nil {
			global.Logger.Error("db error: ", zap.Error(err))
			return err
		}
		if total > 0 {
			return errors.New("权限标识=>" + req.Permission + "=>已存在!")
		}
	}
	if req.ComponentName != "" && oldMenu.ComponentName != req.ComponentName {
		err = global.Db.Model(&permission.Menu{}).Scopes(utils.IsDeleteSoft).Where("component_name = ?", req.ComponentName).Count(&total).Error
		if err != nil {
			global.Logger.Error("db error: ", zap.Error(err))
			return err
		}
		if total > 0 {
			return errors.New("组件名称=>" + req.ComponentName + "=>已存在!")
		}
	}
	if req.IFrame {
		if !strings.HasPrefix(strings.ToLower(req.Path), "http") && !strings.HasPrefix(strings.ToLower(req.Path), "http") {
			return errors.New("外链菜单必须以http://或者https://开头")
		}
	}
	db := global.Db.Begin()
	defer func() {
		if err != nil {
			global.Logger.Error("tran error: ", zap.Error(err))
			db.Rollback()
		} else {
			db.Commit()
		}
	}()
	req.Generate(menu)
	err = db.Updates(menu).Error
	if err != nil {
		return err
	}

	if oldMenu.ParentId != req.ParentId {
		if req.ParentId > 0 {
			parentMenu := &permission.Menu{}
			err = db.Scopes(utils.IsDeleteSoft).Where("id = ?", req.ParentId).First(parentMenu).Error
			if err != nil {
				global.Logger.Error("db error: ", zap.Error(err))
				return err
			}
			err = db.Model(&permission.Menu{}).Scopes(utils.IsDeleteSoft).Where("parent_id = ?", parentMenu.Id).Count(&total).Error
			if err != nil {
				global.Logger.Error("db error: ", zap.Error(err))
				return err
			}
			parentMenu.SubCount = int(total)
			err = db.Updates(parentMenu).Error
			if err != nil {
				global.Logger.Error("db error: ", zap.Error(err))
				return err
			}
		}
		if oldMenu.ParentId > 0 {
			var parentMenu permission.Menu
			err = db.Scopes(utils.IsDeleteSoft).Where("id = ?", oldMenu.ParentId).First(&parentMenu).Error
			if err != nil {
				global.Logger.Error("db error: ", zap.Error(err))
				return err
			}
			err = db.Model(&permission.Menu{}).Scopes(utils.IsDeleteSoft).Where("parent_id = ?", parentMenu.Id).Count(&total).Error
			if err != nil {
				global.Logger.Error("db error: ", zap.Error(err))
				return err
			}
			parentMenu.SubCount = int(total)
			err = db.Updates(&parentMenu).Error
			if err != nil {
				global.Logger.Error("db error: ", zap.Error(err))
				return err
			}
		}
	}
	return nil
}

// Delete
// @description: 删除
// @receiver: menuService
// @param: ids
// @param: updateBy
// @return: err
func (menuService *MenuService) Delete(ids []int64, updateBy string) (err error) {
	allIds, err := menuService.GetChildrenIds(ids, []int64{})
	if err != nil {
		return err
	}

	var menus []permission.Menu
	result := global.Db.Scopes(utils.IsDeleteSoft).Find(&menus, "id in ?", allIds)
	if result.Error != nil {
		return err
	}
	if result.RowsAffected == 0 {
		return errors.New("数据不存在")
	}
	localTime := ext.GetCurrentTime()
	err = global.Db.Model(&permission.Menu{}).Scopes(utils.IsDeleteSoft).Where("id in (?)", allIds).Updates(
		permission.Menu{BaseModel: model.BaseModel{
			UpdateBy:   &updateBy,
			UpdateTime: &localTime,
		}, SoftDeleted: model.SoftDeleted{IsDeleted: true}},
	).Error
	if result.Error != nil {
		return err
	}

	var pIds []int64
	for _, dept := range menus {
		pIds = utils.AppendInt64(pIds, dept.ParentId)
	}
	var updateMenuList []permission.Menu
	err = global.Db.Model(&permission.Menu{}).Scopes(utils.IsDeleteSoft).Where("id in (?)", pIds).Find(&updateMenuList).Error
	if err != nil {
		return err
	}
	if len(updateMenuList) > 0 {
		for _, dept := range updateMenuList {
			var total int64
			err = global.Db.Model(&permission.Menu{}).Scopes(utils.IsDeleteSoft).Where("parent_id = ? ", dept.Id).Count(&total).Error
			if err != nil {
				break
			}
			dept.SubCount = int(total)
			err = global.Db.Model(&permission.Menu{}).Where("id = ?", dept.Id).UpdateColumn("sub_count", dept.SubCount).Error
			if err != nil {
				break
			}
		}
		//err = global.Db.Model(&permission.Department{}).Updates(updateDeptList).Error
	}
	return err
}

// Query
// @description: 查询
// @receiver: menuService
// @param: info
// @return: list
// @return: total
// @return: err
func (menuService *MenuService) Query(info dto.MenuQueryCriteria) (list interface{}, total int64, err error) {

	var menus []permission.Menu
	// 创建db
	db := buildMenuQuery(global.Db.Model(&permission.Menu{}), info)

	err = db.Find(&menus).Error
	if err == nil {
		for i := range menus {
			menus[i].CalculateHasChildren()
			menus[i].CalculateLeaf()
			menus[i].InitChildren()
		}
	}
	total = int64(len(menus))
	return menus, total, err
}

// Download
// @description: 导出
// @receiver: menuService
// @param: info
// @return: filePath
// @return: fileName
// @return: err
func (menuService *MenuService) Download(info dto.MenuQueryCriteria) (filePath string, fileName string, err error) {
	var menus []permission.Menu
	// 创建db并构建查询条件
	err = buildMenuQuery(global.Db.Model(&permission.Menu{}), info).Find(&menus).Error

	if err != nil {
		return
	}
	// 创建一个新的 Excel 文件
	file := xlsx.NewFile()
	sheet, err := file.AddSheet("Menus")
	if err != nil {
		return
	}
	style := xlsx.NewStyle()
	style.Fill = *xlsx.NewFill("solid", "1e90ff", "1e90ff")
	style.ApplyFill = true
	// 表头数据
	header := []string{"ID", "菜单标题", "组件路径", "权限标识符", "是否IFrame", "组件", "组件名称", "菜单父ID", "排序", "Icon图标", "菜单类型", "是否缓存", "是否隐藏", "子菜单个数", "创建时间"}
	// 将表头添加到 Excel 工作表
	row := sheet.AddRow()
	for _, cellValue := range header {
		cell := row.AddCell()
		cell.Value = cellValue
		cell.SetStyle(style)
	}

	// 将数据添加到 Excel 工作表
	for _, menu := range menus {
		row := sheet.AddRow()
		row.WriteSlice(&[]interface{}{
			strconv.FormatInt(menu.Id, 10),
			menu.Title,
			menu.Path,
			menu.Permission,
			func() string {
				if menu.IFrame {
					return "是"
				}
				return "否"
			}(),
			menu.Component,
			menu.ComponentName,
			0,
			menu.Sort,
			menu.Icon,
			func() string {
				if menu.Type == 1 {
					return "目录"
				} else if menu.Type == 2 {
					return "菜单"
				} else if menu.Type == 3 {
					return "按钮"
				}
				return "未知"
			}(),
			func() string {
				if menu.Cache {
					return "是"
				}
				return "否"
			}(),
			func() string {
				if menu.Hidden {
					return "是"
				}
				return "否"
			}(),
			menu.SubCount,
			menu.CreateTime.Format("2006-01-02 15:04:05"),
		}, -1)
	}
	// 保存 Excel 文件到本地
	fileName = "Menus_" + utils.GenerateID().String() + ".xlsx"
	filePath = global.Config.Excel.Dir + fileName
	err = file.Save(filePath)
	return filePath, fileName, err
}

// GetSuperior
// @description: 获取同级、父级菜单
// @receiver: menuService
// @param: id
// @return: menuList
// @return: err
func (menuService *MenuService) GetSuperior(id int64, menuList *[]permission.Menu) error {
	menu := &permission.Menu{}
	var menuMap = make(map[int64][]permission.Menu)
	err := global.Db.Scopes(utils.IsDeleteSoft).First(menu, "id = ? ", id).Error
	if err != nil {
		return err
	}
	menuMap, err = menuService.getSuperior(*menu, []permission.Menu{})

	//var menus = menuMap[0]
	*menuList = menuMap[0]
	//for i := 0; i < len(*menuList); i++ {
	//	err = menuService.getChildrenList((*menuList)[i], menuMap)
	//}
	for i := range *menuList {
		err = menuService.getChildrenList(&(*menuList)[i], menuMap)
	}
	for i := range *menuList {
		(*menuList)[i].CalculateHasChildren()
		(*menuList)[i].CalculateLeaf()
		(*menuList)[i].CalculateLabel()
	}
	return err
}

func (menuService *MenuService) getSuperior(menu permission.Menu, menus []permission.Menu) (menuMap map[int64][]permission.Menu, err error) {
	var menuTmpList []permission.Menu
	menuMap = make(map[int64][]permission.Menu)
	if menu.ParentId == 0 {
		err = global.Db.Scopes(utils.IsDeleteSoft).Where("parent_id = 0 ").Find(&menuTmpList).Error
		if err != nil {
			return
		}
		menus = append(menus, menuTmpList...)
		for _, v := range menus {
			menuMap[v.ParentId] = append(menuMap[v.ParentId], v)
		}
	} else {
		err = global.Db.Scopes(utils.IsDeleteSoft).Where("parent_id = ? ", menu.ParentId).Find(&menuTmpList).Error
		if err != nil {
			return
		}
		menus = append(menus, menuTmpList...)
		var menuTmp permission.Menu
		err = global.Db.Scopes(utils.IsDeleteSoft).Where("id = ? ", menu.ParentId).First(&menuTmp).Error
		if err != nil {
			return
		}
		menuMap, err = menuService.getSuperior(menuTmp, menus)
	}
	return menuMap, err
}

func (menuService *MenuService) GetMenuByParentId(parentId int64, menus *[]permission.Menu) error {
	err := global.Db.Scopes(utils.IsDeleteSoft).Find(&menus, "parent_id = ? ", parentId).Error
	if err != nil {
		return err
	}
	for i := range *menus {
		(*menus)[i].CalculateHasChildren()
		(*menus)[i].CalculateLeaf()
		(*menus)[i].Children = nil
		(*menus)[i].CalculateLabel()
	}
	return nil
}

// GetMenuTree
// @description: 获取用户菜单
// @receiver: menuService
// @param: id
// @param: treeMenus
// @return: error
func (menuService *MenuService) GetMenuTree(id int64, treeMenus *[]response.TreeMenuVo) error {
	var menuTree map[int64][]permission.Menu
	err := menuService.getMenuTreeMap(id, &menuTree)
	if err != nil {
		return err
	}
	var menus = menuTree[0]
	for i := 0; i < len(menus); i++ {
		err = menuService.getChildrenList(&menus[i], menuTree)
		if err != nil {
			return err
		}
	}
	err = menuService.buildTree(menus, treeMenus)
	return err
}

func (menuService *MenuService) buildTree(menus []permission.Menu, treeMenus *[]response.TreeMenuVo) error {
	for _, menu := range menus {
		menuChildren := menu.Children
		treeMenuVo := response.TreeMenuVo{
			Hidden: menu.Hidden,
		}
		if menu.ComponentName == "" {
			treeMenuVo.Name = menu.Title
		} else {
			treeMenuVo.Name = menu.ComponentName
		}
		if menu.ParentId == 0 {
			treeMenuVo.Path = "/" + menu.Path
		} else {
			treeMenuVo.Path = menu.Path
		}
		if !menu.IFrame {
			if menu.ParentId == 0 {
				if menu.Component == "" {
					treeMenuVo.Component = "Layout"
				} else {
					treeMenuVo.Component = menu.Component
				}
			} else if menu.Component != "" {
				treeMenuVo.Component = menu.Component
			}
		}
		treeMenuVo.TreeMenuMeta = response.TreeMenuMeta{Title: menu.Title, Icon: menu.Icon, NoCache: !menu.Cache}

		if len(menuChildren) > 0 {
			treeMenuVo.AlwaysShow = true
			treeMenuVo.Redirect = "noredirect"

			// 递归调用子菜单
			var children []response.TreeMenuVo
			err := menuService.buildTree(menuChildren, &children)
			if err != nil {
				return err
			}
			treeMenuVo.Children = children
		}
		*treeMenus = append(*treeMenus, treeMenuVo)
	}
	return nil
}

func (menuService *MenuService) getMenuTreeMap(userId int64, treeMap *map[int64][]permission.Menu) error {
	var menus []permission.Menu
	var userRoles []permission.UserRole
	var roleMenus []permission.RoleMenu

	// 初始化 map
	*treeMap = make(map[int64][]permission.Menu)

	// 获取用户角色
	err := global.Db.Where("user_id = ?", userId).Find(&userRoles).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		return err
	}

	// 提取角色 ID
	var roleIds []int64
	for i := range userRoles {
		roleIds = append(roleIds, userRoles[i].RoleId)
	}

	// 获取角色关联的菜单
	err = global.Db.Where("role_id in (?)", roleIds).Find(&roleMenus).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		return err
	}

	// 提取菜单 ID
	var menuIds []int64
	for i := range roleMenus {
		menuIds = utils.AppendInt64(menuIds, roleMenus[i].MenuId)
	}

	// 获取菜单
	err = global.Db.Where("id in (?) and type <> '3' ", menuIds).Order("sort").Find(&menus).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		return err
	}

	// 构建 treeMap
	for _, v := range menus {
		(*treeMap)[v.ParentId] = append((*treeMap)[v.ParentId], v)
	}

	return nil
}

func (menuService *MenuService) getChildrenList(menu *permission.Menu, treeMap map[int64][]permission.Menu) (err error) {
	menu.Children = treeMap[menu.Id]
	for i := 0; i < len(menu.Children); i++ {
		err = menuService.getChildrenList(&menu.Children[i], treeMap)
	}
	return err
}

// GetChildrenIds
// @description: 获取所有子集ID
// @receiver: menuService
// @param: ids
// @param: allIds
// @return: allIdList
// @return: err
func (menuService *MenuService) GetChildrenIds(ids []int64, allIds []int64) (allIdList []int64, err error) {
	for _, id := range ids {
		allIds = utils.AppendInt64(allIds, id)
		var menus []permission.Menu
		err = global.Db.Scopes(utils.IsDeleteSoft).Where("parent_id = ? ", id).Find(&menus).Error
		if err == nil && len(menus) > 0 {
			var ids []int64
			for _, d := range menus {
				ids = append(ids, d.Id)
			}
			allIds, err = menuService.GetChildrenIds(ids, allIds)
			if err != nil {
				//return allIds, err
				break
			}
		}
	}
	return allIds, err
}

// buildMenuQuery
// @description: 条件表达式
// @param: db
// @param: info
// @return: *gorm.DB
func buildMenuQuery(db *gorm.DB, info dto.MenuQueryCriteria) *gorm.DB {
	if info.Title != "" {
		db = db.Where("title LIKE ? ", "%"+info.Title+"%")
	}
	if info.ParentId == nil {
		db = db.Where("parent_id = ?", 0)
	} else {
		db = db.Where("parent_id = ?", info.ParentId)
	}
	if len(info.CreateTime) > 0 {
		db = db.Where("create_time >= ? and create_time <= ? ", info.CreateTime[0], info.CreateTime[1])
	}

	db = db.Order("sort").Scopes(utils.IsDeleteSoft)

	return db
}
