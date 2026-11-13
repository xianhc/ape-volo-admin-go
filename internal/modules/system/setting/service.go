package setting

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/xianhc/ape-volo-admin-go/internal/global"
	"github.com/xianhc/ape-volo-admin-go/internal/global/constants/cachePrefix"
	"github.com/xianhc/ape-volo-admin-go/internal/utils"
	"github.com/xianhc/ape-volo-admin-go/internal/utils/excel"
	"github.com/xianhc/ape-volo-admin-go/internal/utils/querycache"
	"go.uber.org/zap"
)

// Service 实现系统设置的业务规则。
//
// Service 只通过 repository 访问数据库，并负责传输模型与实体之间的转换。
type Service struct {
	repository repository
}

// Create 创建系统设置。
//
// 方法校验设置键唯一性，并在缺少主键时生成主键。
func (service *Service) Create(ctx context.Context, req CreateAndUpdateSettingRequest) error {
	exists, err := service.repository.ExistsByName(ctx, req.Name)
	if err != nil {
		return logDatabaseError(err)
	}
	if exists {
		return errors.New("设置键=>=>已存在!")
	}
	setting := newEntity(req.Id, req.Name, req.Value, req.Enabled, req.Description)
	if err := service.repository.Create(ctx, setting); err != nil {
		return logDatabaseError(err)
	}
	return nil
}

// Update 修改系统设置。
//
// 方法先确认目标记录存在，再在设置键变化时校验唯一性。
func (service *Service) Update(ctx context.Context, req CreateAndUpdateSettingRequest) error {
	oldSetting, err := service.repository.FindByID(ctx, req.Id)
	if errors.Is(err, errNotFound) {
		return errors.New("数据不存在或您无权查看！")
	}
	if err != nil {
		return logDatabaseError(err)
	}
	if oldSetting.Name != req.Name {
		exists, err := service.repository.ExistsByName(ctx, req.Name)
		if err != nil {
			return logDatabaseError(err)
		}
		if exists {
			return errors.New("设置键=>=>已存在!")
		}
	}
	setting := newEntity(req.Id, req.Name, req.Value, req.Enabled, req.Description)
	if err := service.repository.Update(ctx, setting); err != nil {
		return logDatabaseError(err)
	}
	return nil
}

// Delete 逻辑删除指定系统设置。
//
// 没有记录受影响时返回业务错误，避免把无效删除报告为成功。
func (service *Service) Delete(ctx context.Context, ids []string) error {
	rowsAffected, err := service.repository.Delete(ctx, ids)
	if err != nil {
		return logDatabaseError(err)
	}
	if rowsAffected == 0 {
		return errors.New("数据不存在")
	}
	return nil
}

// Query 查询系统设置分页列表。
//
// 返回前统一把 Entity 转换为 SettingResponse。
func (service *Service) Query(ctx context.Context, req QuerySettingRequest) ([]SettingResponse, int64, error) {
	settings, total, err := service.repository.List(ctx, newListOptions(req))
	if err != nil {
		return nil, 0, logDatabaseError(err)
	}
	return fromEntities(settings), total, nil
}

// Download 导出符合条件的系统设置。
//
// 导出沿用列表筛选和排序条件，并返回生成文件路径与下载文件名。
func (service *Service) Download(ctx context.Context, req QuerySettingRequest) (filePath string, fileName string, err error) {
	settings, err := service.repository.FindAll(ctx, newListOptions(req))
	if err != nil {
		return "", "", logDatabaseError(err)
	}
	file := excel.NewFile("Settings")
	defer func() { _ = file.Close() }()
	file.SetHeader([]string{"ID", "键", "值", "是否启用", "描述", "创建时间"})
	for _, setting := range settings {
		enabled := "禁用"
		if setting.Enabled {
			enabled = "启用"
		}
		file.AddRow(strconv.FormatInt(setting.Id, 10), setting.Name, setting.Value, enabled, setting.Description, setting.CreationTime.Format("2006-01-02 15:04:05"))
	}
	fileName = "Settings_" + utils.GenerateID().String() + ".xlsx"
	filePath = global.Config.Excel.Dir + fileName
	return filePath, fileName, file.Save(filePath)
}

// FindSettingByName 按设置键查询单个系统设置。
//
// 系统运行设置绕过创建者数据权限；完整记录不进入值缓存。
func (service *Service) FindSettingByName(ctx context.Context, name string) (SettingResponse, error) {
	ctx = utils.WithoutDataScope(ctx)
	setting, err := service.repository.FindByName(ctx, name)
	if errors.Is(err, errNotFound) {
		return SettingResponse{}, err
	}
	if err != nil {
		return SettingResponse{}, logDatabaseError(err)
	}
	return fromEntity(setting), nil
}

// GetBool 读取设置的布尔值；缺少设置时返回 false，非法值返回错误。
// 开关取 Value 字段，转换后的布尔值按名称缓存 30 分钟。
func (service *Service) GetBool(ctx context.Context, name string) (bool, error) {
	var result bool
	err := querycache.GetOrLoad(ctx, querycache.Key(cachePrefix.LoadSettingByName, name), &result, 30*time.Minute, func() error {
		setting, err := service.FindSettingByName(ctx, name)
		if errors.Is(err, errNotFound) {
			return nil
		}
		if err != nil {
			return err
		}
		switch strings.ToLower(strings.TrimSpace(setting.Value)) {
		case "true":
			result = true
		case "false":
			result = false
		default:
			return fmt.Errorf("系统设置 %s 不是有效布尔值", name)
		}
		return nil
	})
	return result, err
}

// GetInt64 读取设置的整数值；缺少设置时返回零，非法值返回错误。
//
// 转换后的整数值按名称缓存 30 分钟，设置写入提交后失效。
func (service *Service) GetInt64(ctx context.Context, name string) (int64, error) {
	var result int64
	err := querycache.GetOrLoad(ctx, querycache.Key(cachePrefix.LoadSettingByName, name), &result, 30*time.Minute, func() error {
		setting, err := service.FindSettingByName(ctx, name)
		if errors.Is(err, errNotFound) {
			return nil
		}
		if err != nil {
			return err
		}
		result, err = strconv.ParseInt(strings.TrimSpace(setting.Value), 10, 64)
		if err != nil {
			return fmt.Errorf("系统设置 %s 不是有效整数", name)
		}
		return nil
	})
	return result, err
}

func newEntity(id int64, name string, value string, enabled bool, description string) *Setting {
	setting := &Setting{Name: name, Value: value, Enabled: enabled, Description: description}
	if id == 0 {
		id = int64(utils.GenerateID())
	}
	setting.Id = id
	return setting
}

func newListOptions(req QuerySettingRequest) queryOptions {
	return queryOptions{
		KeyWords: req.KeyWords,
		Enabled:  req.Enabled,
		Sorting:  req.Sorting,
		Limit:    req.PageSize,
		Offset:   req.PageSize * (req.PageIndex - 1),
	}
}

func logDatabaseError(err error) error {
	global.Logger.Error("db error: ", zap.Error(err))
	return err
}
