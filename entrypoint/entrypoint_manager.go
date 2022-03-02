// Copyright 2021 xgfone
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

package entrypoint

import (
	"fmt"
	"sync"
)

// DefaultManager is the default global entrypoint manager.
var DefaultManager = NewManager()

// AddEntryPoint is equal to DefaultManager.AddEntryPoint(ep).
func AddEntryPoint(ep *EntryPoint) error {
	return DefaultManager.AddEntryPoint(ep)
}

// DelEntryPoint is equal to DefaultManager.DelEntryPoint(name).
func DelEntryPoint(name string) *EntryPoint {
	return DefaultManager.DelEntryPoint(name)
}

// GetEntryPoint is equal to DefaultManager.GetEntryPoint(name).
func GetEntryPoint(name string) *EntryPoint {
	return DefaultManager.GetEntryPoint(name)
}

// GetEntryPoints is equal to DefaultManager.GetEntryPoints().
func GetEntryPoints() []*EntryPoint {
	return DefaultManager.GetEntryPoints()
}

// Manager manages a group of entrypoints.
type Manager struct {
	lock sync.RWMutex
	eps  map[string]*EntryPoint
}

// NewManager returns a new entrypoint manager.
func NewManager() *Manager {
	return &Manager{eps: make(map[string]*EntryPoint, 8)}
}

// AddEntryPoint adds the entrypoint.
func (m *Manager) AddEntryPoint(ep *EntryPoint) (err error) {
	if ep.Name == "" {
		return fmt.Errorf("the entrypoint name is empty")
	}

	m.lock.Lock()
	if _, ok := m.eps[ep.Name]; ok {
		err = fmt.Errorf("the entrypoint named '%s' has existed", ep.Name)
	} else {
		m.eps[ep.Name] = ep
	}
	m.lock.Unlock()

	return
}

// DelEntryPoint deletes and returns the entrypoint by the name.
//
// If the entrypoint does not exist, do nothing and return nil.
func (m *Manager) DelEntryPoint(name string) *EntryPoint {
	m.lock.Lock()
	ep, ok := m.eps[name]
	if ok {
		delete(m.eps, name)
	}
	m.lock.Unlock()
	return ep
}

// GetEntryPoint returns the entrypoint by the name.
//
// If the entrypoint does not exist, return nil.
func (m *Manager) GetEntryPoint(name string) *EntryPoint {
	m.lock.RLock()
	ep := m.eps[name]
	m.lock.RUnlock()
	return ep
}

// GetEntryPoints returns all the entrypoints.
func (m *Manager) GetEntryPoints() []*EntryPoint {
	m.lock.RLock()
	eps := make([]*EntryPoint, 0, len(m.eps))
	for _, ep := range m.eps {
		eps = append(eps, ep)
	}
	m.lock.RUnlock()
	return eps
}
