package main

import (
	"context"

	"github.com/so68/core"
)

func main() {
	app, err := core.NewApplication("./config.yaml")
	if err != nil {
		panic(err)
	}

	// 运行应用
	app.Run(context.Background())
}
