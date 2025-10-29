package core

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/so68/core/cache"
	"github.com/so68/core/config"
	"github.com/so68/core/database"
	"github.com/so68/core/server"
	"github.com/so68/utils/logger"
)

// Core 核心
type Core struct {
	Config *config.AppConfig // 配置
	Logger *slog.Logger      // 日志
	DB     database.Database // 数据库
	Cache  cache.Cache       // 缓存
	Server server.Server     // 服务器
}

// Option 构造可选项
type Option func(*coreOptions)

type coreOptions struct {
	configPath   string
	cfg          *config.AppConfig
	logger       *slog.Logger
	enableDB     bool
	enableCache  bool
	enableServer bool
}

// WithConfigPath 指定配置文件路径
func WithConfigPath(path string) Option {
	return func(o *coreOptions) { o.configPath = path }
}

// WithConfig 直接传入配置
func WithConfig(cfg *config.AppConfig) Option {
	return func(o *coreOptions) { o.cfg = cfg }
}

// WithLogger 传入自定义日志器
func WithLogger(l *slog.Logger) Option {
	return func(o *coreOptions) { o.logger = l }
}

// WithoutDB 禁用数据库
func WithoutDB() Option {
	return func(o *coreOptions) { o.enableDB = false }
}

// WithoutCache 禁用缓存
func WithoutCache() Option {
	return func(o *coreOptions) { o.enableCache = false }
}

// WithoutServer 禁用服务器
func WithoutServer() Option {
	return func(o *coreOptions) { o.enableServer = false }
}

// NewCore 初始化核心
func NewCore(configPath string) (*Core, error) {
	return NewCoreWithOptions(WithConfigPath(configPath))
}

// NewCoreWithOptions 基于可选项初始化核心
func NewCoreWithOptions(opts ...Option) (*Core, error) {
	// 解析可选项
	o := &coreOptions{
		enableDB:     true,
		enableCache:  true,
		enableServer: true,
	}
	for _, opt := range opts {
		opt(o)
	}

	// 加载配置
	var cfg *config.AppConfig
	if o.cfg != nil {
		cfg = o.cfg
	} else {
		loaded, err := config.LoadConfig(o.configPath)
		if err != nil {
			return nil, fmt.Errorf("load config: %w", err)
		}
		cfg = loaded
	}

	// 初始化日志
	var slogLogger *slog.Logger
	if o.logger != nil {
		slogLogger = o.logger
	} else {
		l, err := logger.NewLogger(cfg.Logger)
		if err != nil {
			return nil, fmt.Errorf("init logger: %w", err)
		}
		slogLogger = l
	}

	// 初始化数据库
	var db database.Database
	if o.enableDB {
		dbFactory := database.NewFactory(slogLogger)
		createdDB, err := dbFactory.CreateDatabase(cfg.Database)
		if err != nil {
			return nil, fmt.Errorf("init database: %w", err)
		}
		db = createdDB
	}

	// 初始化缓存
	var c cache.Cache
	if o.enableCache {
		cacheFactory := cache.NewFactory(slogLogger)
		createdCache, err := cacheFactory.CreateCache(cfg.Cache)
		if err != nil {
			return nil, fmt.Errorf("init cache: %w", err)
		}
		c = createdCache
	}

	// 初始化服务器
	var s server.Server
	if o.enableServer {
		s = server.NewServer(slogLogger, cfg)
		err := s.Start()
		if err != nil {
			return nil, fmt.Errorf("start server: %w", err)
		}
	}

	return &Core{
		Config: cfg,
		Logger: slogLogger,
		DB:     db,
		Cache:  c,
		Server: s,
	}, nil
}

// Close 统一释放资源
func (c *Core) Close(ctx context.Context) error {
	var firstErr error

	if c.DB != nil {
		if err := c.DB.Close(ctx); err != nil && firstErr == nil {
			firstErr = fmt.Errorf("close db: %w", err)
		}
	}

	if c.Cache != nil {
		if err := c.Cache.Close(); err != nil && firstErr == nil {
			firstErr = fmt.Errorf("close cache: %w", err)
		}
	}

	return firstErr
}

// Health 聚合健康检查
func (c *Core) Health(ctx context.Context) error {
	if c.DB != nil {
		if err := c.DB.HealthCheck(); err != nil {
			c.Logger.Error("db health check failed", slog.String("component", "db"), slog.Any("error", err))
			return fmt.Errorf("db unhealthy: %w", err)
		}
	} else {
		c.Logger.Info("db disabled or not initialized", slog.String("component", "db"))
	}
	if c.Cache != nil {
		if err := c.Cache.HealthCheck(ctx); err != nil {
			c.Logger.Error("cache health check failed", slog.String("component", "cache"), slog.Any("error", err))
			return fmt.Errorf("cache unhealthy: %w", err)
		}
	} else {
		c.Logger.Info("cache disabled or not initialized", slog.String("component", "cache"))
	}
	return nil
}
