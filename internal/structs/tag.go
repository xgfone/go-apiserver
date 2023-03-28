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

package structs

import (
	"reflect"
	"strings"
)

// GetFieldTag returns the tag information from the struct field.
//
// If the tag value contains "-", return ("", "", ""). Or, fieldName is not empty.
// If tagName is not empty, fieldName is equal to tagName. Or, it is equal to sf.Name.
func GetFieldTag(sf reflect.StructField, tag string) (fieldName, tagName, tagArg string) {
	fieldName = sf.Name
	if tag == "" {
		return
	}

	tagName = sf.Tag.Get(tag)
	if index := strings.IndexByte(tagName, ','); index > -1 {
		tagArg = strings.TrimSpace(tagName[index+1:])
		tagName = strings.TrimSpace(tagName[:index])
	}

	switch tagName {
	case "":
	case "-":
		fieldName = ""
		tagName = ""
		tagArg = ""
	default:
		fieldName = tagName
	}

	return
}
