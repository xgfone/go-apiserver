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
