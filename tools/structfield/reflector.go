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

// Package structfield provides a common policy to call a handler dynamically
// by the struct field tag.
package structfield

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/xgfone/go-apiserver/tools/structfield/handler"
)

func init() {
	Register("validate", handler.NewValidatorHandler(nil))
	Register("default", handler.NewDefaultHandler())
	Register("set", handler.NewSetterHandler(nil, nil))
}

// DefaultReflector is the default global struct field reflector.
var DefaultReflector = NewReflector()

// Register is equal to DefaultReflector.Register(name, handler).
func Register(name string, handler handler.Handler) {
	DefaultReflector.Register(name, handler)
}

// RegisterFunc is equal to DefaultReflector.RegisterFunc(name, handler).
func RegisterFunc(name string, handler handler.Runner) {
	DefaultReflector.RegisterFunc(name, handler)
}

// RegisterSimpleFunc is equal to DefaultReflector.RegisterSimpleFunc(name, handler).
func RegisterSimpleFunc(name string, handler func(reflect.Value, interface{}) error) {
	DefaultReflector.RegisterSimpleFunc(name, handler)
}

// Unregister is equal to DefaultReflector.Unregister(name).
func Unregister(name string) {
	DefaultReflector.Unregister(name)
}

// Reflect is equal to DefaultReflector.Reflect(ctx, structValuePtr).
func Reflect(ctx, structValuePtr interface{}) error {
	return DefaultReflector.Reflect(ctx, structValuePtr)
}

// ReflectValue is equal to DefaultReflector.ReflectValue(ctx, structValue).
func ReflectValue(ctx interface{}, structValue reflect.Value) error {
	return DefaultReflector.ReflectValue(ctx, structValue)
}

type tagKey struct {
	Name  string
	Value string
}

type tagValue struct {
	Value string
	Arg   interface{}
}

// Reflector is used to reflect the tags of the fields of the struct
// and call the field handler by the tag name with the tag value.
type Reflector struct {
	handlers map[string]handler.Handler

	tagCache  atomic.Value
	cacheMap  map[tagKey]tagValue
	cacheLock sync.Mutex
}

// NewReflector returns a new Reflector.
func NewReflector() *Reflector {
	r := &Reflector{
		handlers: make(map[string]handler.Handler, 8),
		cacheMap: make(map[tagKey]tagValue, 32),
	}
	r.updateTags()
	return r
}

// Register registers the field handler with the tag name.
func (r *Reflector) Register(name string, handler handler.Handler) {
	r.handlers[name] = handler
}

// RegisterFunc is equal to r.Register(name, handler).
func (r *Reflector) RegisterFunc(name string, handler handler.Runner) {
	r.Register(name, handler)
}

// RegisterSimpleFunc is the simplified RegisterFunc,
// which only cares about the field value and the tag value.
func (r *Reflector) RegisterSimpleFunc(name string, handler func(reflect.Value, interface{}) error) {
	r.RegisterFunc(name, func(_ interface{}, _, v reflect.Value, _ reflect.StructField, a interface{}) error {
		return handler(v, a)
	})
}

// Unregister unregisters the field handler by the tag name.
func (r *Reflector) Unregister(name string) {
	delete(r.handlers, name)
}

// Reflect reflects all the fields of the struct.
//
// If the field is a struct or slice/array of structs,
// and has a tag named "reflect" with the value "-",
// it stops to reflect the struct field recursively.
func (r *Reflector) Reflect(ctx, structValuePtr interface{}) error {
	if structValuePtr == nil {
		return nil
	}

	v := reflect.ValueOf(structValuePtr)
	switch kind := v.Kind(); kind {
	case reflect.Struct:
	case reflect.Pointer:
		if v.IsNil() {
			return nil
		}

		v = v.Elem()
		if v.Kind() != reflect.Struct {
			return fmt.Errorf("the value %T is not a pointer to struct", structValuePtr)
		}
	default:
		return fmt.Errorf("the value %T is not a struct", structValuePtr)
	}

	return r.ReflectValue(ctx, v)
}

// ReflectValue is the same as Reflect, but uses reflect.Value
// instead of a pointer to a struct.
func (r *Reflector) ReflectValue(ctx interface{}, structValue reflect.Value) error {
	return r.reflectStruct(ctx, structValue, structValue)
}

