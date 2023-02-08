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
	"github.com/xgfone/go-apiserver/tools/setter"
)

// NewSetDefaultHandler returns a handler to set the default value
// of the struct field if it is ZERO, which is registered into DefaultReflector
// with the tag name "default" by default.
//
// For the type of the field, it only supports some base types as follow:
//
//	bool
//	string
//	float32
//	float64
//	int
//	int8
//	int16
//	int32
//	int64
//	uint
//	uint8
//	uint16
//	uint32
//	uint64
//	struct
//	struct slice
//	setter.Setter
//	time.Time      // Format: A. Integer(UTC); B. String(RFC3339)
//	time.Duration  // Format: A. Integer(ms);  B. String(time.ParseDuration)
//	pointer to the types above
//
// If the field type is string or int64, and the tag value is like "now()"
// or "now(layout)", set the default value of the field to the current time
// by helper.Now(). For example,
//
//	type T struct {
//	    StartTime string `default:"now()"`
//	    EndTime   int64  `default:"now()"`
//	}
//
// Notice: If the tag value starts with ".", it represents a field name and
// the default value of current field is set to the value of that field.
// But their types must be consistent, or panic.
func NewSetDefaultHandler() Handler { return setdefault{} }

type setdefault struct{}

func (h setdefault) Parse(s string) (interface{}, error) { return s, nil }
func (h setdefault) Run(ctx interface{}, r, v reflect.Value, t reflect.StructField, a interface{}) error {
	if !v.CanSet() {
		return fmt.Errorf("the field '%s' cannnot be set", t.Name)
	}

	var p reflect.Value
	v = helper.FillNilPtr(v)
	if helper.IsPointer(v) {
		p, v = v, v.Elem()
	} else {
		p = v.Addr()
	}

	if !v.IsZero() {
		return nil
	}

	s := a.(string)
	if len(s) > 0 && s[0] == '.' {
		if s = s[1:]; s == "" {
			return fmt.Errorf("%s: invalid default value", t.Name)
		}

		fieldv, ok := helper.GetStructFieldByName(r, s)
		if !ok {
			panic(fmt.Errorf("not found the struct field '%s'", s))
		}

		if helper.IsPointer(fieldv) {
			fieldv = fieldv.Elem()
		}
		v.Set(fieldv)
		return nil
	}

	if i, ok := p.Interface().(setter.Setter); ok {
		return i.Set(s)
	}

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
			return setter.Set(p.Interface(), helper.Now().Unix())
		}
		return setter.Set(p.Interface(), s)

	case reflect.Bool, reflect.Float32, reflect.Float64,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return setter.Set(p.Interface(), s)

	case reflect.Struct:
		if _, ok := v.Interface().(time.Time); ok {
			return setter.Set(p.Interface(), s)
		}

	default:
		return fmt.Errorf("%s: unsupported type %T", t.Name, v.Interface())
	}

	return nil
}
