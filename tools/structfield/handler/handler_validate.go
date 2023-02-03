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

	"github.com/xgfone/go-apiserver/tools/structfield/formatter"
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
func (h validator) Run(c interface{}, r, v reflect.Value, t reflect.StructField, a interface{}) error {
	builder := h.Builder
	if builder == nil {
		builder = validation.DefaultBuilder
	}

	err := builder.Validate(r, v.Interface(), a.(string))
	if err != nil {
		err = fmt.Errorf("%s: %w", getStructFieldName(builder, t), err)
	}
	return err
}

func getStructFieldName(b *validation.Builder, ft reflect.StructField) (name string) {
	name = formatter.DefaultNameFormatter(ft)
	if name == "" {
		name = ft.Name
	}
	return
}
