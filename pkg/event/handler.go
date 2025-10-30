package event

import (
	"context"
	"fmt"
	"log"
	types "onebot-go2/pkg/const"
)

// ServerInterface 定义 Server 接口，用于避免循环依赖
type ServerInterface interface {
	SendPrivateMsg(userID int64, message types.MessageArray) (*types.SendMessageResponse, error)
	SendGroupMsg(groupID int64, message types.MessageArray) (*types.SendMessageResponse, error)
	SendMsg(params *types.SendMessageParams) (*types.SendMessageResponse, error)
	DeleteMsg(messageID int32) error
	GetMsg(messageID int32) (*types.GetMsgResponse, error)
	SetGroupKick(groupID, userID int64, rejectAddRequest bool) error
	SetGroupBan(groupID, userID int64, duration int64) error
	SetGroupWholeBan(groupID int64, enable bool) error
	SetGroupCard(groupID, userID int64, card string) error
	SetGroupName(groupID int64, groupName string) error
	GetGroupInfo(groupID int64, noCache bool) (*types.GetGroupInfoResponse, error)
	GetGroupMemberInfo(groupID, userID int64, noCache bool) (*types.GetGroupMemberInfoResponse, error)
	GetGroupMemberList(groupID int64) (types.GetGroupMemberListResponse, error)
	GetLoginInfo() (*types.GetLoginInfoResponse, error)
	GetFriendList() (types.GetFriendListResponse, error)
	GetGroupList() (types.GetGroupListResponse, error)
	IsConnected() bool
}

// EventHandler 事件处理器接口
type EventHandler[T any] interface {
	// Handle 处理事件
	Handle(ctx *Context[T]) error
	// Priority 返回处理器优先级，数值越小优先级越高
	Priority() int
	// Name 返回处理器名称，用于日志和调试
	Name() string
}

// Context 事件处理上下文（类似 Gin 的 Context）
type Context[T any] struct {
	context.Context
	// Event 原始事件数据
	Event T
	// Metadata 元数据，可以在处理器间传递数据
	Metadata map[string]interface{}
	// aborted 标记是否中止后续处理器
	aborted bool
	// server OneBot 服务器实例（用于调用 API）
	server interface{}
}

// NewContext 创建新的事件上下文
func NewContext[T any](ctx context.Context, event T) *Context[T] {
	return &Context[T]{
		Context:  ctx,
		Event:    event,
		Metadata: make(map[string]interface{}),
		aborted:  false,
	}
}

// ============ 元数据管理方法 ============

// Set 设置元数据
func (c *Context[T]) Set(key string, value interface{}) {
	c.Metadata[key] = value
}

// Get 获取元数据
func (c *Context[T]) Get(key string) (interface{}, bool) {
	val, ok := c.Metadata[key]
	return val, ok
}

// MustGet 获取元数据，不存在时panic
func (c *Context[T]) MustGet(key string) interface{} {
	val, ok := c.Get(key)
	if !ok {
		panic(fmt.Sprintf("key %s not found in context", key))
	}
	return val
}

// Abort 中止后续处理器执行
func (c *Context[T]) Abort() {
	c.aborted = true
}

// IsAborted 判断是否已中止
func (c *Context[T]) IsAborted() bool {
	return c.aborted
}

// ============ Server 访问方法 ============

// GetServer 获取 Server 实例
func (c *Context[T]) GetServer() ServerInterface {
	if c.server == nil {
		log.Printf("[Context] Warning: Server is not set")
		return nil
	}
	if server, ok := c.server.(ServerInterface); ok {
		return server
	}
	log.Printf("[Context] Warning: Server does not implement ServerInterface")
	return nil
}

// ============ 便捷消息发送方法（类似 Gin）============

// Reply 回复消息（根据事件类型自动判断是私聊还是群聊）
func (c *Context[T]) Reply(message types.MessageArray) (*types.SendMessageResponse, error) {
	server := c.GetServer()
	if server == nil {
		return nil, fmt.Errorf("server not available")
	}

	// 尝试从事件中提取消息信息
	if msgEvent, ok := any(c.Event).(*types.MessageEvent); ok {
		if msgEvent.MessageType == types.MessageTypePrivate {
			return server.SendPrivateMsg(msgEvent.UserID, message)
		} else if msgEvent.MessageType == types.MessageTypeGroup {
			return server.SendGroupMsg(msgEvent.GroupID, message)
		}
	}

	return nil, fmt.Errorf("cannot determine message type from event")
}

