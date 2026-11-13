package role

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/xianhc/ape-volo-admin-go/internal/global"
	"github.com/xianhc/ape-volo-admin-go/internal/global/constants/rolecode"
	"github.com/xianhc/ape-volo-admin-go/internal/modules/organization/department"
	"github.com/xianhc/ape-volo-admin-go/internal/utils"
	"github.com/xianhc/ape-volo-admin-go/internal/utils/excel"
	"go.uber.org/zap"
)

// Service 实现角色管理、关联授权和权限码查询业务。
//
// Service 只通过 Repository 访问数据库，并返回独立 RoleResponse。
type Service struct{ repository repository }

// Create 创建角色及部门关联。
//
// 方法校验名称和角色码唯一性，并在缺少主键时生成主键。
func (service *Service) Create(ctx context.Context, req CreateAndUpdateRoleRequest) error {
	return utils.InTransaction(ctx, func(ctx context.Context) error { return service.create(ctx, req) })
}

func (service *Service) create(ctx context.Context, req CreateAndUpdateRoleRequest) error {
	if err := requireAdmin(ctx); err != nil {
		return err
	}
	if err := service.validateUnique(ctx, req.Name, req.AuthCode, nil); err != nil {
		return err
	}
	role := newRole(req.Id, req.Name, req.Description, req.DataScopeType, req.AuthCode)
	return service.repository.Create(ctx, role, referenceIDs(req.Departments))
}

// Update 修改角色并替换部门关联。
//
// 方法先确认角色存在，再校验发生变化的唯一字段。
func (service *Service) Update(ctx context.Context, req CreateAndUpdateRoleRequest) error {
	return utils.InTransaction(ctx, func(ctx context.Context) error { return service.update(ctx, req) })
}

func (service *Service) update(ctx context.Context, req CreateAndUpdateRoleRequest) error {
	old, err := service.repository.FindByID(ctx, req.Id, true)
	if errors.Is(err, errNotFound) {
		return errors.New("数据不存在或您无权查看！")
	}
	if err != nil {
		return logDatabaseError(err)
	}
	if old.AuthCode != req.AuthCode || old.DataScopeType != req.DataScopeType || !sameDepartmentIDs(old.Departments, referenceIDs(req.Departments)) {
		if err := requireAdmin(ctx); err != nil {
			return err
		}
	}
	if err := service.validateUnique(ctx, req.Name, req.AuthCode, &old); err != nil {
		return err
	}
	return service.repository.Update(ctx, newRole(req.Id, req.Name, req.Description, req.DataScopeType, req.AuthCode), referenceIDs(req.Departments))
}

// Delete 删除未被用户引用的角色。
//
// 任一角色存在用户关联时拒绝整批删除。
func (service *Service) Delete(ctx context.Context, ids []string) error {
	return utils.InTransaction(ctx, func(ctx context.Context) error { return service.delete(ctx, ids) })
}

func (service *Service) delete(ctx context.Context, ids []string) error {
	used, err := service.repository.HasUsers(ctx, ids)
	if err != nil {
		return logDatabaseError(err)
	}
	if used {
		return errors.New("存在用户关联，请解除后再试！")
	}
	_, err = service.repository.Delete(ctx, ids)
	return err
}

// Query 查询角色分页列表。
//
// 返回结果包含菜单关联并转换为独立 Response。
func (service *Service) Query(ctx context.Context, req QueryRoleRequest) ([]RoleResponse, int64, error) {
	roles, total, err := service.repository.List(ctx, listOptions(req))
	if err != nil {
		return nil, 0, logDatabaseError(err)
	}
	return FromEntities(roles), total, nil
}

// All 查询全部角色及完整关联。
//
// 返回结果用于角色选择和授权界面。
func (service *Service) All(ctx context.Context) ([]RoleResponse, error) {
	roles, err := service.repository.All(ctx)
	if err != nil {
		return nil, logDatabaseError(err)
	}
	return FromEntities(roles), nil
}

// GetWithDetails 查询单个角色及完整关联。
//
// 未找到时返回统一不可见错误。
func (service *Service) GetWithDetails(ctx context.Context, id int64) (RoleResponse, error) {
	if err := requireAdmin(ctx); err != nil {
		return RoleResponse{}, err
	}
	role, err := service.repository.FindByID(ctx, id, true)
	if errors.Is(err, errNotFound) {
		return RoleResponse{}, errors.New("数据不存在或您无权查看！")
	}
	if err != nil {
		return RoleResponse{}, logDatabaseError(err)
	}
	return fromEntity(role), nil
}

// UpdateMenus 替换角色菜单授权。
//
// menus 为空时保持现有关联，延续既有接口行为。
func (service *Service) UpdateMenus(ctx context.Context, req CreateAndUpdateRoleRequest) error {
	return utils.InTransaction(ctx, func(ctx context.Context) error { return service.updateMenus(ctx, req) })
}

