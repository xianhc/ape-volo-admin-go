package initialize

import (
	"encoding/json"
	"errors"
	"fmt"
	"go-apevolo/global"
	"go-apevolo/model/message/email"
	"go-apevolo/model/monitor"
	"go-apevolo/model/permission"
	"go-apevolo/model/queued"
	"go-apevolo/model/system"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"io/ioutil"
)

// InitTable 初始化数据表
func InitTable() error {
	db := global.Db
	err := db.AutoMigrate(
		// monitor模块
		monitor.AuditLog{},
		monitor.ExceptionLog{},

		// permission模块
		permission.Department{},
		permission.Job{},
		permission.Menu{},
		permission.Apis{},
		permission.Role{},
		permission.RoleDepartment{},
		permission.RoleMenu{},
		permission.RoleApis{},
		permission.User{},
		permission.UserJob{},
		permission.UserRole{},

		// system模块
		system.AppSecret{},
		system.Dict{},
		system.DictDetail{},
		system.FileRecord{},
		system.Setting{},
		system.TokenBlacklist{},
		system.Task{},
		system.TaskLog{},

		// 消息模块
		email.Account{},
		email.MessageTemplate{},

		//队列
		queued.Email{},
	)
	if err != nil {
		global.Logger.Error("初始化数据库表失败")
		return err
	}
	return nil
}

