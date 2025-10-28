# 事件处理器使用快速示例

## 快速开始

### 1. 创建自定义处理器

```go
package myhandler

import (
    "log"
    "onebot-go2/pkg/event"
    types "onebot-go2/pkg/const"
)

type WelcomeHandler struct{}

func NewWelcomeHandler() *WelcomeHandler {
    return &WelcomeHandler{}
}

func (h *WelcomeHandler) Handle(ctx *event.Context[*types.MessageEvent]) error {
    msg := ctx.Event
    
    // 检查是否是新用户消息
    if msg.RawMessage == "hello" {
        log.Printf("欢迎新用户 %d!", msg.UserID)
        ctx.Set("should_send_welcome", true)
    }
    
    return nil
}

func (h *WelcomeHandler) Priority() int {
    return 50  // 中等优先级
}

func (h *WelcomeHandler) Name() string {
    return "WelcomeHandler"
}
```

### 2. 在main.go中注册

```go
package main

import (
    "log"
    "onebot-go2/internal/server"
    "onebot-go2/pkg/event"
    "myhandler"  // 你的处理器包
    
    "github.com/gin-gonic/gin"
)

func main() {
    // 创建服务器和分发器
    wsServer := server.NewWSServer("your-token")
    dispatcher := wsServer.GetDispatcher()
    
    // 添加中间件
    dispatcher.Use(event.RecoveryMiddleware())
    dispatcher.Use(event.LoggingMiddleware())
    
    // 注册你的处理器
    event.Register(dispatcher, myhandler.NewWelcomeHandler())
    
    // 启动服务器
    r := gin.Default()
    r.GET("/ws", wsServer.HandlerWebsocket)
    
    log.Printf("Server starting on :8080")
    r.Run(":8080")
}
```

### 3. 使用函数式处理器（更简单）

不需要创建struct，直接使用函数：

```go
// 在main.go中直接注册
event.RegisterFunc(dispatcher, "SimpleLogger", 10, 
    func(ctx *event.Context[*types.MessageEvent]) error {
        log.Printf("收到消息: %s", ctx.Event.RawMessage)
        return nil
    })

event.RegisterFunc(dispatcher, "CommandHandler", 50, 
    func(ctx *event.Context[*types.MessageEvent]) error {
        msg := ctx.Event.RawMessage
        
        if msg == "/help" {
            log.Printf("用户请求帮助")
            // 发送帮助信息
        } else if msg == "/status" {
            log.Printf("用户查询状态")
            // 返回状态信息
        }
        
        return nil
    })
```

## 常见场景

### 多个处理器协作

```go
// 处理器1：验证用户权限（优先级10）
event.RegisterFunc(dispatcher, "AuthCheck", 10, 
    func(ctx *event.Context[*types.MessageEvent]) error {
        // 检查权限
        if !hasPermission(ctx.Event.UserID) {
            ctx.Abort()  // 阻止后续处理
            return nil
        }
        ctx.Set("authorized", true)
        return nil
    })

// 处理器2：执行命令（优先级50）
event.RegisterFunc(dispatcher, "CommandExecutor", 50, 
    func(ctx *event.Context[*types.MessageEvent]) error {
        // 检查是否已授权
        if auth, _ := ctx.Get("authorized"); !auth.(bool) {
            return nil
        }
        
        // 执行命令
        executeCommand(ctx.Event.RawMessage)
        return nil
    })
```

### 异步处理（提高性能）

```go
// 对于耗时操作，可以启用异步处理
dispatcher.SetAsync(true)

event.RegisterFunc(dispatcher, "SlowProcessor", 100, 
    func(ctx *event.Context[*types.MessageEvent]) error {
        // 这将在后台goroutine中执行
        time.Sleep(5 * time.Second)
        log.Printf("处理完成")
        return nil
    })
```

## 详细文档

查看 [EVENT_SYSTEM.md](./EVENT_SYSTEM.md) 获取完整的架构说明和高级用法。
