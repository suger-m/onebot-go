# Context 初始化流程文档

本文档详细说明了 OneBot Go2 中 `event.Context` 的初始化和传递流程。

## Context 结构体定义

**位置**: `pkg/event/handler.go:41-52`

```go
type Context[T any] struct {
    context.Context              // 标准库的 Context
    Event    T                   // 原始事件数据（类型安全）
    Metadata map[string]interface{} // 元数据，用于处理器间传递数据
    aborted  bool                 // 是否中止后续处理器
    server   interface{}          // OneBot 服务器实例（用于调用 API）
}
```

## 完整调用链路

```
┌─────────────────────────────────────────────────────────────┐
│ 1. WebSocket 收到消息                                        │
│    bot_server.go:100 - conn.ReadMessage()                   │
└────────────────────┬────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────┐
│ 2. 解析事件                                                  │
│    bot_server.go:127 - ParseEvent(message)                  │
│    返回: *types.MessageEvent / *types.NoticeEvent 等        │
└────────────────────┬────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────┐
│ 3. 分发事件                                                  │
│    bot_server.go:134                                         │
│    s.dispatcher.Dispatch(context.Background(), evt, s)       │
│    传入: 标准 Context + 事件 + WSServer 实例                 │
└────────────────────┬────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────┐
│ 4. Dispatcher.Dispatch                                       │
│    dispatcher.go:94-113                                      │
│    - 查找事件类型对应的处理器                                │
│    - 调用 dispatchToHandlers(ctx, event, wrappers, server)  │
└────────────────────┬────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────┐
│ 5. 遍历处理器                                                │
│    dispatcher.go:116-125                                     │
│    for _, wrapper := range wrappers {                        │
│        invokeHandler(ctx, event, wrapper, server)            │
│    }                                                         │
└────────────────────┬────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────┐
│ 6. 创建 Context 实例 ⭐ 核心初始化点                         │
│    dispatcher.go:143-149                                     │
│                                                              │
│    eventCtx := &Context[interface{}]{                        │
│        Context:  ctx,           // context.Background()      │
│        Event:    event,         // 解析的事件对象            │
│        Metadata: make(map[string]interface{}), // 空 map     │
│        aborted:  false,         // 未中止                    │
│        server:   server,        // WSServer 实例             │
│    }                                                         │
└────────────────────┬────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────┐
│ 7. 应用中间件（洋葱模型）                                    │
│    dispatcher.go:151-161                                     │
│    handler := applyMiddlewares(finalHandler, middlewares)    │
│    - LoggingMiddleware: 记录开始时间                         │
│    - RecoveryMiddleware: 捕获 panic                          │
│    - 其他中间件...                                           │
└────────────────────┬────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────┐
│ 8. 调用处理器                                                │
│    dispatcher.go:152-157                                     │
│    handleMethod.Call([]reflect.Value{reflect.ValueOf(ctx)}) │
│    - 使用反射调用 handler.Handle(ctx)                        │
└────────────────────┬────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────┐
│ 9. 处理器接收 Context                                        │
│    handler/xxx.go                                            │
│                                                              │
│    func (h *Handler) Handle(ctx *event.Context[T]) error {  │
│        // 可以使用 ctx 的所有便捷方法                        │
│        ctx.ReplyText("你好")                                 │
│        ctx.GetGroupInfo(groupID)                             │
│        ctx.BanGroupMember(groupID, userID, 600)              │
│        return nil                                            │
│    }                                                         │
└─────────────────────────────────────────────────────────────┘
```

## Context 各字段的初始化来源

| 字段 | 初始化值 | 来源 | 说明 |
|------|---------|------|------|
| `Context` | `context.Background()` | bot_server.go:134 | 标准库的 Context，用于超时控制等 |
| `Event` | 事件对象 | bot_server.go:127 ParseEvent() | 从 WebSocket 消息解析的具体事件类型 |
| `Metadata` | `make(map[string]interface{})` | dispatcher.go:146 | 新创建的空 map，用于处理器间传递数据 |
| `aborted` | `false` | dispatcher.go:147 | 初始化为 false，可通过 ctx.Abort() 修改 |
| `server` | WSServer 实例 | bot_server.go:134 传入 | WebSocket 服务器实例，提供 API 调用能力 |

## 关键代码片段

### 1. WebSocket 处理入口

**文件**: `internal/server/bot_server.go:127-137`

```go
// 解析为事件
evt, err := ParseEvent(message)
if err != nil {
    log.Printf("Parse event error: %v", err)
    continue
}

// 分发事件到注册的处理器
// 关键：这里传入了 s (WSServer 实例)
if err := s.dispatcher.Dispatch(context.Background(), evt, s); err != nil {
    log.Printf("Error dispatching event: %v", err)
}
```

### 2. Dispatcher 分发方法

**文件**: `pkg/event/dispatcher.go:94-113`

```go
// Dispatch 分发事件
// server 参数就是从这里传入的
func (d *Dispatcher) Dispatch(ctx context.Context, event interface{}, server interface{}) error {
    eventType := reflect.TypeOf(event)

    d.mu.RLock()
    wrappers, exists := d.handlers[eventType]
    d.mu.RUnlock()

    if !exists || len(wrappers) == 0 {
        log.Printf("[EventDispatcher] No handlers registered for event type %s", eventType)
        return nil
    }

    log.Printf("[EventDispatcher] Dispatching event type %s to %d handler(s)", eventType, len(wrappers))

    if d.async {
        go d.dispatchToHandlers(ctx, event, eventType, wrappers, server)
        return nil
    }

    return d.dispatchToHandlers(ctx, event, eventType, wrappers, server)
}
```

