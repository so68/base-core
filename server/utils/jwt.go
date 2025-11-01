package utils

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/so68/core/cache"
	"github.com/so68/utils"
)

// Claims 自定义JWT声明
type Claims struct {
	SessionID string `json:"session_id"` // 会话ID(UUID)
	IP        string `json:"ip"`         // 客户端IP
	UserID    uint   `json:"user_id"`    // 用户ID(管理员ID或用户ID)
	jwt.RegisteredClaims
}

// JWT JWT配置
type JWT struct {
	secretKey string        // 密钥
	expiresIn time.Duration // 过期时间
	cacheKey  string        // 缓存Key
	cache     cache.Cache   // 缓存
}

// NewJWT 创建一个JWT实例
func NewJWT(secretKey string, expiresIn time.Duration) *JWT {
	return &JWT{
		secretKey: secretKey,
		expiresIn: expiresIn,
		cacheKey:  "jwt:token:" + utils.NewRandomGenerator().String(5),
	}
}

// WithCache 为 JWT 启用基于缓存的 Token 存储/校验（可选）
func (j *JWT) WithCache(c cache.Cache) *JWT {
	j.cache = c
	return j
}

// tokenKey 生成在缓存中保存 Token 的 Key
func (j *JWT) tokenKey(token string) string {
	return fmt.Sprintf("%s:%s", j.cacheKey, token)
}

// GenerateToken 生成JWT Token
func (j *JWT) GenerateToken(userID uint, ip string) string {
	claims := &Claims{
		UserID: userID,
		IP:     ip,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.expiresIn)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(j.secretKey))
	if err != nil {
		return ""
	}

	// 将 Token 存入缓存用于后续校验/注销
	if j.cache != nil {
		_ = j.cache.Set(context.Background(), j.tokenKey(signed), userID, j.expiresIn)
	}

	return signed
}

// ParseToken 解析JWT Token
func (j *JWT) ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(j.secretKey), nil
	})
	if err != nil {
		return nil, err
	}

	// 验证Token是否有效
	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	// 如果启用了缓存，校验 Token 是否仍然有效（未被注销）
	if j.cache != nil {
		exists, err := j.cache.Exists(context.Background(), j.tokenKey(tokenString))
		if err == nil && !exists {
			return nil, errors.New("token revoked or expired")
		}
	}

	return token.Claims.(*Claims), nil
}

// RevokeToken 主动使 Token 失效（从缓存中删除）
func (j *JWT) RevokeToken(tokenString string) error {
	if j.cache == nil {
		return nil
	}
	return j.cache.Delete(context.Background(), j.tokenKey(tokenString))
}

// GetRequestToken 获取请求头中的Token
func GetRequestToken(c *gin.Context) string {
	// 优先级：Query > Header
	token := c.Query("token")
	if token != "" {
		return token
	}

	// 获取请求头中的Token
	token = c.GetHeader("Authorization")
	if token != "" {
		tokenList := strings.Split(token, " ")
		if len(tokenList) == 2 && tokenList[0] == "Bearer" {
			return tokenList[1]
		}
	}

	// 如果没有Token，则返回空字符串
	return ""
}
