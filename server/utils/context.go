package utils

import "github.com/gin-gonic/gin"

const (
	ContextUserIDKey = "user_id" // 用户ID
)

// GetContextUserID 获取用户ID
func GetContextUserID(c *gin.Context) uint {
	return c.GetUint(ContextUserIDKey)
}
