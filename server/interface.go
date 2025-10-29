package server

import (
	"context"
)

// Server 服务器接口
type Server interface {
	// 启动服务器
	Start() error
	// 关闭服务器
	Shutdown(ctx context.Context) error
	// 静态文件服务
	Static(relativePath, root string)
}
