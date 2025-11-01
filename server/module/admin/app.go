package admin

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/so68/core"
	"github.com/so68/core/server/middleware"
	"github.com/so68/core/server/module/admin/service"
	"github.com/so68/core/server/utils"
)

// AdminApp 管理员应用
type AdminApp struct {
	relativePath  string                // 相对路径
	app           *core.Application     // 应用
	jwt           *utils.JWT            // JWT实例
	casbinService service.CasbinService // 权限服务
	router        *gin.RouterGroup      // 普通路由
	authRouter    *gin.RouterGroup      // 认证路由
}

// NewAdminApp 创建一个管理员应用
func NewAdminApp(app *core.Application, relativePath string) *AdminApp {
	router := app.Server.NewGroup(relativePath)

	// 使用JWT中间件验证Token - 登陆之后的路由
	jwt := utils.NewJWT(app.Config.JWT.SecretKey, time.Duration(app.Config.JWT.ExpiresIn)*time.Second)
	// 如果启用单点登录，则使用缓存验证Token
	if app.Config.JWT.EnableSingle {
		jwt.WithCache(app.Cache)
	}
	// 权限服务
	casbinService := service.NewCasbinService(app.DB.DB(), app.Cache, app.Logger)

	// 创建管理员应用
	adminApp := &AdminApp{app: app, router: router, relativePath: relativePath, jwt: jwt, casbinService: casbinService}
	adminApp.initAuthRouter().initHandler().initMigrate()
	return adminApp
}

// authRouter 使用JWT中间件验证Token - 登陆之后的路由
func (c *AdminApp) initAuthRouter() *AdminApp {
	authRouter := c.app.Server.Middleware(c.router.Group(""), middleware.NewJWTMiddleware(c.jwt))
	c.authRouter = c.app.Server.Middleware(authRouter, middleware.NewCasbinMiddleware(c.casbinService))
	return c
}

// initHandler 初始化路由处理
func (c *AdminApp) initHandler() *AdminApp {
	InitRouter(c)
	return c
}

// Handler 无验证中间件处理路由
func (c *AdminApp) Handler(name string, method string, path string, handler gin.HandlerFunc) {
	switch method {
	case "GET":
		c.router.GET(path, handler)
	case "POST":
		c.router.POST(path, handler)
	case "PUT":
		c.router.PUT(path, handler)
	case "DELETE":
		c.router.DELETE(path, handler)
	case "PATCH":
		c.router.PATCH(path, handler)
	default:
		c.app.Logger.Error("不支持的方法: " + method)
	}
}

// AuthHandler 认证处理器
func (c *AdminApp) AuthHandler(name string, method string, path string, handler gin.HandlerFunc) {
	switch method {
	case "GET":
		c.authRouter.GET(path, handler)
	case "POST":
		c.authRouter.POST(path, handler)
	case "PUT":
		c.authRouter.PUT(path, handler)
	case "DELETE":
		c.authRouter.DELETE(path, handler)
	case "PATCH":
		c.authRouter.PATCH(path, handler)
	default:
		c.app.Logger.Error("不支持的方法: " + method)
	}

	// 添加权限策略
	c.casbinService.AddPolicy(name, c.relativePath+path, method)
	// 添加角色继承
	c.casbinService.AddRoleInheritance(service.RoleSuperAdmin, name)
}

// initMigrate 初始化迁移
func (c *AdminApp) initMigrate() *AdminApp {
	err := InitMigrate(c.app.DB.DB(), c.app.Logger)
	if err != nil {
		c.app.Logger.Error("初始化迁移失败", "error", err)
	}
	return c
}
