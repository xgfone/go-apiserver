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

// Package binder provides a common struct binder.
package binder

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/xgfone/go-apiserver/helper"
	"github.com/xgfone/go-apiserver/internal/structs"
)

// BindStruct is used to bind the dstStruct to srcMap.
var BindStruct func(dstStruct, srcMap interface{}, tag string) error = bindStruct

func bindStruct(dstStruct, srcMap interface{}, tag string) (err error) {
	switch ms := srcMap.(type) {
	case map[string]interface{}:
		err = BindStructFromMap(dstStruct, tag, ms)

	case map[string]string:
		err = BindStructFromStringMap(dstStruct, tag, ms)

	case map[string][]string:
		err = BindStructFromURLValues(dstStruct, tag, ms)

	case url.Values:
		err = BindStructFromURLValues(dstStruct, tag, ms)

	case http.Header:
		err = BindStructFromHTTPHeader(dstStruct, tag, ms)

	default:
		err = fmt.Errorf("binder.BindStruct: unsupport the type %T", srcMap)
	}
	return
}

// Unmarshaler is an interface to unmarshal itself from the string parameter.
type Unmarshaler interface {
	UnmarshalBind(string) error
}

// BindStructFromFunc binds the fields of the pointer struct to the values
// got by the get function that should return one of nil, string, []string
// or []*multipart.FileHeader. For []*multipart.FileHeader, it is only
//
// Notice: tag is the name of the struct tag. such as "form", "query", etc.
// If the tag value is equal to "-", ignore this field.
//
// Support the types of the struct fields as follow:
//
//   - bool
//   - int
//   - int8
//   - int16
//   - int32
//   - int64
//   - uint
//   - uint8
//   - uint16
//   - uint32
//   - uint64
//   - string
//   - float32
//   - float64
//   - time.Time     // use time.Time.UnmarshalText(), so only support RFC3339 format
//   - time.Duration // use time.ParseDuration()
//
// And any pointer to the type above, and
//
//   - BindUnmarshaler
//   - *multipart.FileHeader
//   - []*multipart.FileHeader
func BindStructFromFunc(structptr interface{}, tag string, get func(name string) interface{}) error {
	pvalue := reflect.ValueOf(structptr)
	if !helper.IsPointer(pvalue) {
		return fmt.Errorf("%T is not a pointer to struct", structptr)
	}

	vvalue := pvalue.Elem()
	if vvalue.Kind() != reflect.Struct {
		return fmt.Errorf("%T is not a pointer to struct", structptr)
	}

	return bindValues(vvalue, tag, get)
}

// BindStructFromMap is the same BindStruct, but use a map instead of a function.
func BindStructFromMap(structptr interface{}, tag string, data map[string]interface{}) error {
	return BindStructFromFunc(structptr, tag, func(name string) interface{} { return data[name] })
}

// BindStructFromStringMap is the same BindStruct, but use a string map instead of a function.
func BindStructFromStringMap(structptr interface{}, tag string, data map[string]string) error {
	return BindStructFromFunc(structptr, tag, func(name string) interface{} {
		if value, ok := data[name]; ok {
			return value
		}
		return nil
	})
}

// BindStructFromHTTPHeader is the same BindStruct, but use a map instead of a http.Header.
func BindStructFromHTTPHeader(structptr interface{}, tag string, data http.Header) error {
	return BindStructFromFunc(structptr, tag, func(name string) interface{} {
		if value, ok := data[textproto.CanonicalMIMEHeaderKey(name)]; ok {
			return value
		}
		return nil
	})
}

// BindStructFromURLValues is the same BindStruct, but use a map instead of a url.Values.
func BindStructFromURLValues(structptr interface{}, tag string, data url.Values) error {
	return BindStructFromFunc(structptr, tag, func(name string) interface{} {
		if value, ok := data[name]; ok {
			return value
		}
		return nil
	})
}

// BindStructFromMultipartFileHeaders binds the struct to the multipart form file headers.
func BindStructFromMultipartFileHeaders(structptr interface{}, tag string, fhs map[string][]*multipart.FileHeader) error {
	if len(fhs) == 0 {
		return nil
	}

	pvalue := reflect.ValueOf(structptr)
	if pvalue.Kind() != reflect.Pointer {
		return fmt.Errorf("BindStructFromMultipartFileHeaders: %T is not a pointer", structptr)
	}

	vvalue := pvalue.Elem()
	if vvalue.Kind() != reflect.Struct {
		return fmt.Errorf("BindStructFromMultipartFileHeaders: %T is not a pointer to struct", structptr)
	}

	stype := vvalue.Type()
	for name, field := range structs.GetAllFieldsWithTag(stype, tag) {
		if !field.IsExported() {
			continue
		}

		if values := fhs[name]; len(values) > 0 {
			fieldValue := vvalue.Field(field.Index)
			switch iface := fieldValue.Interface(); iface.(type) {
			case *multipart.FileHeader:
				fieldValue.Set(reflect.ValueOf(values[0]))
			case []*multipart.FileHeader:
				fieldValue.Set(reflect.ValueOf(values))
			default:
				return fmt.Errorf("BindStructFromMultipartFileHeaders: unsupport to bind %T to []*multipart.FileHeader", iface)
			}
		}
	}

	return nil
}

