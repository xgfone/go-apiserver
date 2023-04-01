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

package binder

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"reflect"

	"github.com/xgfone/go-cast"
	"github.com/xgfone/go-structs/field"
)

// BindStructFromStringMap binds the struct to map[string]string.
//
// For the key name, it is case-sensitive.
func BindStructFromStringMap(structptr interface{}, tag string, data map[string]string) (err error) {
	if len(data) == 0 {
		return nil
	}

	pvalue := reflect.ValueOf(structptr)
	if pvalue.Kind() != reflect.Pointer {
		return fmt.Errorf("BindStructFromStringMap: %T is not a pointer", structptr)
	}

	vvalue := pvalue.Elem()
	if vvalue.Kind() != reflect.Struct {
		return fmt.Errorf("BindStructFromStringMap: %T is not a pointer to struct", structptr)
	}

	for name, field := range field.GetAllFieldsWithTag(vvalue.Type(), tag) {
		if !field.IsExported() {
			continue
		}

		if value, ok := data[name]; ok {
			if field.Type.Kind() == reflect.Slice {
				slice := reflect.MakeSlice(field.Type, 1, 1)
				err = cast.Set(slice.Index(0), value)
				if err == nil {
					vvalue.Field(field.Index).Set(slice)
				}
			} else {
				err = cast.Set(vvalue.Field(field.Index), value)
			}
		}

		if err != nil {
			return
		}
	}

	return
}

// BindStructFromHTTPHeader binds the struct to http.Header.
//
// For the key name, it will use textproto.CanonicalMIMEHeaderKey(s) to normalize it.
func BindStructFromHTTPHeader(structptr interface{}, tag string, data http.Header) error {
	if len(data) == 0 {
		return nil
	}

	return bindStructFromMapStrings(structptr, "BindStructFromHTTPHeader", tag, func(s string) []string {
		return data[textproto.CanonicalMIMEHeaderKey(s)]
	})
}

// BindStructFromURLValues binds the struct to url.Values.
//
// For the key name, it is case-sensitive.
func BindStructFromURLValues(structptr interface{}, tag string, data url.Values) error {
	if len(data) == 0 {
		return nil
	}

	return bindStructFromMapStrings(structptr, "BindStructFromURLValues", tag, func(s string) []string {
		return data[s]
	})
}

func bindStructFromMapStrings(structptr interface{}, fname, tag string, get func(string) []string) (err error) {
	pvalue := reflect.ValueOf(structptr)
	if pvalue.Kind() != reflect.Pointer {
		return fmt.Errorf("%s: %T is not a pointer", fname, structptr)
	}

	vvalue := pvalue.Elem()
	if vvalue.Kind() != reflect.Struct {
		return fmt.Errorf("%s: %T is not a pointer to struct", fname, structptr)
	}

	for name, field := range field.GetAllFieldsWithTag(vvalue.Type(), tag) {
		if !field.IsExported() {
			continue
		}

		if values := get(name); len(values) > 0 {
			if field.Type.Kind() == reflect.Slice {
				num := len(values)
				slice := reflect.MakeSlice(field.Type, num, num)
				for j := 0; j < num; j++ {
					err = cast.Set(slice.Index(j), values[j])
					if err != nil {
						return
					}
				}
				vvalue.Field(field.Index).Set(slice)
			} else {
				err = cast.Set(vvalue.Field(field.Index), values[0])
				if err != nil {
					return
				}
			}
		}
	}

	return
}

// BindStructFromMultipartFileHeaders binds the struct to the multipart form file headers.
//
// For the key name, it is case-sensitive.
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

	for name, field := range field.GetAllFieldsWithTag(vvalue.Type(), tag) {
		if !field.IsExported() {
			continue
		}

		if values := fhs[name]; len(values) > 0 {
			switch {
			case field.Type == multipartFileHeader:
				vvalue.Field(field.Index).Set(reflect.ValueOf(values[0]))
			case field.Type == multipartFileHeaders:
				vvalue.Field(field.Index).Set(reflect.ValueOf(values))
			default:
				return fmt.Errorf("BindStructFromMultipartFileHeaders: unsupport to bind %s to []*multipart.FileHeader", field.Type.String())
			}
		}
	}

	return nil
}

var (
	multipartFileHeader  = reflect.TypeOf((*multipart.FileHeader)(nil))
	multipartFileHeaders = reflect.TypeOf(([]*multipart.FileHeader)(nil))
)
