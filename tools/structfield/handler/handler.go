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

// Package handler provides a handler interface and some implementations.
package handler

import "reflect"

// Handler is an interface to handle the struct field.
type Handler interface {
	Parse(string) (interface{}, error) // Used to optimize: pre-parse and cache the tag value.
	Run(ctx interface{}, fieldType reflect.StructField, fieldValue reflect.Value, arg interface{}) error
}

// HandlerFunc is a handler function.
type HandlerFunc func(interface{}, reflect.StructField, reflect.Value, interface{}) error

// Parse implements the interface Handler, which does nothing
// and returns the original string input as the parsed result.
func (f HandlerFunc) Parse(s string) (interface{}, error) { return s, nil }

// Run implements the interface Handler.
func (f HandlerFunc) Run(ctx interface{}, t reflect.StructField, v reflect.Value, arg interface{}) error {
	return f(ctx, t, v, arg)
}

// NewHandler returns a new Handler from the parse and run functions.
func NewHandler(parse func(string) (interface{}, error), run HandlerFunc) Handler {
	return handler{parse: parse, run: run}
}

type handler struct {
	parse func(string) (interface{}, error)
	run   func(interface{}, reflect.StructField, reflect.Value, interface{}) error
}

func (h handler) Parse(s string) (interface{}, error) { return h.parse(s) }
func (h handler) Run(c interface{}, t reflect.StructField, v reflect.Value, a interface{}) error {
	return h.run(c, t, v, a)
}
