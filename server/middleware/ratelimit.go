package middleware

import (
	"net/http"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/so68/core/config"
	"golang.org/x/time/rate"
)

// ipLimiterStore 按客户端 IP 存储限流器
var ipLimiterStore sync.Map // map[string]*rate.Limiter

func getOrCreateLimiter(ip string, r rate.Limit, burst int) *rate.Limiter {
	if v, ok := ipLimiterStore.Load(ip); ok {
		lim := v.(*rate.Limiter)
		return lim
	}
	lim := rate.NewLimiter(r, burst)
	actual, _ := ipLimiterStore.LoadOrStore(ip, lim)
	return actual.(*rate.Limiter)
}

// NewIPRateLimitMiddleware 基于 x/time/rate 的按 IP 限流中间件
// - 支持包含/排除路径
// - 仅当 cfg.RateLimit 非空且 Rate>0 时才应被启用
func NewIPRateLimitMiddleware(cfg *config.AppConfig) gin.HandlerFunc {
	// 基本参数
	limitPerSec := cfg.RateLimit.Rate
	burst := cfg.RateLimit.Burst
	include := cfg.RateLimit.IncludePaths
	exclude := cfg.RateLimit.ExcludePaths

	// 自动排除静态资源路径（如 /static）
	if s := strings.TrimSpace(cfg.Static); s != "" {
		// 规范化为以 / 开头的前缀
		if !strings.HasPrefix(s, "/") {
			s = "/" + s
		}
		exclude = append(exclude, s)
	}

	shouldCheck := func(path string) bool {
		// 排除优先
		for _, p := range exclude {
			if p != "" && strings.HasPrefix(path, p) {
				return false
			}
		}
		// 若包含列表为空，表示对所有路径生效
		if len(include) == 0 {
			return true
		}
		for _, p := range include {
			if p != "" && strings.HasPrefix(path, p) {
				return true
			}
		}
		return false
	}

	return func(c *gin.Context) {
		if cfg.RateLimit == nil || limitPerSec <= 0 {
			c.Next()
			return
		}
		if !shouldCheck(c.Request.URL.Path) {
			c.Next()
			return
		}

		ip := c.ClientIP()
		limiter := getOrCreateLimiter(ip, rate.Limit(limitPerSec), burst)
		if !limiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limited"})
			return
		}
		c.Next()
	}
}
