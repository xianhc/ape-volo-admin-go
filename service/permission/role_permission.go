package permission

import (
	"errors"
	"go-apevolo/global"
	"go-apevolo/model/permission"
	"go-apevolo/payloads/dto"
	"go-apevolo/payloads/vo"
	"go-apevolo/utils"
	"go-apevolo/utils/ext"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"strconv"
)

type RolePermissionService struct{}

// GetPermissionIdentifier
// @description: 获取权限标识符
// @receiver: rolePermissionService
// @param: userId
// @param: roleIds
// @param: permissionRoles
// @return: error
func (rolePermissionService *RolePermissionService) GetPermissionIdentifier(userId int64, roleIds []int64, permissionIdentifierList *[]string) error {
	var roles []permission.Role
	if len(roleIds) == 0 {
		var user permission.User
		err := global.Db.Where("id = ? ", userId).Preload("Roles").First(&user).Error
		if err != nil {
			global.Logger.Error("db error: ", zap.Error(err))
			return err
		}

		for i := range user.Roles {
			roleIds = append(
				roleIds, user.Roles[i].Id,
			)
		}
	}
	err := global.Db.Where("id in (?) ", roleIds).Preload("Menus").Find(&roles).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		return err
	}
	for _, role := range roles {
		*permissionIdentifierList = utils.AppendIfNotExists(*permissionIdentifierList, role.Permission)
		for _, menu := range role.Menus {
			*permissionIdentifierList = utils.AppendIfNotExists(*permissionIdentifierList, menu.Permission)
		}
	}
	return nil
}

// GetUrlAccessControl
// @description: 获取权限Url
// @receiver: rolePermissionService
// @param: userId
// @param: permissionVos
// @return: error
func (rolePermissionService *RolePermissionService) GetUrlAccessControl(userId int64, urlAccessControlList *[]vo.UrlAccessControl) error {

	var roles []permission.Role
	var user permission.User
	var roleIds []int64
	err := global.Db.Where("id = ? ", userId).Preload("Roles").First(&user).Error
	if err != nil {
		return err
	}

	for i := range user.Roles {
		roleIds = append(roleIds, user.Roles[i].Id)
	}

	err = global.Db.Where("id in (?) ", roleIds).Preload("Apis").Find(&roles).Error
	if err != nil {
		return err
	}
	for _, role := range roles {
		for _, api := range role.Apis {
			*urlAccessControlList = append(*urlAccessControlList, vo.UrlAccessControl{
				Url:    api.Url,
				Method: api.Method,
			})
		}
	}
	return nil
}

// GetAllMenus
// @description: 获取全部菜单
// @receiver: rolePermissionService
// @return: menuTree
// @return: err
func (rolePermissionService *RolePermissionService) GetAllMenus() (menuTree []permission.Menu, err error) {
	var menus []permission.Menu
	treeMap := make(map[int64][]permission.Menu)
	err = global.Db.Scopes(utils.IsDeleteSoft).Find(&menus).Error
	if err != nil {
		return
	}
	var zeroCount int
	for _, v := range menus {
		if v.ParentId == 0 {
			zeroCount++
		}
		treeMap[v.ParentId] = append(treeMap[v.ParentId], v)
	}

	menuTree = treeMap[0]
	for i := 0; i < zeroCount; i++ {
		err = rolePermissionService.getMenuChildren(&menuTree[i], treeMap)
	}
	for i := range menuTree {
		menuTree[i].CalculateHasChildren()
		menuTree[i].CalculateLeaf()
		menuTree[i].CalculateLabel()
	}
	return menuTree, err
}

// GetAllApis
// @description: 获取全部Api
// @receiver: rolePermissionService
// @return: apiTree
// @return: err
func (rolePermissionService *RolePermissionService) GetAllApis() (apiTree []vo.ApisTree, err error) {
	var apis []permission.Apis
	err = global.Db.Scopes(utils.IsDeleteSoft).Order("id").Find(&apis).Error
	if err != nil {
		return
	}

	// 使用map进行GroupBy操作
	groupedData := make(map[string][]permission.Apis)

	for _, api := range apis {
		groupedData[api.Group] = append(groupedData[api.Group], api)
	}

	var index = 0
	for group, groupApis := range groupedData {
		var apisTreesTmp []vo.ApisTree
		for _, api := range groupApis {
			apisTreesTmp = append(apisTreesTmp, vo.ApisTree{
				Id:          ext.Int64ToString(api.Id),
				Label:       api.Description,
				Leaf:        true,
				HasChildren: false,
				Children:    nil,
			})
		}
		index++
		apiTree = append(apiTree, vo.ApisTree{
			Id:          strconv.Itoa(index),
			Label:       group,
			Leaf:        false,
			HasChildren: true,
			Children:    apisTreesTmp,
		})
	}

	return apiTree, err
}

