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

// Package binder provides a common binder, for example, bind a struct to a map.
package binder

import (
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/xgfone/go-cast"
	"github.com/xgfone/go-structs/field"
)

// Unmarshaler is an interface to unmarshal itself from the parameter.
type Unmarshaler interface {
	UnmarshalBind(interface{}) error
}

// BindStructFromMap binds the struct to map[string]interface{}.
//
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
//   - time.Time
//   - time.Duration
//
// And any pointer to the types above, and the interface BindUnmarshaler.
func BindStructFromMap(structptr interface{}, tag string, data map[string]interface{}) error {
	return bindStruct(structptr, data, func(sf reflect.StructField) (name, arg string) {
		switch name, arg = field.GetTag(sf, tag); name {
		case "":
			name = sf.Name
		case "-":
			name = ""
		default:
		}
		return
	})
}

func bindStruct(dstptr, src interface{}, getTag func(reflect.StructField) (name, arg string)) (err error) {
	dstValue := reflect.ValueOf(dstptr)
	if dstValue.Kind() != reflect.Pointer {
		return fmt.Errorf("%T must be a pointer to struct", dstptr)
	} else if dstValue = dstValue.Elem(); dstValue.Kind() != reflect.Struct {
		return fmt.Errorf("%T must be a pointer to struct", dstptr)
	}

	return binder{getTag: getTag}.bind(dstValue.Kind(), dstValue, src)
}

type binder struct {
	getTag func(reflect.StructField) (name, arg string)
}

func (b binder) bind(kind reflect.Kind, value reflect.Value, src interface{}) (err error) {
	if src == nil {
		return
	}
	if !value.CanSet() {
		switch kind {
		case reflect.Pointer, reflect.Interface:
			if !value.Elem().CanAddr() {
				return
			}
		default:
			return
		}
	}

	ptrvalue := value
	if kind != reflect.Pointer {
		ptrvalue = value.Addr()
	}
	if unmarshaler, ok := ptrvalue.Interface().(Unmarshaler); ok {
		return unmarshaler.UnmarshalBind(src)
	}

	switch kind {
	case reflect.Bool:
		err = b.bindBool(value, src)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32:
		err = b.bindInt(value, src)
	case reflect.Int64:
		err = b.bindInt64(value, src)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		err = b.bindUint(value, src)
	case reflect.Float32, reflect.Float64:
		err = b.bindFloat(value, src)
	case reflect.String:
		err = b.bindString(value, src)
	case reflect.Pointer:
		err = b.bindPointer(value, src)
	case reflect.Interface:
		err = b.bindInterface(value, src)
	case reflect.Struct:
		err = b.bindStruct(value, src)
	case reflect.Array:
		err = b.bindArray(value, src)
	case reflect.Slice:
		err = b.bindSlice(value, src)
	case reflect.Map:
		err = b.bindMap(value, src)

	// case reflect.Chan:
	// case reflect.Func:
	// case reflect.Complex64:
	// case reflect.Complex128:
	// case reflect.UnsafePointer:
	default:
		err = fmt.Errorf("unsupport to bind %T to a value", value.Interface())
	}

	return
}

func (b binder) bindBool(dstValue reflect.Value, src interface{}) (err error) {
	v, err := cast.ToBool(src)
	if err == nil {
		dstValue.SetBool(v)
	}
	return
}

func (b binder) bindInt(dstValue reflect.Value, src interface{}) (err error) {
	v, err := cast.ToInt64(src)
	if err == nil {
		dstValue.SetInt(v)
	}
	return
}

func (b binder) bindInt64(dstValue reflect.Value, src interface{}) (err error) {
	if _, ok := dstValue.Interface().(time.Duration); !ok {
		return b.bindInt(dstValue, src)
	}

	v, err := cast.ToDuration(src)
	if err == nil {
		dstValue.SetInt(int64(v))
	}
	return
}

