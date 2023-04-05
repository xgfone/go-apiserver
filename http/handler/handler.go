// Copyright 2021~2023 xgfone
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

// Package handler provides some http handler functions.
package handler

import (
	"net/http"

	"github.com/xgfone/go-atomicvalue"
)

// Pre-define some http handlers.
var (
	Handler200 http.Handler = http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(200)
	})

	Handler400 http.Handler = http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(400)
	})

	Handler404 http.Handler = http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Connection", "close")
		w.WriteHeader(404)
	})
)

// SwitchHandler is a HTTP handler that is used to switch the real handler.
type SwitchHandler struct {
	handler atomicvalue.Value[http.Handler]
}

// NewSwitchHandler returns a new switch handler with the initial handler.
func NewSwitchHandler(handler http.Handler) *SwitchHandler {
	if handler == nil {
		panic("SwitchHandler: the http handler is nil")
	}
	return &SwitchHandler{handler: atomicvalue.NewValue(handler)}
}

// Set sets the http handler to new.
func (sh *SwitchHandler) Set(new http.Handler) { sh.handler.Store(new) }

// Get returns the current handler.
func (sh *SwitchHandler) Get() http.Handler { return sh.handler.Load() }

// Swap stores the new handler and returns the old.
func (sh *SwitchHandler) Swap(new http.Handler) (old http.Handler) {
	if new == nil {
		panic("SwitchHandler.Swap: the new http handler is nil")
	}
	return sh.handler.Swap(new)
}

// ServeHTTP implements the interface http.Handler.
func (sh *SwitchHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	sh.handler.Load().ServeHTTP(w, r)
}
