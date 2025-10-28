package config

import (
	"time"
)

// CacheConfig 缓存配置
type CacheConfig struct {
	Driver string `yaml:"driver" json:"driver"` // 缓存驱动: redis, memory

	// Redis 配置
	Host     string `yaml:"host" json:"host"`         // Redis 主机
	Port     int    `yaml:"port" json:"port"`         // Redis 端口
	Password string `yaml:"password" json:"password"` // Redis 密码
	Database int    `yaml:"database" json:"database"` // Redis 数据库编号
	Prefix   string `yaml:"prefix" json:"prefix"`     // 键前缀

	// 连接池配置
	MaxRetries      int           `yaml:"max_retries" json:"max_retries"`             // 最大重试次数
	MinRetryBackoff time.Duration `yaml:"min_retry_backoff" json:"min_retry_backoff"` // 最小重试间隔
	MaxRetryBackoff time.Duration `yaml:"max_retry_backoff" json:"max_retry_backoff"` // 最大重试间隔
	DialTimeout     time.Duration `yaml:"dial_timeout" json:"dial_timeout"`           // 连接超时
	ReadTimeout     time.Duration `yaml:"read_timeout" json:"read_timeout"`           // 读取超时
	WriteTimeout    time.Duration `yaml:"write_timeout" json:"write_timeout"`         // 写入超时
	PoolSize        int           `yaml:"pool_size" json:"pool_size"`                 // 连接池大小
	MinIdleConns    int           `yaml:"min_idle_conns" json:"min_idle_conns"`       // 最小空闲连接数
	MaxConnAge      time.Duration `yaml:"max_conn_age" json:"max_conn_age"`           // 连接最大生存时间
	PoolTimeout     time.Duration `yaml:"pool_timeout" json:"pool_timeout"`           // 连接池超时
	IdleTimeout     time.Duration `yaml:"idle_timeout" json:"idle_timeout"`           // 空闲超时
	IdleCheckFreq   time.Duration `yaml:"idle_check_freq" json:"idle_check_freq"`     // 空闲检查频率

	// 内存缓存配置
	MaxMemory       int64         `yaml:"max_memory" json:"max_memory"`             // 最大内存使用量（字节）
	CleanupInterval time.Duration `yaml:"cleanup_interval" json:"cleanup_interval"` // 清理间隔
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
