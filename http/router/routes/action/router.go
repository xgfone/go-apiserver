// Copyright 2022 xgfone
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

// Package action implements a route manager based on the action service.
package action

import (
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/xgfone/go-apiserver/http/reqresp"
	"github.com/xgfone/go-apiserver/middleware"
)

var ctxpool = sync.Pool{New: func() interface{} { return &Context{} }}

// Predefine some http headers.
var (
	HeaderAction = "X-Action"
)

// GetContext returns the Context from the http request.
func GetContext(w http.ResponseWriter, r *http.Request) *Context {
	c, _ := reqresp.GetContext(w, r).Reg3.(*Context)
	return c
}

type actionsWrapper struct{ actions map[string]http.Handler }

// Context is the request context.
type Context struct {
	// Notice: It uses Reg3 to store the action context.
	*reqresp.Context

	// If action is empty, it represents that there is no action in the request.
	Action  string
	handler http.Handler
	respond func(*Context, Response) error
}

// Reset resets the context.
func (c *Context) Reset() { *c = Context{} }

// DefaultRouter is the default global action router.
var DefaultRouter = NewDefaultRouter()

// Router is used to manage the routes based on the action service.
type Router struct {
	// Middlewares is used to manage the middlewares of the action handlers,
	// which will wrap the handlers of all the actions and take effect
	// after finding the action and before the action handler is executed.
	Middlewares *middleware.Manager

	// GetAction is used to get the action name from the http request.
	//
	// Default: HeaderAction
	GetAction func(*http.Request) string

	// NotFound is used when the manager is used as http.Handler
	// and does not find the route.
	//
	// Default: c.Failure(ErrInvalidAction)
	NotFound http.Handler

	// HandleResponse is used to wrap the response and handle it by itself,
	// which is used by the methods of Context: Respond, Success, Failure.
	//
	// Default: c.JSON(200, resp)
	HandleResponse func(c *Context, resp Response) error

	alock   sync.RWMutex
	amaps   map[string]http.Handler
	actions atomic.Value
}

// NewRouter returns a new http router based on the action.
func NewRouter() *Router {
	r := &Router{amaps: make(map[string]http.Handler, 16)}
	r.NotFound = http.HandlerFunc(notFoundHandler)
	r.Middlewares = middleware.NewManager(nil)
	r.Middlewares.SetHandler(http.HandlerFunc(r.serveHTTP))
	r.updateActions()
	return r
}

// NewDefaultRouter returns a new default router, which is the same as NewRouter,
// but also adds the middlewares Logger(10) and Recover(20).
func NewDefaultRouter() *Router {
	r := NewRouter()
	r.Middlewares.Use(Logger(10), Recover(20))
	return r
}

// ServeHTTP implements the interface http.Handler.
func (m *Router) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	m.Route(resp, req, m.NotFound)
}

// Route implements the interface router.RouteManager.
func (m *Router) Route(w http.ResponseWriter, r *http.Request, notFound http.Handler) {
	var action string
	if m.GetAction != nil {
		action = m.GetAction(r)
	} else {
		action = r.Header.Get(HeaderAction)
	}

	var ok bool
	var h http.Handler
	if len(action) > 0 {
		h, ok = m.actions.Load().(actionsWrapper).actions[action]
	}

	if ok {
		m.respond(action, h, w, r)
	} else if notFound != nil {
		m.respond(action, notFound, w, r)
	} else if m.NotFound != nil {
		m.respond(action, m.NotFound, w, r)
	} else {
		m.respond(action, http.HandlerFunc(notFoundHandler), w, r)
	}
}

func (m *Router) respond(action string, handler http.Handler,
	w http.ResponseWriter, r *http.Request) {
	ctx := reqresp.GetContext(w, r)
	if ctx == nil {
		ctx = reqresp.DefaultContextAllocator.Acquire()
		ctx.ResponseWriter = reqresp.NewResponseWriter(w)
		ctx.Request = reqresp.SetContext(r, ctx)
		defer reqresp.DefaultContextAllocator.Release(ctx)
	}

	c, ok := ctx.Reg3.(*Context)
	if ok {
		c.Action = action
	} else {
		c = ctxpool.Get().(*Context)
		c.Context = ctx
		c.Action = action
		ctx.Reg3 = c
		defer releaseContext(c)
	}

	c.handler = handler
	c.respond = m.HandleResponse
	m.Middlewares.ServeHTTP(ctx, ctx.Request)
	if !c.WroteHeader() {
		if c.Err == nil {
			c.Success(nil)
		} else {
			c.Failure(c.Err)
		}
	}
}