// ReplyText 回复纯文本消息
func (c *Context[T]) ReplyText(text string) (*types.SendMessageResponse, error) {
	message := types.MessageArray{
		{
			Type: "text",
			Data: map[string]interface{}{"text": text},
		},
	}
	return c.Reply(message)
}

// ReplyWithQuote 回复消息并引用原消息
func (c *Context[T]) ReplyWithQuote(message types.MessageArray) (*types.SendMessageResponse, error) {
	server := c.GetServer()
	if server == nil {
		return nil, fmt.Errorf("server not available")
	}

	// 尝试从事件中提取消息信息
	if msgEvent, ok := any(c.Event).(*types.MessageEvent); ok {
		// 在消息前添加 reply 消息段
		quotedMessage := append(
			types.MessageArray{{
				Type: "reply",
				Data: map[string]interface{}{"id": fmt.Sprintf("%d", msgEvent.MessageID)},
			}},
			message...,
		)

		if msgEvent.MessageType == types.MessageTypePrivate {
			return server.SendPrivateMsg(msgEvent.UserID, quotedMessage)
		} else if msgEvent.MessageType == types.MessageTypeGroup {
			return server.SendGroupMsg(msgEvent.GroupID, quotedMessage)
		}
	}

	return nil, fmt.Errorf("cannot determine message type from event")
}

// SendPrivateMsg 发送私聊消息
func (c *Context[T]) SendPrivateMsg(userID int64, message types.MessageArray) (*types.SendMessageResponse, error) {
	server := c.GetServer()
	if server == nil {
		return nil, fmt.Errorf("server not available")
	}
	return server.SendPrivateMsg(userID, message)
}

// SendGroupMsg 发送群消息
func (c *Context[T]) SendGroupMsg(groupID int64, message types.MessageArray) (*types.SendMessageResponse, error) {
	server := c.GetServer()
	if server == nil {
		return nil, fmt.Errorf("server not available")
	}
	return server.SendGroupMsg(groupID, message)
}

// SendMsg 发送消息（通用）
func (c *Context[T]) SendMsg(params *types.SendMessageParams) (*types.SendMessageResponse, error) {
	server := c.GetServer()
	if server == nil {
		return nil, fmt.Errorf("server not available")
	}
	return server.SendMsg(params)
}

// DeleteMsg 撤回消息
func (c *Context[T]) DeleteMsg(messageID int32) error {
	server := c.GetServer()
	if server == nil {
		return fmt.Errorf("server not available")
	}
	return server.DeleteMsg(messageID)
}

// GetMsg 获取消息
func (c *Context[T]) GetMsg(messageID int32) (*types.GetMsgResponse, error) {
	server := c.GetServer()
	if server == nil {
		return nil, fmt.Errorf("server not available")
	}
	return server.GetMsg(messageID)
}

// ============ 群管理便捷方法 ============

// KickGroupMember 踢出群成员
func (c *Context[T]) KickGroupMember(groupID, userID int64, rejectAddRequest bool) error {
	server := c.GetServer()
	if server == nil {
		return fmt.Errorf("server not available")
	}
	return server.SetGroupKick(groupID, userID, rejectAddRequest)
}

// BanGroupMember 禁言群成员
func (c *Context[T]) BanGroupMember(groupID, userID int64, duration int64) error {
	server := c.GetServer()
	if server == nil {
		return fmt.Errorf("server not available")
	}
	return server.SetGroupBan(groupID, userID, duration)
}

// UnbanGroupMember 解除禁言
func (c *Context[T]) UnbanGroupMember(groupID, userID int64) error {
	return c.BanGroupMember(groupID, userID, 0)
}

// BanAllGroupMembers 全员禁言
func (c *Context[T]) BanAllGroupMembers(groupID int64) error {
	server := c.GetServer()
	if server == nil {
		return fmt.Errorf("server not available")
	}
	return server.SetGroupWholeBan(groupID, true)
}

// UnbanAllGroupMembers 解除全员禁言
func (c *Context[T]) UnbanAllGroupMembers(groupID int64) error {
	server := c.GetServer()
	if server == nil {
		return fmt.Errorf("server not available")
	}
	return server.SetGroupWholeBan(groupID, false)
}

// SetGroupCard 设置群名片
func (c *Context[T]) SetGroupCard(groupID, userID int64, card string) error {
	server := c.GetServer()
	if server == nil {
		return fmt.Errorf("server not available")
	}
	return server.SetGroupCard(groupID, userID, card)
}

// SetGroupName 设置群名
func (c *Context[T]) SetGroupName(groupID int64, groupName string) error {
	server := c.GetServer()
	if server == nil {
		return fmt.Errorf("server not available")
	}
	return server.SetGroupName(groupID, groupName)
}

