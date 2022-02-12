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

package tcp

import (
	"context"
	"net"

	"github.com/xgfone/go-apiserver/internal/atomic"
)

type handlerWrapper struct{ Handler }

var _ Handler = &SwitchHandler{}

// SwitchHandler is a switch handler, which can switch the handler to any one.
type SwitchHandler struct{ handler atomic.Value }

func newSwitchHandler(handler Handler) (sh SwitchHandler) {
	if handler == nil {
		panic("SwitchHandler: the tcp handler is nil")
	}

	sh.handler.Store(handlerWrapper{handler})
	return
}

// NewSwitchHandler returns a new switch handler with the initial handler.
func NewSwitchHandler(handler Handler) *SwitchHandler {
	sh := newSwitchHandler(handler)
	return &sh
}

// Get returns the current handler.
func (sh *SwitchHandler) Get() Handler {
	return sh.handler.Load().(handlerWrapper).Handler
}

// Swap stores the new handler and returns the old.
func (sh *SwitchHandler) Swap(new Handler) (old Handler) {
	if new == nil {
		panic("SwitchHandler: the new handler is nil")
	}
	return sh.handler.Swap(handlerWrapper{new}).(handlerWrapper).Handler
}

// OnConnection implements the interface Handler, which will forward the call
// to the inner handler.
func (sh *SwitchHandler) OnConnection(c net.Conn) { sh.Get().OnConnection(c) }

// OnServerExit implements the interface Handler, which will forward the call
// to the inner handler.
func (sh *SwitchHandler) OnServerExit(err error) { sh.Get().OnServerExit(err) }

// OnShutdown implements the interface Handler, which will forward the call
// to the inner handler.
func (sh *SwitchHandler) OnShutdown(c context.Context) { sh.Get().OnShutdown(c) }
