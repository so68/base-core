package config

import (
	"time"
)

// CacheConfig 缓存配置
type CacheConfig struct {
	Driver string `yaml:"driver"` // 缓存驱动: redis, memory

	// Redis 配置
	Host     string `yaml:"host"`     // Redis 主机
	Port     int    `yaml:"port"`     // Redis 端口
	Password string `yaml:"password"` // Redis 密码
	Database int    `yaml:"database"` // Redis 数据库编号
	Prefix   string `yaml:"prefix"`   // 键前缀

	// 连接池配置
	MaxRetries      int           `yaml:"maxRetries"`      // 最大重试次数
	MinRetryBackoff time.Duration `yaml:"minRetryBackoff"` // 最小重试间隔
	MaxRetryBackoff time.Duration `yaml:"maxRetryBackoff"` // 最大重试间隔
	DialTimeout     time.Duration `yaml:"dialTimeout"`     // 连接超时
	ReadTimeout     time.Duration `yaml:"readTimeout"`     // 读取超时
	WriteTimeout    time.Duration `yaml:"writeTimeout"`    // 写入超时
	PoolSize        int           `yaml:"poolSize"`        // 连接池大小
	MinIdleConns    int           `yaml:"minIdleConns"`    // 最小空闲连接数
	MaxConnAge      time.Duration `yaml:"maxConnAge"`      // 连接最大生存时间
	PoolTimeout     time.Duration `yaml:"poolTimeout"`     // 连接池超时
	IdleTimeout     time.Duration `yaml:"idleTimeout"`     // 空闲超时
	IdleCheckFreq   time.Duration `yaml:"idleCheckFreq"`   // 空闲检查频率

	// 内存缓存配置
	MaxMemory       int64         `yaml:"maxMemory"`       // 最大内存使用量（字节）
	CleanupInterval time.Duration `yaml:"cleanupInterval"` // 清理间隔
}

// DefaultCacheConfig 返回默认缓存配置
func DefaultCacheConfig() *CacheConfig {
	return &CacheConfig{
		Driver:   "redis",
		Host:     "localhost",
		Port:     6379,
		Database: 0,
		Prefix:   "",

		MaxRetries:      3,
		MinRetryBackoff: 8 * time.Millisecond,
		MaxRetryBackoff: 512 * time.Millisecond,
		DialTimeout:     5 * time.Second,
		ReadTimeout:     3 * time.Second,
		WriteTimeout:    3 * time.Second,
		PoolSize:        10,
		MinIdleConns:    5,
		MaxConnAge:      time.Hour,
		PoolTimeout:     4 * time.Second,
		IdleTimeout:     5 * time.Minute,
		IdleCheckFreq:   1 * time.Minute,

		MaxMemory:       100 * 1024 * 1024, // 100MB
		CleanupInterval: 10 * time.Minute,
	}
}

// SetDefaults 设置默认配置值
func (c *CacheConfig) SetDefaults() {
	if c.Driver == "" {
		c.Driver = "redis"
	}
	if c.Host == "" {
		c.Host = "localhost"
	}
	if c.Port == 0 {
		c.Port = 6379
	}
	if c.Prefix == "" {
		c.Prefix = ""
	}
	if c.MaxRetries == 0 {
		c.MaxRetries = 3
	}
	if c.MinRetryBackoff == 0 {
		c.MinRetryBackoff = 8 * time.Millisecond
	}
	if c.MaxRetryBackoff == 0 {
		c.MaxRetryBackoff = 512 * time.Millisecond
	}
	if c.DialTimeout == 0 {
		c.DialTimeout = 5 * time.Second
	}
	if c.ReadTimeout == 0 {
		c.ReadTimeout = 3 * time.Second
	}
	if c.WriteTimeout == 0 {
		c.WriteTimeout = 3 * time.Second
	}
	if c.PoolSize == 0 {
		c.PoolSize = 10
	}
	if c.MinIdleConns == 0 {
		c.MinIdleConns = 5
	}
	if c.MaxConnAge == 0 {
		c.MaxConnAge = time.Hour
	}
	if c.PoolTimeout == 0 {
		c.PoolTimeout = 4 * time.Second
	}
	if c.IdleTimeout == 0 {
		c.IdleTimeout = 5 * time.Minute
	}
	if c.IdleCheckFreq == 0 {
		c.IdleCheckFreq = 1 * time.Minute
	}
	if c.MaxMemory == 0 {
		c.MaxMemory = 100 * 1024 * 1024 // 100MB
	}
	if c.CleanupInterval == 0 {
		c.CleanupInterval = 10 * time.Minute
	}
}
