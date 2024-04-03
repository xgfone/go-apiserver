// Copyright 2024 xgfone
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

// MapValues returns all the values of the map.
//
// DEPRECATED!!!
func MapValues[M ~map[K]V, K comparable, V any](maps M) []V {
	values := make([]V, 0, len(maps))
	for _, v := range maps {
		values = append(values, v)
	}
	return values
}

// MapKeys returns all the keys of the map.
//
// DEPRECATED!!!
func MapKeys[M ~map[K]V, K comparable, V any](maps M) []K {
	keys := make([]K, 0, len(maps))
	for k := range maps {
		keys = append(keys, k)
	}
	return keys
}

// MapKeysFunc returns all the keys of the map by the conversion function.
//
// DEPRECATED!!!
func MapKeysFunc[M ~map[K]V, T any, K comparable, V any](maps M, convert func(K) T) []T {
	keys := make([]T, 0, len(maps))
	for k := range maps {
		keys = append(keys, convert(k))
	}
	return keys
}

// MapValues returns all the values of the map by the conversion function.
//
// DEPRECATED!!!
func MapValuesFunc[M ~map[K]V, T any, K comparable, V any](maps M, convert func(V) T) []T {
	values := make([]T, 0, len(maps))
	for _, v := range maps {
		values = append(values, convert(v))
	}
	return values
}

// DEPRECATED!!!
func ToMapWithIndex[S ~[]E, K comparable, V, E any](s S, convert func(int, E) (K, V)) map[K]V {
	_len := len(s)
	maps := make(map[K]V, _len)
	for i := 0; i < _len; i++ {
		k, v := convert(i, s[i])
		maps[k] = v
	}
	return maps
}

// DEPRECATED!!!
func ToMap[S ~[]E, K comparable, V, E any](s S, convert func(E) (K, V)) map[K]V {
	return ToMapWithIndex(s, func(_ int, e E) (K, V) { return convert(e) })
}

// ToSetMap converts a slice s to a set map.
//
// DEPRECATED!!!
func ToSetMap[S ~[]T, T comparable](s S) map[T]struct{} {
	return ToMap(s, func(e T) (T, struct{}) { return e, struct{}{} })
}

// ToBoolMap converts a slice s to a bool map.
//
// DEPRECATED!!!
func ToBoolMap[S ~[]T, T comparable](s S) map[T]bool {
	return ToMap(s, func(e T) (T, bool) { return e, true })
}

// ToSetMapFunc converts a slice s to a set map by a conversion function.
//
// DEPRECATED!!!
func ToSetMapFunc[S ~[]T, K comparable, T any](s S, convert func(T) K) map[K]struct{} {
	return ToMap(s, func(e T) (K, struct{}) { return convert(e), struct{}{} })
}

// ToBoolMapFunc converts a slice s to a bool map by a conversion function.
//
// DEPRECATED!!!
func ToBoolMapFunc[S ~[]T, K comparable, T any](s S, convert func(T) K) map[K]bool {
	return ToMap(s, func(e T) (K, bool) { return convert(e), true })
}
