// Copyright 2026 xgfone
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package middleware

import (
	"io"
	"net/http"

	"github.com/xgfone/go-apiserver/http/reqresp"
)

var _ Middleware = ContextHandler(nil)

// ContextHandler is a handler function that processes an HTTP request
// using the reqresp.Context object. It returns an error if the request
// cannot be processed successfully.
type ContextHandler reqresp.HandlerWithError

// Handler implements the Middleware interface.
func (ch ContextHandler) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if c := reqresp.GetContext(r.Context()); c == nil {
			w.WriteHeader(500)
			_, _ = io.WriteString(w, "missing reqresp.Context")
		} else if err := ch(c); err != nil {
			c.AppendError(err)
		} else {
			next.ServeHTTP(w, r)
		}
	})
}

// InsertContextHandler inserts a set of context handlers as middlewares
// into the front and return a new middleware slice.
func (ms Middlewares) InsertContextHandler(hs ...ContextHandler) Middlewares {
	return addMiddlewares(ms, ms.Insert, hs)
}

// AppendContextHandler appends a set of context handlers as middlewares
// and return a new middleware slice.
func (ms Middlewares) AppendContextHandler(hs ...ContextHandler) Middlewares {
	return addMiddlewares(ms, ms.Append, hs)
}

// InsertContextHandler inserts a set of context handlers as middlewares into the front.
func (m *Manager) InsertContextHandler(hs ...ContextHandler) {
	managerAddMiddlewares(m, m.mdws.Insert, hs)
}

// AppendContextHandler appends a set of context handlers as middlewares.
func (m *Manager) AppendContextHandler(hs ...ContextHandler) {
	managerAddMiddlewares(m, m.mdws.Append, hs)
}

// NewWithContextHandler returns a new middleware that executes the given ContextHandler.
func NewWithContextHandler(name string, priority int, handler ContextHandler) Middleware {
	return New(name, priority, handler.Handler)
}

// Or creates a middleware that executes the given ContextHandlers in order
// and stops at the first handler that returns nil error (logical OR).
//
// Parameters:
//   - name: The name of the middleware for identification and debugging.
//   - priority: The priority of the middleware. Lower values have higher priority.
//   - handlers: One or more ContextHandler functions to execute.
//
// Returns:
//   - A Middleware that will execute the handlers in logical OR fashion.
//
// Behavior:
//   - If any handler returns nil error, the chain stops and the request
//     proceeds to the next handler in the middleware chain.
//   - If all handlers return errors, the last error is appended to the context
//     and the request does not proceed to the next handler.
//   - If no handlers are provided, the middleware acts as a pass-through
//     and the next handler is called directly.
//   - If exactly one handler is provided, it's optimized to directly use
//     the handler's Handler method without additional wrapping.
//   - If any handler is nil, the function panics.
func Or(name string, priority int, handlers ...ContextHandler) Middleware {
	if len(handlers) == 1 {
		if handlers[0] == nil {
			panic("middleware: ContextHandler must not be nil")
		}
		return NewWithContextHandler(name, priority, handlers[0])
	}
	return New(name, priority, contextHandlersToMiddlewareFunc(handlers, false))
}

// And creates a middleware that executes the given ContextHandlers in order
// and requires all handlers to return nil error for the request to proceed
// (logical AND).
//
// Parameters:
//   - name: The name of the middleware for identification and debugging.
//   - priority: The priority of the middleware. Lower values have higher priority.
//   - handlers: One or more ContextHandler functions to execute.
//
// Returns:
//   - A Middleware that will execute the handlers in logical AND fashion.
//
// Behavior:
//   - All handlers must return nil error for the request to proceed to the
//     next handler in the middleware chain.
//   - If any handler returns an error, the chain stops immediately, the error
//     is appended to the context, and the request does not proceed to the next handler.
//   - If no handlers are provided, the middleware acts as a pass-through
//     and the next handler is called directly.
//   - If exactly one handler is provided, it's optimized to directly use
//     the handler's Handler method without additional wrapping.
//   - If any handler is nil, the function panics.
func And(name string, priority int, handlers ...ContextHandler) Middleware {
	if len(handlers) == 1 {
		if handlers[0] == nil {
			panic("middleware: ContextHandler must not be nil")
		}
		return NewWithContextHandler(name, priority, handlers[0])
	}
	return New(name, priority, contextHandlersToMiddlewareFunc(handlers, true))
}

// contextHandlersToMiddlewareFunc converts a slice of ContextHandlers into a MiddlewareFunc
// with the specified execution strategy.
//
// Parameters:
//   - handlers: The ContextHandlers to execute.
//   - fast: If true, uses AND mode (stops on first error).
//     If false, uses OR mode (stops on first success).
//
// Returns:
//   - A MiddlewareFunc that wraps the handlers with the specified execution strategy.
//
// Behavior:
//   - If any handler in the slice is nil, the function panics.
//   - If no handlers are provided, returns a middleware function that passes through
//     to the next handler without any processing.
//   - For non-empty handler lists, delegates to convertContextHandlers for execution.
func contextHandlersToMiddlewareFunc(handlers []ContextHandler, fast bool) MiddlewareFunc {
	for _, h := range handlers {
		if h == nil {
			panic("middleware: ContextHandler must not be nil")
		}
	}

	return func(next http.Handler) http.Handler {
		if len(handlers) == 0 {
			return next
		}
		return convertContextHandlers(handlers, next, fast)
	}
}

// convertContextHandlers creates an http.HandlerFunc that executes
// the given ContextHandlers with the specified execution strategy.
//
// Parameters:
//   - handlers: The ContextHandlers to execute.
//   - next: The next http.Handler in the chain to call.
//   - fast: If true, uses AND mode (stops on first error).
//     If false, uses OR mode (stops on first success).
//
// Returns:
//   - An http.HandlerFunc that executes the handlers.
//
// Behavior:
//   - Retrieves the reqresp.Context from the request context.
//   - If no context is found, returns HTTP 500 with error message.
//   - Executes handlers according to the specified mode:
//   - fast=true (AND): Stops on first error, appends error to context.
//     If all handlers succeed, calls the next handler.
//   - fast=false (OR): Stops on first success, calls the next handler.
//     If all handlers fail, appends the last error to context.
func convertContextHandlers(handlers []ContextHandler, next http.Handler, fast bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c := reqresp.GetContext(r.Context())
		if c == nil {
			w.WriteHeader(500)
			_, _ = io.WriteString(w, "missing reqresp.Context")
			return
		}

		var err error
		for i := range handlers {
			err = handlers[i](c)
			if shouldReturn(fast, err) {
				break
			}
		}

		if err != nil {
			c.AppendError(err)
		} else {
			next.ServeHTTP(w, r)
		}
	}
}

// 1. If fast is true, return true when first error is encountered.
// 2. If fast is false, return true when first success is encountered.
func shouldReturn(fast bool, err error) bool {
	return fast == (err != nil)
}
