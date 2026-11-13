package dict

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/xianhc/ape-volo-admin-go/internal/global"
	"github.com/xianhc/ape-volo-admin-go/internal/global/constants/cachePrefix"
	"github.com/xianhc/ape-volo-admin-go/internal/utils"
	"github.com/xianhc/ape-volo-admin-go/internal/utils/excel"
	"github.com/xianhc/ape-volo-admin-go/internal/utils/querycache"
	"go.uber.org/zap"
)

// Service 实现字典主表与明细的业务规则。
//
// Service 只通过 Repository 访问数据库，并统一完成 Entity 到 Response 的转换。
type Service struct {
	repository repository
}

// CreateDict 创建字典主表。
//
// 方法校验名称唯一性，并在缺少主键时生成主键。
func (service *Service) CreateDict(ctx context.Context, req CreateAndUpdateDictRequest) error {
	exists, err := service.repository.DictExistsByName(ctx, req.Name)
	if err != nil {
		return logDatabaseError(err)
	}
	if exists {
		return errors.New("字典名称=>" + req.Name + "=>已存在!")
	}
	dict := &Dict{Name: req.Name, Description: req.Description}
	dict.Id = req.Id
	if dict.Id == 0 {
		dict.Id = int64(utils.GenerateID())
	}
	if err := service.repository.CreateDict(ctx, dict); err != nil {
		return logDatabaseError(err)
	}
	return nil
}

// UpdateDict 修改字典主表。
//
// 方法先确认记录存在，再在名称变化时校验唯一性。
func (service *Service) UpdateDict(ctx context.Context, req CreateAndUpdateDictRequest) error {
	old, err := service.repository.FindDictByID(ctx, req.Id)
	if errors.Is(err, errNotFound) {
		return errors.New("数据不存在或您无权查看！")
	}
	if err != nil {
		return logDatabaseError(err)
	}
	if old.Name != req.Name {
		exists, err := service.repository.DictExistsByName(ctx, req.Name)
		if err != nil {
			return logDatabaseError(err)
		}
		if exists {
			return errors.New("字典名称=>" + req.Name + "=>已存在!")
		}
	}
	dict := &Dict{Name: req.Name, Description: req.Description}
	dict.Id = req.Id
	if err := service.repository.UpdateDict(ctx, dict); err != nil {
		return logDatabaseError(err)
	}
	return nil
}

// DeleteDicts 逻辑删除一组字典主表。
//
// 没有记录受影响时返回统一的不可见错误。
func (service *Service) DeleteDicts(ctx context.Context, ids []string) error {
	rows, err := service.repository.DeleteDicts(ctx, ids)
	if err != nil {
		return logDatabaseError(err)
	}
	if rows == 0 {
		return errors.New("数据不存在或您无权查看！")
	}
	return nil
}

// QueryDicts 查询字典分页列表及明细。
//
// 返回前统一转换为独立的 DictResponse。
func (service *Service) QueryDicts(ctx context.Context, req QueryDictRequest) ([]DictResponse, int64, error) {
	dicts, total, err := service.repository.ListDicts(ctx, listOptions(req))
	if err != nil {
		return nil, 0, logDatabaseError(err)
	}
	return fromEntities(dicts), total, nil
}

// DownloadDicts 导出符合条件的字典及明细。
//
// 每条明细对应导出文件中的一行。
func (service *Service) DownloadDicts(ctx context.Context, req QueryDictRequest) (filePath, fileName string, err error) {
	dicts, err := service.repository.FindAllDicts(ctx, listOptions(req))
	if err != nil {
		return "", "", logDatabaseError(err)
	}
	file := excel.NewFile("Dicts")
	defer func() { _ = file.Close() }()
	file.SetHeader([]string{"ID", "字典名称", "字典描述", "字典标签", "字典值", "创建时间"})
	for _, dict := range dicts {
		for _, detail := range dict.DictDetails {
			file.AddRow(strconv.FormatInt(dict.Id, 10), dict.Name, dict.Description, detail.Label, detail.Value, detail.CreationTime.Format("2006-01-02 15:04:05"))
		}
	}
	fileName = "Dicts_" + utils.GenerateID().String() + ".xlsx"
	filePath = global.Config.Excel.Dir + fileName
	return filePath, fileName, file.Save(filePath)
}

