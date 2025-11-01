package admin

import (
	"github.com/so68/core/server/module/admin/handler"
)

// InitRouter 初始化路由
func InitRouter(app *AdminApp) {
	indexHandler := handler.NewIndexHandler(app.app.Logger, app.app.DB.DB(), app.app.Cache, app.jwt, app.app.Config.Static, app.app.Config.MaxHeader)
	adminHandler := handler.NewAdminHandler(app.app.Logger, app.app.DB.DB(), app.app.Cache)

	// 通用路由
	app.Handler("管理员登录", "POST", "/login", indexHandler.Login)

	// 管理员路由
	app.AuthHandler("管理员列表", "GET", "/admin/index", adminHandler.Index)
	app.AuthHandler("创建管理员", "POST", "/admin/create", adminHandler.Create)
	app.AuthHandler("更新管理员", "PUT", "/admin/update", adminHandler.Update)
	app.AuthHandler("Token更新管理员", "PUT", "/admin/token/update", adminHandler.TokenUpdate)
	app.AuthHandler("Token更新管理员密码", "PUT", "/admin/token/password/update", adminHandler.TokenPasswordUpdate)
	app.AuthHandler("删除管理员", "DELETE", "/admin/delete", adminHandler.Delete)
}
