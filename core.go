package core

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/so68/core/cache"
	"github.com/so68/core/config"
	"github.com/so68/core/database"
	"github.com/so68/core/server"
	"github.com/so68/utils/logger"
)

// Application 应用
type Application struct {
	Config *config.AppConfig // 配置
	Logger *slog.Logger      // 日志
	DB     database.Database // 数据库
	Cache  cache.Cache       // 缓存
	Server server.Server     // 服务器

	serverErrChan <-chan error // 服务器错误通道（StartAsync 使用）
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

// NewApplication 初始化应用
func NewApplication(configPath string, opts ...Option) (*Application, error) {
	return NewApplicationWithOptions(append(opts, WithConfigPath(configPath))...)
}

// NewApplicationWithOptions 基于可选项初始化应用
func NewApplicationWithOptions(opts ...Option) (*Application, error) {
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

	// 初始化服务器（仅构建，不启动）
	var s server.Server
	if o.enableServer {
		s = server.NewServer(slogLogger, cfg)
	}

	return &Application{
		Config: cfg,
		Logger: slogLogger,
		DB:     db,
		Cache:  c,
		Server: s,
	}, nil
}

// Close 统一释放资源
func (a *Application) Close(ctx context.Context) error {
	var firstErr error

	// 先停服务
	if a.Server != nil {
		if err := a.Server.Shutdown(ctx); err != nil {
			firstErr = fmt.Errorf("shutdown server: %w", err)
		}
	}

	if a.DB != nil {
		if firstErr == nil {
			if err := a.DB.Close(ctx); err != nil {
				firstErr = fmt.Errorf("close db: %w", err)
			}
		}
	}

	if a.Cache != nil {
		if firstErr == nil {
			if err := a.Cache.Close(); err != nil {
				firstErr = fmt.Errorf("close cache: %w", err)
			}
		}
	}

	return firstErr
}

// Start 启动核心组件（非阻塞启动 Server）
func (a *Application) Start(ctx context.Context) error {
	if a.Server != nil && a.serverErrChan == nil {
		a.serverErrChan = a.Server.StartAsync()
	}
	return nil
}

// Run 运行直到上下文取消或服务器报错（优雅关闭）
func (a *Application) Run(ctx context.Context) error {
	// 确保已启动
	if err := a.Start(ctx); err != nil {
		return err
	}

	if a.Server == nil {
		// 没有 Server，直接等待 ctx 结束
		<-ctx.Done()
		return a.Close(ctx)
	}

	// 等待退出或错误
	select {
	case <-ctx.Done():
		return a.Close(ctx)
	case err, ok := <-a.serverErrChan:
		if !ok {
			return nil
		}
		// http.ErrServerClosed 视为正常退出
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}

// Health 聚合健康检查
func (a *Application) Health(ctx context.Context) error {
	if a.DB != nil {
		if err := a.DB.HealthCheck(); err != nil {
			a.Logger.Error("db health check failed", slog.String("component", "db"), slog.Any("error", err))
			return fmt.Errorf("db unhealthy: %w", err)
		}
	} else {
		a.Logger.Info("db disabled or not initialized", slog.String("component", "db"))
	}
	if a.Cache != nil {
		if err := a.Cache.HealthCheck(ctx); err != nil {
			a.Logger.Error("cache health check failed", slog.String("component", "cache"), slog.Any("error", err))
			return fmt.Errorf("cache unhealthy: %w", err)
		}
	} else {
		a.Logger.Info("cache disabled or not initialized", slog.String("component", "cache"))
	}
	return nil
}
