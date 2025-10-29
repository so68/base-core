# base-core

base 框架核心部件。

## 特性
- 统一初始化：Config / Logger / DB / Cache
- 可选项构造：可禁用 DB/Cache，或注入自定义 Logger/Config
- 统一生命周期：`Close(ctx)`、`Health(ctx)`

## 快速开始

1) 安装

```bash
go get github.com/so68/core
```

2) 配置
- 参考 `example/config.yaml` 或简化版 `example/config.simple.yaml`

3) 最小示例

```go
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
    app, err := core.NewCore("./example/config.yaml")
    if err != nil {
        panic(err)
    }
    defer func() {
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        _ = app.Close(ctx)
    }()

    app.Logger.Info("core started")
    _ = app.Health(context.Background())

    ch := make(chan os.Signal, 1)
    signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
    <-ch
}
```

## 进阶：可选项构造

```go
app, err := core.NewCoreWithOptions(
    core.WithConfigPath("./example/config.yaml"),
    // core.WithConfig(myCfg),
    // core.WithLogger(myLogger),
    // core.WithoutDB(),
    // core.WithoutCache(),
)
```

## 健康检查与优雅关闭
- 健康检查：`app.Health(ctx)` 聚合 DB/Cache
- 关闭：`app.Close(ctx)` 统一释放资源
