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
	"sync"
)

// Builder is used to build a new middleware with the middleware config.
type Builder func(name string, config map[string]interface{}) (Middleware, error)

// BuilderManager is used to manage the middleware builder.
type BuilderManager struct{ builders sync.Map }

// NewBuilderManager returns a new middleware builder manager.
func NewBuilderManager() *BuilderManager { return &BuilderManager{} }

// RegisterBuilder registers a new middleware builder typed typ.
func (m *BuilderManager) RegisterBuilder(typ string, builder Builder) (err error) {
	if typ == "" {
		panic("the middleware builder type is emtpy")
	} else if builder == nil {
		panic("the middleware builder is nil")
	}

	if _, loaded := m.builders.LoadOrStore(typ, builder); loaded {
		err = fmt.Errorf("the middleware builder typed '%s' has existed", typ)
	}
	return
}

// UnregisterBuilder unregisters the middleware builder by the type.
func (m *BuilderManager) UnregisterBuilder(typ string) {
	if typ == "" {
		panic("the middleware builder type is emtpy")
	}
	m.builders.Delete(typ)
}

// GetBuilder returns the middleware builder by the type.
//
// If the middleware builder does not exist, return nil.
func (m *BuilderManager) GetBuilder(typ string) Builder {
	if value, ok := m.builders.Load(typ); ok {
		return value.(Builder)
	}
	return nil
}

// GetBuilders returns all the middleware builders.
func (m *BuilderManager) GetBuilders() map[string]Builder {
	builders := make(map[string]Builder, 32)
	m.builders.Range(func(key, value interface{}) bool {
		builders[key.(string)] = value.(Builder)
		return true
	})
	return builders
}

// Build uses the builder typed typ to build a middleware named name
// with the config.
func (m *BuilderManager) Build(typ, name string, config map[string]interface{}) (Middleware, error) {
	if builder := m.GetBuilder(typ); builder != nil {
		return builder(name, config)
	}
	return nil, fmt.Errorf("no the middleware builder typed '%s'", typ)
}
