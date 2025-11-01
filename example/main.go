package main

import (
	"context"

	"github.com/so68/core"
	"github.com/so68/core/example/internal/admin"
)

func main() {
	app, err := core.NewApplication("./config.yaml")
	if err != nil {
		panic(err)
	}

	// 初始化 admin 服务器
	if err := admin.NewAdminServer(app); err != nil {
		panic(err)
	}

	// 运行应用
	app.Run(context.Background())
}
