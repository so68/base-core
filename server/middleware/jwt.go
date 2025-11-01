package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/so68/core/server/utils"
)

// NewJWTMiddleware 创建一个 jwt 中间件
func NewJWTMiddleware(jwt *utils.JWT) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := utils.GetRequestToken(c)
		claims, err := jwt.ParseToken(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		// 如果IP不匹配，则返回401
		if claims.IP != c.ClientIP() {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "IP not match"})
			return
		}

		// 设置用户ID
		c.Set(utils.ContextUserIDKey, claims.UserID)
		c.Next()
	}
}
