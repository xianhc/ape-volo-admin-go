package session

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"time"

	"github.com/xianhc/ape-volo-admin-go/internal/global"
	"github.com/xianhc/ape-volo-admin-go/internal/payloads/auth"
	"github.com/xianhc/ape-volo-admin-go/internal/utils/cache"
)

const sessionKeyPrefix = "auth:session:"

// Session 表示一次只存在于缓存中的登录会话。
//
// 缓存中没有会话即视为未登录，不从数据库重建。
type Session struct {
	Id               string    `json:"id"`
	UserId           int64     `json:"userId"`
	AccessTokenId    string    `json:"accessTokenId"`
	RefreshTokenHash string    `json:"refreshTokenHash"`
	LoginInfo        string    `json:"loginInfoJson"`
	CreatedAt        time.Time `json:"createdAtUtc"`
	ExpiresAt        time.Time `json:"expiresAtUtc"`
}

// LoginUserInfo 还原会话中的登录用户信息。
//
// LoginInfo 为空或 JSON 损坏时返回错误。
func (session *Session) LoginUserInfo() (*auth.LoginUserInfo, error) {
	if session.LoginInfo == "" {
		return nil, errors.New("会话中缺少登录用户信息")
	}
	var info auth.LoginUserInfo
	if err := json.Unmarshal([]byte(session.LoginInfo), &info); err != nil {
		return nil, err
	}
	return &info, nil
}

// Service 管理登录会话的创建、轮换、查询和撤销。
//
// 零值即可使用，缓存后端由 global.Cache 提供。
type Service struct{}

// Create 创建具有绝对有效期的会话。
//
// 比较替换保证同一会话标识不会覆盖已有记录。
func (service *Service) Create(ctx context.Context, session *Session) (bool, error) {
	remaining := time.Until(session.ExpiresAt)
	if session.UserId <= 0 || remaining <= 0 {
		return false, nil
	}
	session.CreatedAt = session.CreatedAt.UTC()
	session.ExpiresAt = session.ExpiresAt.UTC()
	return global.Cache.CompareExchange(ctx, sessionKey(session.Id), nil, session, remaining)
}

// FindActive 读取仍然有效的会话。
//
// 缓存缺失或绝对过期均返回 nil。
func (service *Service) FindActive(ctx context.Context, id string) (*Session, error) {
	if id == "" {
		return nil, nil
	}
	var stored struct {
		Session
		LegacyVersion json.RawMessage `json:"userVersion"`
	}
	if err := global.Cache.Get(ctx, sessionKey(id), &stored); err != nil {
		if errors.Is(err, cache.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	// 包含 userVersion 的缓存记录不作为有效会话，读取时撤销。
	if len(stored.LegacyVersion) != 0 {
		return nil, service.RevokeSession(ctx, id)
	}
	session := stored.Session
	if !session.ExpiresAt.After(time.Now()) {
		return nil, nil
	}
	return &session, nil
}

// Rotate 原子轮换刷新令牌哈希与访问令牌标识。
//
// previousHash 不匹配或并发竞争失败时返回 false。
func (service *Service) Rotate(ctx context.Context, id, previousHash, nextHash, accessTokenId string) (bool, error) {
	current, err := service.FindActive(ctx, id)
	if err != nil || current == nil {
		return false, err
	}
	if current.RefreshTokenHash != previousHash {
		return false, nil
	}
	remaining := time.Until(current.ExpiresAt)
	if remaining <= 0 {
		return false, nil
	}
	next := *current
	next.RefreshTokenHash = nextHash
	next.AccessTokenId = accessTokenId
	return global.Cache.CompareExchange(ctx, sessionKey(id), current, &next, remaining)
}

// RevokeSession 撤销单个会话。
//
// 空会话标识视为无需处理。
func (*Service) RevokeSession(ctx context.Context, id string) error {
	if id == "" {
		return nil
	}
	return global.Cache.Del(ctx, sessionKey(id))
}

// RevokeUser 撤销用户的全部会话。
//
// 枚举会话并删除属于目标用户的记录，不保存额外的撤销版本键。
func (service *Service) RevokeUser(ctx context.Context, userId int64) error {
	if userId <= 0 {
		return nil
	}
	sessions, err := service.ListActive(ctx)
	if err != nil {
		return err
	}
	for _, session := range sessions {
		if session.UserId == userId {
			if err := service.RevokeSession(ctx, session.Id); err != nil {
				return err
			}
		}
	}
	return nil
}

// ListActive 列出全部有效会话。
//
// 损坏或过期会话由 FindActive 过滤。
func (service *Service) ListActive(ctx context.Context) ([]Session, error) {
	keys, err := global.Cache.Keys(ctx, sessionKeyPrefix+"*")
	if err != nil {
		return nil, err
	}
	sessions := make([]Session, 0, len(keys))
	for _, key := range keys {
		session, findErr := service.FindActive(ctx, key[len(sessionKeyPrefix):])
		if findErr != nil {
			return nil, findErr
		}
		if session != nil {
			sessions = append(sessions, *session)
		}
	}
	return sessions, nil
}

// NewID 生成会话与令牌使用的 32 位十六进制标识。
//
// 随机源失败时 panic，避免在不安全状态下继续签发凭据。
func NewID() string { return newID() }

func sessionKey(id string) string { return sessionKeyPrefix + id }
func newID() string {
	buffer := make([]byte, 16)
	if _, err := rand.Read(buffer); err != nil {
		panic("生成会话标识失败: " + err.Error())
	}
	return hex.EncodeToString(buffer)
}
