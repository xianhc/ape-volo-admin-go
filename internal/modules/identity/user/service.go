package user

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/xianhc/ape-volo-admin-go/internal/global"
	"github.com/xianhc/ape-volo-admin-go/internal/global/constants/rolecode"
	"github.com/xianhc/ape-volo-admin-go/internal/modules/authorization/session"
	"github.com/xianhc/ape-volo-admin-go/internal/modules/identity/role"
	"github.com/xianhc/ape-volo-admin-go/internal/modules/organization/department"
	"github.com/xianhc/ape-volo-admin-go/internal/utils"
	"github.com/xianhc/ape-volo-admin-go/internal/utils/excel"
	"go.uber.org/zap"
)

// Service 实现用户管理、认证读取和个人中心业务。
//
// Service 只通过 Repository 访问用户主表与关联表。
type Service struct {
	repository repository
	sessions   session.Service
}

// Create 创建用户及岗位、角色关联。
//
// 初始密码从配置读取，未配置时使用 123456 并进行 bcrypt 哈希。
func (service *Service) Create(ctx context.Context, req CreateAndUpdateUserRequest) error {
	return utils.InTransaction(ctx, func(ctx context.Context) error { return service.create(ctx, req) })
}

func (service *Service) create(ctx context.Context, req CreateAndUpdateUserRequest) error {
	if err := requireAdmin(ctx); err != nil {
		return err
	}
	if err := service.validateUnique(ctx, req.UserName, req.NickName, req.Email, req.Phone, 0); err != nil {
		return err
	}
	user := newUser(req.Id, req.UserName, req.NickName, req.Email, req.Enabled, req.Phone, req.GenderCode, req.Dept.Id)
	user.Password = utils.BcryptHash(defaultPassword())
	if err := service.repository.Create(ctx, user, referenceIDs(req.Positions), referenceIDs(req.Roles)); err != nil {
		return logDatabaseError(err)
	}
	service.revoke(ctx, user.Id, "修改用户后撤销会话失败")
	return nil
}

// Update 修改用户及岗位、角色关联。
//
// 方法先确认记录存在，再校验用户名、昵称、邮箱和电话唯一性。
func (service *Service) Update(ctx context.Context, req CreateAndUpdateUserRequest) error {
	return utils.InTransaction(ctx, func(ctx context.Context) error { return service.update(ctx, req) })
}

func (service *Service) update(ctx context.Context, req CreateAndUpdateUserRequest) error {
	old, err := service.repository.FindByID(ctx, req.Id, true, false)
	if errors.Is(err, errNotFound) {
		return errors.New("数据不存在")
	} else if err != nil {
		return logDatabaseError(err)
	}
	if old.DeptId != req.Dept.Id || !sameRoleIDs(old.Roles, referenceIDs(req.Roles)) {
		if err := requireAdmin(ctx); err != nil {
			return err
		}
	}
	if err := service.validateUnique(ctx, req.UserName, req.NickName, req.Email, req.Phone, req.Id); err != nil {
		return err
	}
	user := newUser(req.Id, req.UserName, req.NickName, req.Email, req.Enabled, req.Phone, req.GenderCode, req.Dept.Id)
	if err := service.repository.Update(ctx, user, referenceIDs(req.Positions), referenceIDs(req.Roles)); err != nil {
		return logDatabaseError(err)
	}
	service.revoke(ctx, user.Id, "修改用户后撤销会话失败")
	return nil
}

// Delete 删除用户并撤销其全部会话。
//
// 当前登录用户不能删除自己。
func (service *Service) Delete(ctx context.Context, ids []string, currentUserId int64) error {
	return utils.InTransaction(ctx, func(ctx context.Context) error { return service.delete(ctx, ids, currentUserId) })
}

func (service *Service) delete(ctx context.Context, ids []string, currentUserId int64) error {
	if utils.ContainsValue(ids, strconv.FormatInt(currentUserId, 10)) {
		return errors.New("禁止删除自己")
	}
	users, err := service.repository.FindByIDs(ctx, ids)
	if err != nil {
		return logDatabaseError(err)
	}
	if len(users) == 0 {
		return errors.New("数据不存在或您无权查看！")
	}
	if _, err = service.repository.Delete(ctx, ids); err != nil {
		return logDatabaseError(err)
	}
	for _, user := range users {
		service.revoke(ctx, user.Id, "删除用户后撤销会话失败")
	}
	return nil
}

// Query 查询用户分页列表。
//
// 返回结果预加载部门、岗位和角色并转换为脱敏 Response。
func (service *Service) Query(ctx context.Context, req QueryUserRequest) ([]UserResponse, int64, error) {
	options, err := service.listOptions(ctx, req)
	if err != nil {
		return nil, 0, err
	}
	users, total, err := service.repository.List(ctx, options)
	if err != nil {
		return nil, 0, logDatabaseError(err)
	}
	return fromEntities(users), total, nil
}