// InitTableData 初始化表数据
func InitTableData() error {
	// 用户
	var user permission.User
	err := global.Db.First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		fileData, err := ioutil.ReadFile(fmt.Sprintf("./resource/db/%s.tsv", user.TableName()))
		if err != nil {
			global.Logger.Error("无法读取文件", zap.Error(err))
			return err
		}
		var users []permission.User
		if err := json.Unmarshal(fileData, &users); err != nil {
			global.Logger.Error("解析User数据失败", zap.Error(err))
			return err
		}
		err = global.Db.Create(users).Error
		if err != nil {
			return err
		}
	} else {
		if err != nil {
			global.Logger.Error("db error ", zap.Error(err))
			return err
		}
	}

	// 角色
	var role permission.Role
	err = global.Db.First(&role).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		fileData, err := ioutil.ReadFile(fmt.Sprintf("./resource/db/%s.tsv", role.TableName()))
		if err != nil {
			global.Logger.Error("无法读取文件", zap.Error(err))
			return err
		}
		var roles []permission.Role
		if err := json.Unmarshal(fileData, &roles); err != nil {
			global.Logger.Error("解析Role数据失败", zap.Error(err))
			return err
		}
		err = global.Db.Create(roles).Error
		if err != nil {
			return err
		}
	} else {
		if err != nil {
			global.Logger.Error("db error ", zap.Error(err))
			return err
		}
	}

	// 菜单
	var menu permission.Menu
	err = global.Db.First(&menu).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		fileData, err := ioutil.ReadFile(fmt.Sprintf("./resource/db/%s.tsv", menu.TableName()))
		if err != nil {
			global.Logger.Error("无法读取文件", zap.Error(err))
			return err
		}
		var menus []permission.Menu
		if err := json.Unmarshal(fileData, &menus); err != nil {
			global.Logger.Error("解析Menu数据失败", zap.Error(err))
			return err
		}
		err = global.Db.Create(menus).Error
		if err != nil {
			return err
		}
	} else {
		if err != nil {
			global.Logger.Error("db error ", zap.Error(err))
			return err
		}
	}

	// 部门
	var dept permission.Department
	err = global.Db.First(&dept).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		fileData, err := ioutil.ReadFile(fmt.Sprintf("./resource/db/%s.tsv", dept.TableName()))
		if err != nil {
			global.Logger.Error("无法读取文件", zap.Error(err))
			return err
		}
		var depts []permission.Department
		if err := json.Unmarshal(fileData, &depts); err != nil {
			global.Logger.Error("解析Department数据失败", zap.Error(err))
			return err
		}
		err = global.Db.Create(depts).Error
		if err != nil {
			return err
		}
	} else {
		if err != nil {
			global.Logger.Error("db error ", zap.Error(err))
			return err
		}
	}

	// 岗位
	var job permission.Job
	err = global.Db.First(&job).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		fileData, err := ioutil.ReadFile(fmt.Sprintf("./resource/db/%s.tsv", job.TableName()))
		if err != nil {
			global.Logger.Error("无法读取文件", zap.Error(err))
			return err
		}
		var jobs []permission.Job
		if err := json.Unmarshal(fileData, &jobs); err != nil {
			global.Logger.Error("解析Job数据失败", zap.Error(err))
			return err
		}
		err = global.Db.Create(jobs).Error
		if err != nil {
			return err
		}
	} else {
		if err != nil {
			global.Logger.Error("db error ", zap.Error(err))
			return err
		}
	}

	// 系统设置
	var setting system.Setting
	err = global.Db.First(&setting).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		fileData, err := ioutil.ReadFile(fmt.Sprintf("./resource/db/%s.tsv", setting.TableName()))
		if err != nil {
			global.Logger.Error("无法读取文件", zap.Error(err))
			return err
		}
		var settings []system.Setting
		if err := json.Unmarshal(fileData, &settings); err != nil {
			global.Logger.Error("解析Setting数据失败", zap.Error(err))
			return err
		}
		err = global.Db.Create(settings).Error
		if err != nil {
			return err
		}
	} else {
		if err != nil {
			global.Logger.Error("db error ", zap.Error(err))
			return err
		}
	}

	// 字典
	var dict system.Dict
	err = global.Db.First(&dict).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		fileData, err := ioutil.ReadFile(fmt.Sprintf("./resource/db/%s.tsv", dict.TableName()))
		if err != nil {
			global.Logger.Error("无法读取文件", zap.Error(err))
			return err
		}
		var dicts []system.Dict
		if err := json.Unmarshal(fileData, &dicts); err != nil {
			global.Logger.Error("解析Dict数据失败", zap.Error(err))
			return err
		}
		err = global.Db.Create(dicts).Error
		if err != nil {
			return err
		}
	} else {
		if err != nil {
			global.Logger.Error("db error ", zap.Error(err))
			return err
		}
	}

	// 字典详情
	var dictDetail system.DictDetail
	err = global.Db.First(&dictDetail).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		fileData, err := ioutil.ReadFile(fmt.Sprintf("./resource/db/%s.tsv", dictDetail.TableName()))
		if err != nil {
			global.Logger.Error("无法读取文件", zap.Error(err))
			return err
		}
		var dictDetails []system.DictDetail
		if err := json.Unmarshal(fileData, &dictDetails); err != nil {
			global.Logger.Error("解析DictDetail数据失败", zap.Error(err))
			return err
		}
		err = global.Db.Create(dictDetails).Error
		if err != nil {
			return err
		}
	} else {
		if err != nil {
			global.Logger.Error("db error ", zap.Error(err))
			return err
		}
	}

	// 用户与角色
	var userRole permission.UserRole
	err = global.Db.First(&userRole).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		fileData, err := ioutil.ReadFile(fmt.Sprintf("./resource/db/%s.tsv", userRole.TableName()))
		if err != nil {
			global.Logger.Error("无法读取文件", zap.Error(err))
			return err
		}
		var userRoles []permission.UserRole
		if err := json.Unmarshal(fileData, &userRoles); err != nil {
			global.Logger.Error("解析UserRole数据失败", zap.Error(err))
			return err
		}
		err = global.Db.Create(userRoles).Error
		if err != nil {
			return err
		}
	} else {
		if err != nil {
			global.Logger.Error("db error ", zap.Error(err))
			return err
		}
	}

	// 用户与岗位
	var userJob permission.UserJob
	err = global.Db.First(&userJob).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		fileData, err := ioutil.ReadFile(fmt.Sprintf("./resource/db/%s.tsv", userJob.TableName()))
		if err != nil {
			global.Logger.Error("无法读取文件", zap.Error(err))
			return err
		}
		var userJobs []permission.UserJob
		if err := json.Unmarshal(fileData, &userJobs); err != nil {
			global.Logger.Error("解析UserJob数据失败", zap.Error(err))
			return err
		}
		err = global.Db.Create(userJobs).Error
		if err != nil {
			return err
		}
	} else {
		if err != nil {
			global.Logger.Error("db error ", zap.Error(err))
			return err
		}
	}

	// 角色与菜单
	var roleMenu permission.RoleMenu
	err = global.Db.First(&roleMenu).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		fileData, err := ioutil.ReadFile(fmt.Sprintf("./resource/db/%s.tsv", roleMenu.TableName()))
		if err != nil {
			global.Logger.Error("无法读取文件", zap.Error(err))
			return err
		}
		var roleMenus []permission.RoleMenu
		if err := json.Unmarshal(fileData, &roleMenus); err != nil {
			global.Logger.Error("解析RoleMenu数据失败", zap.Error(err))
			return err
		}
		err = global.Db.Create(roleMenus).Error
		if err != nil {
			return err
		}
	} else {
		if err != nil {
			global.Logger.Error("db error ", zap.Error(err))
			return err
		}
	}

	// Apis
	var apis permission.Apis
	err = global.Db.First(&apis).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		fileData, err := ioutil.ReadFile(fmt.Sprintf("./resource/db/%s.tsv", apis.TableName()))
		if err != nil {
			global.Logger.Error("无法读取文件", zap.Error(err))
			return err
		}
		var apisList []permission.Apis
		if err := json.Unmarshal(fileData, &apisList); err != nil {
			global.Logger.Error("解析Apis数据失败", zap.Error(err))
			return err
		}
		err = global.Db.Create(apisList).Error
		if err != nil {
			return err
		}
	} else {
		if err != nil {
			global.Logger.Error("db error ", zap.Error(err))
			return err
		}
	}

	// 角色与Apis
	var roleApis permission.RoleApis
	err = global.Db.First(&roleApis).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		fileData, err := ioutil.ReadFile(fmt.Sprintf("./resource/db/%s.tsv", roleApis.TableName()))
		if err != nil {
			global.Logger.Error("无法读取文件", zap.Error(err))
			return err
		}
		var roleApisList []permission.RoleApis
		if err := json.Unmarshal(fileData, &roleApisList); err != nil {
			global.Logger.Error("解析Apis数据失败", zap.Error(err))
			return err
		}
		err = global.Db.Create(roleApisList).Error
		if err != nil {
			return err
		}
	} else {
		if err != nil {
			global.Logger.Error("db error ", zap.Error(err))
			return err
		}
	}

	// 邮箱账户
	var emailAccount email.Account
	err = global.Db.First(&emailAccount).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		fileData, err := ioutil.ReadFile(fmt.Sprintf("./resource/db/%s.tsv", emailAccount.TableName()))
		if err != nil {
			global.Logger.Error("无法读取文件", zap.Error(err))
			return err
		}
		var emailAccountList []email.Account
		if err := json.Unmarshal(fileData, &emailAccountList); err != nil {
			global.Logger.Error("解析Apis数据失败", zap.Error(err))
			return err
		}
		err = global.Db.Create(emailAccountList).Error
		if err != nil {
			return err
		}
	} else {
		if err != nil {
			global.Logger.Error("db error ", zap.Error(err))
			return err
		}
	}

	// 邮件模板
	var emailMessageTemplate email.MessageTemplate
	err = global.Db.First(&emailMessageTemplate).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		fileData, err := ioutil.ReadFile(fmt.Sprintf("./resource/db/%s.tsv", emailMessageTemplate.TableName()))
		if err != nil {
			global.Logger.Error("无法读取文件", zap.Error(err))
			return err
		}
		var emailMessageTemplateList []email.MessageTemplate
		if err := json.Unmarshal(fileData, &emailMessageTemplateList); err != nil {
			global.Logger.Error("解析Apis数据失败", zap.Error(err))
			return err
		}
		err = global.Db.Create(emailMessageTemplateList).Error
		if err != nil {
			return err
		}
	} else {
		if err != nil {
			global.Logger.Error("db error ", zap.Error(err))
			return err
		}
	}

	// 调度任务
	var task system.Task
	err = global.Db.First(&task).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		fileData, err := ioutil.ReadFile(fmt.Sprintf("./resource/db/%s.tsv", task.TableName()))
		if err != nil {
			global.Logger.Error("无法读取文件", zap.Error(err))
			return err
		}
		var taskList []system.Task
		if err := json.Unmarshal(fileData, &taskList); err != nil {
			global.Logger.Error("解析调度任务数据失败", zap.Error(err))
			return err
		}
		err = global.Db.Create(taskList).Error
		if err != nil {
			return err
		}
	} else {
		if err != nil {
			global.Logger.Error("db error ", zap.Error(err))
			return err
		}
	}

	return nil
}
