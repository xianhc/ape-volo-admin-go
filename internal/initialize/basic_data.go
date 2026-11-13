package initialize

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"reflect"
	"strings"

	"github.com/xianhc/ape-volo-admin-go/internal/global"
	menumodule "github.com/xianhc/ape-volo-admin-go/internal/modules/identity/menu"
	rolemodule "github.com/xianhc/ape-volo-admin-go/internal/modules/identity/role"
	usermodule "github.com/xianhc/ape-volo-admin-go/internal/modules/identity/user"
	"github.com/xianhc/ape-volo-admin-go/internal/modules/message/emailaccount"
	"github.com/xianhc/ape-volo-admin-go/internal/modules/message/emailtemplate"
	"github.com/xianhc/ape-volo-admin-go/internal/modules/message/queuedemail"
	"github.com/xianhc/ape-volo-admin-go/internal/modules/organization/department"
	"github.com/xianhc/ape-volo-admin-go/internal/modules/organization/position"
	"github.com/xianhc/ape-volo-admin-go/internal/modules/system/appsecret"
	dictmodule "github.com/xianhc/ape-volo-admin-go/internal/modules/system/dict"
	"github.com/xianhc/ape-volo-admin-go/internal/modules/system/filerecord"
	jobmodule "github.com/xianhc/ape-volo-admin-go/internal/modules/system/job"
	settingmodule "github.com/xianhc/ape-volo-admin-go/internal/modules/system/setting"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// InitTable 初始化数据表