// Download 导出符合条件的用户。
//
// 导出包含角色、部门、岗位和脱敏用户字段。
func (service *Service) Download(ctx context.Context, req QueryUserRequest) (filePath, fileName string, err error) {
	options, err := service.listOptions(ctx, req)
	if err != nil {
		return "", "", err
	}
	users, err := service.repository.FindAll(ctx, options)
	if err != nil {
		return "", "", logDatabaseError(err)
	}
	file := excel.NewFile("Users")
	defer func() { _ = file.Close() }()
	file.SetHeader([]string{"ID", "用户名称", "角色名称", "用户昵称", "用户电话", "用户邮箱", "是否激活", "部门名称", "岗位名称", "性别", "创建时间"})
	for _, user := range users {
		roleNames := make([]string, 0, len(user.Roles))
		for _, role := range user.Roles {
			roleNames = append(roleNames, role.Name)
		}
		positionNames := make([]string, 0, len(user.Positions))
		for _, position := range user.Positions {
			positionNames = append(positionNames, position.Name)
		}
		enabled := "否"
		if user.Enabled {
			enabled = "是"
		}
		file.AddRow(strconv.FormatInt(user.Id, 10), user.UserName, strings.Join(roleNames, ","), user.NickName, user.Phone, user.Email, enabled, user.Dept.Name, strings.Join(positionNames, ","), user.GenderCode, user.CreationTime.Format("2006-01-02 15:04:05"))
	}
	fileName = "Users_" + utils.GenerateID().String() + ".xlsx"
	filePath = global.Config.Excel.Dir + fileName
	return filePath, fileName, file.Save(filePath)
}

// AuthByName 按用户名读取认证用户。
//
// Password 仅保留在内部 AuthUser DTO，Public 为脱敏响应模型。
func (service *Service) AuthByName(ctx context.Context, name string) (UserResponse, string, error) {
	user, err := service.repository.FindByName(ctx, name)
	if errors.Is(err, errNotFound) {
		return UserResponse{}, "", errors.New("数据不存在或您无权查看！")
	}
	if err != nil {
		return UserResponse{}, "", logDatabaseError(err)
	}
	return fromEntity(user), user.Password, nil
}

// AuthByID 按主键读取带完整关联的认证用户。
//
// 查询绕过数据权限，确保当前用户可以读取自己的档案。
func (service *Service) AuthByID(ctx context.Context, id int64) (UserResponse, string, error) {
	user, err := service.repository.FindByID(ctx, id, true, true)
	if errors.Is(err, errNotFound) {
		return UserResponse{}, "", errors.New("数据不存在或您无权查看！")
	}
	if err != nil {
		return UserResponse{}, "", logDatabaseError(err)
	}
	return fromEntity(user), user.Password, nil
}

func (service *Service) listOptions(ctx context.Context, req QueryUserRequest) (queryOptions, error) {
	options := queryOptions{Id: req.Id, UserName: req.UserName, NickName: req.NickName, KeyWords: req.KeyWords, Enabled: req.Enabled, CreateTime: req.CreateTime, Sorting: req.Sorting, Limit: req.PageSize, Offset: req.PageSize * (req.PageIndex - 1)}
	if req.DepartmentIdArray != "" {
		for _, raw := range strings.Split(req.DepartmentIdArray, ",") {
			id, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
			if err != nil || id <= 0 {
				return options, errors.New("departmentIdArray 必须是逗号分隔的正整数主键")
			}
			options.DeptIds = utils.AppendInt64(options.DeptIds, id)
		}
	} else if req.DeptId != nil {
		ids, err := (&department.Service{}).GetChildrenIDs(ctx, []int64{*req.DeptId})
		if err != nil {
			return options, err
		}
		options.DeptIds = ids
	}
	return options, nil
}
func (service *Service) validateUnique(ctx context.Context, userName, nickName, email, phone string, exclude int64) error {
	checks := []struct{ field, value, label string }{{"user_name", userName, "用户名称"}, {"nick_name", nickName, "昵称"}, {"email", email, "邮箱"}, {"phone", phone, "电话"}}
	for _, check := range checks {
		total, err := service.repository.CountByField(ctx, check.field, check.value, exclude)
		if err != nil {
			return logDatabaseError(err)
		}
		if total > 0 {
			return errors.New(check.label + "=>" + check.value + "=>已存在!")
		}
	}
	return nil
}
func (service *Service) revoke(ctx context.Context, userId int64, message string) {
	utils.AfterCommit(ctx, "session-revoke:"+strconv.FormatInt(userId, 10), func(ctx context.Context) {
		if err := service.sessions.RevokeUser(ctx, userId); err != nil {
			global.Logger.Error(message, zap.Error(err))
		}
	})
}
func defaultPassword() string {
	if password := global.Config.System.UserDefaultPassword; password != "" {
		return password
	}
	return "123456"
}
func newUser(id int64, userName, nickName, email string, enabled bool, phone string, gender int, deptId int64) *User {
	user := &User{UserName: userName, NickName: nickName, Email: email, Enabled: enabled, Phone: phone, GenderCode: gender, DeptId: deptId}
	if id == 0 {
		id = int64(utils.GenerateID())
	}
	user.Id = id
	return user
}
func referenceIDs(refs []IDReference) []int64 {
	ids := make([]int64, 0, len(refs))
	for _, ref := range refs {
		ids = append(ids, ref.Id)
	}
	return ids
}

func sameRoleIDs(current []role.Role, requested []int64) bool {
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
func logDatabaseError(err error) error { global.Logger.Error("db error: ", zap.Error(err)); return err }
