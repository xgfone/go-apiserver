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
func MapValues[M ~map[K]V, K comparable, V any](maps M) []V {
	values := make([]V, 0, len(maps))
	for _, v := range maps {
		values = append(values, v)
	}
	return values
}

// MapKeys returns all the keys of the map.
func MapKeys[M ~map[K]V, K comparable, V any](maps M) []K {
	keys := make([]K, 0, len(maps))
	for k := range maps {
		keys = append(keys, k)
	}
	return keys
}

// MapKeysFunc returns all the keys of the map by the conversion function.
func MapKeysFunc[M ~map[K]V, T any, K comparable, V any](maps M, convert func(K) T) []T {
	keys := make([]T, 0, len(maps))
	for k := range maps {
		keys = append(keys, convert(k))
	}
	return keys
}

// MapValues returns all the values of the map by the conversion function.
func MapValuesFunc[M ~map[K]V, T any, K comparable, V any](maps M, convert func(V) T) []T {
	values := make([]T, 0, len(maps))
	for _, v := range maps {
		values = append(values, convert(v))
	}
	return values
}

// ToSetMap converts a slice s to a set map.
func ToSetMap[S ~[]T, T comparable](s S) map[T]struct{} {
	m := make(map[T]struct{}, len(s))
	for _, k := range s {
		m[k] = struct{}{}
	}
	return m
}

// ToBoolMap converts a slice s to a bool map.
func ToBoolMap[S ~[]T, T comparable](s S) map[T]bool {
	m := make(map[T]bool, len(s))
	for _, k := range s {
		m[k] = true
	}
	return m
}

// ToSetMapFunc converts a slice s to a set map by a conversion function.
func ToSetMapFunc[S ~[]T, K comparable, T any](s S, convert func(T) K) map[K]struct{} {
	m := make(map[K]struct{}, len(s))
	for _, k := range s {
		m[convert(k)] = struct{}{}
	}
	return m
}

// ToBoolMapFunc converts a slice s to a bool map by a conversion function.
func ToBoolMapFunc[S ~[]T, K comparable, T any](s S, convert func(T) K) map[K]bool {
	m := make(map[K]bool, len(s))
	for _, k := range s {
		m[convert(k)] = true
	}
	return m
}