func bindValues(val reflect.Value, tag string, get func(string) interface{}) (err error) {
	valType := val.Type()
	for i, num := 0, valType.NumField(); i < num; i++ {
		field := valType.Field(i)
		fieldName := field.Tag.Get(tag)
		switch fieldName = strings.TrimSpace(fieldName); fieldName {
		case "":
			fieldName = field.Name
		case "-":
			continue
		}

		fieldValue := val.Field(i)
		fieldKind := fieldValue.Kind()
		if field.Anonymous && fieldKind == reflect.Struct {
			if err = bindValues(fieldValue, tag, get); err != nil {
				return err
			}
			continue
		} else if !fieldValue.CanSet() {
			continue
		}

		switch value := get(fieldName).(type) {
		case nil:
			continue

		case string:
			if fieldKind == reflect.Slice {
				kind := field.Type.Elem().Kind()
				slice := reflect.MakeSlice(field.Type, 1, 1)
				err = setWithProperType(kind, slice.Index(0), value)
				if err != nil {
					return
				}
				fieldValue.Set(slice)
			} else {
				err = setWithProperType(fieldKind, fieldValue, value)
				if err != nil {
					return
				}
			}

		case []string:
			if fieldKind == reflect.Slice {
				num := len(value)
				kind := field.Type.Elem().Kind()
				slice := reflect.MakeSlice(field.Type, num, num)
				for j := 0; j < num; j++ {
					err = setWithProperType(kind, slice.Index(j), value[j])
					if err != nil {
						return
					}
				}
				fieldValue.Set(slice)
			} else {
				err = setWithProperType(fieldKind, fieldValue, value[0])
				if err != nil {
					return
				}
			}

		case []*multipart.FileHeader:
			if len(value) == 0 {
				continue
			}

			switch fieldValue.Interface().(type) {
			case *multipart.FileHeader:
				fieldValue.Set(reflect.ValueOf(value[0]))
			case []*multipart.FileHeader:
				fieldValue.Set(reflect.ValueOf(value))
			default:
				continue
			}

		default:
			return fmt.Errorf("unsupported value type %T", value)
		}
	}

	return
}

var binderType = reflect.TypeOf((*Unmarshaler)(nil)).Elem()

func bindUnmarshaler(kind reflect.Kind, val reflect.Value, value string) (ok bool, err error) {
	if kind != reflect.Pointer && kind != reflect.Interface {
		val = val.Addr()
	}

	if unmarshaler, ok := val.Interface().(Unmarshaler); ok {
		return true, unmarshaler.UnmarshalBind(value)
	}
	return false, nil
}

func setWithProperType(kind reflect.Kind, value reflect.Value, input string) error {
	if kind == reflect.Pointer && value.IsNil() {
		value.Set(reflect.New(value.Type().Elem()))
	} else if kind == reflect.Interface && value.IsNil() {
		panic("the bind struct field interface value must not be nil")
	}

	if ok, err := bindUnmarshaler(kind, value, input); ok {
		return err
	}

	switch kind {
	case reflect.Pointer:
		value = value.Elem()
		return setWithProperType(value.Kind(), value, input)
	case reflect.Int:
		return setIntField(input, 0, value)
	case reflect.Int8:
		return setIntField(input, 8, value)
	case reflect.Int16:
		return setIntField(input, 16, value)
	case reflect.Int32:
		return setIntField(input, 32, value)
	case reflect.Int64:
		if _, ok := value.Interface().(time.Duration); ok {
			v, err := time.ParseDuration(input)
			if err == nil {
				value.SetInt(int64(v))
			}
			return err
		}
		return setIntField(input, 64, value)
	case reflect.Uint:
		return setUintField(input, 0, value)
	case reflect.Uint8:
		return setUintField(input, 8, value)
	case reflect.Uint16:
		return setUintField(input, 16, value)
	case reflect.Uint32:
		return setUintField(input, 32, value)
	case reflect.Uint64:
		return setUintField(input, 64, value)
	case reflect.Bool:
		return setBoolField(input, value)
	case reflect.Float32:
		return setFloatField(input, 32, value)
	case reflect.Float64:
		return setFloatField(input, 64, value)
	case reflect.String:
		value.SetString(input)
	default:
		if _, ok := value.Interface().(time.Time); ok {
			if input == "" {
				return nil
			}
			return value.Addr().Interface().(*time.Time).UnmarshalText([]byte(input))
		}
		return fmt.Errorf("unknown field type '%T'", value.Interface())
	}
	return nil
}

func setIntField(value string, bitSize int, field reflect.Value) error {
	if value == "" {
		return nil
	}

	intVal, err := strconv.ParseInt(value, 10, bitSize)
	if err == nil {
		field.SetInt(intVal)
	}
	return err
}

func setUintField(value string, bitSize int, field reflect.Value) error {
	if value == "" {
		return nil
	}

	uintVal, err := strconv.ParseUint(value, 10, bitSize)
	if err == nil {
		field.SetUint(uintVal)
	}
	return err
}

func setBoolField(value string, field reflect.Value) error {
	if value == "" {
		return nil
	}

	boolVal, err := strconv.ParseBool(value)
	if err == nil {
		field.SetBool(boolVal)
	}
	return err
}

func setFloatField(value string, bitSize int, field reflect.Value) error {
	if value == "" {
		return nil
	}

	floatVal, err := strconv.ParseFloat(value, bitSize)
	if err == nil {
		field.SetFloat(floatVal)
	}
	return err
}
