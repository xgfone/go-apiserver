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

package middleware

import (
	"fmt"
)

var builders map[string]Builder

// Builder is used to build a new middleware with the middleware config.
type Builder func(name string, config map[string]interface{}) (Middleware, error)

// RegisterBuilder registers a new middleware builder typed typ.
func RegisterBuilder(typ string, builder Builder) (err error) {
	if typ == "" {
		panic("the middleware builder type is emtpy")
	} else if builder == nil {
		panic("the middleware builder is nil")
	}

	if _, ok := builders[typ]; ok {
		err = fmt.Errorf("the middleware builder typed '%s' has existed", typ)
	} else {
		builders[typ] = builder
	}

	return
}

// UnregisterBuilder unregisters the middleware builder by the type.
func UnregisterBuilder(typ string) { delete(builders, typ) }

// GetBuilder returns the middleware builder by the type.
//
// If the middleware builder does not exist, return nil.
func GetBuilder(typ string) Builder { return builders[typ] }

// GetBuilderTypes returns the types of all the middleware builders.
func GetBuilderTypes() (types []string) {
	types = make([]string, 0, len(builders))
	for _type := range builders {
		types = append(types, _type)
	}
	return
}

// Build uses the builder typed typ to build a middleware named name
// with the config.
func Build(typ, name string, config map[string]interface{}) (Middleware, error) {
	if builder := GetBuilder(typ); builder != nil {
		return builder(name, config)
	}
	return nil, fmt.Errorf("no the middleware builder typed '%s'", typ)
}
