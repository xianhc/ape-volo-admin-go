package permission

import (
	"errors"
	"github.com/tealeg/xlsx"
	"go-apevolo/global"
	"go-apevolo/model"
	"go-apevolo/model/permission"
	"go-apevolo/payloads/dto"
	"go-apevolo/payloads/request"
	"go-apevolo/utils"
	"go-apevolo/utils/ext"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"math"
	"strconv"
	"strings"
)

type RoleService struct{}

// Create
// @description: 创建
// @receiver: roleService
// @param: req
// @return: error
func (roleService *RoleService) Create(req *dto.CreateUpdateRoleDto) error {
	role := &permission.Role{}
	var total int64
	err := global.Db.Model(permission.Role{}).Scopes(utils.IsDeleteSoft).Where("name = ?", req.Name).Count(&total).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		return err
	}
	if total > 0 {
		return errors.New("名称=>" + req.Name + "=>已存在!")
	}

	err = global.Db.Model(permission.Role{}).Scopes(utils.IsDeleteSoft).Where("permission = ?", req.Permission).Count(&total).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		return err
	}
	if total > 0 {
		return errors.New("标识=>" + req.Permission + "=>已存在!")
	}
	req.Generate(role)
	db := global.Db.Begin()
	defer func() {
		if err != nil {
			global.Logger.Error("tran error: ", zap.Error(err))
			db.Rollback()
		} else {
			db.Commit()
		}
	}()

	err = db.Create(role).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		return err
	}
	if len(req.Departments) > 0 {
		var roleDepartments []permission.RoleDepartment
		for _, d := range req.Departments {
			rd := permission.RoleDepartment{RoleId: role.Id, DeptId: d.Id}
			roleDepartments = append(roleDepartments, rd)
		}
		err = db.Create(&roleDepartments).Error
		if err != nil {
			global.Logger.Error("db error: ", zap.Error(err))
			return err
		}
	}
	return nil
}

// Update
// @description: 修改
// @receiver: roleService
// @param: req
// @return: error
func (roleService *RoleService) Update(req *dto.CreateUpdateRoleDto) error {
	oldRole := &permission.Role{}
	role := &permission.Role{}

	err := global.Db.Scopes(utils.IsDeleteSoft).First(&oldRole, req.Id).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("数据不存在或您无权查看！")
		}
		return err
	}
	var total int64
	if oldRole.Name != req.Name {
		err = global.Db.Model(permission.Role{}).Scopes(utils.IsDeleteSoft).Where("name = ?", req.Name).Count(&total).Error
		if err != nil {
			global.Logger.Error("db error: ", zap.Error(err))
			return err
		}
		if total > 0 {
			return errors.New("名称=>" + req.Name + "=>已存在!")
		}
	}
	if oldRole.Permission != req.Permission {
		err = global.Db.Model(permission.Role{}).Scopes(utils.IsDeleteSoft).Where("permission = ?", req.Permission).Count(&total).Error
		if err != nil {
			global.Logger.Error("db error: ", zap.Error(err))
			return err
		}
		if total > 0 {
			return errors.New("标识=>" + req.Permission + "=>已存在!")
		}
	}
	req.Generate(role)
	db := global.Db.Begin()
	defer func() {
		if err != nil {
			global.Logger.Error("tran error: ", zap.Error(err))
			db.Rollback()
		} else {
			db.Commit()
		}
	}()

	err = db.Create(role).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		return err
	}

	err = db.Where("role_id = ?", role.Id).Delete(&permission.RoleDepartment{}).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		return err
	}
	if len(req.Departments) > 0 {
		var roleDepartments []permission.RoleDepartment
		for _, d := range req.Departments {
			rd := permission.RoleDepartment{RoleId: role.Id, DeptId: d.Id}
			roleDepartments = append(roleDepartments, rd)
		}
		err = db.Create(&roleDepartments).Error
		if err != nil {
			global.Logger.Error("db error: ", zap.Error(err))
			return err
		}
	}
	return nil
}

// Delete
// @description: 删除
// @receiver: roleService
// @param: idArray
// @param: userId
// @param: updateBy
// @return: error
func (roleService *RoleService) Delete(idArray request.IdCollection, userId int64, updateBy string) error {
	var roles []permission.Role
	err := global.Db.Scopes(utils.IsDeleteSoft).Preload("Users").Find(&roles, "id in ?", idArray.IdArray).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		return err
	}
	for i := range roles {
		if len(roles[i].Users) > 0 {
			return errors.New("存在用户关联，请解除后再试！")
		}
	}

	minLevel := math.MaxInt32 // 初始化为最大整数值
	for _, role := range roles {
		if role.Level < minLevel {
			minLevel = role.Level
		}
	}

	_, err = roleService.VerificationUserRoleLevelAsync(userId, &minLevel)
	if err != nil {
		return err
	}

	localTime := ext.GetCurrentTime()
	err = global.Db.Model(&permission.Role{}).Scopes(utils.IsDeleteSoft).Where("id in (?)", idArray.IdArray).Updates(
		permission.Role{BaseModel: model.BaseModel{
			UpdateBy:   &updateBy,
			UpdateTime: &localTime,
		}, SoftDeleted: model.SoftDeleted{IsDeleted: true}},
	).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		return err
	}
	return nil
}

