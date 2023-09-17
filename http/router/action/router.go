// Copyright 2022~2023 xgfone
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
	"maps"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/xgfone/go-apiserver/http/middleware"
	"github.com/xgfone/go-apiserver/http/reqresp"
	"github.com/xgfone/go-apiserver/result"
)

// HeaderAction represents the http header to store the action method.
var HeaderAction = "X-Action"

func notFoundHandler(resp http.ResponseWriter, req *http.Request) {
	c := reqresp.GetContext(resp, req)
	if len(c.Action) == 0 {
		c.Failure(result.ErrBadRequestInvalidAction.WithMessage("missing the action"))
	} else {
		c.Failure(result.ErrBadRequestInvalidAction.WithMessage("action '%s' is unsupported", c.Action))
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

	// Middlewares is used to manage the middlewares and applied to each route
	// when registering it. So, the middlewares will be run after routing
	// and never be run if not found the route.
	Middlewares *middleware.Manager

	alock   sync.RWMutex
	amaps   map[string]http.Handler
	actions atomic.Value
}

// NewRouter returns a new http router based on the action.
func NewRouter() *Router {
	r := &Router{amaps: make(map[string]http.Handler, 16)}
	r.Middlewares = middleware.NewManager(nil)
	r.NotFound = http.HandlerFunc(notFoundHandler)
	r.actions.Store(map[string]http.Handler(nil))
	r.updateActions()
	return r
}

// ServeHTTP implements the interface http.Handler.
func (r *Router) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	r.Route(resp, req, r.NotFound)
}

// Route implements the interface router.RouteManager.
func (r *Router) Route(rw http.ResponseWriter, req *http.Request, notFound http.Handler) {
	ctx := reqresp.GetContext(rw, req)
	if ctx == nil {
		ctx = reqresp.DefaultContextAllocator.Acquire()
		ctx.ResponseWriter = reqresp.NewResponseWriter(rw)
		ctx.Request = reqresp.SetContext(req, ctx)
		defer reqresp.DefaultContextAllocator.Release(ctx)
	}

	if len(ctx.Action) == 0 {
		if r.GetAction == nil {
			ctx.Action = req.Header.Get(HeaderAction)
		} else {
			ctx.Action = r.GetAction(req)
		}
	}

	var ok bool
	if len(ctx.Action) > 0 {
		ctx.Handler, ok = r.actions.Load().(map[string]http.Handler)[ctx.Action]
	}

	switch true {
	case ok:
	case notFound != nil:
		ctx.Handler = notFound
	case r.NotFound != nil:
		ctx.Handler = r.NotFound
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
func (r *Router) GetHandler(action string) (handler http.Handler) {
	if len(action) == 0 {
		panic("action name is empty")
	}

	r.alock.RLock()
	handler = r.amaps[action]
	r.alock.RUnlock()
	return
}

// GetHandlers returns the handlers of all the actions.
func (r *Router) GetHandlers() (handlers map[string]http.Handler) {
	r.alock.RLock()
	handlers = maps.Clone(r.amaps)
	r.alock.RUnlock()
	return
}

// GetActions returns the names of all the actions.
func (r *Router) GetActions() []string {
	r.alock.RLock()
	actions := make([]string, 0, len(r.amaps))
	for action := range r.amaps {
		actions = append(actions, action)
	}
	r.alock.RUnlock()
	return actions
}

// RegisterContextFuncWithError is the same as RegisterContextFunc,
// but supports to return an error.
func (r *Router) RegisterContextFuncWithError(action string, f reqresp.HandlerWithError) (ok bool) {
	return r.RegisterFunc(action, func(w http.ResponseWriter, r *http.Request) {
		c := reqresp.GetContext(w, r)
		c.UpdateError(f(c))
	})
}

// RegisterContextFunc is the same as RegisterFunc, but use Context instead.
func (r *Router) RegisterContextFunc(action string, f reqresp.Handler) (ok bool) {
	return r.RegisterFunc(action, func(w http.ResponseWriter, r *http.Request) {
		f(reqresp.GetContext(w, r))
	})
}

// RegisterFunc is the same as Register, but use the function as the handler.
func (r *Router) RegisterFunc(action string, handler http.HandlerFunc) (ok bool) {
	return r.Register(action, handler)
}

// Register registers the action and the handler.
func (r *Router) Register(action string, handler http.Handler) (ok bool) {
	handler = r.checkAction(action, handler)

	r.alock.Lock()
	if _, ok := r.amaps[action]; !ok {
		r.amaps[action] = handler
		r.updateActions()
	}
	r.alock.Unlock()

	return
}

// Update updates the given actions and handlers, which will add the action
// if it does not exist, or update it to the new.
func (r *Router) Update(actions map[string]http.Handler) {
	if len(actions) == 0 {
		return
	}

	for action, handler := range actions {
		actions[action] = r.checkAction(action, handler)
	}

	r.alock.Lock()
	maps.Copy(r.amaps, actions)
	r.updateActions()
	r.alock.Unlock()
}

// Reset discards all the original actions and resets them to actions.
func (r *Router) Reset(actions map[string]http.Handler) {
	for action, handler := range actions {
		actions[action] = r.checkAction(action, handler)
	}

	r.alock.Lock()
	clear(r.amaps)
	maps.Copy(r.amaps, actions)
	r.updateActions()
	r.alock.Unlock()
}

// Unregister unregisters the given action.
func (r *Router) Unregister(action string) (ok bool) {
	if action == "" {
		panic("action name is empty")
	}

	r.alock.Lock()
	if _, ok := r.amaps[action]; ok {
		delete(r.amaps, action)
		r.updateActions()
	}
	r.alock.Unlock()

	return
}

func (r *Router) updateActions() { r.actions.Store(maps.Clone(r.amaps)) }
func (r *Router) checkAction(action string, handler http.Handler) http.Handler {
	if len(action) == 0 {
		panic("action name is empty")
	}

	if handler == nil {
		panic("action handler is nil")
	}

	return r.Middlewares.Handler(handler)
}
