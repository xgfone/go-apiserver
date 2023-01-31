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
	"strings"
	"time"

	"github.com/xgfone/go-apiserver/helper"
)

// NewSetDefaultHandler returns a handler to set the default value
// of the struct field, which is registered into DefaultReflector
// with the tag name "default" by default.
//
// If the field type is string or int64, and the tag value is like "now()"
// or "now(layout)", set the default value of the field to the current time by helper.Now().
func NewSetDefaultHandler() Handler { return setdefault{} }

type setdefault struct{}

func (h setdefault) Parse(s string) (interface{}, error) { return s, nil }
func (h setdefault) Run(c interface{}, r, v reflect.Value, t reflect.StructField, a interface{}) error {
	if !v.CanSet() {
		return fmt.Errorf("the field '%s' cannnot be set", t.Name)
	}

	var p reflect.Value
	if helper.IsPointer(v) {
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		p, v = v, v.Elem()
	} else {
		p = v.Addr()
	}

	if !v.IsZero() {
		return nil
	}

	s := a.(string)
	switch v.Kind() {
	case reflect.String:
		if strings.HasPrefix(s, "now(") && strings.HasSuffix(s, ")") {
			if layout := s[4 : len(s)-1]; layout == "" {
				s = helper.Now().Format(time.RFC3339)
			} else {
				s = helper.Now().Format(layout)
			}
		}
		v.SetString(s)

	case reflect.Int64:
		if strings.HasPrefix(s, "now(") && strings.HasSuffix(s, ")") {
			return helper.Set(p.Interface(), helper.Now().Unix())
		}
		return helper.Set(p.Interface(), s)

	case reflect.Bool, reflect.Float32, reflect.Float64,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return helper.Set(p.Interface(), s)

	default:
		if _, ok := v.Interface().(time.Time); ok {
			return helper.Set(p.Interface(), s)
		}
		if i, ok := p.Interface().(helper.DefaultSetter); ok {
			return i.SetDefault(s)
		}
		return fmt.Errorf("%s: unsupported type %T", t.Name, v.Interface())
	}

	return nil
}
