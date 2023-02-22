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

// Package maps provides some convenient map functions.
package maps

// Add adds the key-value pair into the maps if the key does not exist.
// Or, do nothing and return false.
func Add[T ~map[K]V, K comparable, V any](maps T, k K, v V) (ok bool) {
	_, exist := maps[k]
	if ok = !exist; ok {
		maps[k] = v
	}
	return
}

// Pop removes the element by the key and returns the removed value.
func Pop[T ~map[K]V, K comparable, V any](maps T, k K) (v V, ok bool) {
	if v, ok = maps[k]; ok {
		delete(maps, k)
	}
	return
}

// Delete removes the element by the key.
func Delete[T ~map[K]V, K comparable, V any](maps T, k K) (ok bool) {
	if _, ok = maps[k]; ok {
		delete(maps, k)
	}
	return
}

// Values returns all the values of the map.
func Values[T ~map[K]V, K comparable, V any](maps T) []V {
	values := make([]V, 0, len(maps))
	for _, v := range maps {
		values = append(values, v)
	}
	return values
}

// Keys returns all the keys of the map.
func Keys[T ~map[K]V, K comparable, V any](maps T) []K {
	keys := make([]K, 0, len(maps))
	for k := range maps {
		keys = append(keys, k)
	}
	return keys
}

// Clone clones the map and returns the new.
func Clone[T ~map[K]V, K comparable, V any](maps T) T {
	if maps == nil {
		return nil
	}

	newmap := make(T, len(maps))
	for k, v := range maps {
		newmap[k] = v
	}
	return newmap
}

// Convert converts the map from map[K1]V1 to map[K1]V2.
func Convert[T ~map[K1]V1, K1, K2 comparable, V1, V2 any](maps T, convert func(K1, V1) (K2, V2)) map[K2]V2 {
	if maps == nil {
		return nil
	}

	newmap := make(map[K2]V2, len(maps))
	for k1, v1 := range maps {
		k2, v2 := convert(k1, v1)
		newmap[k2] = v2
	}
	return newmap
}

// ConvertValues clones the map, converts the value from V1 to V2, and returns the new.
func ConvertValues[T ~map[K]V1, K comparable, V1, V2 any](maps T, convert func(V1) V2) map[K]V2 {
	if maps == nil {
		return nil
	}

	newmap := make(map[K]V2, len(maps))
	for k, v := range maps {
		newmap[k] = convert(v)
	}
	return newmap
}
