package event

import (
	"context"
	"fmt"
)

// EventHandler 事件处理器接口
type EventHandler[T any] interface {
	// Handle 处理事件
	Handle(ctx *Context[T]) error
	// Priority 返回处理器优先级，数值越小优先级越高
	Priority() int
	// Name 返回处理器名称，用于日志和调试
	Name() string
}

// Context 事件处理上下文
type Context[T any] struct {
	context.Context
	// Event 原始事件数据
	Event T
	// Metadata 元数据，可以在处理器间传递数据
	Metadata map[string]interface{}
	// aborted 标记是否中止后续处理器
	aborted bool
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
