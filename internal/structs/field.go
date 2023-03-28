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
	"sync"
)

var (
	fieldlock  sync.Mutex
	fieldtypes sync.Map
)

type mapkey struct {
	Type reflect.Type
	Tag  string
}

// GetAllFields returns all the field types of the struct.
func GetAllFields(stype reflect.Type) []reflect.StructField {
	if stype.Kind() != reflect.Struct {
		panic("GetAllFields: reflect.Type is not a struct type")
	}

	key := mapkey{Type: stype}
	if value, ok := fieldtypes.Load(key); ok {
		return value.([]reflect.StructField)
	}

	fieldlock.Lock()
	defer fieldlock.Unlock()

	_len := stype.NumField()
	fields := make([]reflect.StructField, _len)
	for i := 0; i < _len; i++ {
		fields[i] = stype.Field(i)
	}

	fieldtypes.Store(key, fields)
	return fields
}

// Field represents the information of the struct field.
type Field struct {
	reflect.StructField
	FieldName string // If TagName is not empty, it is equal to TagName. Or, StructField.Name.
	TagName   string
	TagArg    string
	Index     int
}

// GetAllFieldsWithTag returns all the field types of the struct with the tag,
// which will filter the fields that the tag value contains "-".
func GetAllFieldsWithTag(stype reflect.Type, tag string) map[string]Field {
	if stype.Kind() != reflect.Struct {
		panic("GetAllFieldsWithTag: reflect.Type is not a struct type")
	}
	if tag == "" {
		panic("GetAllFieldsWithTag: the tag must not be empty")
	}

	key := mapkey{Type: stype, Tag: tag}
	if value, ok := fieldtypes.Load(key); ok {
		return value.(map[string]Field)
	}

	fieldlock.Lock()
	defer fieldlock.Unlock()

	_len := stype.NumField()
	fields := make(map[string]Field, _len)
	for i := 0; i < _len; i++ {
		field := Field{StructField: stype.Field(i), Index: i}
		field.FieldName, field.TagName, field.TagArg = GetFieldTag(field.StructField, tag)
		if len(field.FieldName) > 0 {
			fields[field.FieldName] = field
		}
	}

	fieldtypes.Store(key, fields)
	return fields
}

// GetFieldValueByName returns the struct field value by the name.
//
// fieldName maybe starts with ".".
func GetFieldValueByName(structValue interface{}, fieldName string) (fieldValue reflect.Value, ok bool) {
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
	case reflect.Pointer:
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
		if v == (reflect.Value{}) {
			return reflect.Value{}, false
		}
	}

	return v, true
}
