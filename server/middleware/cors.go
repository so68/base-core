package middleware

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/so68/core/config"
)

// NewCORSMiddleware 使用 gin-contrib/cors，并将数组配置直接映射
func NewCORSMiddleware(cfg *config.AppConfig) gin.HandlerFunc {
	if cfg == nil || cfg.Cors == nil {
		return func(c *gin.Context) { c.Next() }
	}

	c := cfg.Cors
	conf := cors.Config{
		AllowCredentials: c.AllowCredentials,
		AllowMethods:     c.AllowMethods,
		AllowHeaders:     c.AllowHeaders,
		ExposeHeaders:    c.ExposeHeaders,
		AllowOrigins:     c.AllowOrigins,
		MaxAge:           time.Duration(c.MaxAge) * time.Second,
	}
	return cors.New(conf)
}
