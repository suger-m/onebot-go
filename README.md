# OneBot Go2 - 拉格朗日 OneBot Bot 端实现

一个功能完整、易用性强的 OneBot 11 标准 Bot 端实现，基于 Go 语言开发。

## 特性

### 核心功能

- **完整的 API 支持** - 实现了 OneBot 11 标准的所有主要 API
- **类 Gin Context** - 提供类似 Gin 框架的便捷 API 调用方法
- **事件驱动架构** - 基于泛型的类型安全事件处理系统
- **中间件支持** - 可插拔的中间件系统（日志、恢复、超时、限流等）
- **命令路由** - 内置命令处理器，支持命令注册和路由
- **消息构造器** - Builder 模式构造复杂消息

### 架构优势

1. **类型安全** - 充分利用 Go 泛型，编译期类型检查
2. **易于扩展** - 清晰的处理器接口，方便添加新功能
3. **生产就绪** - 完善的错误处理、日志记录和优雅关闭
4. **文档完善** - 详细的代码注释和使用示例

## 快速开始

### 安装依赖

```bash
go mod download
```

### 配置环境变量

```bash
export ONEBOT_TOKEN="your-token-here"  # OneBot 鉴权 token
export PORT="8080"                      # HTTP 服务器端口（可选）
```

### 运行程序

```bash
go run cmd/main.go
```

服务器将在 `http://localhost:8080` 启动，WebSocket 端点为 `ws://localhost:8080/ws`。

### 连接 OneBot 客户端

配置拉格朗日或其他 OneBot 11 客户端，连接到 `ws://localhost:8080/ws`。

## 使用示例

### 1. 类 Gin 的便捷方法

```go
// 在事件处理器中使用 Context 的便捷方法
event.RegisterFunc(dispatcher, "MyHandler", 50, func(ctx *event.Context[*types.MessageEvent]) error {
    // 回复消息（自动判断私聊/群聊）
    ctx.ReplyText("你好！")

    // 引用回复
    ctx.ReplyWithQuote(message.Text("这是引用回复"))

    // 发送群消息
    ctx.SendGroupMsg(groupID, message.Text("群消息"))

    // 发送私聊消息
    ctx.SendPrivateMsg(userID, message.Text("私聊消息"))

    // 群管理操作
    ctx.BanGroupMember(groupID, userID, 600)  // 禁言 10 分钟
    ctx.UnbanGroupMember(groupID, userID)     // 解除禁言
    ctx.KickGroupMember(groupID, userID, false)  // 踢出群

    // 获取信息
    groupInfo, _ := ctx.GetGroupInfo(groupID)
    memberInfo, _ := ctx.GetGroupMemberInfo(groupID, userID)
    friendList, _ := ctx.GetFriendList()

    return nil
})
```

### 2. 消息构造器

```go
import "onebot-go2/pkg/message"

// 链式构造消息
msg := message.NewBuilder().
    At(userID).
    Text(" ").
    Text("你好！").
    Image("https://example.com/image.jpg").
    Build()

ctx.Reply(msg)

// 快捷方法
message.Text("纯文本")
message.AtText(userID, "提到你")
message.ImageText("https://example.com/img.jpg", "图片说明")
```

### 3. 命令处理器

```go
// 创建命令处理器
cmdHandler := handler.NewCommandHandler("/")

// 注册自定义命令
cmdHandler.Register("hello", func(ctx *event.Context[*types.MessageEvent]) error {
    _, err := ctx.ReplyText("Hello, World!")
    return err
})

// 注册到分发器
event.Register(dispatcher, cmdHandler)
```

### 4. 自定义事件处理器

```go
type MyHandler struct {
    priority int
}

func (h *MyHandler) Handle(ctx *event.Context[*types.MessageEvent]) error {
    // 获取消息内容
    rawMsg, _ := ctx.GetRawMessage()

    // 判断消息类型
    if ctx.IsGroupMessage() {
        groupID, _ := ctx.GetGroupID()
        log.Printf("群消息: %d - %s", groupID, rawMsg)
    }

    // 回复消息
    return ctx.ReplyText("收到消息: " + rawMsg)
}

func (h *MyHandler) Priority() int { return h.priority }
func (h *MyHandler) Name() string { return "MyHandler" }

// 注册处理器
event.Register(dispatcher, &MyHandler{priority: 50})
```

### 5. 使用中间件

```go
// 日志中间件
dispatcher.Use(event.LoggingMiddleware())

// 恢复中间件（防止 panic）
dispatcher.Use(event.RecoveryMiddleware())

// 超时中间件
dispatcher.Use(event.TimeoutMiddleware(5 * time.Second))

// 限流中间件
dispatcher.Use(event.RateLimitMiddleware(100, time.Minute))

// 自定义中间件
dispatcher.Use(func(next event.HandlerFunc[interface{}]) event.HandlerFunc[interface{}] {
    return func(ctx *event.Context[interface{}]) error {
        log.Println("Before handler")
        err := next(ctx)
        log.Println("After handler")
        return err
    }
})
```

## API 文档

### Context 便捷方法

#### 消息发送
- `Reply(message)` - 回复消息（自动判断类型）
- `ReplyText(text)` - 回复纯文本
- `ReplyWithQuote(message)` - 引用回复
- `SendPrivateMsg(userID, message)` - 发送私聊消息
- `SendGroupMsg(groupID, message)` - 发送群消息
- `DeleteMsg(messageID)` - 撤回消息

