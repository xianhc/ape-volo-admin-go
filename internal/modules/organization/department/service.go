package department

import (
	"context"
	"errors"
	"strconv"

	"github.com/xianhc/ape-volo-admin-go/internal/global"
	"github.com/xianhc/ape-volo-admin-go/internal/utils"
	"github.com/xianhc/ape-volo-admin-go/internal/utils/excel"
	"go.uber.org/zap"
)

// Service 实现部门树、删除校验、查询和导出业务。
//
// Service 只通过 Repository 访问数据库，不依赖 Role 或 User Entity。
type Service struct{ repository repository }

// Create 创建部门。
//
// 方法校验名称唯一性和父部门存在性，并在缺少主键时生成主键。
func (service *Service) Create(ctx context.Context, req CreateAndUpdateDepartmentRequest) error {
	exists, err := service.repository.ExistsByName(ctx, req.Name)
	if err != nil {
		return logDatabaseError(err)
	}
	if exists {
		return errors.New("部门名称=>" + req.Name + "=>已存在!")
	}
	if req.ParentId != nil && *req.ParentId != 0 {
		exists, err := service.repository.ParentExists(ctx, *req.ParentId)
		if err != nil {
			return logDatabaseError(err)
		}
		if !exists {
			return errors.New("上级部门不存在")
		}
	}
	return service.repository.Create(ctx, newEntity(req.Id, req.Name, req.ParentId, req.Sort, req.Enabled))
}

// Update 修改部门。
//
// 方法先确认记录存在，再在名称变化时校验唯一性。
func (service *Service) Update(ctx context.Context, req CreateAndUpdateDepartmentRequest) error {
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
			return errors.New("部门名称=>" + req.Name + "=>已存在!")
		}
	}
	return service.repository.Update(ctx, newEntity(req.Id, req.Name, req.ParentId, req.Sort, req.Enabled))
}

// Delete 删除部门及其所有启用子部门。
//
// 存在用户或角色关联时拒绝整批删除。
func (service *Service) Delete(ctx context.Context, ids []int64) error {
	allIds, err := service.GetChildrenIDs(ctx, ids)
	if err != nil {
		return logDatabaseError(err)
	}
	departments, err := service.repository.FindByIDs(ctx, allIds)
	if err != nil {
		return logDatabaseError(err)
	}
	if len(departments) == 0 {
		return errors.New("数据不存在")
	}
	used, err := service.repository.HasUsers(ctx, allIds)
	if err != nil {
		return logDatabaseError(err)
	}
	if used {
		return errors.New("存在用户关联，请解除后再试！")
	}
	used, err = service.repository.HasRoles(ctx, allIds)
	if err != nil {
		return logDatabaseError(err)
	}
	if used {
		return errors.New("存在角色关联，请解除后再试！")
	}
	_, err = service.repository.Delete(ctx, allIds)
	return err
}

// GetChildrenIDs 展开指定部门及所有启用子部门主键。
//
// 结构查询绕过数据权限，避免创建者过滤截断部门层级。
func (service *Service) GetChildrenIDs(ctx context.Context, ids []int64) ([]int64, error) {
	result := make([]int64, 0)
	var walk func([]int64) error
	walk = func(current []int64) error {
		for _, id := range current {
			result = utils.AppendInt64(result, id)
			children, err := service.repository.EnabledChildren(ctx, id)
			if err != nil {
				return err
			}
			childIDs := make([]int64, 0, len(children))
			for _, child := range children {
				childIDs = append(childIDs, child.Id)
			}
			if err := walk(childIDs); err != nil {
				return err
			}
		}
		return nil
	}
	return result, walk(ids)
}

// Query 查询部门列表。
//
// 根节点查询默认分页，子节点查询保持不分页；结果批量填充树派生字段。
func (service *Service) Query(ctx context.Context, req QueryDepartmentRequest) ([]DepartmentResponse, int64, error) {
	if req.PageSize == 0 {
		req.PageIndex, req.PageSize = 1, 10
	}
	options := listOptions(req)
	options.Paginate = req.ParentId == nil
	departments, total, err := service.repository.List(ctx, options)
	if err != nil {
		return nil, 0, logDatabaseError(err)
	}
	ids := make([]int64, 0, len(departments))
	for _, department := range departments {
		ids = append(ids, department.Id)
	}
	parentIds, err := service.repository.ParentIDs(ctx, ids)
	if err != nil {
		return nil, 0, logDatabaseError(err)
	}
	has := make(map[int64]struct{}, len(parentIds))
	for _, id := range parentIds {
		has[id] = struct{}{}
	}
	for index := range departments {
		_, departments[index].HasChildren = has[departments[index].Id]
		departments[index].Leaf = !departments[index].HasChildren
		departments[index].Label = departments[index].Name
	}
	return FromEntities(departments), total, nil
}

// All 查询全部部门。
//
// 返回结果按 Sort 排序并转换为独立 Response。
func (service *Service) All(ctx context.Context) ([]DepartmentResponse, error) {
	departments, err := service.repository.All(ctx)
	if err != nil {
		return nil, logDatabaseError(err)
	}
	return FromEntities(departments), nil
}

// Download 导出符合条件的部门。
//
// 导出保持既有列顺序和值格式。
func (service *Service) Download(ctx context.Context, req QueryDepartmentRequest) (filePath, fileName string, err error) {
	departments, err := service.repository.FindAll(ctx, listOptions(req))
	if err != nil {
		return "", "", logDatabaseError(err)
	}
	file := excel.NewFile("Depts")
	defer func() { _ = file.Close() }()
	file.SetHeader([]string{"ID", "部门名称", "部门父ID", "排序", "是否启用", "创建时间"})
	for _, department := range departments {
		enabled := "禁用"
		if department.Enabled {
			enabled = "启用"
		}
		file.AddRow(strconv.FormatInt(department.Id, 10), department.Name, department.ParentKey(), department.Sort, enabled, department.CreationTime.Format("2006-01-02 15:04:05"))
	}
	fileName = "Depts_" + utils.GenerateID().String() + ".xlsx"
	filePath = global.Config.Excel.Dir + fileName
	return filePath, fileName, file.Save(filePath)
}

func newEntity(id int64, name string, parentId *int64, sort int, enabled bool) *Department {
	department := &Department{Name: name, ParentId: parentId, Sort: sort, Enabled: enabled}
	if id == 0 {
		id = int64(utils.GenerateID())
	}
	department.Id = id
	return department
}
func listOptions(req QueryDepartmentRequest) queryOptions {
	name := req.Name
	if name == "" {
		name = req.DeptName
	}
	return queryOptions{Name: name, Enabled: req.Enabled, ParentId: req.ParentId, CreateTime: req.CreateTime, Sorting: req.Sorting, Limit: req.PageSize, Offset: req.PageSize * (req.PageIndex - 1)}
}
func logDatabaseError(err error) error { global.Logger.Error("db error: ", zap.Error(err)); return err }
