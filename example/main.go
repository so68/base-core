package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/so68/core"
)

func main() {
	app, err := core.NewCore("./config.yaml")
	if err != nil {
		panic(err)
	}

	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = app.Close(ctx)
	}()

	app.Logger.Info("App Core started")

	// 健康检查示例
	_ = app.Health(context.Background())

	// 等待优雅退出
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
}
