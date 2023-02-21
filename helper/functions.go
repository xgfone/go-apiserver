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
	"time"
)

// Now is used to customize the time Now.
var Now = time.Now

// ScannerFunc is a scanner function.
type ScannerFunc func(src interface{}) error

// Scan implements the interface sql.Scanner.
func (f ScannerFunc) Scan(src interface{}) error { return f(src) }

// Indirect returns the underlying value of the pointer or interface
// if the input value is a pointer or interface. Or, return the input.
//
// Return nil if the input value is a pointer(nil), or interface(nil).
func Indirect(value interface{}) interface{} {
	if value == nil {
		return nil
	}

	switch vf := reflect.ValueOf(value); vf.Kind() {
	case reflect.Pointer, reflect.Interface:
		if vf.IsNil() {
			return nil
		}
		return Indirect(vf.Elem().Interface())

	default:
		return value
	}
}

// MapKeys returns all the keys of the map.
func MapKeys[T ~map[K]V, K comparable, V any](maps T) []K {
	keys := make([]K, 0, len(maps))
	for k := range maps {
		keys = append(keys, k)
	}
	return keys
}

// CloneMap clones the map and returns the new.
func CloneMap[T ~map[K]V, K comparable, V any](maps T) T {
	newmap := make(T, len(maps))
	for k, v := range maps {
		newmap[k] = v
	}
	return newmap
}

// CloneSlice clones the slice and returns the new.
func CloneSlice[T ~[]E, E any](slice T) T {
	newslice := make(T, len(slice))
	copy(newslice, slice)
	return newslice
}