### 3. Context 创建核心代码

**文件**: `pkg/event/dispatcher.go:143-161`

```go
// 创建事件上下文，传入 server
eventCtx := &Context[interface{}]{
    Context:  ctx,           // 标准 context.Context
    Event:    event,         // 原始事件数据
    Metadata: make(map[string]interface{}), // 空的元数据 map
    aborted:  false,         // 未中止
    server:   server,        // 添加 server 引用 ⭐ 关键
}

// 应用中间件
finalHandler := func(c *Context[interface{}]) error {
    results := handleMethod.Call([]reflect.Value{reflect.ValueOf(eventCtx)})
    if len(results) > 0 && !results[0].IsNil() {
        return results[0].Interface().(error)
    }
    return nil
}

// 中间件洋葱模型
handler := applyMiddlewares(finalHandler, d.middlewares)
return handler(eventCtx)
```

## 实际流程示例

假设收到一条消息 "你好"，完整流程如下：

### Step 1: 接收原始消息

```json
{
  "time": 1234567890,
  "self_id": 123456,
  "post_type": "message",
  "message_type": "group",
  "group_id": 789012,
  "user_id": 345678,
  "message": [{"type": "text", "data": {"text": "你好"}}],
  "raw_message": "你好"
}
```

### Step 2: 解析为事件对象

```go
&types.MessageEvent{
    Event: types.Event{
        Time:     1234567890,
        SelfID:   123456,
        PostType: "message",
    },
    MessageType: "group",
    GroupID:     789012,
    UserID:      345678,
    RawMessage:  "你好",
    Message:     [...],
}
```

### Step 3: 创建 Context

```go
&event.Context[interface{}]{
    Context: context.Background(),
    Event: &types.MessageEvent{
        GroupID:    789012,
        UserID:     345678,
        RawMessage: "你好",
        // ... 其他字段
    },
    Metadata: map[string]interface{}{},
    aborted:  false,
    server:   wsServer, // 包含所有 API 方法
}
```

### Step 4: 处理器使用

```go
func (h *MyHandler) Handle(ctx *event.Context[*types.MessageEvent]) error {
    // 获取消息内容
    rawMsg, _ := ctx.GetRawMessage()  // "你好"

    // 判断消息类型
    if ctx.IsGroupMessage() {
        groupID, _ := ctx.GetGroupID()  // 789012

        // 回复消息（内部调用 ctx.server.SendGroupMsg）
        _, err := ctx.ReplyText("你好！我收到了你的消息")
        return err
    }

    return nil
}
```

## Context 便捷方法的实现原理

Context 的便捷方法（如 `ReplyText`）都是通过访问 `ctx.server` 来实现的：

```go
// ReplyText 的实现
func (c *Context[T]) ReplyText(text string) (*types.SendMessageResponse, error) {
    message := types.MessageArray{
        {
            Type: "text",
            Data: map[string]interface{}{"text": text},
        },
    }
    return c.Reply(message)  // 调用 Reply
}

// Reply 的实现
func (c *Context[T]) Reply(message types.MessageArray) (*types.SendMessageResponse, error) {
    server := c.GetServer()  // 获取 server
    if server == nil {
        return nil, fmt.Errorf("server not available")
    }

    // 尝试从事件中提取消息信息
    if msgEvent, ok := any(c.Event).(*types.MessageEvent); ok {
        if msgEvent.MessageType == types.MessageTypePrivate {
            // 内部调用 server.SendPrivateMsg
            return server.SendPrivateMsg(msgEvent.UserID, message)
        } else if msgEvent.MessageType == types.MessageTypeGroup {
            // 内部调用 server.SendGroupMsg
            return server.SendGroupMsg(msgEvent.GroupID, message)
        }
    }

    return nil, fmt.Errorf("cannot determine message type from event")
}
```

## 设计优势

### 1. 延迟创建
Context 在事件分发时才创建，而不是提前创建，节省内存和资源。

### 2. 类型安全
使用泛型 `Context[T]` 确保事件类型的安全性，编译期检查错误。

### 3. 依赖注入
通过参数传入 server，避免循环依赖和全局变量。

### 4. 中间件友好
创建后立即应用中间件，形成标准的洋葱模型：

```
Request → MW1前 → MW2前 → Handler → MW2后 → MW1后 → Response
```

### 5. 处理器间通信
通过 `Metadata` 字段，处理器可以在执行链中传递数据：

```go
// 处理器 1
ctx.Set("user_level", "admin")

// 处理器 2
level, _ := ctx.Get("user_level")
```

### 6. 流程控制
通过 `Abort()` 方法可以中止后续处理器的执行：

```go
// 权限检查处理器
if !hasPermission {
    ctx.Abort()
    return ctx.ReplyText("权限不足")
}
// 后续处理器不会执行
```

## 总结

Context 的初始化是一个精心设计的流程，它：

1. **在正确的时机创建**（事件分发时）
2. **包含所有必要的信息**（事件数据、server 实例、元数据）
3. **提供便捷的 API**（类似 Gin 的方法）
4. **支持扩展**（中间件、元数据传递）
5. **类型安全**（Go 泛型）

这使得编写事件处理器变得简单直观，同时保持了代码的可维护性和扩展性。
