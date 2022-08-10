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

// Package formatter is used to format the name of the struct field.
package formatter

import (
	"reflect"
	"strings"
)

// DefaultNameFormatter is the default struct field name formatter.
var DefaultNameFormatter = NewNameFormatterByTags("json", "query", "header")

// NameFormatter is a function to format the struct field name.
type NameFormatter func(reflect.StructField) string

// NewNameFormatterByTags returns a struct field name formatter,
// which looks up the first found not-empty tag value and return the original
// field name if no the tags.
func NewNameFormatterByTags(tags ...string) NameFormatter {
	return func(sf reflect.StructField) string {
		return FormatNameByTags(sf, tags...)
	}
}

// FormatNameByTags looks up the first found not-empty tag value
// and return the original field name if no the tags.
func FormatNameByTags(sf reflect.StructField, tags ...string) string {
	for _, tag := range tags {
		if v := strings.TrimSpace(sf.Tag.Get(tag)); v != "" {
			return v
		}
	}
	return sf.Name
}
