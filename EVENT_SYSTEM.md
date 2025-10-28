# 事件处理系统文档

## 概述

本事件处理系统是一个灵活、类型安全的事件分发架构，支持：
- 多处理器注册：同一事件类型可注册多个处理器
- 优先级控制：按优先级顺序执行处理器
- 中间件支持：日志、恢复、超时、限流等
- 同步/异步处理：可选的异步事件处理
- 类型安全：使用Go泛型确保类型安全

## 架构组件

### 1. EventHandler[T] 接口

事件处理器接口，所有处理器都需要实现：

```go
type EventHandler[T any] interface {
    Handle(ctx *Context[T]) error  // 处理事件
    Priority() int                 // 优先级（数值越小优先级越高）
    Name() string                  // 处理器名称
}
```

### 2. Context[T] 上下文

事件处理上下文，提供事件数据和元数据传递：

```go
type Context[T any] struct {
    Context  context.Context
    Event    T                          // 事件数据
    Metadata map[string]interface{}     // 元数据
}

// 方法
ctx.Set(key, value)      // 设置元数据
ctx.Get(key)             // 获取元数据
ctx.Abort()              // 中止后续处理器
ctx.IsAborted()          // 检查是否已中止
```

### 3. Dispatcher 分发器

事件分发器，负责注册和分发事件：

```go
dispatcher := event.NewDispatcher()

// 注册处理器
event.Register(dispatcher, handler)

// 注册函数式处理器
event.RegisterFunc(dispatcher, "HandlerName", priority, func(ctx *event.Context[*types.MessageEvent]) error {
    // 处理逻辑
    return nil
})

// 添加中间件
dispatcher.Use(event.LoggingMiddleware())

// 分发事件
dispatcher.Dispatch(context.Background(), event)

// 设置异步处理
dispatcher.SetAsync(true)
```

## 使用示例

### 1. 创建自定义处理器

```go
package handler

import (
    "log"
    "onebot-go2/pkg/event"
    types "onebot-go2/pkg/const"
)

type MyMessageHandler struct{}

func NewMyMessageHandler() *MyMessageHandler {
    return &MyMessageHandler{}
}

func (h *MyMessageHandler) Handle(ctx *event.Context[*types.MessageEvent]) error {
    msg := ctx.Event
    log.Printf("Received message: %s", msg.RawMessage)
    
    // 设置元数据供其他处理器使用
    ctx.Set("processed", true)
    
    return nil
}

func (h *MyMessageHandler) Priority() int {
    return 10  // 优先级：10
}

func (h *MyMessageHandler) Name() string {
    return "MyMessageHandler"
}
```

### 2. 注册处理器

```go
func main() {
    wsServer := server.NewWSServer("token")
    dispatcher := wsServer.GetDispatcher()
    
    // 添加中间件
    dispatcher.Use(event.RecoveryMiddleware())
    dispatcher.Use(event.LoggingMiddleware())
    
    // 注册多个处理器（按优先级执行）
    event.Register(dispatcher, NewMyMessageHandler())        // 优先级 10
    event.Register(dispatcher, handler.NewMessageLogHandler()) // 优先级 10
    event.Register(dispatcher, handler.NewMessageEchoHandler()) // 优先级 50
    
    // 使用函数式处理器
    event.RegisterFunc(dispatcher, "QuickHandler", 100, 
        func(ctx *event.Context[*types.MessageEvent]) error {
            log.Printf("Quick processing: %s", ctx.Event.RawMessage)
            return nil
        })
}
```

### 3. 使用中间件

系统提供多种内置中间件：

```go
// 日志中间件
dispatcher.Use(event.LoggingMiddleware())

// 异常恢复中间件
dispatcher.Use(event.RecoveryMiddleware())

// 超时中间件
dispatcher.Use(event.TimeoutMiddleware(5 * time.Second))

// 限流中间件
dispatcher.Use(event.RateLimitMiddleware(10)) // 每秒最多10个

// 过滤中间件
dispatcher.Use(event.FilterMiddleware(func(ctx *event.Context[interface{}]) bool {
    // 返回true继续处理，false跳过
    return true
}))

// 指标收集中间件
dispatcher.Use(event.MetricsMiddleware(func(duration time.Duration, err error) {
    log.Printf("Handler took %v, error: %v", duration, err)
}))
```

### 4. 自定义中间件

```go
func MyCustomMiddleware() event.Middleware {
    return func(next event.HandlerFunc[interface{}]) event.HandlerFunc[interface{}] {
        return func(ctx *event.Context[interface{}]) error {
            // 前置处理
            log.Println("Before handler")
            
            // 调用下一个处理器
            err := next(ctx)
            
            // 后置处理
            log.Println("After handler")
            
            return err
        }
    }
}

dispatcher.Use(MyCustomMiddleware())
```

### 5. 处理器间数据传递

```go
// 处理器1：设置数据
func (h *Handler1) Handle(ctx *event.Context[*types.MessageEvent]) error {
    ctx.Set("user_role", "admin")
    ctx.Set("processed_at", time.Now())
    return nil
}

// 处理器2：使用数据
func (h *Handler2) Handle(ctx *event.Context[*types.MessageEvent]) error {
    role, exists := ctx.Get("user_role")
    if exists && role == "admin" {
        // 特殊处理逻辑
    }
    return nil
}
```

### 6. 中止后续处理器

```go
func (h *FilterHandler) Handle(ctx *event.Context[*types.MessageEvent]) error {
    if containsBannedWord(ctx.Event.RawMessage) {
        ctx.Abort()  // 中止后续处理器执行
        return nil
    }
    return nil
}
```

## 最佳实践

1. **优先级分配**
   - 0-20: 预处理（日志、验证、过滤）
   - 21-50: 核心业务逻辑
   - 51-100: 后处理（通知、清理）

2. **错误处理**
   - 处理器返回error不会中止后续处理器
   - 使用`ctx.Abort()`显式中止
   - 添加RecoveryMiddleware防止panic

3. **性能优化**
   - 使用异步处理提高吞吐量：`dispatcher.SetAsync(true)`
   - 注意异步模式下错误处理
   - 合理设置优先级避免不必要的处理

4. **线程安全**
   - 处理器可能并发执行，注意共享状态
   - 使用Context传递请求级别的数据
   - 避免在处理器中修改全局变量

## 完整示例

参考 `cmd/app/main.go` 和 `internal/handler/message.go` 查看完整使用示例。
