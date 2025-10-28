package handler

import (
	"log"
	"onebot-go2/pkg/event"
	types "onebot-go2/pkg/const"
	"strings"
)

// MessageLogHandler 消息日志处理器
type MessageLogHandler struct{}

func NewMessageLogHandler() *MessageLogHandler {
	return &MessageLogHandler{}
}

func (h *MessageLogHandler) Handle(ctx *event.Context[*types.MessageEvent]) error {
	msg := ctx.Event
	log.Printf("[MessageLogHandler] Received message from user %d: %s", msg.UserID, msg.RawMessage)
	
	// 在上下文中设置一些元数据供其他处理器使用
	ctx.Set("logged", true)
	ctx.Set("message_length", len(msg.RawMessage))
	
	return nil
}

func (h *MessageLogHandler) Priority() int {
	return 10 // 优先级较高，先记录日志
}

func (h *MessageLogHandler) Name() string {
	return "MessageLogHandler"
}

// MessageEchoHandler 消息回显处理器
type MessageEchoHandler struct{}

func NewMessageEchoHandler() *MessageEchoHandler {
	return &MessageEchoHandler{}
}

func (h *MessageEchoHandler) Handle(ctx *event.Context[*types.MessageEvent]) error {
	msg := ctx.Event
	
	// 检查是否是回显命令
	if strings.HasPrefix(msg.RawMessage, "/echo ") {
		content := strings.TrimPrefix(msg.RawMessage, "/echo ")
		log.Printf("[MessageEchoHandler] Echo command detected: %s", content)
		
		ctx.Set("should_reply", true)
		ctx.Set("reply_content", content)
	}
	
	return nil
}

func (h *MessageEchoHandler) Priority() int {
	return 50 // 中等优先级
}

func (h *MessageEchoHandler) Name() string {
	return "MessageEchoHandler"
}

// MessageFilterHandler 消息过滤处理器
type MessageFilterHandler struct {
	bannedWords []string
}

func NewMessageFilterHandler(bannedWords []string) *MessageFilterHandler {
	return &MessageFilterHandler{
		bannedWords: bannedWords,
	}
}

func (h *MessageFilterHandler) Handle(ctx *event.Context[*types.MessageEvent]) error {
	msg := ctx.Event
	
	// 检查是否包含禁用词
	for _, word := range h.bannedWords {
		if strings.Contains(msg.RawMessage, word) {
			log.Printf("[MessageFilterHandler] Message contains banned word: %s", word)
			ctx.Set("filtered", true)
			ctx.Abort() // 中止后续处理器
			return nil
		}
	}
	
	return nil
}

func (h *MessageFilterHandler) Priority() int {
	return 20 // 较高优先级，尽早过滤
}

func (h *MessageFilterHandler) Name() string {
	return "MessageFilterHandler"
}

// NoticeHandler 通知事件处理器
type NoticeHandler struct{}

func NewNoticeHandler() *NoticeHandler {
	return &NoticeHandler{}
}

func (h *NoticeHandler) Handle(ctx *event.Context[*types.NoticeEvent]) error {
	notice := ctx.Event
	log.Printf("[NoticeHandler] Received notice type: %s from user %d", notice.NoticeType, notice.UserID)
	return nil
}

func (h *NoticeHandler) Priority() int {
	return 10
}

func (h *NoticeHandler) Name() string {
	return "NoticeHandler"
}

// RequestHandler 请求事件处理器
type RequestHandler struct{}

func NewRequestHandler() *RequestHandler {
	return &RequestHandler{}
}

func (h *RequestHandler) Handle(ctx *event.Context[*types.RequestEvent]) error {
	req := ctx.Event
	log.Printf("[RequestHandler] Received request type: %s from user %d", req.RequestType, req.UserID)
	return nil
}

func (h *RequestHandler) Priority() int {
	return 10
}

func (h *RequestHandler) Name() string {
	return "RequestHandler"
}
