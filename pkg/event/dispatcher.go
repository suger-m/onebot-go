package event

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"sort"
	"sync"
)

// Dispatcher 事件分发器
type Dispatcher struct {
	handlers  map[reflect.Type][]handlerWrapper
	mu        sync.RWMutex
	middlewares []Middleware
	async     bool
	errorHandler ErrorHandler
}

type handlerWrapper struct {
	handler  interface{}
	priority int
	name     string
}

// ErrorHandler 错误处理函数
type ErrorHandler func(err error, eventType reflect.Type, handlerName string)

// NewDispatcher 创建新的事件分发器
func NewDispatcher() *Dispatcher {
	return &Dispatcher{
		handlers: make(map[reflect.Type][]handlerWrapper),
		async:    false,
		errorHandler: func(err error, eventType reflect.Type, handlerName string) {
			log.Printf("[EventDispatcher] Error in handler %s for event %s: %v", handlerName, eventType, err)
		},
	}
}

// SetAsync 设置是否异步处理事件
func (d *Dispatcher) SetAsync(async bool) *Dispatcher {
	d.async = async
	return d
}

// SetErrorHandler 设置错误处理器
func (d *Dispatcher) SetErrorHandler(handler ErrorHandler) *Dispatcher {
	d.errorHandler = handler
	return d
}

// Use 添加全局中间件
func (d *Dispatcher) Use(middleware Middleware) *Dispatcher {
	d.middlewares = append(d.middlewares, middleware)
	return d
}

// register 内部注册方法（非泛型）
func (d *Dispatcher) register(eventType reflect.Type, handler interface{}, priority int, name string) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	
	d.handlers[eventType] = append(d.handlers[eventType], handlerWrapper{
		handler:  handler,
		priority: priority,
		name:     name,
	})
	
	// 按优先级排序
	sort.Slice(d.handlers[eventType], func(i, j int) bool {
		return d.handlers[eventType][i].priority < d.handlers[eventType][j].priority
	})
	
	log.Printf("[EventDispatcher] Registered handler %s for event type %s with priority %d",
		name, eventType, priority)
	
	return nil
}

// Register 注册事件处理器（泛型函数）
func Register[T any](d *Dispatcher, handler EventHandler[T]) error {
	var event T
	eventType := reflect.TypeOf(event)
	return d.register(eventType, handler, handler.Priority(), handler.Name())
}

// RegisterFunc 注册函数类型处理器（泛型函数）
func RegisterFunc[T any](d *Dispatcher, name string, priority int, handler HandlerFunc[T]) error {
	return Register(d, NewSimpleHandler(name, priority, handler))
}

// Dispatch 分发事件
func (d *Dispatcher) Dispatch(ctx context.Context, event interface{}) error {
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
		go d.dispatchToHandlers(ctx, event, eventType, wrappers)
		return nil
	}
	
	return d.dispatchToHandlers(ctx, event, eventType, wrappers)
}

func (d *Dispatcher) dispatchToHandlers(ctx context.Context, event interface{}, eventType reflect.Type, wrappers []handlerWrapper) error {
	for _, wrapper := range wrappers {
		if err := d.invokeHandler(ctx, event, wrapper); err != nil {
			if d.errorHandler != nil {
				d.errorHandler(err, eventType, wrapper.name)
			}
		}
	}
	return nil
}

func (d *Dispatcher) invokeHandler(ctx context.Context, event interface{}, wrapper handlerWrapper) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic in handler %s: %v", wrapper.name, r)
		}
	}()
	
	// 使用反射调用处理器
	handlerValue := reflect.ValueOf(wrapper.handler)
	handleMethod := handlerValue.MethodByName("Handle")
	
	if !handleMethod.IsValid() {
		return fmt.Errorf("handler %s does not have Handle method", wrapper.name)
	}
	
	// 创建事件上下文
	eventCtx := &Context[interface{}]{
		Context:  ctx,
		Event:    event,
		Metadata: make(map[string]interface{}),
		aborted:  false,
	}
	
	// 应用中间件
	finalHandler := func(c *Context[interface{}]) error {
		results := handleMethod.Call([]reflect.Value{reflect.ValueOf(eventCtx)})
		if len(results) > 0 && !results[0].IsNil() {
			return results[0].Interface().(error)
		}
		return nil
	}
	
	handler := applyMiddlewares(finalHandler, d.middlewares)
	return handler(eventCtx)
}

// GetHandlerCount 获取指定事件类型的处理器数量
func (d *Dispatcher) GetHandlerCount(eventType reflect.Type) int {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return len(d.handlers[eventType])
}

// GetHandlerCountForEvent 通过事件实例获取处理器数量
func (d *Dispatcher) GetHandlerCountForEvent(event interface{}) int {
	return d.GetHandlerCount(reflect.TypeOf(event))
}
