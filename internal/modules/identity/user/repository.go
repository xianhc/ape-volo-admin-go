package user

import (
	"context"
	"errors"
	"time"

	"github.com/xianhc/ape-volo-admin-go/internal/modules/organization/position"
	"github.com/xianhc/ape-volo-admin-go/internal/utils"
	"gorm.io/gorm"
)

// errNotFound 表示未找到可见的用户。
//
// repository 用该错误隔离 GORM 的记录不存在错误。
var errNotFound = errors.New("user not found")

// queryOptions 描述用户列表查询的数据库参数。
//
// DeptIds 为空时不按部门过滤。
type queryOptions struct {
	Id            *int64
	UserName      string
	NickName      string
	KeyWords      string
	Enabled       *bool
	DeptIds       []int64
	CreateTime    []time.Time
	Sorting       string
	Limit, Offset int
}

// repository 封装用户主表和用户关联表的数据库访问。
//
// 零值即可使用，写操作复用请求事务。
type repository struct{}

// CountByField 统计排除指定主键后的字段值数量。
//
// excludeId 为零时不排除任何记录。
func (repository) CountByField(ctx context.Context, field string, value any, excludeId int64) (int64, error) {
	var total int64
	db := utils.FromContext(ctx).Model(&User{}).Where(field+" = ?", value)
	if excludeId > 0 {
		db = db.Where("id <> ?", excludeId)
	}
	err := db.Count(&total).Error
	return total, err
}

// FindByID 按主键读取用户和可选关联。
//
// withoutScope 为 true 时绕过数据权限，供认证和个人中心使用。
func (repository) FindByID(ctx context.Context, id int64, details, withoutScope bool) (User, error) {
	if withoutScope {
		ctx = utils.WithoutDataScope(ctx)
	}
	var user User
	db := utils.FromContext(ctx)
	if details {
		db = db.Preload("Positions").Preload("Roles").Preload("Dept")
	}
	err := db.First(&user, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return User{}, errNotFound
	}
	return user, err
}

// FindByName 按用户名读取认证所需用户。
//
// 查询绕过数据权限，但不预加载关联。
func (repository) FindByName(ctx context.Context, name string) (User, error) {
	ctx = utils.WithoutDataScope(ctx)
	var user User
	err := utils.FromContext(ctx).Where("user_name = ?", name).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return User{}, errNotFound
	}
	return user, err
}

// Create 创建用户并写入岗位与角色关联。
//
// 调用方事务保证主表和关联表原子写入。
func (repository) Create(ctx context.Context, user *User, positionIds, roleIds []int64) error {
	db := utils.FromContext(ctx)
	if err := db.Create(user).Error; err != nil {
		return err
	}
	return replaceRelations(db, user.Id, positionIds, roleIds)
}

// Update 更新用户并替换岗位与角色关联。
//
// 调用方事务保证主表和关联表原子写入。
func (repository) Update(ctx context.Context, user *User, positionIds, roleIds []int64) error {
	db := utils.FromContext(ctx)
	if err := db.Select("UserName", "NickName", "Email", "Enabled", "DeptId", "Phone", "GenderCode", "LastModifierId", "LastModificationTime").Updates(user).Error; err != nil {
		return err
	}
	return replaceRelations(db, user.Id, positionIds, roleIds)
}

// Delete 逻辑删除指定用户。
//
// 返回受影响行数和数据库错误。
func (repository) Delete(ctx context.Context, ids []string) (int64, error) {
	result := utils.FromContext(ctx).Delete(&User{}, ids)
	return result.RowsAffected, result.Error
}

// FindByIDs 查询指定用户集合。
//
// 结果供删除后会话撤销使用。
func (repository) FindByIDs(ctx context.Context, ids []string) ([]User, error) {
	users := make([]User, 0)
	err := utils.FromContext(ctx).Find(&users, "id in ?", ids).Error
	return users, err
}