func (b binder) bindUint(dstValue reflect.Value, src interface{}) (err error) {
	v, err := cast.ToUint64(src)
	if err == nil {
		dstValue.SetUint(v)
	}
	return
}

func (b binder) bindFloat(dstValue reflect.Value, src interface{}) (err error) {
	v, err := cast.ToFloat64(src)
	if err == nil {
		dstValue.SetFloat(v)
	}
	return
}

func (b binder) bindString(dstValue reflect.Value, src interface{}) (err error) {
	v, err := cast.ToString(src)
	if err == nil {
		dstValue.SetString(v)
	}
	return
}

func (b binder) bindPointer(dstValue reflect.Value, src interface{}) (err error) {
	if dstValue.IsNil() {
		dstValue.Set(reflect.New(dstValue.Type().Elem()))
	}
	dstValue = dstValue.Elem()
	return b.bind(dstValue.Kind(), dstValue, src)
}

func (b binder) bindInterface(dstValue reflect.Value, src interface{}) (err error) {
	if dstValue.IsValid() && dstValue.Elem().IsValid() { // Interface is set to a specific value.
		elem := dstValue.Elem()
		bindElem := elem

		// If we can't address this element, then its not writable. Instead,
		// we make a copy of the value (which is a pointer and therefore
		// writable), decode into that, and replace the whole value.
		var copied bool
		if !elem.CanAddr() {
			if elem.Kind() == reflect.Pointer {
				// (xgf) If it is a pointer and the element is addressable,
				// we should not new one, and still use the old.
				if elem.Elem().CanAddr() {
					// (xgf) We use the old pointer to check
					// whether it has implemented the interface Unmarshaler.
					bindElem = elem
				} else {
					copied = true
				}
			}
		}
		if copied {
			bindElem = reflect.New(elem.Type()) // v = new(T)
			bindElem.Elem().Set(elem)           // *v = elem
		}

		err = b.bind(bindElem.Kind(), elem, src)
		if err != nil || !copied {
			return
		}

		dstValue.Set(elem.Elem()) // elem is copied.
		return
	}

	srcValue := reflect.ValueOf(src)
	dstType := dstValue.Type()

	// If the input data is a pointer, and the assigned type is the dereference
	// of that exact pointer, then indirect it so that we can assign it.
	// Example: *string to string
	if srcValue.Kind() == reflect.Pointer && srcValue.Type().Elem() == dstType {
		srcValue = reflect.Indirect(srcValue)
	}

	if !srcValue.IsValid() {
		srcValue = reflect.Zero(dstType)
	}

	srcType := srcValue.Type()
	if !srcType.AssignableTo(dstType) {
		return fmt.Errorf("cannot assign %s to %s", srcType.String(), dstType.String())
	}

	dstValue.Set(srcValue)
	return
}

func (b binder) bindArray(dstValue reflect.Value, src interface{}) (err error) {
	return b._bindList(dstValue, src, true)
}

func (b binder) bindSlice(dstValue reflect.Value, src interface{}) (err error) {
	return b._bindList(dstValue, src, false)
}

func (b binder) _bindList(dstValue reflect.Value, src interface{}, isArray bool) (err error) {
	dstType := dstValue.Type()
	ekind := dstType.Elem().Kind()

	var _len int
	var bind func(reflect.Value, int) error
	switch vs := src.(type) {
	case []interface{}:
		_len = len(vs)
		bind = func(v reflect.Value, i int) error { return b.bind(ekind, v, vs[i]) }

	case []string:
		_len = len(vs)
		bind = func(v reflect.Value, i int) error { return b.bind(ekind, v, vs[i]) }

	default:
		srcValue := reflect.ValueOf(src)
		switch srcValue.Kind() {
		case reflect.Array, reflect.Slice:
			_len = srcValue.Len()
			bind = func(v reflect.Value, i int) error {
				return b.bind(ekind, v, srcValue.Index(i).Interface())
			}
		default:
			return errors.New("cannot bind a slice type to a non-array/slice type")
		}
	}

	elems := dstValue
	if isArray {
		dstlen := dstValue.Len()
		if dstlen == 0 {
			return
		}
		if _len < dstlen {
			_len = dstlen
		}
	} else {
		elems = reflect.MakeSlice(dstType, _len, _len)
	}

	for i := 0; i < _len; i++ {
		if err = bind(elems.Index(i), i); err != nil {
			return
		}
	}

	if !isArray {
		dstValue.Set(elems)
	}
	return
}

