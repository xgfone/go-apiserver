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

package handler

import (
	"fmt"
	"reflect"

	"github.com/xgfone/go-apiserver/helper"
	"github.com/xgfone/go-apiserver/tools/setter"
)

// NewSetterHandler returns a handler to set the struct field to something
// by the set function, which is registered into DefaultReflector
// with the tag name "set" and the nil parser and set function by default.
//
// If parser is nil, use the default, which returns the input directly.
// If set is nil, use the default, which will assert the struct field
// to the interface setter.Setter.
func NewSetterHandler(parser Parser, set Runner) Handler {
	if set == nil {
		set = usesetter
	}
	if parser == nil {
		parser = func(s string) (interface{}, error) { return s, nil }
	}
	return setHandler{parser: parser, setter: set}
}

type setHandler struct {
	parser Parser
	setter Runner
}

func (h setHandler) Parse(s string) (interface{}, error) { return h.parser(s) }
func (h setHandler) Run(c interface{}, root, value reflect.Value, sf reflect.StructField, arg interface{}) error {
	if !value.CanSet() {
		return fmt.Errorf("the field '%s' cannnot be set", sf.Name)
	}

	var ptr reflect.Value
	if value = helper.FillNilPtr(value); helper.IsPointer(value) {
		ptr = value
	} else {
		ptr = value.Addr()
	}

	return h.setter(c, root, ptr, sf, arg)
}

func usesetter(_ interface{}, root, fieldptr reflect.Value, sf reflect.StructField, arg interface{}) error {
	if setter, ok := fieldptr.Interface().(setter.Setter); ok {
		return setter.Set(arg)
	}
	panic(fmt.Errorf("%s(%T) has not implemented the interface setter.Setter", sf.Name, fieldptr.Interface()))
}
