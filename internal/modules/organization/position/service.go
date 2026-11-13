package position

import (
	"context"
	"errors"
	"strconv"

	"github.com/xianhc/ape-volo-admin-go/internal/global"
	"github.com/xianhc/ape-volo-admin-go/internal/utils"
	"github.com/xianhc/ape-volo-admin-go/internal/utils/excel"
	"go.uber.org/zap"
)

// Service 实现岗位的业务规则。
//
// Service 只通过 Repository 访问岗位和用户岗位关联数据。
type Service struct{ repository repository }

// Create 创建岗位。
//
// 方法校验名称唯一性，并在缺少主键时生成主键。
func (service *Service) Create(ctx context.Context, req CreateAndUpdatePositionRequest) error {
	exists, err := service.repository.ExistsByName(ctx, req.Name)
	if err != nil {
		return logDatabaseError(err)
	}
	if exists {
		return errors.New("岗位名称=>" + req.Name + "=>已存在!")
	}
	position := newEntity(req.Id, req.Name, req.Sort, req.Enabled)
	if err := service.repository.Create(ctx, position); err != nil {
		return logDatabaseError(err)
	}
	return nil
}

// Update 修改岗位。
//
// 方法先确认记录存在，再在名称变化时校验唯一性。
func (service *Service) Update(ctx context.Context, req CreateAndUpdatePositionRequest) error {
	old, err := service.repository.FindByID(ctx, req.Id)
	if errors.Is(err, errNotFound) {
		return errors.New("数据不存在或您无权查看！")
	}
	if err != nil {
		return logDatabaseError(err)
	}
	if old.Name != req.Name {
		exists, err := service.repository.ExistsByName(ctx, req.Name)
		if err != nil {
			return logDatabaseError(err)
		}
		if exists {
			return errors.New("岗位名称=>" + req.Name + "=>已存在!")
		}
	}
	return service.repository.Update(ctx, newEntity(req.Id, req.Name, req.Sort, req.Enabled))
}

// Delete 删除未被用户引用的岗位。
//
// 任一岗位存在用户关联时拒绝整批删除。
func (service *Service) Delete(ctx context.Context, ids []string) error {
	positions, err := service.repository.FindByIDs(ctx, ids)
	if err != nil {
		return logDatabaseError(err)
	}
	if len(positions) == 0 {
		return errors.New("数据不存在")
	}
	positionIds := make([]int64, 0, len(positions))
	for _, position := range positions {
		positionIds = append(positionIds, position.Id)
	}
	used, err := service.repository.HasUsers(ctx, positionIds)
	if err != nil {
		return logDatabaseError(err)
	}
	if used {
		return errors.New("数据被使用,无法删除")
	}
	_, err = service.repository.Delete(ctx, ids)
	return err
}

// Query 查询岗位分页列表。
//
// 返回前统一把 Entity 转换为 PositionResponse。
func (service *Service) Query(ctx context.Context, req QueryPositionRequest) ([]PositionResponse, int64, error) {
	positions, total, err := service.repository.List(ctx, listOptions(req))
	if err != nil {
		return nil, 0, logDatabaseError(err)
	}
	return FromEntities(positions), total, nil
}

// All 查询全部启用岗位。
//
// 返回结果按排序字段升序，并转换为独立 Response。
func (service *Service) All(ctx context.Context) ([]PositionResponse, error) {
	positions, err := service.repository.AllEnabled(ctx)
	if err != nil {
		return nil, logDatabaseError(err)
	}
	return FromEntities(positions), nil
}

// Download 导出符合条件的岗位。
//
// 导出保持既有列顺序和值格式。
func (service *Service) Download(ctx context.Context, req QueryPositionRequest) (filePath, fileName string, err error) {
	positions, err := service.repository.FindAll(ctx, listOptions(req))
	if err != nil {
		return "", "", logDatabaseError(err)
	}
	file := excel.NewFile("Positions")
	defer func() { _ = file.Close() }()
	file.SetHeader([]string{"ID", "岗位名称", "排序", "状态", "创建时间"})
	for _, position := range positions {
		state := "禁用"
		if position.Enabled {
			state = "启用"
		}
		file.AddRow(strconv.FormatInt(position.Id, 10), position.Name, position.Sort, state, position.CreationTime.Format("2006-01-02 15:04:05"))
	}
	fileName = "Jobs_" + utils.GenerateID().String() + ".xlsx"
	filePath = global.Config.Excel.Dir + fileName
	return filePath, fileName, file.Save(filePath)
}

func newEntity(id int64, name string, sort int, enabled bool) *Position {
	position := &Position{Name: name, Sort: sort, Enabled: enabled}
	if id == 0 {
		id = int64(utils.GenerateID())
	}
	position.Id = id
	return position
}
func listOptions(req QueryPositionRequest) queryOptions {
	return queryOptions{Name: req.JobName, Enabled: req.Enabled, CreateTime: req.CreateTime, Sorting: req.Sorting, Limit: req.PageSize, Offset: req.PageSize * (req.PageIndex - 1)}
}
func logDatabaseError(err error) error { global.Logger.Error("db error: ", zap.Error(err)); return err }
