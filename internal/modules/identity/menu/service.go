package menu

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/xianhc/ape-volo-admin-go/internal/global"
	"github.com/xianhc/ape-volo-admin-go/internal/global/constants/cachePrefix"
	"github.com/xianhc/ape-volo-admin-go/internal/global/constants/rolecode"
	"github.com/xianhc/ape-volo-admin-go/internal/payloads/response"
	"github.com/xianhc/ape-volo-admin-go/internal/utils"
	"github.com/xianhc/ape-volo-admin-go/internal/utils/excel"
	"github.com/xianhc/ape-volo-admin-go/internal/utils/querycache"
	"go.uber.org/zap"
)

// Service 实现菜单校验、树查询、用户路由和导出业务。
//
// Service 只通过 Repository 访问数据库，并在内部管理菜单查询缓存。
type Service struct{ repository repository }

// FillHasChildren 为菜单实体列表填充树控件派生字段。
//
// 更新传入切片中的 HasChildren 与 Leaf，不修改数据库。
func (service *Service) FillHasChildren(ctx context.Context, menus []Menu) error {
	return service.fillHasChildren(ctx, menus)
}

// Create 创建菜单。
//
// 方法校验标题、权限码、组件名称、外链路径和父菜单存在性。
func (service *Service) Create(ctx context.Context, req CreateAndUpdateMenuRequest) error {
	if req.AuthCode != "" {
		if err := requireAdmin(ctx); err != nil {
			return err
		}
	}
	if err := service.validate(ctx, req.Title, req.AuthCode, req.ComponentName, req.Path, req.MenuType, nil); err != nil {
		return err
	}
	if req.ParentId != nil && *req.ParentId > 0 {
		if _, err := service.repository.FindByID(ctx, *req.ParentId); err != nil {
			return logDatabaseError(err)
		}
	}
	return service.repository.Create(ctx, entityFromRequest(req))
}

// Update 修改菜单。
//
// 唯一性校验会排除未变化的原字段值。
func (service *Service) Update(ctx context.Context, req CreateAndUpdateMenuRequest) error {
	old, err := service.repository.FindByID(ctx, req.Id)
	if errors.Is(err, errNotFound) {
		return errors.New("数据不存在或您无权查看！")
	}
	if err != nil {
		return logDatabaseError(err)
	}
	if old.AuthCode != req.AuthCode || (old.AuthCode != "" && (old.MenuType != req.MenuType || old.Enabled != req.Enabled)) {
		if err := requireAdmin(ctx); err != nil {
			return err
		}
	}
	if err := service.validate(ctx, req.Title, req.AuthCode, req.ComponentName, req.Path, req.MenuType, &old); err != nil {
		return err
	}
	return service.repository.Update(ctx, entityFromRequest(req))
}

// Delete 删除菜单及其全部子菜单。
//
// 没有匹配记录时返回“数据不存在”。
func (service *Service) Delete(ctx context.Context, ids []int64) error {
	allIds, err := service.ChildrenIDs(ctx, ids)
	if err != nil {
		return err
	}
	if len(allIds) == 0 {
		return errors.New("数据不存在")
	}
	_, err = service.repository.Delete(ctx, allIds)
	return err
}

// Query 查询直接菜单列表。
//
// 返回结果批量填充 HasChildren、Leaf、Label 和空 Children。
func (service *Service) Query(ctx context.Context, req QueryMenuRequest) ([]MenuResponse, int64, error) {
	menus, err := service.repository.Query(ctx, req.Title, req.ParentId, req.CreateTime)
	if err != nil {
		return nil, 0, logDatabaseError(err)
	}
	if err := service.fillHasChildren(ctx, menus); err != nil {
		return nil, 0, err
	}
	for index := range menus {
		menus[index].Children = []Menu{}
		menus[index].Label = menus[index].Title
	}
	return FromEntities(menus), int64(len(menus)), nil
}