func (service *Service) updateMenus(ctx context.Context, req CreateAndUpdateRoleRequest) error {
	if err := requireAdmin(ctx); err != nil {
		return err
	}
	if _, err := service.repository.FindByID(ctx, req.Id, false); errors.Is(err, errNotFound) {
		return errors.New("数据不存在或您无权查看！")
	} else if err != nil {
		return logDatabaseError(err)
	}
	return service.repository.ReplaceMenus(ctx, req.Id, referenceIDs(req.Menus))
}

// PermissionCodes 查询用户持有的菜单权限码。
//
// roleIds 可为空，Repository 会按用户主键查询角色关联。
func (service *Service) PermissionCodes(ctx context.Context, userId int64, roleIds []int64) ([]string, error) {
	codes, err := service.repository.PermissionCodes(ctx, userId, roleIds)
	if err != nil {
		return nil, logDatabaseError(err)
	}
	return codes, nil
}

// GetPermissionIdentifier 查询用户权限码并写入目标切片。
//
// 该方法保留现有认证与权限中间件的调用形状，内部委托 PermissionCodes。
func (service *Service) GetPermissionIdentifier(ctx context.Context, userId int64, roleIds []int64, target *[]string) error {
	codes, err := service.PermissionCodes(ctx, userId, roleIds)
	if err != nil {
		return err
	}
	*target = codes
	return nil
}

// RoleCodes 读取启用用户的非空角色码。
//
// 角色码属于认证元数据，读取绕过创建者数据权限并去重。
func (service *Service) RoleCodes(ctx context.Context, userId int64) ([]string, error) {
	codes, err := service.repository.RoleCodes(utils.WithoutDataScope(ctx), userId)
	if err != nil {
		return nil, logDatabaseError(err)
	}
	result := make([]string, 0, len(codes))
	for _, code := range codes {
		if strings.TrimSpace(code) != "" {
			result = append(result, code)
		}
	}
	return result, nil
}

// Download 导出符合条件的角色。
//
// 导出包含数据范围和关联部门名称。
func (service *Service) Download(ctx context.Context, req QueryRoleRequest) (filePath, fileName string, err error) {
	roles, err := service.repository.FindForExport(ctx, listOptions(req))
	if err != nil {
		return "", "", logDatabaseError(err)
	}
	file := excel.NewFile("Roles")
	defer func() { _ = file.Close() }()
	file.SetHeader([]string{"ID", "角色名称", "角色描述", "数据范围", "数据部门", "角色代码", "创建时间"})
	for _, role := range roles {
		names := make([]string, 0, len(role.Departments))
		for _, dept := range role.Departments {
			names = append(names, dept.Name)
		}
		file.AddRow(strconv.FormatInt(role.Id, 10), role.Name, role.Description, role.DataScopeLabel(), strings.Join(names, ","), role.AuthCode, role.CreationTime.Format("2006-01-02 15:04:05"))
	}
	fileName = "Roles_" + utils.GenerateID().String() + ".xlsx"
	filePath = global.Config.Excel.Dir + fileName
	return filePath, fileName, file.Save(filePath)
}

func (service *Service) validateUnique(ctx context.Context, name, code string, old *Role) error {
	if old == nil || old.Name != name {
		total, err := service.repository.CountByField(ctx, "name", name)
		if err != nil {
			return logDatabaseError(err)
		}
		if total > 0 {
			return errors.New("名称=>" + name + "=>已存在!")
		}
	}
	if old == nil || old.AuthCode != code {
		total, err := service.repository.CountByField(ctx, "auth_code", code)
		if err != nil {
			return logDatabaseError(err)
		}
		if total > 0 {
			return errors.New("标识=>" + code + "=>已存在!")
		}
	}
	return nil
}
func newRole(id int64, name, description string, scope int, code string) *Role {
	role := &Role{Name: name, Description: description, DataScopeType: scope, AuthCode: code}
	if id == 0 {
		id = int64(utils.GenerateID())
	}
	role.Id = id
	return role
}
func referenceIDs(refs []IDReference) []int64 {
	ids := make([]int64, 0, len(refs))
	for _, ref := range refs {
		ids = append(ids, ref.Id)
	}
	return ids
}

func sameDepartmentIDs(current []department.Department, requested []int64) bool {
	if len(current) != len(requested) {
		return false
	}
	ids := make(map[int64]struct{}, len(current))
	for _, item := range current {
		ids[item.Id] = struct{}{}
	}
	for _, id := range requested {
		if _, ok := ids[id]; !ok {
			return false
		}
	}
	return true
}

func requireAdmin(ctx context.Context) error {
	if !utils.HasRole(ctx, rolecode.Admin) {
		return errors.New("抱歉，您访问权限等级不够")
	}
	return nil
}
func listOptions(req QueryRoleRequest) queryOptions {
	return queryOptions{Name: req.RoleName, CreateTime: req.CreateTime, Sorting: req.Sorting, Limit: req.PageSize, Offset: req.PageSize * (req.PageIndex - 1)}
}
func logDatabaseError(err error) error { global.Logger.Error("db error: ", zap.Error(err)); return err }
