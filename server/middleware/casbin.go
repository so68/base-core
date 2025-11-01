package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/so68/core/server/module/admin/service"
)

// NewCasbinMiddleware 创建一个 casbin 中间件
func NewCasbinMiddleware(casbinService service.CasbinService) gin.HandlerFunc {
	return func(c *gin.Context) {
		adminRole, err := casbinService.GetContextRole(c)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		// 检查角色是否具有继承权限
		if !casbinService.HasRoleInheritancesEnforce(adminRole, c.Request.URL.Path, c.Request.Method) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "无权限访问"})
			return
		}
		c.Next()
	}
}