// Download 导出符合条件的菜单。
//
// 导出保持既有列顺序和值格式。
func (service *Service) Download(ctx context.Context, req QueryMenuRequest) (filePath, fileName string, err error) {
	menus, err := service.repository.Query(ctx, req.Title, req.ParentId, req.CreateTime)
	if err != nil {
		return "", "", logDatabaseError(err)
	}
	file := excel.NewFile("Menus")
	defer func() { _ = file.Close() }()
	file.SetHeader([]string{"ID", "菜单标题", "组件路径", "权限标识符", "菜单类型", "组件", "组件名称", "菜单父ID", "排序", "Icon图标", "是否启用", "是否缓存", "是否隐藏", "创建时间"})
	for _, menu := range menus {
		external, menuType, keep, hidden := "否", "未知", "否", "否"
		if menu.IsExternalLink() {
			external = "是"
		}
		if menu.MenuType == 1 {
			menuType = "目录"
		} else if menu.MenuType == 2 {
			menuType = "菜单"
		} else if menu.MenuType == 3 {
			menuType = "按钮"
		}
		if menu.KeepAlive {
			keep = "是"
		}
		if menu.Hidden {
			hidden = "是"
		}
		file.AddRow(strconv.FormatInt(menu.Id, 10), menu.Title, menu.Path, menu.AuthCode, external, menu.Component, menu.ComponentName, 0, menu.Sort, menu.Icon, menuType, keep, hidden, menu.CreationTime.Format("2006-01-02 15:04:05"))
	}
	fileName = "Menus_" + utils.GenerateID().String() + ".xlsx"
	filePath = global.Config.Excel.Dir + fileName
	return filePath, fileName, file.Save(filePath)
}

// BuildTree 构建当前用户的前端路由菜单。
//
// 结果按用户主键缓存两小时，按钮菜单不会进入路由树。
func (service *Service) BuildTree(ctx context.Context, userId int64) ([]response.TreeMenuVo, error) {
	result := make([]response.TreeMenuVo, 0)
	key := querycache.Key(cachePrefix.UserMenuById, userId)
	err := querycache.GetOrLoadAlways(ctx, key, &result, 2*time.Hour, func() error {
		menus, err := service.repository.MenusForUser(ctx, userId)
		if err != nil {
			return err
		}
		grouped := groupMenus(menus)
		roots := grouped[0]
		buildTreeEntities(roots, grouped)
		result = buildRouteTree(roots)
		return nil
	})
	return result, err
}

// AllTree 查询全部菜单授权树。
//
// 返回完整树结构供角色菜单授权弹窗使用。
func (service *Service) AllTree(ctx context.Context) ([]MenuResponse, error) {
	if err := requireAdmin(ctx); err != nil {
		return nil, err
	}
	menus, err := service.repository.All(ctx)
	if err != nil {
		return nil, logDatabaseError(err)
	}
	grouped := groupMenus(menus)
	roots := grouped[0]
	buildTreeEntities(roots, grouped)
	return FromEntities(roots), nil
}

