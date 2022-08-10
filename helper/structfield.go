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

package helper

import (
	"reflect"
	"strings"
)

var zeroReflectValue reflect.Value

// GetStructFieldByName returns the field value by the name.
//
// fieldName maybe starts with ".".
func GetStructFieldByName(structValue interface{}, fieldName string) (fieldValue reflect.Value, ok bool) {
	fieldName = strings.TrimPrefix(fieldName, ".")
	if fieldName == "" {
		return
	}

	var v reflect.Value
	switch sv := structValue.(type) {
	case nil:
		return

	case reflect.Value:
		v = sv

	default:
		v = reflect.ValueOf(structValue)
	}

	switch v.Kind() {
	case reflect.Struct:
	case reflect.Ptr:
		if v.IsNil() {
			return
		}

		if v = v.Elem(); v.Kind() != reflect.Struct {
			return
		}
	default:
		return
	}

	for len(fieldName) > 0 {
		name := fieldName
		index := strings.IndexByte(fieldName, '.')
		if index < 0 {
			fieldName = ""
		} else {
			name = fieldName[:index]
			fieldName = fieldName[index+1:]
		}

		v = v.FieldByName(name)
		if v == zeroReflectValue {
			return reflect.Value{}, false
		}
	}

	return v, true
}
