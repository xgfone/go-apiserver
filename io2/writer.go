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

	"github.com/xgfone/go-atomicvalue"
)

var (
	_ io.WriteCloser = new(SwitchWriter)
	_ io.WriteCloser = new(SafeWriter)
)

// RunWriter executes the function f with the writer w.
// If w has implemented the interface { Run(func(io.Writer)) },
// call the Run method with f instead.
func RunWriter(w io.Writer, f func(io.Writer)) {
	if wr, ok := w.(interface{ Run(func(io.Writer)) }); ok {
		wr.Run(f)
	} else {
		f(w)
	}
}

// SyncWriter calls the Sync method if w has implemented the interface
// { Sync() error } or { Flush() error }, or call the Run method to finish
// to sync if w has implemented the interface { Run(func(io.Writer)) }.
// Or, do nothing and return nil.
func SyncWriter(w io.Writer) (err error) {
	switch _w := w.(type) {
	case interface{ Sync() error }:
		err = _w.Sync()

	case interface{ Flush() error }:
		err = _w.Flush()

	case interface{ Run(func(io.Writer)) }:
		_w.Run(func(w io.Writer) { err = SyncWriter(w) })
	}

	return
}

// SwitchWriter is a writer proxy, which can be switch the writer to another
// at run-time.
type SwitchWriter struct {
	w atomicvalue.Value[io.Writer]
}

// NewSwitchWriter returns a new SwitchWriter with w.
func NewSwitchWriter(w io.Writer) *SwitchWriter {
	if w == nil {
		panic("SwitchWriter: io.Writer is nil")
	}
	return &SwitchWriter{w: atomicvalue.NewValue(w)}
}

// Write implements the interface io.Writer.
func (w *SwitchWriter) Write(b []byte) (int, error) {
	return w.Get().Write(b)
}

// Close implements the interface io.Closer.
func (w *SwitchWriter) Close() error {
	return Close(w.w)
}

// Sync calls the Sync method if the inner writer has implemented the interface
// { Sync() error }. Or, do nothing and return nil.
func (w *SwitchWriter) Sync() error {
	if ws, ok := w.Get().(interface{ Sync() error }); ok {
		return ws.Sync()
	}
	return nil
}

// Run executes the function f with the inner writer.
func (w *SwitchWriter) Run(f func(io.Writer)) {
	f(w.Get())
}

// Get returns the wrapped writer.
func (w *SwitchWriter) Get() io.Writer {
	return w.w.Load()
}

// Swap swaps the old writer with the new writer.
func (w *SwitchWriter) Swap(new io.Writer) (old io.Writer) {
	if new == nil {
		panic("SwitchWriter.Swap: io.Writer is nil")
	}
	return w.w.Swap(new)
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
