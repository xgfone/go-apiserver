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

	"github.com/xgfone/go-apiserver/validation"
)

// NewValidatorHandler returns a handler to validate whether the field value
// is valid, which is registered into DefaultReflector
// with the tag name "validate" by default.
//
// Because the reflector walks the sub-fields of the struct slice field
// recursively, so the validation rule "array(structure)" should not be used.
//
// If builder is nil, use validation.DefaultBuilder instead.
func NewValidatorHandler(builder *validation.Builder) Handler {
	return validator{builder}
}

type validator struct {
	*validation.Builder
}

func (h validator) Parse(s string) (interface{}, error) { return s, nil }
func (h validator) Run(c interface{}, t reflect.StructField, v reflect.Value, a interface{}) error {
	builder := h.Builder
	if builder == nil {
		builder = validation.DefaultBuilder
	}

	err := builder.Validate(v.Interface(), a.(string))
	if err != nil {
		err = fmt.Errorf("%s: %w", getStructFieldName(builder, t), err)
	}
	return err
}

var lookupStructFieldName = validation.LookupStructFieldNameByTags("json", "query")

func getStructFieldName(b *validation.Builder, ft reflect.StructField) (name string) {
	if b.LookupStructFieldName == nil {
		name = lookupStructFieldName(ft)
	} else {
		name = b.LookupStructFieldName(ft)
	}

	if name == "" {
		name = ft.Name
	}

	return
}