// Query
// @description: 查询
// @receiver: roleService
// @param: info
// @return: list
// @return: total
// @return: err
func (roleService *RoleService) Query(info *dto.RoleQueryCriteria, list *[]permission.Role, count *int64) error {
	limit := info.PageSize
	offset := info.PageSize * (info.PageIndex - 1)
	// 创建db
	db := buildRoleQuery(global.Db.Model(&permission.Role{}), info)
	err := db.Count(count).Limit(limit).Offset(offset).Preload("Menus").Preload("Apis").Find(list).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		return err
	}
	return nil
}

// Download
// @description: 导出
// @receiver: roleService
// @param: info
// @return: filePath
// @return: fileName
// @return: err
func (roleService *RoleService) Download(info *dto.RoleQueryCriteria) (filePath string, fileName string, err error) {

	var roles []permission.Role
	// 创建db
	err = buildRoleQuery(global.Db.Model(&permission.Role{}), info).Preload("Departments").Find(&roles).Error
	if err != nil {
		return
	}
	// 创建一个新的 Excel 文件
	file := xlsx.NewFile()
	sheet, err := file.AddSheet("Roles")
	if err != nil {
		return
	}
	style := xlsx.NewStyle()
	style.Fill = *xlsx.NewFill("solid", "1e90ff", "1e90ff")
	style.ApplyFill = true
	// 表头数据
	header := []string{"ID", "角色名称", "角色等级", "角色描述", "数据范围", "数据部门", "角色代码", "创建时间"}
	// 将表头添加到 Excel 工作表
	row := sheet.AddRow()
	for _, cellValue := range header {
		cell := row.AddCell()
		cell.Value = cellValue
		cell.SetStyle(style)
	}

	// 将数据添加到 Excel 工作表
	for _, role := range roles {
		row := sheet.AddRow()
		row.WriteSlice(&[]interface{}{
			strconv.FormatInt(role.Id, 10),
			role.Name,
			role.Level,
			role.Description,
			role.DataScope,
			strings.Join(func() []string {
				var deptNames []string
				for _, dept := range role.Departments {
					deptNames = append(deptNames, dept.Name)
				}
				return deptNames
			}(), ","),
			role.Permission,
			role.CreateTime.Format("2006-01-02 15:04:05"),
		}, -1)
	}
	// 保存 Excel 文件到本地
	fileName = "Roles_" + utils.GenerateID().String() + ".xlsx"
	filePath = global.Config.Excel.Dir + fileName
	err = file.Save(filePath)
	return filePath, fileName, err
}

// QuerySingle
// @description: 获取单个角色
// @receiver: roleService
// @param: roleId
// @return: role
// @return: err
func (roleService *RoleService) QuerySingle(roleId int64, role *permission.Role) error {
	err := global.Db.Where("id = ? ", roleId).Preload("Menus").Preload("Apis").Preload("Departments").First(role).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("数据不存在或您无权查看！")
		}
		return err
	}
	return nil
}

// GetAllRole
// @description: 获取全部角色
// @receiver: roleService
// @return: roles
// @return: err
func (roleService *RoleService) GetAllRole(list *[]permission.Role) error {
	err := global.Db.Preload("Menus").Preload("Departments").Find(list).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		return err
	}
	return nil
}

// VerificationUserRoleLevelAsync
// @description: 验证用户等级权限
// @receiver: roleService
// @param: userId
// @param: level
// @return: minLevel
// @return: err
func (roleService *RoleService) VerificationUserRoleLevelAsync(userId int64, level *int) (minLevel int, err error) {
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

// buildRoleQuery
// @description: 条件表达式
// @param: db
// @param: info
// @return: *gorm.DB
func buildRoleQuery(db *gorm.DB, info *dto.RoleQueryCriteria) *gorm.DB {
	if info.RoleName != "" {
		db = db.Where("name LIKE ?", "%"+info.RoleName+"%")
	}
	if len(info.CreateTime) > 0 {
		db = db.Where("create_time >= ? and create_time <= ? ", info.CreateTime[0], info.CreateTime[1])
	}
	if len(info.SortFields) > 0 {
		for _, sort := range info.SortFields {
			db = db.Order(sort)
		}
	}
	db = db.Scopes(utils.IsDeleteSoft)

	return db
}