func (b binder) bindMap(dstValue reflect.Value, src interface{}) (err error) {
	dstType := dstValue.Type()
	keyType := dstType.Key()
	valueType := dstType.Elem()

	var dstmaps reflect.Value
	switch srcmaps := src.(type) {
	case map[string]interface{}:
		dstmaps = reflect.MakeMapWithSize(dstType, len(srcmaps))
		for key, value := range srcmaps {
			err = b._bindMapIndex(dstmaps, keyType, valueType, key, value)
			if err != nil {
				return
			}
		}

	case map[string]string:
		dstmaps = reflect.MakeMapWithSize(dstType, len(srcmaps))
		for key, value := range srcmaps {
			err = b._bindMapIndex(dstmaps, keyType, valueType, key, value)
			if err != nil {
				return
			}
		}

	default:
		srcValue := reflect.ValueOf(src)
		if srcValue.Kind() != reflect.Map {
			return errors.New("cannot bind a map type to a non-map type")
		}

		dstmaps = reflect.MakeMapWithSize(dstType, srcValue.Len())
		for iter := srcValue.MapRange(); iter.Next(); {
			key, value := iter.Key().Interface(), iter.Value().Interface()
			err = b._bindMapIndex(dstmaps, keyType, valueType, key, value)
			if err != nil {
				return
			}
		}
	}

	dstValue.Set(dstmaps)
	return
}

func (b binder) _bindMapIndex(dstmap reflect.Value, keyType, valueType reflect.Type, key, value interface{}) (err error) {
	srckey := reflect.New(keyType)
	err = b.bind(keyType.Kind(), srckey.Elem(), key)
	if err != nil {
		return
	}

	dstvalue := reflect.New(valueType)
	err = b.bind(valueType.Kind(), dstvalue.Elem(), value)
	if err != nil {
		return
	}

	dstmap.SetMapIndex(srckey.Elem(), dstvalue.Elem())
	return
}

func (b binder) bindStruct(dstStructValue reflect.Value, src interface{}) (err error) {
	if _, ok := dstStructValue.Interface().(time.Time); ok {
		var v time.Time
		if v, err = cast.ToTime(src); err == nil {
			dstStructValue.Set(reflect.ValueOf(v))
		}
		return
	}

	fields := field.GetAllFields(dstStructValue.Type())
	for index, field := range fields {
		err = b.bindField(dstStructValue.Field(index), field, src)
		if err != nil {
			return
		}
	}
	return
}

func (b binder) bindField(fieldValue reflect.Value, fieldType reflect.StructField, src interface{}) (err error) {
	if !fieldValue.CanSet() {
		return
	}

	fieldName, tagArg := b.getTag(fieldType)
	if fieldName == "" {
		return
	}

	fieldKind := fieldValue.Kind()
	if fieldKind == reflect.Struct {
		if fieldType.Anonymous || tagArg == "squash" {
			return b.bindStruct(fieldValue, src)
		}
	}

	srcValue := reflect.ValueOf(src)
	if srcValue.Kind() != reflect.Map {
		return fmt.Errorf("unsupport to bind a struct to %T", src)
	}

	value := srcValue.MapIndex(reflect.ValueOf(fieldName))
	if !value.IsValid() {
		return
	}

	return b.bind(fieldKind, fieldValue, value.Interface())
}