// ChildrenIDs 展开指定菜单及所有子菜单主键。
//
// 结果去重并包含输入主键自身。
func (service *Service) ChildrenIDs(ctx context.Context, ids []int64) ([]int64, error) {
	result := make([]int64, 0)
	var walk func([]int64) error
	walk = func(current []int64) error {
		for _, id := range current {
			result = utils.AppendInt64(result, id)
			children, err := service.repository.Children(ctx, id)
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

func (service *Service) fillHasChildren(ctx context.Context, menus []Menu) error {
	ids := make([]int64, 0, len(menus))
	for _, menu := range menus {
		ids = append(ids, menu.Id)
	}
	parents, err := service.repository.ParentIDs(ctx, ids)
	if err != nil {
		return logDatabaseError(err)
	}
	has := make(map[int64]struct{}, len(parents))
	for _, id := range parents {
		has[id] = struct{}{}
	}
	for i := range menus {
		_, menus[i].HasChildren = has[menus[i].Id]
		menus[i].Leaf = !menus[i].HasChildren
	}
	return nil
}
func (service *Service) validate(ctx context.Context, title, authCode, componentName, path string, menuType int, old *Menu) error {
	requiresAuthCode := menuType == MenuTypeMenu || menuType == MenuTypeButton
	checks := []struct {
		field, value, message string
		changed               bool
	}{{"title", title, "菜单标题=>" + title + "=>已存在!", old == nil || old.Title != title}, {"auth_code", authCode, "权限标识=>" + authCode + "=>已存在!", requiresAuthCode && authCode != "" && (old == nil || old.AuthCode != authCode)}, {"component_name", componentName, "组件名称=>" + componentName + "=>已存在!", componentName != "" && (old == nil || old.ComponentName != componentName)}}
	for _, check := range checks {
		if !check.changed {
			continue
		}
		total, err := service.repository.CountByField(ctx, check.field, check.value)
		if err != nil {
			return logDatabaseError(err)
		}
		if total > 0 {
			return errors.New(check.message)
		}
	}
	if requiresAuthCode && authCode == "" {
		return errors.New("权限标识为必填")
	}
	if menuType == MenuTypeExternalLink {
		lower := strings.ToLower(path)
		if !strings.HasPrefix(lower, "http://") && !strings.HasPrefix(lower, "https://") {
			return errors.New("外链菜单必须以http://或者https://开头")
		}
	}
	return nil
}
func entityFromRequest(req CreateAndUpdateMenuRequest) *Menu {
	menu := &Menu{
		Title:         req.Title,
		Path:          req.Path,
		AuthCode:      req.AuthCode,
		Component:     req.Component,
		ComponentName: req.ComponentName,
		ParentId:      req.ParentId,
		Sort:          req.Sort,
		Icon:          req.Icon,
		MenuType:      req.MenuType,
		Enabled:       req.Enabled,
		KeepAlive:     req.KeepAlive,
		Hidden:        req.Hidden,
	}
	menu.Id = req.Id
	if menu.Id == 0 {
		menu.Id = int64(utils.GenerateID())
	}
	return menu
}
func groupMenus(menus []Menu) map[int64][]Menu {
	grouped := make(map[int64][]Menu)
	for _, menu := range menus {
		grouped[menu.ParentKey()] = append(grouped[menu.ParentKey()], menu)
	}
	return grouped
}
func buildTreeEntities(nodes []Menu, grouped map[int64][]Menu) {
	for i := range nodes {
		nodes[i].Children = grouped[nodes[i].Id]
		nodes[i].HasChildren = len(nodes[i].Children) > 0
		nodes[i].Leaf = !nodes[i].HasChildren
		nodes[i].Label = nodes[i].Title
		buildTreeEntities(nodes[i].Children, grouped)
	}
}
func buildRouteTree(menus []Menu) []response.TreeMenuVo {
	result := make([]response.TreeMenuVo, 0, len(menus))
	for _, menu := range menus {
		item := response.TreeMenuVo{Hidden: menu.Hidden, Name: menu.ComponentName, TreeMenuMeta: response.TreeMenuMeta{Title: menu.Title, Icon: menu.Icon, NoCache: !menu.KeepAlive}}
		if item.Name == "" {
			item.Name = menu.Title
		}
		if menu.IsRoot() {
			item.Path = "/" + menu.Path
		} else {
			item.Path = menu.Path
		}
		if !menu.IsExternalLink() && menu.Component != "" {
			item.Component = menu.Component
		}
		if len(menu.Children) > 0 {
			item.AlwaysShow = true
			item.Redirect = "noredirect"
			item.Children = buildRouteTree(menu.Children)
		}
		result = append(result, item)
	}
	return result
}
func logDatabaseError(err error) error { global.Logger.Error("db error: ", zap.Error(err)); return err }

func requireAdmin(ctx context.Context) error {
	if !utils.HasRole(ctx, rolecode.Admin) {
		return errors.New("抱歉，您访问权限等级不够")
	}
	return nil
}