// GetDictByName 按名称查询字典及明细。
//
// 结果按名称缓存 30 分钟；名称为空或记录不存在时不回填缓存。
func (service *Service) GetDictByName(ctx context.Context, name string) (DictResponse, error) {
	if name == "" {
		return DictResponse{}, errors.New("字典名称不能为空")
	}
	var result DictResponse
	key := querycache.Key(cachePrefix.LoadDictByName, name)
	err := querycache.GetOrLoad(ctx, key, &result, 30*time.Minute, func() error {
		dict, err := service.repository.FindDictByName(ctx, name)
		if errors.Is(err, errNotFound) {
			return errors.New("数据不存在或您无权查看！")
		}
		if err != nil {
			return logDatabaseError(err)
		}
		result = fromEntity(dict)
		return nil
	})
	return result, err
}

// CreateDetail 创建字典明细。
//
// 同一字典下标签和值组合重复时拒绝创建。
func (service *Service) CreateDetail(ctx context.Context, req CreateAndUpdateDictDetailRequest) error {
	exists, err := service.repository.DetailExists(ctx, req.Dict.Id, req.Label, req.Value)
	if err != nil {
		return logDatabaseError(err)
	}
	if exists {
		return errors.New("字典详情标签=>" + req.Label + "=>已存在!")
	}
	detail := detailEntity(req.Id, req.Label, req.Value, req.DictSort, req.Dict.Id)
	if err := service.repository.CreateDetail(ctx, detail); err != nil {
		return logDatabaseError(err)
	}
	return nil
}

// UpdateDetail 修改字典明细。
//
// 方法确认目标记录存在后更新允许修改的字段。
func (service *Service) UpdateDetail(ctx context.Context, req CreateAndUpdateDictDetailRequest) error {
	_, err := service.repository.FindDetailByID(ctx, req.Id)
	if errors.Is(err, errNotFound) {
		return errors.New("数据不存在或您无权查看！")
	}
	if err != nil {
		return logDatabaseError(err)
	}
	detail := detailEntity(req.Id, req.Label, req.Value, req.DictSort, req.Dict.Id)
	if err := service.repository.UpdateDetail(ctx, detail); err != nil {
		return logDatabaseError(err)
	}
	return nil
}

// DeleteDetail 逻辑删除字典明细。
//
// 没有记录受影响时返回统一的不可见错误。
func (service *Service) DeleteDetail(ctx context.Context, id int64) error {
	rows, err := service.repository.DeleteDetail(ctx, id)
	if err != nil {
		return logDatabaseError(err)
	}
	if rows == 0 {
		return errors.New("数据不存在或您无权查看！")
	}
	return nil
}

// QueryDetails 按字典名称查询排序后的明细。
//
// 返回值通过 DictId 标识所属字典。
func (service *Service) QueryDetails(ctx context.Context, name string) ([]DictDetailResponse, error) {
	dict, err := service.repository.FindDetailsByDictName(ctx, name)
	if err != nil {
		return nil, logDatabaseError(err)
	}
	return detailsFromEntities(dict.DictDetails), nil
}

func listOptions(req QueryDictRequest) queryOptions {
	return queryOptions{KeyWords: req.KeyWords, Sorting: req.Sorting, Limit: req.PageSize, Offset: req.PageSize * (req.PageIndex - 1)}
}
func detailEntity(id int64, label, value string, sort int, dictId int64) *DictDetail {
	detail := &DictDetail{DictId: dictId, Label: label, Value: value, DictSort: sort}
	if id == 0 {
		id = int64(utils.GenerateID())
	}
	detail.Id = id
	return detail
}
func logDatabaseError(err error) error { global.Logger.Error("db error: ", zap.Error(err)); return err }
