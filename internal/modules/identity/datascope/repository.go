package datascope

import (
	"context"
	"errors"

	rolemodule "github.com/xianhc/ape-volo-admin-go/internal/modules/identity/role"
	usermodule "github.com/xianhc/ape-volo-admin-go/internal/modules/identity/user"
	"github.com/xianhc/ape-volo-admin-go/internal/modules/organization/department"
	"github.com/xianhc/ape-volo-admin-go/internal/utils"
	"gorm.io/gorm"
)

var errNotFound = errors.New("data scope user not found")

type repository struct{}

func (repository) findUser(ctx context.Context, userId int64) (usermodule.User, error) {
	var user usermodule.User
	err := utils.FromContext(ctx).Preload("Roles").First(&user, userId).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return usermodule.User{}, errNotFound
	}
	return user, err
}

func (repository) roleDepartments(ctx context.Context, roleId int64) ([]rolemodule.RoleDepartment, error) {
	links := make([]rolemodule.RoleDepartment, 0)
	err := utils.FromContext(ctx).Where("role_id = ?", roleId).Find(&links).Error
	return links, err
}

func (repository) userIdsInDepartments(ctx context.Context, deptIds []int64) ([]int64, error) {
	if len(deptIds) == 0 {
		return nil, nil
	}
	ids := make([]int64, 0)
	err := utils.FromContext(ctx).Model(&usermodule.User{}).Where("dept_id in ?", deptIds).Pluck("id", &ids).Error
	return ids, err
}
func (repository) departmentAndChildrenIDs(ctx context.Context, rootId int64) ([]int64, error) {
	all := []int64{rootId}
	frontier := []int64{rootId}
	for len(frontier) > 0 {
		children := make([]int64, 0)
		err := utils.FromContext(ctx).Model(&department.Department{}).Where("enabled = ? and parent_id in ?", true, frontier).Pluck("id", &children).Error
		if err != nil {
			return nil, err
		}
		if len(children) == 0 {
			break
		}
		all = append(all, children...)
		frontier = children
	}
	return all, nil
}
