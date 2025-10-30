package server

import (
	"context"
)

// Server 服务器接口
type Server interface {
	// 启动服务器
	Start() error
	// 非阻塞启动，返回错误通道用于监听启动/运行期错误
	StartAsync() <-chan error
	// 关闭服务器
	Shutdown(ctx context.Context) error
}
