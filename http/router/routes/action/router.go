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
	"github.com/xgfone/go-apiserver/result"
)

// HeaderAction represents the http header to store the action method.
var HeaderAction = "X-Action"

type actionsWrapper struct{ actions map[string]http.Handler }

func notFoundHandler(resp http.ResponseWriter, req *http.Request) {
	c := reqresp.GetContext(resp, req)
	if len(c.Action) == 0 {
		c.Failure(result.ErrInvalidAction.WithMessage("missing the action"))
	} else {
		c.Failure(result.ErrInvalidAction.WithMessage("action '%s' is unsupported", c.Action))
	}
}

// DefaultRouter is the default global action router.
var DefaultRouter = NewRouter()

// Router is used to manage the routes based on the action service.
type Router struct {
	// GetAction is used to get the action name from the http request.
	//
	// Default: HeaderAction
	GetAction func(*http.Request) string

	// NotFound is used when the manager is used as http.Handler
	// and does not find the route.
	//
	// Default: c.Failure(result.ErrInvalidAction)
	NotFound http.Handler

	alock   sync.RWMutex
	amaps   map[string]http.Handler
	actions atomic.Value
}

// NewRouter returns a new http router based on the action.
func NewRouter() *Router {
	r := &Router{amaps: make(map[string]http.Handler, 16)}
	r.NotFound = http.HandlerFunc(notFoundHandler)
	r.updateActions()
	return r
}

// ServeHTTP implements the interface http.Handler.
func (m *Router) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	m.Route(resp, req, m.NotFound)
}

// Route implements the interface router.RouteManager.
func (m *Router) Route(w http.ResponseWriter, r *http.Request, notFound http.Handler) {
	ctx := reqresp.GetContext(w, r)
	if ctx == nil {
		ctx = reqresp.DefaultContextAllocator.Acquire()
		ctx.ResponseWriter = reqresp.NewResponseWriter(w)
		ctx.Request = reqresp.SetContext(r, ctx)
		defer reqresp.DefaultContextAllocator.Release(ctx)
	}

	if m.GetAction != nil {
		ctx.Action = m.GetAction(r)
	} else {
		ctx.Action = r.Header.Get(HeaderAction)
	}

	var ok bool
	if len(ctx.Action) > 0 {
		ctx.Handler, ok = m.actions.Load().(actionsWrapper).actions[ctx.Action]
	}

	switch true {
	case ok:
	case notFound != nil:
		ctx.Handler = notFound
	case m.NotFound != nil:
		ctx.Handler = m.NotFound
	default:
		ctx.Handler = http.HandlerFunc(notFoundHandler)
	}

	ctx.Handler.ServeHTTP(ctx.ResponseWriter, ctx.Request)
	if !ctx.WroteHeader() {
		ctx.Failure(ctx.Err)
	}
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

// RegisterContextFuncWithError is the same as RegisterContextFunc,
// but supports to return an error.
func (m *Router) RegisterContextFuncWithError(action string, f reqresp.HandlerWithError) (ok bool) {
	return m.RegisterFunc(action, func(w http.ResponseWriter, r *http.Request) {
		c := reqresp.GetContext(w, r)
		c.UpdateError(f(c))
	})
}

// RegisterContextFunc is the same as RegisterFunc, but use Context instead.
func (m *Router) RegisterContextFunc(action string, f reqresp.Handler) (ok bool) {
	return m.RegisterFunc(action, func(w http.ResponseWriter, r *http.Request) {
		f(reqresp.GetContext(w, r))
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