func (r *Reflector) reflectStruct(ctx interface{}, root, v reflect.Value) (err error) {
	t := v.Type()
	for i, _len := 0, v.NumField(); i < _len; i++ {
		if err = r.reflectField(ctx, root, v.Field(i), t.Field(i)); err != nil {
			return err
		}
	}
	return
}

func (r *Reflector) reflectField(ctx interface{}, root, v reflect.Value, t reflect.StructField) (err error) {
	if !t.IsExported() {
		return
	}

	stop, err := r.walkTag(ctx, root, v, t, string(t.Tag))
	if err == nil && !stop {
		switch v.Kind() {
		case reflect.Struct:
			err = r.reflectStruct(ctx, root, v)

		case reflect.Pointer:
			if !v.IsNil() {
				if v = v.Elem(); v.Kind() == reflect.Struct {
					err = r.reflectStruct(ctx, root, v)
				}
			}

		case reflect.Array, reflect.Slice:
			for i, _len := 0, v.Len(); i < _len; i++ {
				if vf := v.Index(i); vf.Kind() == reflect.Struct {
					if err = r.reflectStruct(ctx, root, vf); err != nil {
						break
					}
				}
			}
		}
	}

	return
}

func (r *Reflector) updateTags() {
	tags := make(map[tagKey]tagValue, len(r.cacheMap))
	for key, value := range r.cacheMap {
		tags[key] = value
	}
	r.tagCache.Store(tags)
}

func (r *Reflector) loadTags(key tagKey) (value tagValue, ok bool) {
	value, ok = r.tagCache.Load().(map[tagKey]tagValue)[key]
	return
}

func (r *Reflector) getTagArg(handler handler.Handler, name, qvalue string) tagValue {
	key := tagKey{Name: name, Value: qvalue}
	if tvalue, ok := r.loadTags(key); ok {
		return tvalue
	}

	r.cacheLock.Lock()
	defer r.cacheLock.Unlock()

	if tvalue, ok := r.loadTags(key); ok {
		return tvalue
	}

	value, err := strconv.Unquote(qvalue)
	if err != nil {
		panic(fmt.Errorf("invalid tag '%s' value: %s", name, err))
	}

	arg, err := handler.Parse(value)
	if err != nil {
		panic(fmt.Errorf("invalid tag '%s' value '%s': %s", name, value, err))
	}

	tvalue := tagValue{Value: qvalue, Arg: arg}
	r.cacheMap[key] = tvalue
	r.updateTags()

	return tvalue
}

func unquote(s string) string {
	if _s, err := strconv.Unquote(s); err == nil {
		return strings.TrimSpace(_s)
	}
	return s
}

func (r *Reflector) do(ctx interface{}, root, v reflect.Value, t reflect.StructField, name, value string, stop *bool) (err error) {
	if name == "reflect" && (value == `"-"` || unquote(value) == "-") {
		*stop = true
		return
	}

	if h, ok := r.handlers[name]; ok {
		err = h.Run(ctx, root, v, t, r.getTagArg(h, name, value).Arg)
	}

	return
}

// copy and modify from https://github.com/golang/go/blob/go1.18.4/src/reflect/type.go
func (r *Reflector) walkTag(ctx interface{}, root, v reflect.Value, t reflect.StructField, tag string) (stop bool, err error) {
	for tag != "" {
		// Skip leading space.
		i := 0
		for i < len(tag) && tag[i] == ' ' {
			i++
		}
		tag = tag[i:]
		if tag == "" {
			break
		}

		// Scan to colon. A space, a quote or a control character is a syntax error.
		// Strictly speaking, control chars include the range [0x7f, 0x9f], not just
		// [0x00, 0x1f], but in practice, we ignore the multi-byte control characters
		// as it is simpler to inspect the tag's bytes than the tag's runes.
		i = 0
		for i < len(tag) && tag[i] > ' ' && tag[i] != ':' && tag[i] != '"' && tag[i] != 0x7f {
			i++
		}
		if i == 0 || i+1 >= len(tag) || tag[i] != ':' || tag[i+1] != '"' {
			break
		}
		name := string(tag[:i])
		tag = tag[i+1:]

		// Scan quoted string to find value.
		i = 1
		for i < len(tag) && tag[i] != '"' {
			if tag[i] == '\\' {
				i++
			}
			i++
		}
		if i >= len(tag) {
			break
		}
		qvalue := string(tag[:i+1])
		tag = tag[i+1:]

		// (xgfone): Poll the key-value tag.
		if err = r.do(ctx, root, v, t, name, qvalue, &stop); err != nil {
			break
		}
	}

	return
}
