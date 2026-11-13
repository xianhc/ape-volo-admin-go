package appsecret

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/xianhc/ape-volo-admin-go/internal/global"
	"github.com/xianhc/ape-volo-admin-go/internal/utils"
	"github.com/xianhc/ape-volo-admin-go/internal/utils/excel"
	"go.uber.org/zap"
)

// Service 实现应用密钥的业务规则。
//
// Service 只通过 repository 访问数据库，并负责请求、实体与响应之间的转换。
type Service struct {
	repository repository
}

// Create 创建应用密钥。
//
// 未指定主键时生成主键、应用 ID 和密钥，并在写入前校验应用名称唯一性。
func (service *Service) Create(ctx context.Context, req CreateAndUpdateAppSecretRequest) error {
	exists, err := service.repository.ExistsByName(ctx, req.AppName)
	if err != nil {
		return logDatabaseError(err)
	}
	if exists {
		return errors.New("应用=>" + req.AppName + "=>已存在!")
	}

	appSecret := &AppSecret{
		AppName: req.AppName,
		Remark:  req.Remark,
	}
	if req.Id != 0 {
		appSecret.Id = req.Id
	} else {
		appSecret.Id = int64(utils.GenerateID())
		stringId := strconv.FormatInt(appSecret.Id, 10)
		appSecret.AppId = time.Now().Format("20060102") + stringId[len(stringId)-8:]
		appSecret.AppSecretKey = utils.BcryptHash(appSecret.AppId + stringId)
	}
	if err := service.repository.Create(ctx, appSecret); err != nil {
		return logDatabaseError(err)
	}
	return nil
}

// Update 修改应用密钥的名称与备注。
//
// 方法先确认目标记录可见且存在，再在名称变化时执行唯一性校验。
func (service *Service) Update(ctx context.Context, req CreateAndUpdateAppSecretRequest) error {
	oldAppSecret, err := service.repository.FindByID(ctx, req.Id)
	if errors.Is(err, errNotFound) {
		return errors.New("数据不存在或您无权查看！")
	}
	if err != nil {
		return logDatabaseError(err)
	}
	if oldAppSecret.AppName != req.AppName {
		exists, err := service.repository.ExistsByName(ctx, req.AppName)
		if err != nil {
			return logDatabaseError(err)
		}
		if exists {
			return errors.New("应用名称=>" + req.AppName + "=>已存在!")
		}
	}

	appSecret := &AppSecret{AppName: req.AppName, Remark: req.Remark}
	appSecret.Id = req.Id
	if err := service.repository.Update(ctx, appSecret); err != nil {
		return logDatabaseError(err)
	}
	return nil
}

// Delete 逻辑删除指定应用密钥。
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

// Query 查询应用密钥分页列表。
//
// Entity 在返回 Handler 前统一转换为 AppSecretResponse，避免暴露持久化模型。
func (service *Service) Query(ctx context.Context, req QueryAppSecretRequest) ([]AppSecretResponse, int64, error) {
	options := newListOptions(req)
	appSecrets, total, err := service.repository.List(ctx, options)
	if err != nil {
		return nil, 0, logDatabaseError(err)
	}
	return fromEntities(appSecrets), total, nil
}

// Download 导出符合条件的应用密钥。
//
// 导出沿用列表筛选和排序条件，并返回生成文件的路径与下载文件名。
func (service *Service) Download(ctx context.Context, req QueryAppSecretRequest) (filePath string, fileName string, err error) {
	appSecrets, err := service.repository.FindAll(ctx, newListOptions(req))
	if err != nil {
		return "", "", logDatabaseError(err)
	}

	file := excel.NewFile("AppSecret")
	defer func() { _ = file.Close() }()
	file.SetHeader([]string{"ID", "应用密钥", "应用名称", "备注", "创建时间"})
	for _, appSecret := range appSecrets {
		file.AddRow(
			strconv.FormatInt(appSecret.Id, 10),
			appSecret.AppSecretKey,
			appSecret.AppName,
			appSecret.Remark,
			appSecret.CreationTime.Format("2006-01-02 15:04:05"),
		)
	}
	fileName = "AppSecret_" + utils.GenerateID().String() + ".xlsx"
	filePath = global.Config.Excel.Dir + fileName
	return filePath, fileName, file.Save(filePath)
}

func newListOptions(req QueryAppSecretRequest) queryOptions {
	return queryOptions{
		KeyWords:   req.KeyWords,
		CreateTime: req.CreateTime,
		Sorting:    req.Sorting,
		Limit:      req.PageSize,
		Offset:     req.PageSize * (req.PageIndex - 1),
	}
}

func logDatabaseError(err error) error {
	global.Logger.Error("db error: ", zap.Error(err))
	return err
}
