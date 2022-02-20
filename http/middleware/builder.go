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

// BuilderFunc is a function to build the http middleware.
type BuilderFunc func(name string, config map[string]interface{}) (Middleware, error)

// Builder is used to build a new middleware with the middleware config.
type Builder interface {
	Build(name string, config map[string]interface{}) (Middleware, error)
	Type() string
}

type builder struct {
	new BuilderFunc
	typ string
}

func (b builder) Type() string { return b.typ }
func (b builder) Build(n string, c map[string]interface{}) (Middleware, error) {
	return b.new(n, c)
}

// NewBuilder returns a new http middleware builder.
func NewBuilder(typ string, build BuilderFunc) Builder {
	return builder{typ: typ, new: build}
}

// BuilderManager is used to manage the middleware builder.
type BuilderManager struct {
	builders map[string]Builder
	lock     sync.RWMutex
}

// NewBuilderManager returns a new middleware builder manager.
func NewBuilderManager() *BuilderManager {
	return &BuilderManager{builders: make(map[string]Builder, 8)}
}

// AddBuilder adds the middleware builder.
func (m *BuilderManager) AddBuilder(b Builder) (err error) {
	typ := b.Type()
	m.lock.Lock()
	if _, ok := m.builders[typ]; ok {
		err = fmt.Errorf("the middleware builder typed '%s' has existed", typ)
	} else {
		m.builders[typ] = b
	}
	m.lock.Unlock()
	return
}

// DelBuilder removes and returns the middleware builder by the type.
//
// If the middleware builder does not exist, do nothing and return nil.
func (m *BuilderManager) DelBuilder(typ string) Builder {
	m.lock.Lock()
	builder, ok := m.builders[typ]
	if ok {
		delete(m.builders, typ)
	}
	m.lock.Unlock()
	return builder
}

// GetBuilder returns the middleware builder by the type.
//
// If the middleware builder does not exist, return nil.
func (m *BuilderManager) GetBuilder(typ string) Builder {
	m.lock.RLock()
	builder := m.builders[typ]
	m.lock.RUnlock()
	return builder
}

// GetBuilders returns all the middleware builders.
func (m *BuilderManager) GetBuilders() []Builder {
	m.lock.RLock()
	builders := make([]Builder, 0, len(m.builders))
	for _, builder := range m.builders {
		builders = append(builders, builder)
	}
	m.lock.RUnlock()
	return builders
}

// Build uses the builder typed typ to build a middleware named name
// with the config.
func (m *BuilderManager) Build(typ, name string, config map[string]interface{}) (Middleware, error) {
	if builder := m.GetBuilder(typ); builder != nil {
		return builder.Build(name, config)
	}
	return nil, fmt.Errorf("no the http middleware builder typed '%s'", typ)
}
