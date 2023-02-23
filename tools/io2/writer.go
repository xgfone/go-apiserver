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

package io2

import (
	"io"
	"sync"
	"sync/atomic"
)

var _ io.WriteCloser = new(SwitchWriter)

type wrappedWriter struct{ io.Writer }

// SwitchWriter is a writer proxy, which can be switch the writer to another
// at run-time.
type SwitchWriter struct {
	w atomic.Value
}

// NewSwitchWriter returns a new SwitchWriter with w.
func NewSwitchWriter(w io.Writer) *SwitchWriter {
	sw := new(SwitchWriter)
	sw.w.Store(wrappedWriter{w})
	return sw
}

// Write implements the interface io.Writer.
func (w *SwitchWriter) Write(b []byte) (int, error) {
	return w.Get().Write(b)
}

// Close implements the interface io.Closer.
func (w *SwitchWriter) Close() error {
	return Close(w.w)
}

// Get returns the wrapped writer.
func (w *SwitchWriter) Get() io.Writer {
	return w.w.Load().(wrappedWriter).Writer
}

// Swap swaps the old writer with the new writer.
func (w *SwitchWriter) Swap(new io.Writer) (old io.Writer) {
	return w.w.Swap(wrappedWriter{new}).(wrappedWriter).Writer
}

// SafeWriter is a writer to write the data concurrently and safely.
type SafeWriter struct {
	m sync.Mutex
	w io.Writer
}

// NewSafeWriter returns a new safe writer.
func NewSafeWriter(w io.Writer) *SafeWriter {
	if w == nil {
		panic("SafeWriter: io.Writer is nil")
	}
	return &SafeWriter{w: w}
}

// Write implements the interface io.Writer.
func (w *SafeWriter) Write(b []byte) (int, error) {
	w.m.Lock()
	defer w.m.Unlock()
	return w.w.Write(b)
}

// Close implements the interface io.Closer.
func (w *SafeWriter) Close() error {
	w.m.Lock()
	defer w.m.Unlock()
	return Close(w.w)
}

// Sync calls the Sync method if the inner writer has implemented the interface
// { Sync() error }. Or, do nothing and return nil.
func (w *SafeWriter) Sync() error {
	w.m.Lock()
	defer w.m.Unlock()

	if ws, ok := w.w.(interface{ Sync() error }); ok {
		return ws.Sync()
	}
	return nil
}

// Run executes the function f with the inner writer.
func (w *SafeWriter) Run(f func(io.Writer)) {
	w.m.Lock()
	defer w.m.Unlock()
	f(w.w)
}

// Get returns the wrapped writer.
func (w *SafeWriter) Get() io.Writer {
	return w.w
}