func InitTable() error {
	db := global.Db
	err := db.AutoMigrate(
		// 组织模块
		department.Department{},
		position.Position{},
		position.UserPosition{},

		// 身份模块
		menumodule.Menu{},
		rolemodule.Role{},
		rolemodule.RoleDepartment{},
		menumodule.RoleMenu{},
		usermodule.User{},
		usermodule.UserRole{},

		// system模块
		appsecret.AppSecret{},
		dictmodule.Dict{},
		dictmodule.DictDetail{},
		filerecord.FileRecord{},
		settingmodule.Setting{},
		jobmodule.Job{},

		// 消息模块
		emailaccount.Account{},
		emailtemplate.MessageTemplate{},

		//队列
		queuedemail.QueuedEmail{},
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
	var user usermodule.User
	err := global.Db.First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		fileData, err := ioutil.ReadFile(fmt.Sprintf("./resource/db/%s.json", user.TableName()))
		if err != nil {
			global.Logger.Error("无法读取文件", zap.Error(err))
			return err
		}
		var users []usermodule.User
		if err := unmarshalSeed(fileData, &users); err != nil {
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
	var role rolemodule.Role
	err = global.Db.First(&role).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		fileData, err := ioutil.ReadFile(fmt.Sprintf("./resource/db/%s.json", role.TableName()))
		if err != nil {
			global.Logger.Error("无法读取文件", zap.Error(err))
			return err
		}
		var roles []rolemodule.Role
		if err := unmarshalSeed(fileData, &roles); err != nil {
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
	var menu menumodule.Menu
	err = global.Db.First(&menu).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		fileData, err := ioutil.ReadFile(fmt.Sprintf("./resource/db/%s.json", menu.TableName()))
		if err != nil {
			global.Logger.Error("无法读取文件", zap.Error(err))
			return err
		}
		var menus []menumodule.Menu
		if err := unmarshalSeed(fileData, &menus); err != nil {
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
	var dept department.Department
	err = global.Db.First(&dept).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		fileData, err := ioutil.ReadFile(fmt.Sprintf("./resource/db/%s.json", dept.TableName()))
		if err != nil {
			global.Logger.Error("无法读取文件", zap.Error(err))
			return err
		}
		var depts []department.Department
		if err := unmarshalSeed(fileData, &depts); err != nil {
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
	var job position.Position
	err = global.Db.First(&job).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		fileData, err := ioutil.ReadFile(fmt.Sprintf("./resource/db/%s.json", job.TableName()))
		if err != nil {
			global.Logger.Error("无法读取文件", zap.Error(err))
			return err
		}
		var jobs []position.Position
		if err := unmarshalSeed(fileData, &jobs); err != nil {
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
	var setting settingmodule.Setting
	err = global.Db.First(&setting).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		fileData, err := ioutil.ReadFile(fmt.Sprintf("./resource/db/%s.json", setting.TableName()))
		if err != nil {
			global.Logger.Error("无法读取文件", zap.Error(err))
			return err
		}
		var settings []settingmodule.Setting
		if err := unmarshalSeed(fileData, &settings); err != nil {
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
	var dict dictmodule.Dict
	err = global.Db.First(&dict).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		fileData, err := ioutil.ReadFile(fmt.Sprintf("./resource/db/%s.json", dict.TableName()))
		if err != nil {
			global.Logger.Error("无法读取文件", zap.Error(err))
			return err
		}
		var dicts []dictmodule.Dict
		if err := unmarshalSeed(fileData, &dicts); err != nil {
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
	var dictDetail dictmodule.DictDetail
	err = global.Db.First(&dictDetail).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		fileData, err := ioutil.ReadFile(fmt.Sprintf("./resource/db/%s.json", dictDetail.TableName()))
		if err != nil {
			global.Logger.Error("无法读取文件", zap.Error(err))
			return err
		}
		var dictDetails []dictmodule.DictDetail
		if err := unmarshalSeed(fileData, &dictDetails); err != nil {
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
	var userRole usermodule.UserRole
	err = global.Db.First(&userRole).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		fileData, err := ioutil.ReadFile(fmt.Sprintf("./resource/db/%s.json", userRole.TableName()))
		if err != nil {
			global.Logger.Error("无法读取文件", zap.Error(err))
			return err
		}
		var userRoles []usermodule.UserRole
		if err := unmarshalSeed(fileData, &userRoles); err != nil {
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
	var userJob position.UserPosition
	err = global.Db.First(&userJob).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		fileData, err := ioutil.ReadFile(fmt.Sprintf("./resource/db/%s.json", userJob.TableName()))
		if err != nil {
			global.Logger.Error("无法读取文件", zap.Error(err))
			return err
		}
		var userJobs []position.UserPosition
		if err := unmarshalSeed(fileData, &userJobs); err != nil {
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
	var roleMenu menumodule.RoleMenu
	err = global.Db.First(&roleMenu).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		fileData, err := ioutil.ReadFile(fmt.Sprintf("./resource/db/%s.json", roleMenu.TableName()))
		if err != nil {
			global.Logger.Error("无法读取文件", zap.Error(err))
			return err
		}
		var roleMenus []menumodule.RoleMenu
		if err := unmarshalSeed(fileData, &roleMenus); err != nil {
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

	// 邮箱账户
	var emailAccount emailaccount.Account
	err = global.Db.First(&emailAccount).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		fileData, err := ioutil.ReadFile(fmt.Sprintf("./resource/db/%s.json", emailAccount.TableName()))
		if err != nil {
			global.Logger.Error("无法读取文件", zap.Error(err))
			return err
		}
		var emailAccountList []emailaccount.Account
		if err := unmarshalSeed(fileData, &emailAccountList); err != nil {
			global.Logger.Error("解析邮箱账户数据失败", zap.Error(err))
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
	var emailMessageTemplate emailtemplate.MessageTemplate
	err = global.Db.First(&emailMessageTemplate).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		fileData, err := ioutil.ReadFile(fmt.Sprintf("./resource/db/%s.json", emailMessageTemplate.TableName()))
		if err != nil {
			global.Logger.Error("无法读取文件", zap.Error(err))
			return err
		}
		var emailMessageTemplateList []emailtemplate.MessageTemplate
		if err := unmarshalSeed(fileData, &emailMessageTemplateList); err != nil {
			global.Logger.Error("解析邮件模板数据失败", zap.Error(err))
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
	var task jobmodule.Job
	err = global.Db.First(&task).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		fileData, err := ioutil.ReadFile(fmt.Sprintf("./resource/db/%s.json", task.TableName()))
		if err != nil {
			global.Logger.Error("无法读取文件", zap.Error(err))
			return err
		}
		var taskList []jobmodule.Job
		if err := unmarshalSeed(fileData, &taskList); err != nil {
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

// seedKeyRenames 上游实体与 Go 模型之间仅存的字段命名差异。
//
// 其余字段名与上游实体一致（JSON 匹配本身大小写不敏感），
// 这两个改名在解析前统一映射，避免逐表维护两份数据。
var seedKeyRenames = map[string]string{
	"cache":    "keepAlive", // 菜单：上游 Cache -> Go KeepAlive
	"isactive": "enabled",   // 邮件模板：上游 IsActive -> Go Enabled
}

// unmarshalSeed 解析上游导出的种子 JSON。
//
// 上游实体用 long 主键，导出的 JSON 是数字；Go 模型的 id/parentId/creatorId 等
// 字段带 `,string` 标签（对外契约要求 long 序列化为字符串），直接反序列化会报
// 「cannot unmarshal number into Go struct field」。这里按目标类型反射出这些字段，
// 统一把数字补成字符串；布尔字段必须在种子中使用 true/false，不接受数字。
func unmarshalSeed(fileData []byte, target interface{}) error {
	fileData = bytes.TrimPrefix(fileData, []byte("\xef\xbb\xbf"))

	var records []map[string]json.RawMessage
	if err := json.Unmarshal(fileData, &records); err != nil {
		return err
	}

	stringKeys := seedStringFields(reflect.TypeOf(target).Elem().Elem())
	for i, record := range records {
		normalized := make(map[string]json.RawMessage, len(record))
		for key, value := range record {
			name := strings.ToLower(key)
			if renamed, ok := seedKeyRenames[name]; ok {
				key = renamed
				name = strings.ToLower(renamed)
			}
			trimmed := bytes.TrimSpace(value)
			if stringKeys[name] && len(trimmed) > 0 && !bytes.Equal(trimmed, []byte("null")) && trimmed[0] != '"' {
				value = json.RawMessage(`"` + string(trimmed) + `"`)
			}
			normalized[key] = value
		}
		records[i] = normalized
	}

	normalized, err := json.Marshal(records)
	if err != nil {
		return err
	}
	return json.Unmarshal(normalized, target)
}

// seedStringFields 收集结构体（含匿名内嵌）中声明为 `,string` 的 json 字段名。
func seedStringFields(t reflect.Type) map[string]bool {
	stringKeys := make(map[string]bool)

	var walk func(reflect.Type)
	walk = func(rt reflect.Type) {
		if rt.Kind() == reflect.Ptr {
			rt = rt.Elem()
		}
		if rt.Kind() != reflect.Struct {
			return
		}
		for i := 0; i < rt.NumField(); i++ {
			field := rt.Field(i)
			if field.Anonymous {
				walk(field.Type)
				continue
			}
			tag := field.Tag.Get("json")
			if tag == "" || tag == "-" {
				continue
			}
			parts := strings.Split(tag, ",")
			if parts[0] == "" {
				continue
			}
			name := strings.ToLower(parts[0])
			for _, option := range parts[1:] {
				if option == "string" {
					stringKeys[name] = true
				}
			}
		}
	}
	walk(t)
	return stringKeys
}
