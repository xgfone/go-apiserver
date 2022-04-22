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
	"encoding/json"
	"fmt"
	"unicode/utf8"

	"github.com/xgfone/go-apiserver/helper"
	"github.com/xgfone/go-apiserver/validation"
)

// CountString is used to count the number of the characters in the string.
var CountString func(string) int = utf8.RuneCountInString

// OneOf returns a new Validator to chech whether the string value is one
// of the given strings.
func OneOf(values ...string) validation.Validator {
	if len(values) == 0 {
		panic("OneOf: the values must be empty")
	}

	bs, err := json.Marshal(values)
	if err != nil {
		panic(err)
	}

	desc := fmt.Sprintf("oneof(%s)", string(bs[1:len(bs)-1]))
	return validation.NewValidator(desc, func(i interface{}) error {
		switch v := helper.Indirect(i).(type) {
		case string:
			if !helper.InStrings(v, values) {
				return fmt.Errorf("the string '%s' is not one of %v", v, values)
			}

		case fmt.Stringer:
			if s := v.String(); !helper.InStrings(s, values) {
				return fmt.Errorf("the string '%s' is not one of %v", s, values)
			}

		default:
			return fmt.Errorf("expect a string, but got %T", i)
		}

		return nil
	})
}
