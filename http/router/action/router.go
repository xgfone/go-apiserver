// Copyright 2023 xgfone
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

// Package action implements a router based on the action service.
package action

import (
	"net/http"

	"github.com/xgfone/go-apiserver/http/middleware"
	"github.com/xgfone/go-apiserver/http/reqresp"
	"github.com/xgfone/go-apiserver/result"
	"github.com/xgfone/go-toolkit/codeint"
)

// HeaderAction is the http header to store the action method.
var HeaderAction = "X-Action"

// DefaultRouter is the default global action router.
var DefaultRouter = NewRouter()

// Router is used to manage the routes based on the action service.
type Router struct {
	// GetAction is used to get the action name from the http request.
	//
	// Default: get action name from the request header HeaderAction.
	GetAction func(*http.Request) string

	// NotFound is called when the handler is not found by the action.
	//
	// Default: call result.ErrBadRequestInvalidAction.Respond
	NotFound http.Handler

	// Middlewares is used to manage the middlewares and applied
	// to each action handler when registering it.
	//
	// So, the middlewares will be run after routing
	// but never be run if the handler is not found.
	Middlewares *middleware.Manager

	handlers map[string]http.Handler
}

// NewRouter returns a new http router based on the action.
func NewRouter() *Router {
	return &Router{
		Middlewares: middleware.NewManager(nil),
		NotFound:    http.HandlerFunc(notFoundHandler),
		handlers:    make(map[string]http.Handler, 16),
	}
}

// Actions returns all the actions.
func (r *Router) Actions() []string {
	actions := make([]string, 0, len(r.handlers))
	for action := range r.handlers {
		actions = append(actions, action)
	}
	return actions
}

// GetHandlers returns all the action handlers, which is read-only.
func (r *Router) Handlers() map[string]http.Handler {
	return r.handlers
}

// ServeHTTP implements the interface http.Handler.
func (r *Router) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	var action string
	if r.GetAction == nil {
		action = req.Header.Get(HeaderAction)
	} else {
		action = r.GetAction(req)
	}

	var handler http.Handler
	if action != "" {
		handler = r.handlers[action]
		if c := reqresp.GetContext(req.Context()); c != nil {
			c.Action = action
		}
	}

	switch {
	case handler != nil:
		handler.ServeHTTP(rw, req)

	case r.NotFound != nil:
		r.NotFound.ServeHTTP(rw, req)

	default:
		notFoundHandler(rw, req)
	}
}

func notFoundHandler(resp http.ResponseWriter, req *http.Request) {
	c := reqresp.GetContext(req.Context())
	if len(c.Action) == 0 {
		result.Err(codeint.ErrBadRequest.WithMessage("missing the action")).Respond(c)
	} else {
		result.Err(codeint.ErrBadRequest.WithMessagef("action '%s' is unsupported", c.Action)).Respond(c)
	}
}

func (r *Router) checkAction(action string, handler http.Handler) http.Handler {
	if len(action) == 0 {
		panic("action name is empty")
	} else if handler == nil {
		panic("action handler is nil")
	}
	return r.Middlewares.Handler(handler)
}

// Register registers the action and the handler.
//
// If exist, override it.
func (r *Router) Register(action string, handler http.Handler) {
	r.handlers[action] = r.checkAction(action, handler)
}

// RegisterFunc is the same as Register, but use the function as the handler.
func (r *Router) RegisterFunc(action string, handler http.HandlerFunc) {
	r.Register(action, handler)
}

// RegisterContext is the same as RegisterFunc, but use Context instead.
func (r *Router) RegisterContext(action string, handler reqresp.Handler) {
	r.Register(action, handler)
}

// RegisterContextWithError is the same as RegisterContext,
// but supports the handler to return an error.
func (r *Router) RegisterContextWithError(action string, handler reqresp.HandlerWithError) {
	r.Register(action, handler)
}
