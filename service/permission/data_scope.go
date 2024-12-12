package permission

import (
	"errors"
	"go-apevolo/global"
	"go-apevolo/model/permission"
	"go-apevolo/utils"
)

type DeptScopeService struct{}

func (deptScopeService *DeptScopeService) GetDataScopeDeptList(userId int64, deptId int64) (allIdList []int64, err error) {

	var roleDepts []permission.RoleDepartment
	var userRoles []permission.UserRole
	var roles []permission.Role

	err = global.Db.Where("user_id = ?", userId).Find(&userRoles).Error
	if err != nil {
		return nil, err
	}
	var roleIds []int64
	for _, ur := range userRoles {
		roleIds = append(roleIds, ur.RoleId)
	}
	err = global.Db.Scopes(utils.IsDeleteSoft).Where("id in ?", roleIds).Find(&roles).Error
	if err != nil {
		return nil, err
	}
	if len(roles) == 0 {
		return nil, errors.New("无可用角色")
	}
	var isAll = false
	for _, role := range roles {
		if role.DataScope == "全部" {
			isAll = true
			break
		}
	}
	if isAll {
		return allIdList, nil
	}
	service := DeptService{}
	for _, role := range roles {
		if role.DataScope == "本级" {
			ids, err := service.GetChildrenIds([]int64{deptId}, []int64{})
			if err != nil {
				break
			}
			for _, id := range ids {
				allIdList = utils.AppendInt64(allIdList, id)
			}
		} else if role.DataScope == "自定义" {
			err = global.Db.Where("role_id = ?", role.Id).Find(&roleDepts).Error
			if err != nil {
				break
			}
			var ids []int64
			for _, rd := range roleDepts {
				ids = utils.AppendInt64(ids, rd.DeptId)
			}
			ids, err := service.GetChildrenIds(ids, []int64{})
			if err != nil {
				break
			}
			for _, id := range ids {
				allIdList = utils.AppendInt64(allIdList, id)
			}
		} else {
			allIdList = utils.AppendInt64(allIdList, 0)
		}
	}
	if len(allIdList) == 0 {
		allIdList = utils.AppendInt64(allIdList, 0)
	}
	return allIdList, nil
}
