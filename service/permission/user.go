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
	"strconv"
	"strings"
)

type UserService struct{}

// Create
// @description: 创建
// @receiver: userService
// @param: req
// @return: error
func (userService *UserService) Create(req *dto.CreateUpdateUserReq) error {
	var user permission.User
	var total int64
	err := global.Db.Model(permission.User{}).Scopes(utils.IsDeleteSoft).Where("username = ?", req.Username).Count(&total).Error
	if err != nil {
		return err
	}
	if total > 0 {
		return errors.New("用户名称=>" + req.Username + "=>已存在!")
	}

	err = global.Db.Model(permission.User{}).Scopes(utils.IsDeleteSoft).Where("nick_name = ?", req.NickName).Count(&total).Error
	if err != nil {
		return err
	}
	if total > 0 {
		return errors.New("昵称=>" + req.NickName + "=>已存在!")
	}

	err = global.Db.Model(permission.User{}).Scopes(utils.IsDeleteSoft).Where("email = ?", req.Email).Count(&total).Error
	if err != nil {
		return err
	}
	if total > 0 {
		return errors.New("邮箱=>" + req.Email + "=>已存在!")
	}

	err = global.Db.Model(permission.User{}).Scopes(utils.IsDeleteSoft).Where("phone = ?", req.Phone).Count(&total).Error
	if err != nil {
		return err
	}
	if total > 0 {
		return errors.New("电话=>" + req.Phone + "=>已存在!")
	}

	req.Generate(&user)
	user.Password = utils.BcryptHash("123456")

	db := global.Db.Begin()
	defer func() {
		if err != nil {
			global.Logger.Error("tran error: ", zap.Error(err))
			db.Rollback()
		} else {
			db.Commit()
		}
	}()

	err = db.Create(&user).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		return err
	}
	var userJobs []permission.UserJob
	for _, j := range req.Jobs {
		uj := permission.UserJob{UserId: user.Id, JobId: j.Id}
		userJobs = append(userJobs, uj)
	}
	err = db.Create(&userJobs).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		return err
	}
	var userRoles []permission.UserRole
	for _, r := range req.Roles {
		ur := permission.UserRole{UserId: user.Id, RoleId: r.Id}
		userRoles = append(userRoles, ur)
	}
	err = db.Create(&userRoles).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		return err
	}
	return nil
}

// Update
// @description: 更新
// @receiver: userService
// @param: req
// @return: error
func (userService *UserService) Update(req *dto.CreateUpdateUserReq) error {
	var user permission.User
	var oldUser permission.User
	err := global.Db.Scopes(utils.IsDeleteSoft).Where("id = ?", req.Id).First(&oldUser).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("数据不存在")
		}
		return err
	}
	var total int64
	if oldUser.Username != req.Username {
		err = global.Db.Model(permission.User{}).Scopes(utils.IsDeleteSoft).Where("username = ?", req.Username).Count(&total).Error
		if err != nil {
			return err
		}
		if total > 0 {
			return errors.New("用户名称=>" + req.Username + "=>已存在!")
		}
	}
	if oldUser.NickName != req.NickName {
		err = global.Db.Model(permission.User{}).Scopes(utils.IsDeleteSoft).Where("nick_name = ?", req.NickName).Count(&total).Error
		if err != nil {
			return err
		}
		if total > 0 {
			return errors.New("昵称=>" + req.NickName + "=>已存在!")
		}
	}
	if oldUser.Email != req.Email {
		err = global.Db.Model(permission.User{}).Scopes(utils.IsDeleteSoft).Where("email = ?", req.Email).Count(&total).Error
		if err != nil {
			return err
		}
		if total > 0 {
			return errors.New("邮箱=>" + req.Email + "=>已存在!")
		}
	}
	if oldUser.Phone != req.Phone {
		err = global.Db.Model(permission.User{}).Scopes(utils.IsDeleteSoft).Where("phone = ?", req.Phone).Count(&total).Error
		if err != nil {
			return err
		}
		if total > 0 {
			return errors.New("电话=>" + req.Phone + "=>已存在!")
		}
	}
	req.Generate(&user)
	db := global.Db.Begin()
	defer func() {
		if err != nil {
			global.Logger.Error("tran error: ", zap.Error(err))
			db.Rollback()
		} else {
			db.Commit()
		}
	}()

	err = global.Db.Updates(&user).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		return err
	}

	err = db.Where("user_id = ?", user.Id).Delete(&permission.UserJob{}).Error
	if err != nil {
		return err
	}

	var userJobs []permission.UserJob
	for _, j := range req.Jobs {
		uj := permission.UserJob{UserId: user.Id, JobId: j.Id}
		userJobs = append(userJobs, uj)
	}
	err = db.Create(&userJobs).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		return err
	}

	err = db.Where("user_id = ?", user.Id).Delete(&permission.UserRole{}).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		return err
	}

	var userRoles []permission.UserRole
	for _, r := range req.Roles {
		ur := permission.UserRole{UserId: user.Id, RoleId: r.Id}
		userRoles = append(userRoles, ur)
	}
	err = db.Create(&userRoles).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		return err
	}
	return nil
}

