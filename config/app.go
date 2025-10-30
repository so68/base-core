package config

import (
	"time"

	"github.com/so68/utils/logger"
)

// AppConfig 应用基础配置
type AppConfig struct {
	Name   string `yaml:"name"`   // 应用名称
	Host   string `yaml:"host"`   // 服务主机
	Port   int    `yaml:"port"`   // 服务端口
	Debug  bool   `yaml:"debug"`  // 是否开启调试模式
	Static string `yaml:"static"` // 静态文件目录（例如：static 或 public）

	// 超时配置
	ReadTimeout  string `yaml:"readTimeout"`  // 读取超时时间
	WriteTimeout string `yaml:"writeTimeout"` // 写入超时时间
	IdleTimeout  string `yaml:"idleTimeout"`  // 空闲超时时间
	MaxHeader    int    `yaml:"maxHeader"`    // 最大请求头大小(bytes)

	// Cors配置
	Cors *CorsConfig `yaml:"cors"`

	// 限流配置
	RateLimit *RateLimitConfig `yaml:"rateLimit"`

	// JWT配置
	JWT *JWTConfig `yaml:"jwt"`

	// 日志配置
	Logger *logger.Config `yaml:"logger"`

	// 缓存配置
	Cache *CacheConfig `yaml:"cache"`

	// 数据库配置
	Database *DatabaseConfig `yaml:"database"`
}

// CorsConfig Cors配置
type CorsConfig struct {
	AllowOrigins     []string `yaml:"allowOrigins"`     // 允许的跨域请求源
	AllowMethods     []string `yaml:"allowMethods"`     // 允许的请求方法
	AllowHeaders     []string `yaml:"allowHeaders"`     // 允许的请求头
	ExposeHeaders    []string `yaml:"expose_headers"`   // 暴露的响应头
	AllowCredentials bool     `yaml:"allowCredentials"` // 是否允许携带凭证
	MaxAge           int      `yaml:"maxAge"`           // 预检请求的缓存时间
}

// JWTConfig JWT配置
type JWTConfig struct {
	ExpiresIn  int    `yaml:"expiresIn"`  // JWT 过期时间(秒)
	SecretKey  string `yaml:"secretKey"`  // JWT 密钥
	HeaderName string `yaml:"headerName"` // JWT 头部名称
	Scheme     string `yaml:"scheme"`     // JWT 方案
}

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	Rate         int      `yaml:"rate"`         // 每秒请求数限制
	Burst        int      `yaml:"burst"`        // 突发请求数限制
	IncludePaths []string `yaml:"includePaths"` // 包含限流的路径
	ExcludePaths []string `yaml:"excludePaths"` // 排除限流的路径
}

// DefaultAppConfig 返回默认应用配置
func DefaultAppConfig() *AppConfig {
	return &AppConfig{
		Name:         "Taozijun Network Technology Co., Ltd.",
		Debug:        true,
		Host:         "0.0.0.0",
		Port:         8000,
		Static:       "static",
		ReadTimeout:  "30s",
		WriteTimeout: "30s",
		IdleTimeout:  "60s",
		MaxHeader:    1 << 20, // 1MB

		Cors: &CorsConfig{
			AllowOrigins:     []string{"*"},
			AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
			AllowHeaders:     []string{"Authorization", "Content-Type"},
			AllowCredentials: false,
			MaxAge:           600, // 10 minutes
		},

		JWT: &JWTConfig{
			ExpiresIn:  3600, // 1 hour
			SecretKey:  "not-secret-key",
			HeaderName: "Authorization",
			Scheme:     "Bearer",
		},
		RateLimit: &RateLimitConfig{
			Rate:         100,
			Burst:        200,
			IncludePaths: []string{},
			ExcludePaths: []string{},
		},
		Logger:   logger.DefaultConfig(),
		Cache:    DefaultCacheConfig(),
		Database: DefaultDatabaseConfig(),
	}
}

// SetDefaults 设置默认配置值
func (c *AppConfig) SetDefaults() {
	// 基础配置
	if c.Name == "" {
		c.Name = "Taozijun Network Technology Co., Ltd."
	}
	if c.Host == "" {
		c.Host = "0.0.0.0"
	}
	if c.Port == 0 {
		c.Port = 8080
	}
	if c.Static == "" {
		c.Static = "static"
	}
	if c.ReadTimeout == "" {
		c.ReadTimeout = "30s"
	}
	if c.WriteTimeout == "" {
		c.WriteTimeout = "30s"
	}
	if c.IdleTimeout == "" {
		c.IdleTimeout = "60s"
	}
	if c.MaxHeader == 0 {
		c.MaxHeader = 1 << 20 // 1MB
	}

	// Cors 配置
	if c.Cors == nil {
		c.Cors = &CorsConfig{}
	}
	if len(c.Cors.AllowOrigins) == 0 {
		c.Cors.AllowOrigins = []string{"*"}
	}
	if len(c.Cors.AllowMethods) == 0 {
		c.Cors.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	}
	if len(c.Cors.AllowHeaders) == 0 {
		c.Cors.AllowHeaders = []string{"Authorization", "Content-Type"}
	}
	if c.Cors.MaxAge == 0 {
		c.Cors.MaxAge = 600 // 10 minutes
	}

	// JWT 配置
	if c.JWT == nil {
		c.JWT = &JWTConfig{}
	}
	if c.JWT.ExpiresIn == 0 {
		c.JWT.ExpiresIn = 3600 // 1 hour
	}
	if c.JWT.SecretKey == "" {
		c.JWT.SecretKey = "not-secret-key"
	}
	if c.JWT.HeaderName == "" {
		c.JWT.HeaderName = "Authorization"
	}
	if c.JWT.Scheme == "" {
		c.JWT.Scheme = "Bearer"
	}

	// RateLimit
	if c.RateLimit == nil {
		c.RateLimit = &RateLimitConfig{}
	}
	if c.RateLimit.Rate == 0 {
		c.RateLimit.Rate = 100
	}
	if c.RateLimit.Burst == 0 {
		c.RateLimit.Burst = 200
	}
	if c.RateLimit.IncludePaths == nil {
		c.RateLimit.IncludePaths = []string{}
	}
	if c.RateLimit.ExcludePaths == nil {
		c.RateLimit.ExcludePaths = []string{}
	}

	// 子配置默认值
	if c.Cache != nil {
		c.Cache.SetDefaults()
	} else {
		c.Cache = DefaultCacheConfig()
	}
	if c.Database != nil {
		c.Database.SetDefaults()
	} else {
		c.Database = DefaultDatabaseConfig()
	}
	if c.Logger != nil {
		c.Logger.SetDefaults()
	} else {
		c.Logger = logger.DefaultConfig()
	}

	// 如果是不是Debug模式
	if !c.Debug {
		c.Logger.Level = logger.LevelInfo
		c.Database.LogLevel = "warn"
		c.Cors.MaxAge = 86400 // 24 hours
	}
}

// ParseDuration 解析时间
func (c *AppConfig) ParseDuration(value string) time.Duration {
	duration, err := time.ParseDuration(value)
	if err != nil {
		// 默认5秒
		return 5 * time.Second
	}
	return duration
}
