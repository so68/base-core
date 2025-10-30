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

	// 健康检查示例（非严格）
	_ = app.Health(context.Background())

	// 通过上下文驱动运行与优雅退出
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 捕获信号
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-ch
		cancel()
	}()

	app.Logger.Info("App Core starting")
	if err := app.Run(ctx); err != nil {
		// http.ErrServerClosed 已在内部视为正常，无需特殊处理
		app.Logger.Error("app stopped with error", "error", err)
	}
}
