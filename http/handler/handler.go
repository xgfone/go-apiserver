// Copyright 2021 xgfone
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

// Package handler provides some http handler and middleware functions.
package handler

import (
	"net/http"

	"github.com/xgfone/go-apiserver/helper"
	"github.com/xgfone/go-apiserver/internal/atomic"
)

// Pre-define some http handlers.
var (
	Handler200 http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})

	Handler400 http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
	})

	Handler404 http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Connection", "close")
		w.WriteHeader(404)
	})
)

/// ----------------------------------------------------------------------- ///

// WrappedHandlerFunc is the function to handle the request with the wrapped handler.
type WrappedHandlerFunc func(http.Handler, http.ResponseWriter, *http.Request)

// WrappedHandler is a http handler which wraps and returns the inner handler.
type WrappedHandler interface {
	WrappedHandler() http.Handler
	http.Handler
}

type wrappedHandler struct {
	Handler http.Handler
	Handle  WrappedHandlerFunc
}

func (wh wrappedHandler) Close() error                 { return helper.Close(wh.Handler) }
func (wh wrappedHandler) WrappedHandler() http.Handler { return wh.Handler }
func (wh wrappedHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	wh.Handle(wh.Handler, w, r)
}

// WrapHandler returns a new WrappedHandler with the wrapped handler
// and the handling function.
func WrapHandler(handler http.Handler, handlerFunc WrappedHandlerFunc) WrappedHandler {
	return wrappedHandler{Handler: handler, Handle: handlerFunc}
}

// UnwrapHandler unwraps the wrapped innest http handler from handler if it has
// implemented the interface WrappedHandler. Or return the original handler.
func UnwrapHandler(handler http.Handler) http.Handler {
	for {
		if wh, ok := handler.(WrappedHandler); ok {
			handler = wh.WrappedHandler()
		} else {
			break
		}
	}
	return handler
}

/// ----------------------------------------------------------------------- ///

type httpHandlerWrapper struct{ http.Handler }

// SwitchHandler is a HTTP handler that is used to switch the real handler.
type SwitchHandler struct{ handler atomic.Value }

// NewSwitchHandler returns a new switch handler with the initial handler.
func NewSwitchHandler(handler http.Handler) *SwitchHandler {
	if handler == nil {
		panic("SwitchHandler: the http handler is nil")
	}

	sh := &SwitchHandler{}
	sh.handler.Store(httpHandlerWrapper{handler})
	return sh
}

// Close implements the interface io.Closer.
func (sh *SwitchHandler) Close() error { return helper.Close(sh.Get()) }

// Set sets the http handler to new.
func (sh *SwitchHandler) Set(new http.Handler) {
	sh.handler.Store(httpHandlerWrapper{new})
}

// Get returns the current handler.
func (sh *SwitchHandler) Get() http.Handler {
	return sh.handler.Load().(httpHandlerWrapper).Handler
}

// Swap stores the new handler and returns the old.
func (sh *SwitchHandler) Swap(new http.Handler) (old http.Handler) {
	if new == nil {
		panic("SwitchHandler.Swap(): the new http handler is nil")
	}
	return sh.handler.Swap(httpHandlerWrapper{new}).(httpHandlerWrapper).Handler
}

// ServeHTTP implements the interface http.Handler.
func (sh *SwitchHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	sh.handler.Load().(httpHandlerWrapper).Handler.ServeHTTP(w, r)
}

// WrappedHandler implements the interface WrappedHandler.
func (sh *SwitchHandler) WrappedHandler() http.Handler { return sh.Get() }