// List 查询一页用户及完整关联。
//
// 返回分页前总数。
func (repository) List(ctx context.Context, o queryOptions) ([]User, int64, error) {
	users := make([]User, 0)
	var total int64
	db := buildQuery(utils.FromContext(ctx).Model(&User{}), o)
	if err := db.Count(&total).Limit(o.Limit).Offset(o.Offset).Preload("Positions").Preload("Roles").Preload("Dept").Find(&users).Error; err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

// FindAll 查询符合条件的全部用户及关联。
//
// 该方法用于导出。
func (repository) FindAll(ctx context.Context, o queryOptions) ([]User, error) {
	users := make([]User, 0)
	err := buildQuery(utils.FromContext(ctx).Model(&User{}), o).Preload("Positions").Preload("Roles").Preload("Dept").Find(&users).Error
	return users, err
}

// UpdateFields 按主键更新指定字段。
//
// 个人中心使用该方法更新有限字段。
func (repository) UpdateFields(ctx context.Context, id int64, fields map[string]any) error {
	return utils.FromContext(utils.WithoutDataScope(ctx)).Model(&User{}).Where("id = ?", id).Updates(fields).Error
}

// ReplacePositions 替换用户岗位关联。
//
// 调用方事务保证删除和新增原子完成。
func (repository) ReplacePositions(ctx context.Context, userId int64, ids []int64) error {
	db := utils.FromContext(ctx)
	if err := db.Where("user_id = ?", userId).Delete(&position.UserPosition{}).Error; err != nil {
		return err
	}
	if len(ids) == 0 {
		return nil
	}
	rows := make([]position.UserPosition, 0, len(ids))
	for _, id := range ids {
		rows = append(rows, position.UserPosition{UserId: userId, PositionId: id})
	}
	return db.Create(&rows).Error
}

// ReplaceRoles 替换用户角色关联。
//
// 调用方事务保证删除和新增原子完成。
func (repository) ReplaceRoles(ctx context.Context, userId int64, ids []int64) error {
	db := utils.FromContext(ctx)
	if err := db.Where("user_id = ?", userId).Delete(&UserRole{}).Error; err != nil {
		return err
	}
	if len(ids) == 0 {
		return nil
	}
	rows := make([]UserRole, 0, len(ids))
	for _, id := range ids {
		rows = append(rows, UserRole{UserId: userId, RoleId: id})
	}
	return db.Create(&rows).Error
}

func replaceRelations(db *gorm.DB, userId int64, positionIds, roleIds []int64) error {
	if err := db.Where("user_id = ?", userId).Delete(&position.UserPosition{}).Error; err != nil {
		return err
	}
	if len(positionIds) > 0 {
		rows := make([]position.UserPosition, 0, len(positionIds))
		for _, id := range positionIds {
			rows = append(rows, position.UserPosition{UserId: userId, PositionId: id})
		}
		if err := db.Create(&rows).Error; err != nil {
			return err
		}
	}
	if err := db.Where("user_id = ?", userId).Delete(&UserRole{}).Error; err != nil {
		return err
	}
	if len(roleIds) > 0 {
		rows := make([]UserRole, 0, len(roleIds))
		for _, id := range roleIds {
			rows = append(rows, UserRole{UserId: userId, RoleId: id})
		}
		if err := db.Create(&rows).Error; err != nil {
			return err
		}
	}
	return nil
}
func buildQuery(db *gorm.DB, o queryOptions) *gorm.DB {
	if o.Id != nil {
		db = db.Where("id = ?", *o.Id)
	}
	if o.UserName != "" {
		db = db.Where("user_name LIKE ?", "%"+o.UserName+"%")
	}
	if o.NickName != "" {
		db = db.Where("nick_name LIKE ?", "%"+o.NickName+"%")
	}
	if o.Enabled != nil {
		db = db.Where("enabled = ?", o.Enabled)
	}
	if o.KeyWords != "" {
		db = db.Where("user_name LIKE ? or nick_name LIKE ? or email LIKE ?", "%"+o.KeyWords+"%", "%"+o.KeyWords+"%", "%"+o.KeyWords+"%")
	}
	if len(o.CreateTime) != 0 && len(o.CreateTime) != 2 {
		_ = db.AddError(errors.New("创建时间必须包含起止两个值"))
		return db
	}
	if len(o.CreateTime) == 2 {
		db = db.Where("creation_time >= ? and creation_time <= ?", o.CreateTime[0], o.CreateTime[1])
	}
	if len(o.DeptIds) > 0 {
		db = db.Where("dept_id in ?", o.DeptIds)
	}
	if o.Sorting != "" {
		db = utils.ApplySorting(db, o.Sorting)
	}
	return db
}
