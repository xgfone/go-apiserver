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
	Run(ctx interface{}, rootStructValue, fieldValue reflect.Value, fieldType reflect.StructField, arg interface{}) error
}

// Parser is the function to pre-parse the field tag value.
type Parser func(string) (interface{}, error)

// Runner is the function to handle the struct field.
type Runner func(interface{}, reflect.Value, reflect.Value, reflect.StructField, interface{}) error

// SimpleRunner converts a simple function to Runner.
func SimpleRunner(f func(field reflect.Value, arg interface{}) error) Runner {
	return func(_ interface{}, _, vf reflect.Value, _ reflect.StructField, arg interface{}) error {
		return f(vf, arg)
	}
}

// Parse implements the interface Handler, which does nothing
// and returns the original string input as the parsed result.
func (f Runner) Parse(s string) (interface{}, error) { return s, nil }

// Run implements the interface Handler.
func (f Runner) Run(ctx interface{}, r, v reflect.Value, t reflect.StructField, arg interface{}) error {
	return f(ctx, r, v, t, arg)
}

// NewHandler returns a new Handler from the parse and run functions.
func NewHandler(parse Parser, run Runner) Handler {
	return handler{parse: parse, run: run}
}

type handler struct {
	parse Parser
	run   Runner
}

func (h handler) Parse(s string) (interface{}, error) { return h.parse(s) }
func (h handler) Run(c interface{}, r, v reflect.Value, t reflect.StructField, a interface{}) error {
	return h.run(c, r, v, t, a)
}
