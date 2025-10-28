package event

import (
	"log"
	"time"
)

// Middleware 中间件类型
type Middleware func(next HandlerFunc[interface{}]) HandlerFunc[interface{}]

// applyMiddlewares 应用中间件链
func applyMiddlewares(handler HandlerFunc[interface{}], middlewares []Middleware) HandlerFunc[interface{}] {
	// 从后向前应用中间件
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}

// LoggingMiddleware 日志中间件
func LoggingMiddleware() Middleware {
	return func(next HandlerFunc[interface{}]) HandlerFunc[interface{}] {
		return func(ctx *Context[interface{}]) error {
			start := time.Now()
			log.Printf("[EventMiddleware] Before handling event")
			
			err := next(ctx)
			
			duration := time.Since(start)
			if err != nil {
				log.Printf("[EventMiddleware] After handling event (duration: %v, error: %v)", duration, err)
			} else {
				log.Printf("[EventMiddleware] After handling event (duration: %v)", duration)
			}
			
			return err
		}
	}
}

// RecoveryMiddleware 异常恢复中间件
func RecoveryMiddleware() Middleware {
	return func(next HandlerFunc[interface{}]) HandlerFunc[interface{}] {
		return func(ctx *Context[interface{}]) (err error) {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("[EventMiddleware] Recovered from panic: %v", r)
					if e, ok := r.(error); ok {
						err = e
					}
				}
			}()
			
			return next(ctx)
		}
	}
}

// TimeoutMiddleware 超时中间件
func TimeoutMiddleware(timeout time.Duration) Middleware {
	return func(next HandlerFunc[interface{}]) HandlerFunc[interface{}] {
		return func(ctx *Context[interface{}]) error {
			done := make(chan error, 1)
			
			go func() {
				done <- next(ctx)
			}()
			
			select {
			case err := <-done:
				return err
			case <-time.After(timeout):
				log.Printf("[EventMiddleware] Handler timeout after %v", timeout)
				ctx.Abort()
				return nil
			}
		}
	}
}

// FilterMiddleware 过滤中间件 - 根据条件决定是否执行处理器
func FilterMiddleware(filter func(ctx *Context[interface{}]) bool) Middleware {
	return func(next HandlerFunc[interface{}]) HandlerFunc[interface{}] {
		return func(ctx *Context[interface{}]) error {
			if !filter(ctx) {
				log.Printf("[EventMiddleware] Event filtered out")
				ctx.Abort()
				return nil
			}
			return next(ctx)
		}
	}
}

// RateLimitMiddleware 限流中间件
func RateLimitMiddleware(maxPerSecond int) Middleware {
	ticker := time.NewTicker(time.Second / time.Duration(maxPerSecond))
	
	return func(next HandlerFunc[interface{}]) HandlerFunc[interface{}] {
		return func(ctx *Context[interface{}]) error {
			<-ticker.C
			return next(ctx)
		}
	}
}

// MetricsMiddleware 指标收集中间件
func MetricsMiddleware(collector func(duration time.Duration, err error)) Middleware {
	return func(next HandlerFunc[interface{}]) HandlerFunc[interface{}] {
		return func(ctx *Context[interface{}]) error {
			start := time.Now()
			err := next(ctx)
			duration := time.Since(start)
			
			collector(duration, err)
			return err
		}
	}
}