// ============ 信息获取便捷方法 ============

// GetGroupInfo 获取群信息
func (c *Context[T]) GetGroupInfo(groupID int64) (*types.GetGroupInfoResponse, error) {
	server := c.GetServer()
	if server == nil {
		return nil, fmt.Errorf("server not available")
	}
	return server.GetGroupInfo(groupID, false)
}

// GetGroupMemberInfo 获取群成员信息
func (c *Context[T]) GetGroupMemberInfo(groupID, userID int64) (*types.GetGroupMemberInfoResponse, error) {
	server := c.GetServer()
	if server == nil {
		return nil, fmt.Errorf("server not available")
	}
	return server.GetGroupMemberInfo(groupID, userID, false)
}

// GetGroupMemberList 获取群成员列表
func (c *Context[T]) GetGroupMemberList(groupID int64) (types.GetGroupMemberListResponse, error) {
	server := c.GetServer()
	if server == nil {
		return nil, fmt.Errorf("server not available")
	}
	return server.GetGroupMemberList(groupID)
}

// GetLoginInfo 获取登录号信息
func (c *Context[T]) GetLoginInfo() (*types.GetLoginInfoResponse, error) {
	server := c.GetServer()
	if server == nil {
		return nil, fmt.Errorf("server not available")
	}
	return server.GetLoginInfo()
}

// GetFriendList 获取好友列表
func (c *Context[T]) GetFriendList() (types.GetFriendListResponse, error) {
	server := c.GetServer()
	if server == nil {
		return nil, fmt.Errorf("server not available")
	}
	return server.GetFriendList()
}

// GetGroupList 获取群列表
func (c *Context[T]) GetGroupList() (types.GetGroupListResponse, error) {
	server := c.GetServer()
	if server == nil {
		return nil, fmt.Errorf("server not available")
	}
	return server.GetGroupList()
}

// ============ 事件相关便捷方法 ============

// GetMessageEvent 获取消息事件（如果当前事件是消息事件）
func (c *Context[T]) GetMessageEvent() (*types.MessageEvent, bool) {
	if msgEvent, ok := any(c.Event).(*types.MessageEvent); ok {
		return msgEvent, true
	}
	return nil, false
}

// GetGroupID 从事件中获取群ID（如果有）
func (c *Context[T]) GetGroupID() (int64, bool) {
	if msgEvent, ok := c.GetMessageEvent(); ok {
		if msgEvent.MessageType == types.MessageTypeGroup {
			return msgEvent.GroupID, true
		}
	}
	return 0, false
}

// GetUserID 从事件中获取用户ID（如果有）
func (c *Context[T]) GetUserID() (int64, bool) {
	if msgEvent, ok := c.GetMessageEvent(); ok {
		return msgEvent.UserID, true
	}
	return 0, false
}

// GetMessageID 从事件中获取消息ID（如果有）
func (c *Context[T]) GetMessageID() (int32, bool) {
	if msgEvent, ok := c.GetMessageEvent(); ok {
		return msgEvent.MessageID, true
	}
	return 0, false
}

// GetRawMessage 从事件中获取原始消息文本（如果有）
func (c *Context[T]) GetRawMessage() (string, bool) {
	if msgEvent, ok := c.GetMessageEvent(); ok {
		return msgEvent.RawMessage, true
	}
	return "", false
}

// IsGroupMessage 判断是否为群消息
func (c *Context[T]) IsGroupMessage() bool {
	if msgEvent, ok := c.GetMessageEvent(); ok {
		return msgEvent.MessageType == types.MessageTypeGroup
	}
	return false
}

// IsPrivateMessage 判断是否为私聊消息
func (c *Context[T]) IsPrivateMessage() bool {
	if msgEvent, ok := c.GetMessageEvent(); ok {
		return msgEvent.MessageType == types.MessageTypePrivate
	}
	return false
}

// HandlerFunc 处理器函数类型
type HandlerFunc[T any] func(ctx *Context[T]) error

// SimpleHandler 简单处理器实现
type SimpleHandler[T any] struct {
	name     string
	priority int
	handler  HandlerFunc[T]
}

// NewSimpleHandler 创建简单处理器
func NewSimpleHandler[T any](name string, priority int, handler HandlerFunc[T]) EventHandler[T] {
	return &SimpleHandler[T]{
		name:     name,
		priority: priority,
		handler:  handler,
	}
}

func (h *SimpleHandler[T]) Handle(ctx *Context[T]) error {
	return h.handler(ctx)
}

func (h *SimpleHandler[T]) Priority() int {
	return h.priority
}

func (h *SimpleHandler[T]) Name() string {
	return h.name
}
