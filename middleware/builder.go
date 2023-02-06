// Copyright 2022~2023 xgfone
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

package middleware

import "fmt"

var builders = make(map[string]Builder, 16)

// Builder is used to build a new middleware with the middleware config.
type Builder func(name string, config map[string]interface{}) (Middleware, error)

// RegisterBuilder registers a new middleware builder named name.
func RegisterBuilder(name string, builder Builder) (err error) {
	if name == "" {
		panic("the middleware builder name is emtpy")
	} else if builder == nil {
		panic("the middleware builder is nil")
	}

	if _, ok := builders[name]; ok {
		err = fmt.Errorf("the middleware builder named '%s' has existed", name)
	} else {
		builders[name] = builder
	}

	return
}

// UnregisterBuilder unregisters the middleware builder by the name.
func UnregisterBuilder(name string) { delete(builders, name) }

// GetBuilder returns the middleware builder by the name.
//
// If the middleware builder does not exist, return nil.
func GetBuilder(name string) Builder { return builders[name] }

// GetBuilderNames returns the names of all the middleware builders.
func GetBuilderNames() (names []string) {
	names = make([]string, 0, len(builders))
	for name := range builders {
		names = append(names, name)
	}
	return
}

// Build uses the builder named name to build a middleware named name
// with the config.
func Build(name string, config map[string]interface{}) (Middleware, error) {
	if builder := GetBuilder(name); builder != nil {
		return builder(name, config)
	}
	return nil, fmt.Errorf("no the middleware builder named '%s'", name)
}
