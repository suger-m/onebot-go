package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"

	types "onebot-go2/pkg/const"
	"onebot-go2/pkg/event"
	"onebot-go2/internal/handler"
	"onebot-go2/internal/server"
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.Println("=== OneBot Go2 Bot Starting ===")

	// 创建 WebSocket 服务器
	token := os.Getenv("ONEBOT_TOKEN")
	if token == "" {
		token = "your-token-here"
		log.Printf("Warning: Using default token. Set ONEBOT_TOKEN environment variable for production.")
	}

	wsServer := server.NewWSServer(token)
	dispatcher := wsServer.GetDispatcher()

	// ============ 配置中间件 ============
	log.Println("Configuring middlewares...")

	// 1. 恢复中间件 - 防止 panic 导致程序崩溃
	dispatcher.Use(event.RecoveryMiddleware())

	// 2. 日志中间件 - 记录每个事件的处理时间
	dispatcher.Use(event.LoggingMiddleware())

	// 3. 超时中间件 - 防止处理器执行过长时间
	// dispatcher.Use(event.TimeoutMiddleware(5 * time.Second))

	// 4. 限流中间件 - 防止过载（可选）
	// dispatcher.Use(event.RateLimitMiddleware(100, time.Minute))

	// ============ 注册事件处理器 ============
	log.Println("Registering event handlers...")

	// 1. 消息日志处理器 - 记录所有消息
	event.Register(dispatcher, handler.NewMessageLogHandler())

	// 2. 消息过滤器 - 过滤不当内容（可选）
	// event.Register(dispatcher, handler.NewMessageFilterHandler())

	// 3. 命令处理器 - 处理 /help, /ping, /echo 等命令
	event.Register(dispatcher, handler.NewDefaultCommandHandler())

	// 4. 简单回复处理器示例
	event.RegisterFunc(dispatcher, "SimpleReplyHandler", 100, func(ctx *event.Context[*types.MessageEvent]) error {
		// 示例：回复包含"你好"的消息
		rawMsg, _ := ctx.GetRawMessage()
		if rawMsg == "你好" {
			_, err := ctx.ReplyText("你好！我是 OneBot Go2 Bot")
			return err
		}
		return nil
	})

	// 5. 通知事件处理器
	event.Register(dispatcher, handler.NewNoticeHandler())

	// 6. 请求事件处理器
	event.Register(dispatcher, handler.NewRequestHandler())

	log.Println("All handlers registered successfully")

	// ============ 启动 Gin HTTP 服务器 ============
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// WebSocket 端点
	r.GET("/ws", wsServer.HandlerWebsocket)

	// 健康检查端点
	r.GET("/health", func(c *gin.Context) {
		status := "disconnected"
		if wsServer.IsConnected() {
			status = "connected"
		}
		c.JSON(200, gin.H{
			"status":  "ok",
			"onebot":  status,
			"version": "1.0.0",
		})
	})

	// 获取服务器端口
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// 启动服务器
	log.Printf("Starting HTTP server on port %s", port)
	log.Printf("WebSocket endpoint: ws://localhost:%s/ws", port)
	log.Printf("Health check: http://localhost:%s/health", port)
	log.Println("Waiting for OneBot client connection...")

	// 优雅关闭
	go func() {
		if err := r.Run(":" + port); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// 等待中断信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("=== OneBot Go2 Bot Shutting Down ===")
}
