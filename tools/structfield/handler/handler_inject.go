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

package handler

import (
	"fmt"
	"reflect"

	"github.com/xgfone/go-apiserver/helper"
)

// Injector is an interface to inject something into itself.
type Injector interface {
	Inject(interface{}) error
}

// NewInjectHandler returns a handler to inject something to the struct field
// value that must have implemented the interface Injector by the tag,
// which is registered into DefaultReflector with the tag name "inject"
// and the nil pre-parser by default.
//
// If preParser is nil, do nothing and return the original tag value,
// which is a string if called by structfield.Reflector.
func NewInjectHandler(preParser func(string) (interface{}, error)) Handler {
	if preParser == nil {
		preParser = func(s string) (interface{}, error) { return s, nil }
	}
	return injector{parser: preParser}
}

type injector struct {
	parser func(string) (interface{}, error)
}

func (h injector) Parse(s string) (interface{}, error) { return h.parser(s) }
func (h injector) Run(c interface{}, r, v reflect.Value, t reflect.StructField, a interface{}) error {
	if v = helper.FillNilPtr(v); v.Kind() != reflect.Pointer {
		v = v.Addr()
	}

	if helper.Implements(v, (*Injector)(nil)) {
		return v.Interface().(Injector).Inject(a)
	}

	panic(fmt.Errorf("%T has not implemented the interface Injector", v.Interface()))
}
