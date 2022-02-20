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

	"github.com/xgfone/go-apiserver/http/handler"
	"github.com/xgfone/go-apiserver/http/reqresp"
)

var ctxpool = sync.Pool{New: func() interface{} { return &Context{} }}

// Predefine some http headers.
var (
	HeaderAction = "X-Action"
)

// GetContext returns the Context from the http request.
func GetContext(req *http.Request) *Context {
	return reqresp.GetContext(req).Any.(*Context)
}

type actionsWrapper struct{ actions map[string]http.Handler }

// Context is the request context.
type Context struct {
	*reqresp.Context

	Action string
}

// Reset resets the context.
func (c *Context) Reset() {
	c.Context = nil
	c.Action = ""
}

// RouteManager is used to manage the routes based on the action service.
type RouteManager struct {
	// GetAction is used to get the action name from the http request.
	//
	// Default: HeaderAction
	GetAction func(*http.Request) string

	// NotFound is used when the manager is used as http.Handler
	// and does not find the route.
	//
	// Default: handler.Handler404
	NotFound http.Handler

	alock   sync.RWMutex
	amaps   map[string]http.Handler
	actions atomic.Value
}

// NewRouteManager returns a new http router based on the action.
func NewRouteManager() *RouteManager {
	r := &RouteManager{amaps: make(map[string]http.Handler, 16)}
	r.NotFound = http.HandlerFunc(notFound)
	r.updateActions()
	return r
}

// ServeHTTP implements the interface http.Handler.
func (m *RouteManager) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	m.Route(resp, req, m.NotFound)
}

// Route implements the interface router.RouteManager.
func (m *RouteManager) Route(w http.ResponseWriter, r *http.Request, notFound http.Handler) {
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
		m.respond(action, handler.Handler404, w, r)
	}
}

func (m *RouteManager) respond(action string, handler http.Handler,
	w http.ResponseWriter, r *http.Request) {
	ctx, new := reqresp.GetOrNewContext(r)
	if new {
		if rw, ok := w.(reqresp.ResponseWriter); ok {
			ctx.ResponseWriter = rw
		} else {
			ctx.ResponseWriter = reqresp.NewResponseWriter(w)
		}

		defer reqresp.DefaultContextAllocator.Release(ctx)
		r = reqresp.SetContext(r, ctx)
		w = ctx.ResponseWriter
	}

	c, ok := ctx.Any.(*Context)
	if ok {
		c.Action = action
	} else {
		c := ctxpool.Get().(*Context)
		c.Context = ctx
		c.Action = action
		ctx.Any = c
		defer releaseContext(c)
	}

	handler.ServeHTTP(w, r)
}

func releaseContext(c *Context) {
	c.Context.Any = nil
	c.Reset()
	ctxpool.Put(c)
}

/* ------------------------------------------------------------------------- */

// GetHandler returns the handler of the action.
//
// If the action does not exist, return nil.
func (m *RouteManager) GetHandler(action string) (handler http.Handler) {
	if len(action) == 0 {
		panic("action name is empty")
	}

	m.alock.RLock()
	handler = m.amaps[action]
	m.alock.RUnlock()
	return
}

// GetHandlers returns the handlers of all the actions.
func (m *RouteManager) GetHandlers() (handlers map[string]http.Handler) {
	m.alock.RLock()
	handlers = make(map[string]http.Handler, len(m.amaps))
	for action, handler := range m.amaps {
		handlers[action] = handler
	}
	m.alock.RUnlock()
	return
}

// GetActions returns the names of all the actions.
func (m *RouteManager) GetActions() (actions []string) {
	m.alock.RLock()
	actions = make([]string, 0, len(m.amaps))
	for action := range m.amaps {
		actions = append(actions, action)
	}
	m.alock.RUnlock()
	return
}

// RegisterContextFunc is the same as RegisterFunc, but use Context instead.
func (m *RouteManager) RegisterContextFunc(action string, f func(*Context)) (ok bool) {
	return m.RegisterFunc(action, func(rw http.ResponseWriter, r *http.Request) {
		f(GetContext(r))
	})
}

// RegisterFunc is the same as Register, but use the function as the handler.
func (m *RouteManager) RegisterFunc(action string, handler http.HandlerFunc) (ok bool) {
	return m.Register(action, handler)
}

// Register registers the action and the handler.
func (m *RouteManager) Register(action string, handler http.Handler) (ok bool) {
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
func (m *RouteManager) Update(actions map[string]http.Handler) {
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
func (m *RouteManager) Reset(actions map[string]http.Handler) {
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
func (m *RouteManager) Unregister(action string) (ok bool) {
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

func (m *RouteManager) updateActions() {
	actions := make(map[string]http.Handler, len(m.amaps))
	for action, handler := range m.amaps {
		actions[action] = handler
	}
	m.actions.Store(actionsWrapper{actions: actions})
}

func (m *RouteManager) checkAction(action string, handler http.Handler) {
	if len(action) == 0 {
		panic("action name is empty")
	}

	if handler == nil {
		panic("action handler is nil")
	}
}
