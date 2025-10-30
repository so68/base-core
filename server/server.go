package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/so68/core/config"
)

// ginServer 实现 Server 接口的最小功能版
type ginServer struct {
	cfg        *config.AppConfig
	logger     *slog.Logger
	engine     *gin.Engine
	httpServer *http.Server
}

// NewServer 创建一个最小可用的 Gin 服务实例
func NewServer(logger *slog.Logger, cfg *config.AppConfig) Server {
	if !cfg.Debug {
		gin.SetMode(gin.ReleaseMode)
	}
	engine := gin.New()
	engine.Use(gin.Recovery())
	return &ginServer{logger: logger, engine: engine, cfg: cfg}
}

// buildHTTPServer 构建 http.Server（不启动）
func (s *ginServer) buildHTTPServer() *http.Server {
	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port),
		Handler: s.engine,

		// 最佳实践：设置读写超时，防止慢客户端攻击或连接长时间占用
		ReadTimeout:    s.cfg.ParseDuration(s.cfg.ReadTimeout),  // 读取整个请求（Header + Body）的超时时间
		WriteTimeout:   s.cfg.ParseDuration(s.cfg.WriteTimeout), // 写入响应的超时时间
		IdleTimeout:    s.cfg.ParseDuration(s.cfg.IdleTimeout),  // Keep-Alive 连接的空闲超时时间
		MaxHeaderBytes: s.cfg.MaxHeader,
	}
	s.httpServer = server
	return server
}

// Start 启动 HTTP 服务（阻塞直到关闭）
func (s *ginServer) Start() error {
	if s.httpServer == nil {
		s.buildHTTPServer()
	}
	s.logger.Info("Server started successfully", slog.String("addr", fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port)))
	if err := s.httpServer.ListenAndServe(); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
	return nil
}

// StartAsync 非阻塞启动，返回错误通道
func (s *ginServer) StartAsync() <-chan error {
	ch := make(chan error, 1)
	go func() {
		err := s.Start()
		ch <- err
		close(ch)
	}()
	return ch
}

// Shutdown 优雅关闭 HTTP 服务
func (s *ginServer) Shutdown(ctx context.Context) error {
	if s.httpServer == nil {
		return nil
	}
	return s.httpServer.Shutdown(ctx)
}

// Static 提供静态文件目录
func (s *ginServer) Static(relativePath, root string) {
	s.engine.Static(relativePath, root)
}
