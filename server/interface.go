package server

import (
	"context"

	"github.com/gin-gonic/gin"
)

type RouterMethod string

const (
	RouterMethodGet    RouterMethod = "GET"    // GET方法
	RouterMethodPost   RouterMethod = "POST"   // POST方法
	RouterMethodPut    RouterMethod = "PUT"    // PUT方法
	RouterMethodDelete RouterMethod = "DELETE" // DELETE方法
	RouterMethodPatch  RouterMethod = "PATCH"  // PATCH方法
)

// RouterHandler 路由处理器
type RouterHandler struct {
	Path    string          // 路径
	Method  RouterMethod    // 方法
	Handler gin.HandlerFunc // 处理器
}

// Server 服务器接口
type Server interface {
	// 启动服务器
	Start() error
	// 非阻塞启动，返回错误通道用于监听启动/运行期错误
	StartAsync() <-chan error
	// 关闭服务器
	Shutdown(ctx context.Context) error

	// 创建路由组
	NewGroup(relativePath string) *gin.RouterGroup
	// 添加中间件
	Middleware(group *gin.RouterGroup, middlewares ...gin.HandlerFunc) *gin.RouterGroup
	// 注册路由
	Register(group *gin.RouterGroup, handlers ...RouterHandler)
}