#### 群管理
- `KickGroupMember(groupID, userID, rejectAddRequest)` - 踢出群成员
- `BanGroupMember(groupID, userID, duration)` - 禁言群成员
- `UnbanGroupMember(groupID, userID)` - 解除禁言
- `BanAllGroupMembers(groupID)` - 全员禁言
- `UnbanAllGroupMembers(groupID)` - 解除全员禁言
- `SetGroupCard(groupID, userID, card)` - 设置群名片
- `SetGroupName(groupID, groupName)` - 设置群名

#### 信息获取
- `GetGroupInfo(groupID)` - 获取群信息
- `GetGroupMemberInfo(groupID, userID)` - 获取群成员信息
- `GetGroupMemberList(groupID)` - 获取群成员列表
- `GetLoginInfo()` - 获取登录号信息
- `GetFriendList()` - 获取好友列表
- `GetGroupList()` - 获取群列表

#### 事件辅助
- `GetMessageEvent()` - 获取消息事件
- `GetGroupID()` - 获取群 ID
- `GetUserID()` - 获取用户 ID
- `GetMessageID()` - 获取消息 ID
- `GetRawMessage()` - 获取原始消息文本
- `IsGroupMessage()` - 是否为群消息
- `IsPrivateMessage()` - 是否为私聊消息

### 完整 API 列表

服务器端（`WSServer`）支持的 API：

**消息 API**
- `SendPrivateMsg` / `SendGroupMsg` / `SendMsg`
- `DeleteMsg`, `GetMsg`, `GetForwardMsg`
- `SendLike`

**群管理 API**
- `SetGroupKick`, `SetGroupBan`, `SetGroupAnonymousBan`
- `SetGroupWholeBan`, `SetGroupAdmin`, `SetGroupAnonymous`
- `SetGroupCard`, `SetGroupName`, `SetGroupLeave`
- `SetGroupSpecialTitle`

**请求处理 API**
- `SetFriendAddRequest`, `SetGroupAddRequest`

**信息获取 API**
- `GetLoginInfo`, `GetStrangerInfo`, `GetFriendList`
- `GetGroupInfo`, `GetGroupList`
- `GetGroupMemberInfo`, `GetGroupMemberList`
- `GetGroupHonorInfo`
- `GetCookies`, `GetCsrfToken`, `GetCredentials`
- `GetStatus`, `GetVersionInfo`

## 项目结构

```
onebot-go2/
├── cmd/                    # 应用入口
│   └── main.go            # 主程序
├── internal/              # 内部实现
│   ├── handler/          # 事件处理器
│   │   ├── message.go    # 消息处理器
│   │   └── command.go    # 命令处理器
│   └── server/           # 服务器实现
│       └── bot_server.go # WebSocket 服务器 + API 实现
├── pkg/                   # 公共库
│   ├── const/            # 常量和类型
│   │   ├── types.go      # OneBot 类型定义
│   │   └── api.go        # API 常量
│   ├── event/            # 事件系统
│   │   ├── dispatcher.go  # 事件分发器
│   │   ├── handler.go     # Context 和处理器接口
│   │   └── middleware.go  # 中间件
│   └── message/          # 消息工具
│       └── builder.go     # 消息构造器
├── config.yaml            # 配置文件示例
└── go.mod                 # 依赖管理
```

## 内置命令

- `/help` - 显示帮助信息
- `/ping` - 测试响应
- `/echo <文本>` - 回复相同文本
- `/info` - 显示群/用户信息
- `/ban <@用户> <时长>` - 禁言用户（仅管理员）
- `/unban <@用户>` - 解除禁言（仅管理员）
- `/quote <文本>` - 引用回复
- `/image <URL>` - 发送图片

## 配置

参考 `config.yaml` 文件进行配置。主要配置项：

- **服务器配置** - 端口、监听地址
- **OneBot 配置** - Token、超时时间
- **中间件配置** - 启用/禁用各种中间件
- **命令配置** - 命令前缀、启用的命令列表
- **管理员配置** - 管理员 QQ 号列表

## 开发指南

### 添加新的处理器

1. 在 `internal/handler/` 下创建新文件
2. 实现 `EventHandler[T]` 接口
3. 在 `main.go` 中注册处理器

### 添加新的 API

1. 在 `pkg/const/types.go` 中定义参数和响应类型
2. 在 `pkg/const/api.go` 中添加 API 常量
3. 在 `internal/server/bot_server.go` 中实现 API 方法
4. 在 `pkg/event/handler.go` 的 `ServerInterface` 中添加方法签名
5. 在 `Context` 中添加便捷方法（可选）

## 依赖

- [gin-gonic/gin](https://github.com/gin-gonic/gin) - HTTP 框架
- [gorilla/websocket](https://github.com/gorilla/websocket) - WebSocket 支持
- [google/uuid](https://github.com/google/uuid) - UUID 生成

## 许可证

MIT License

## 贡献

欢迎提交 Issue 和 Pull Request！

## 相关链接

- [OneBot 11 标准](https://github.com/botuniverse/onebot-11)
- [拉格朗日](https://github.com/LagrangeDev/Lagrange.Core)
