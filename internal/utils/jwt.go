package utils

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/xianhc/ape-volo-admin-go/internal/global"
)

// JwtAuthOptions 令牌签发与校验。
type JwtAuthOptions struct {
	SigningKey []byte
}

// clockSkew 校验时间声明时允许的时钟偏移。
const clockSkew = 30 * time.Second

// JWT 头部的 typ 值，用于区分访问令牌与刷新令牌。
const (
	tokenTypeAccess  = "at+jwt"
	tokenTypeRefresh = "rt+jwt"
)

// NewJwt 用当前配置构造签发器。
func NewJwt() *JwtAuthOptions {
	return &JwtAuthOptions{SigningKey: []byte(global.Config.JwtAuthOptions.SecurityKey)}
}

// newTokenID 生成 jti，使用 crypto/rand 保证不可预测。
func newTokenID() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("生成 token id 失败: %w", err)
	}
	return hex.EncodeToString(buf), nil
}

func tokenTypeOf(use string) string {
	if use == TokenUseRefresh {
		return tokenTypeRefresh
	}
	return tokenTypeAccess
}

// IssueTokenPair 为一次会话签发访问令牌与刷新令牌。
//
// 访问令牌的过期时刻取「当前时间 + 配置的有效期」与会话绝对截止时间中的较小者，
// 避免访问令牌活得比会话还久。
func (j *JwtAuthOptions) IssueTokenPair(sessionId string, userId int64, name string, deptId int64,
	accessTokenId string, sessionExpiresAt time.Time) (*JwtToken, error) {

	now := time.Now()
	accessExpires := now.Add(time.Duration(global.Config.JwtAuthOptions.AccessTokenLifetimeMs) * time.Millisecond)
	if sessionExpiresAt.Before(accessExpires) {
		accessExpires = sessionExpiresAt
	}

	refreshTokenId, err := newTokenID()
	if err != nil {
		return nil, err
	}

	accessToken, err := j.sign(sessionId, userId, name, deptId, accessTokenId,
		TokenUseAccess, now, accessExpires)
	if err != nil {
		return nil, err
	}

	refreshToken, err := j.sign(sessionId, userId, name, deptId, refreshTokenId,
		TokenUseRefresh, now, sessionExpiresAt)
	if err != nil {
		return nil, err
	}

	return &JwtToken{
		SessionId:           sessionId,
		AccessToken:         accessToken,
		Expires:             accessExpires.UnixMilli(),
		TokenType:           "Bearer",
		RefreshToken:        refreshToken,
		RefreshTokenExpires: sessionExpiresAt.UnixMilli(),
	}, nil
}

func (j *JwtAuthOptions) sign(sessionId string, userId int64, name string, deptId int64,
	tokenId, use string, now, expires time.Time) (string, error) {

	claims := ApeClaims{
		Sid:      sessionId,
		TokenUse: use,
		Name:     name,
		DeptId:   deptId,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatInt(userId, 10),
			ID:        tokenId,
			Issuer:    global.Config.JwtAuthOptions.Issuer,
			Audience:  jwt.ClaimStrings{global.Config.JwtAuthOptions.Audience},
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now.Add(-clockSkew)),
			ExpiresAt: jwt.NewNumericDate(expires),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token.Header["typ"] = tokenTypeOf(use)
	return token.SignedString(j.SigningKey)
}

// ReadJwtToken 解析并校验访问令牌。
func (j *JwtAuthOptions) ReadJwtToken(tokenString string) (*ApeClaims, error) {
	return j.ParseToken(tokenString, TokenUseAccess, true)
}

// ParseToken 解析指定用途的令牌。
//
// validateLifetime 为 false 时跳过时间校验，仅用于登出：
// 退出时访问令牌可能刚好过期，但仍需要识别出它属于哪个会话。
func (j *JwtAuthOptions) ParseToken(tokenString, expectedUse string, validateLifetime bool) (*ApeClaims, error) {
	if tokenString == "" {
		return nil, errors.New("令牌为空")
	}

	claims := &ApeClaims{}
	options := []jwt.ParserOption{
		// 限定只接受 HS256：不限定算法时，攻击者可把 alg 改为 none
		// 或用公钥当 HMAC 密钥（算法混淆）。
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithLeeway(clockSkew),
	}
	if !validateLifetime {
		options = append(options, jwt.WithoutClaimsValidation())
	}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return j.SigningKey, nil
	}, options...)
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, jwt.ErrTokenUnverifiable
	}

	// typ 与 token_use 双重校验：只改其中一个都不足以把刷新令牌当访问令牌用
	if typ, _ := token.Header["typ"].(string); typ != tokenTypeOf(expectedUse) {
		return nil, errors.New("令牌类型不匹配")
	}
	if claims.TokenUse != expectedUse {
		return nil, errors.New("令牌用途不匹配")
	}
	return claims, nil
}

// UserID 返回令牌中的用户 ID。
func UserID(claims *ApeClaims) int64 {
	if claims == nil || claims.Subject == "" {
		return 0
	}
	id, err := strconv.ParseInt(claims.Subject, 10, 64)
	if err != nil || id <= 0 {
		return 0
	}
	return id
}

// SessionID 返回令牌中的会话标识。
func SessionID(claims *ApeClaims) string {
	if claims == nil {
		return ""
	}
	return claims.Sid
}
