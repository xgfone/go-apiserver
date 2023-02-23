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

package log

import (
	"context"
	"io"
	"sync/atomic"

	"github.com/xgfone/go-apiserver/helper"
	"golang.org/x/exp/slog"
)

// NewJSONHandler returns a new json handler.
//
// If w is nil, use Writer instead.
func NewJSONHandler(w io.Writer, level Leveler) slog.Handler {
	return slog.HandlerOptions{
		Level:       level,
		AddSource:   true,
		ReplaceAttr: replaceSourceAttr,
	}.NewJSONHandler(w)
}

func replaceSourceAttr(groups []string, a slog.Attr) slog.Attr {
	if a.Key == slog.SourceKey {
		a.Value = slog.StringValue(helper.TrimPkgFile(a.Value.String()))
	}
	return a
}

type handlerWrapper struct{ Handler slog.Handler }

// SwitchHandler is a log handler to switch the handler.
type SwitchHandler struct {
	handler atomic.Value
}

// NewSwitchHandler returns a new slog handler.
func NewSwitchHandler(handler slog.Handler) *SwitchHandler {
	if handler == nil {
		panic("SwitchHandler: slog handler is nil")
	}

	sh := new(SwitchHandler)
	sh.handler.Store(handlerWrapper{handler})
	return sh
}

// Swap swaps the old handler with the new.
func (h *SwitchHandler) Swap(new slog.Handler) (old slog.Handler) {
	return h.handler.Swap(handlerWrapper{new}).(handlerWrapper).Handler
}

// Get returns the wrapped handler.
func (h *SwitchHandler) Get() slog.Handler {
	return h.handler.Load().(handlerWrapper).Handler
}

// Enabled implements the interface slog.Handler#Enabled.
func (h *SwitchHandler) Enabled(c context.Context, l Level) bool {
	return h.Get().Enabled(c, l)
}

// Handle implements the interface slog.Handler#Handle.
func (h *SwitchHandler) Handle(r slog.Record) error {
	return h.Get().Handle(r)
}

// WithAttrs implements the interface slog.Handler#WithAttrs.
func (h *SwitchHandler) WithAttrs(attrs []Attr) slog.Handler {
	return h.Get().WithAttrs(attrs)
}

// WithGroup implements the interface slog.Handler#WithGroup.
func (h *SwitchHandler) WithGroup(name string) slog.Handler {
	return h.Get().WithGroup(name)
}
