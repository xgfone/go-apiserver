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

package validators

import (
	"unicode/utf8"

	"github.com/xgfone/go-apiserver/validation"
	"github.com/xgfone/go-apiserver/validation/internal"
)

// CountString is used to count the number of the characters in the string.
var CountString func(string) int = utf8.RuneCountInString

// OneOf is equal to OneOfWithName("oneof", values...).
func OneOf(values ...string) validation.Validator {
	return OneOfWithName("oneof", values...)
}

// OneOfWithName returns a new Validator with the validator name
// to chech whether the string value is one of the given strings.
func OneOfWithName(name string, values ...string) validation.Validator {
	return internal.NewOneOf(name, values...)
}