func (rolePermissionService *RolePermissionService) getMenuChildren(menu *permission.Menu, treeMap map[int64][]permission.Menu) (err error) {
	menu.Children = treeMap[menu.Id]
	for i := range menu.Children {
		menu.Children[i].CalculateHasChildren()
		menu.Children[i].CalculateLeaf()
		menu.Children[i].CalculateLabel()
	}
	for i := 0; i < len(menu.Children); i++ {
		err = rolePermissionService.getMenuChildren(&menu.Children[i], treeMap)
	}
	return err
}

// UpdateRolesMenus
// @description: 更新角色菜单关联
// @receiver: rolePermissionService
// @param: req
// @param: userId
// @return: err
func (rolePermissionService *RolePermissionService) UpdateRolesMenus(req *dto.CreateUpdateRoleDto, userId int64) error {
	var role permission.Role
	var roleMenu permission.RoleMenu
	var roleMenuList []permission.RoleMenu
	err := global.Db.Scopes(utils.IsDeleteSoft).First(&role, req.Id).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("数据不存在或您无权查看！")
		}
		return err
	}
	_, err = rolePermissionService.verificationUserRoleLevelAsync(userId, &role.Level)
	if err != nil {
		return err
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
	if req.Menus != nil && len(req.Menus) > 0 {
		err = db.Where("role_id = ?", req.Id).Delete(&roleMenu).Error
		if err != nil {
			global.Logger.Error("db error: ", zap.Error(err))
			return err
		}
		for _, menu := range req.Menus {
			roleMenuList = append(roleMenuList, permission.RoleMenu{RoleId: req.Id, MenuId: menu.Id})
		}

		err = db.Create(&roleMenuList).Error
		if err != nil {
			global.Logger.Error("db error: ", zap.Error(err))
			return err
		}
	}
	return nil
}

// UpdateRolesApis
// @description: 更新角色Apis关联
// @receiver: rolePermissionService
// @param: req
// @param: userId
// @return: err
func (rolePermissionService *RolePermissionService) UpdateRolesApis(req *dto.CreateUpdateRoleDto, userId int64) error {

	var role permission.Role
	var roleApis permission.RoleApis
	var roleApisList []permission.RoleApis
	err := global.Db.Scopes(utils.IsDeleteSoft).First(&role, req.Id).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("数据不存在或您无权查看！")
		}
		return err
	}
	_, err = rolePermissionService.verificationUserRoleLevelAsync(userId, &role.Level)
	if err != nil {
		return err
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
	if req.Apis != nil && len(req.Apis) > 0 {
		err = db.Where("role_id = ?", req.Id).Delete(&roleApis).Error
		if err != nil {
			global.Logger.Error("db error: ", zap.Error(err))
			return err
		}
		for _, api := range req.Apis {
			if api.Id > 10000 {
				roleApisList = append(roleApisList, permission.RoleApis{RoleId: req.Id, ApisId: api.Id})
			}
		}

		err = db.Create(&roleApisList).Error
		if err != nil {
			global.Logger.Error("db error: ", zap.Error(err))
			return err
		}
	}
	return nil
}

// verificationUserRoleLevelAsync
// @description: 验证用户等级权限
// @receiver: rolePermissionService
// @param: userId
// @param: level
// @return: minLevel
// @return: err
func (rolePermissionService *RolePermissionService) verificationUserRoleLevelAsync(userId int64, level *int) (minLevel int, err error) {
	var userRoles []permission.UserRole

	err = global.Db.Where("user_id = ?", userId).Find(&userRoles).Error
	if err != nil {
		return -1, err
	}

	var rIds []int64

	for _, ur := range userRoles {
		rIds = utils.AppendInt64(rIds, ur.RoleId)
	}
	var role permission.Role

	err = global.Db.Scopes(utils.IsDeleteSoft).Where("id in ?", rIds).Order("level asc").First(&role).Error

	if err != nil {
		return -1, err //errors.New("您无权修改或删除比你角色等级更高的数据！")
	}
	if level != nil && *level < role.Level {
		return -1, errors.New("您无权修改或删除比你角色等级更高的数据！")
	}
	return role.Level, err
}