func releaseContext(c *Context) {
	c.Reset()
	ctxpool.Put(c)
}

func (m *Router) serveHTTP(w http.ResponseWriter, r *http.Request) {
	GetContext(w, r).handler.ServeHTTP(w, r)
}

/* ------------------------------------------------------------------------- */

// GetHandler returns the handler of the action.
//
// If the action does not exist, return nil.
func (m *Router) GetHandler(action string) (handler http.Handler) {
	if len(action) == 0 {
		panic("action name is empty")
	}

	m.alock.RLock()
	handler = m.amaps[action]
	m.alock.RUnlock()
	return
}

// GetHandlers returns the handlers of all the actions.
func (m *Router) GetHandlers() (handlers map[string]http.Handler) {
	m.alock.RLock()
	handlers = make(map[string]http.Handler, len(m.amaps))
	for action, handler := range m.amaps {
		handlers[action] = handler
	}
	m.alock.RUnlock()
	return
}

// GetActions returns the names of all the actions.
func (m *Router) GetActions() (actions []string) {
	m.alock.RLock()
	actions = make([]string, 0, len(m.amaps))
	for action := range m.amaps {
		actions = append(actions, action)
	}
	m.alock.RUnlock()
	return
}

// RegisterContextFunc is the same as RegisterFunc, but use Context instead.
func (m *Router) RegisterContextFunc(action string, f func(*Context)) (ok bool) {
	return m.RegisterFunc(action, func(w http.ResponseWriter, r *http.Request) {
		f(GetContext(w, r))
	})
}

// RegisterFunc is the same as Register, but use the function as the handler.
func (m *Router) RegisterFunc(action string, handler http.HandlerFunc) (ok bool) {
	return m.Register(action, handler)
}

// Register registers the action and the handler.
func (m *Router) Register(action string, handler http.Handler) (ok bool) {
	m.checkAction(action, handler)

	m.alock.Lock()
	_, ok = m.amaps[action]
	if ok = !ok; ok {
		m.amaps[action] = handler
		m.updateActions()
	}
	m.alock.Unlock()

	return
}

// Update updates the given actions and handlers, which will add the action
// if it does not exist, or update it to the new.
func (m *Router) Update(actions map[string]http.Handler) {
	if len(actions) == 0 {
		return
	}

	for action, handler := range actions {
		m.checkAction(action, handler)
	}

	m.alock.Lock()
	for action, handler := range actions {
		m.amaps[action] = handler
	}
	m.updateActions()
	m.alock.Unlock()
}

// Reset discards all the original actions and resets them to actions.
func (m *Router) Reset(actions map[string]http.Handler) {
	for action, handler := range actions {
		m.checkAction(action, handler)
	}

	m.alock.Lock()
	for action := range m.amaps {
		delete(m.amaps, action)
	}

	for action, handler := range actions {
		m.amaps[action] = handler
	}
	m.updateActions()
	m.alock.Unlock()
}

// Unregister unregisters the given action.
func (m *Router) Unregister(action string) (ok bool) {
	if action == "" {
		panic("action name is empty")
	}

	m.alock.Lock()
	if _, ok = m.amaps[action]; ok {
		delete(m.amaps, action)
		m.updateActions()
	}
	m.alock.Unlock()

	return
}

func (m *Router) updateActions() {
	actions := make(map[string]http.Handler, len(m.amaps))
	for action, handler := range m.amaps {
		actions[action] = handler
	}
	m.actions.Store(actionsWrapper{actions: actions})
}

func (m *Router) checkAction(action string, handler http.Handler) {
	if len(action) == 0 {
		panic("action name is empty")
	}

	if handler == nil {
		panic("action handler is nil")
	}
}
