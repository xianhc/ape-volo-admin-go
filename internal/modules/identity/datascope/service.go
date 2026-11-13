package datascope

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/xianhc/ape-volo-admin-go/internal/global/constants/cachePrefix"
	rolemodule "github.com/xianhc/ape-volo-admin-go/internal/modules/identity/role"
	usermodule "github.com/xianhc/ape-volo-admin-go/internal/modules/identity/user"
	"github.com/xianhc/ape-volo-admin-go/internal/utils"
	"github.com/xianhc/ape-volo-admin-go/internal/utils/querycache"
)

// Snapshot 表示一次请求可见的创建者范围。
//
// AllowAll 为 true 时不过滤，Restricted 为 true 且 UserIds 为空时查询应恒为空。
type Snapshot struct {
	Restricted bool    `json:"-"`
	AllowAll   bool    `json:"allowAll"`
	UserIds    []int64 `json:"userIds"`
}

// Service 解析用户角色配置的数据权限快照。
//
// 结果按用户缓存一小时，并在请求内由 ORM 回调复用。
type Service struct{ repository repository }

type snapshotCacheKey struct{}
type snapshotCache struct {
	mu     sync.Mutex
	userId int64
	value  *Snapshot
}

// WithSnapshotCache 让同一请求的查询和写入共用一次数据权限解析结果。
func WithSnapshotCache(ctx context.Context) context.Context {
	if _, ok := ctx.Value(snapshotCacheKey{}).(*snapshotCache); ok {
		return ctx
	}
	return context.WithValue(ctx, snapshotCacheKey{}, &snapshotCache{})
}

// Resolve 返回用户的数据权限快照。
//
// 非正用户主键表示后台场景，不施加数据权限过滤。
func (service *Service) Resolve(ctx context.Context, userId int64) (*Snapshot, error) {
	if userId <= 0 {
		return &Snapshot{}, nil
	}
	if memo, ok := ctx.Value(snapshotCacheKey{}).(*snapshotCache); ok {
		memo.mu.Lock()
		defer memo.mu.Unlock()
		if memo.value != nil && memo.userId == userId {
			return memo.value, nil
		}
		value, err := service.resolve(ctx, userId)
		if err == nil {
			memo.userId, memo.value = userId, value
		}
		return value, err
	}
	return service.resolve(ctx, userId)
}

func (service *Service) resolve(ctx context.Context, userId int64) (*Snapshot, error) {
	key := querycache.Key(cachePrefix.UserDataScopeById, userId)
	snapshot := Snapshot{Restricted: true, UserIds: make([]int64, 0)}
	err := querycache.GetOrLoad(ctx, key, &snapshot, time.Hour, func() error {
		computed, err := service.compute(utils.WithoutDataScope(ctx), userId)
		if err == nil {
			snapshot = *computed
		}
		return err
	})
	return &snapshot, err
}

func (service *Service) compute(ctx context.Context, userId int64) (*Snapshot, error) {
	user, err := service.repository.findUser(ctx, userId)
	if err != nil {
		if errors.Is(err, errNotFound) {
			return &Snapshot{Restricted: true, UserIds: make([]int64, 0)}, nil
		}
		return nil, err
	}
	if len(user.Roles) == 0 {
		return &Snapshot{Restricted: true, UserIds: make([]int64, 0)}, nil
	}
	snapshot := &Snapshot{Restricted: true, UserIds: make([]int64, 0)}
	for _, role := range user.Roles {
		if role.DataScopeType == rolemodule.DataScopeAll {
			snapshot.AllowAll = true
			return snapshot, nil
		}
	}
	seen := map[int64]struct{}{}
	for index := range user.Roles {
		ids, err := service.userIdsByRole(ctx, &user, &user.Roles[index])
		if err != nil {
			return nil, err
		}
		for _, id := range ids {
			if _, ok := seen[id]; ok {
				continue
			}
			seen[id] = struct{}{}
			snapshot.UserIds = append(snapshot.UserIds, id)
		}
	}
	return snapshot, nil
}
func (service *Service) userIdsByRole(ctx context.Context, user *usermodule.User, role *rolemodule.Role) ([]int64, error) {
	switch role.DataScopeType {
	case rolemodule.DataScopeMySelf:
		return []int64{user.Id}, nil
	case rolemodule.DataScopeMyDept:
		return service.repository.userIdsInDepartments(ctx, []int64{user.DeptId})
	case rolemodule.DataScopeMyDeptAndBelow:
		ids, err := service.repository.departmentAndChildrenIDs(ctx, user.DeptId)
		if err != nil {
			return nil, err
		}
		return service.repository.userIdsInDepartments(ctx, ids)
	case rolemodule.DataScopeCustomize:
		links, err := service.repository.roleDepartments(ctx, role.Id)
		if err != nil {
			return nil, err
		}
		ids := make([]int64, 0, len(links))
		for _, link := range links {
			ids = utils.AppendInt64(ids, link.DeptId)
		}
		return service.repository.userIdsInDepartments(ctx, ids)
	default:
		return nil, nil
	}
}
