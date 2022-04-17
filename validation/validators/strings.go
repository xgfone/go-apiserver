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
	"fmt"

	"github.com/xgfone/go-apiserver/helper"
	"github.com/xgfone/go-apiserver/validation"
)

// OneOf returns a new Validator to chech whether the string value is one
// of the given strings.
func OneOf(values ...string) validation.Validator {
	if len(values) == 0 {
		panic("OneOf: the values must be empty")
	}

	return validation.NewValidator("oneof", func(i interface{}) error {
		if s, ok := i.(string); ok {
			if helper.InStrings(s, values) {
				return nil
			}
			return fmt.Errorf("the string '%s' is not in %v", s, values)
		}
		return fmt.Errorf("expect a string, but got %T", i)
	})
}
