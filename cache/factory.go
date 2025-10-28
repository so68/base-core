package cache

import (
	"fmt"
	"log/slog"

	"github.com/so68/core/config"
)

// Factory 缓存工厂
type Factory struct {
	logger *slog.Logger
}

// NewFactory 创建缓存工厂
func NewFactory(logger *slog.Logger) *Factory {
	return &Factory{
		logger: logger,
	}
}

// CreateCache 根据配置创建缓存连接
func (f *Factory) CreateCache(cfg *config.CacheConfig) (Cache, error) {
	switch cfg.Driver {
	case "redis":
		return NewRedisCache(cfg, f.logger)
	case "memory":
		return NewMemoryCache(cfg, f.logger)
	default:
		return nil, fmt.Errorf("unsupported cache driver: %s", cfg.Driver)
	}
}

// CreateRedisCache 创建 Redis 缓存连接
func (f *Factory) CreateRedisCache(host string, port int, password string, database int) (Cache, error) {
	cfg := &config.CacheConfig{
		Driver:   "redis",
		Host:     host,
		Port:     port,
		Password: password,
		Database: database,
	}

	cfg.SetDefaults()
	return f.CreateCache(cfg)
}

// CreateMemoryCache 创建内存缓存连接
func (f *Factory) CreateMemoryCache() (Cache, error) {
	cfg := &config.CacheConfig{
		Driver: "memory",
	}

	cfg.SetDefaults()
	return f.CreateCache(cfg)
}