// Delete
// @description: 删除
// @receiver: userService
// @param: idArray
// @param: userId
// @param: updateBy
// @return: error
func (userService *UserService) Delete(idArray request.IdCollection, userId int64, updateBy string) error {
	if utils.ContainsValue(idArray.IdArray, strconv.FormatInt(userId, 10)) {
		return errors.New("禁止删除自己")
	}
	var users []permission.User
	err := global.Db.Scopes(utils.IsDeleteSoft).Find(&users, "id in ?", idArray.IdArray).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		return err
	}
	if len(users) == 0 {
		return errors.New("数据不存在或您无权查看！")
	}

	localTime := ext.GetCurrentTime()
	err = global.Db.Model(&permission.User{}).Scopes(utils.IsDeleteSoft).Where("id in (?)", idArray.IdArray).Updates(
		permission.User{BaseModel: model.BaseModel{
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
// @receiver: userService
// @param: info
// @param: userId
// @param: deptId
// @param: list
// @param: total
// @return: error
func (userService *UserService) Query(info *dto.UserQueryCriteria, userId int64, deptId int64, list *[]permission.User, count *int64) error {
	limit := info.PageSize
	offset := info.PageSize * (info.PageIndex - 1)
	//var userList []permission.User
	db := buildUserQuery(global.Db.Model(&permission.User{}), info, userId, deptId)

	err := db.Count(count).Limit(limit).Offset(offset).Preload("Jobs").Preload("Roles").Preload("Dept").Find(list).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		return err
	}
	return nil
}

// Download
// @description: 导出
// @receiver: userService
// @param: info
// @param: userId
// @param: deptId
// @return: filePath
// @return: fileName
// @return: err
func (userService *UserService) Download(info *dto.UserQueryCriteria, userId int64, deptId int64) (filePath string, fileName string, err error) {
	var users []permission.User
	db := buildUserQuery(global.Db.Model(&permission.User{}), info, userId, deptId)

	err = db.Preload("Jobs").Preload("Roles").Preload("Dept").Find(&users).Error
	if err != nil {
		return
	}

	// 创建一个新的 Excel 文件
	file := xlsx.NewFile()
	sheet, err := file.AddSheet("Users")
	if err != nil {
		return
	}
	style := xlsx.NewStyle()
	style.Fill = *xlsx.NewFill("solid", "1e90ff", "1e90ff")
	style.ApplyFill = true
	// 表头数据
	header := []string{"ID", "用户名称", "角色名称", "用户昵称", "用户电话", "用户邮箱", "是否激活", "部门名称", "岗位名称", "性别", "创建时间"}
	// 将表头添加到 Excel 工作表
	row := sheet.AddRow()
	for _, cellValue := range header {
		cell := row.AddCell()
		cell.Value = cellValue
		cell.SetStyle(style)
	}

	// 将数据添加到 Excel 工作表
	for _, user := range users {
		row := sheet.AddRow()
		row.WriteSlice(&[]interface{}{
			strconv.FormatInt(user.Id, 10),
			user.Username,
			strings.Join(func() []string {
				var roleNames []string
				for _, role := range user.Roles {
					roleNames = append(roleNames, role.Name)
				}
				return roleNames
			}(), ","),
			user.NickName,
			user.Phone,
			user.Email,
			func() string {
				if user.Enabled {
					return "是"
				}
				return "否"
			}(),
			user.Dept.Name,
			strings.Join(func() []string {
				var jobNames []string
				for _, job := range user.Jobs {
					jobNames = append(jobNames, job.Name)
				}
				return jobNames
			}(), ","),
			user.Gender,
			user.CreateTime.Format("2006-01-02 15:04:05"),
		}, -1)
	}
	// 保存 Excel 文件到本地
	fileName = "Users_" + utils.GenerateID().String() + ".xlsx"
	filePath = global.Config.Excel.Dir + fileName
	err = file.Save(filePath)
	return filePath, fileName, err
}

// @description: 根据名称查找用户
// @receiver: userService
// @param: username
// @return: err

func (userService *UserService) QueryByName(username string, user *permission.User) error {
	err := global.Db.Where("username = ?", username).First(user).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("数据不存在或您无权查看！")
		}
		return err
	}
	return nil
}

// @description: 根据Id查找用户
// @receiver: userService
// @param: userId
// @return: err

func (userService *UserService) QueryById(userId int64, user *permission.User) error {

	err := global.Db.Where("id = ?", userId).Preload("Jobs").Preload("Roles").Preload("Dept").First(user).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("数据不存在或您无权查看！")
		}
		return err
	}
	return nil
}

// buildUserQuery
// @description: 条件表达式
// @param: db
// @param: info
// @param: userId
// @param: deptId
// @return: *gorm.DB
func buildUserQuery(db *gorm.DB, info *dto.UserQueryCriteria, userId int64, deptId int64) *gorm.DB {
	if info.Enabled != nil {
		db = db.Where("enabled = ?", info.Enabled)
	}
	if info.KeyWords != "" {
		db = db.Where("username LIKE ? or nick_name LIKE ? or email LIKE ?", "%"+info.KeyWords+"%", "%"+info.KeyWords+"%", "%"+info.KeyWords+"%")
	}
	if len(info.CreateTime) > 0 {
		db = db.Where("create_time >= ? and create_time <= ? ", info.CreateTime[0], info.CreateTime[1])
	}
	if info.DeptId != nil {
		ids := []int64{*info.DeptId}
		service := DeptService{}
		childrenIds, _ := service.GetChildrenIds(ids, []int64{})
		db = db.Where("dept_id in ?", childrenIds)
	}
	//service := DeptScopeService{}
	//deptList, err := service.GetDataScopeDeptList(userId, deptId)
	//if err != nil {
	//	return db, err
	//}
	//if len(deptList) > 0 {
	//	db = db.Where("dept_id in ?", deptList)
	//}
	if len(info.SortFields) > 0 {
		for _, sort := range info.SortFields {
			db = db.Order(sort)
		}
	}
	return db.Scopes(utils.IsDeleteSoft)
}
