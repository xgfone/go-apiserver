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

// Package tlsconfig provides a manager to manage the TLS configs.
package tlsconfig

import (
	"crypto/tls"
	"fmt"
	"sync"
)

// DefaultManager is the default tls.Config manager.
var DefaultManager = NewManager()

// Updater is used to add or delete the TLS config.
type Updater interface {
	AddTLSConfig(name string, config *tls.Config)
	DelTLSConfig(name string)
}

// Manager is used to manage a set of tls.Config.
type Manager struct {
	configs  sync.Map
	updaters sync.Map
}

// NewManager returns a new TLS config manager.
func NewManager() *Manager { return &Manager{} }

// GetTLSConfigs returns all the tls configs.
func (m *Manager) GetTLSConfigs() map[string]*tls.Config {
	configs := make(map[string]*tls.Config, 32)
	m.configs.Range(func(key, value interface{}) bool {
		configs[key.(string)] = value.(*tls.Config)
		return true
	})
	return configs
}

// GetTLSConfig returns the tls config by the given name.
//
// If the tls config does not exist, return nil.
func (m *Manager) GetTLSConfig(name string) *tls.Config {
	if name == "" {
		panic("the tls config name is empty")
	}

	if value, ok := m.configs.Load(name); ok {
		return value.(*tls.Config)
	}
	return nil
}

// AddTLSConfig adds the given tls config.
func (m *Manager) AddTLSConfig(name string, config *tls.Config) {
	if name == "" {
		panic("the tls config name is empty")
	}
	if config == nil {
		panic("the tls config is nil")
	}

	m.configs.Store(name, config)
	m.updaters.Range(func(_, value interface{}) bool {
		value.(Updater).AddTLSConfig(name, config)
		return true
	})
}

// DelTLSConfig deletes the tls config by the given name.
func (m *Manager) DelTLSConfig(name string) {
	if name == "" {
		panic("the tls config name is empty")
	}

	m.configs.Delete(name)
	m.updaters.Range(func(_, value interface{}) bool {
		value.(Updater).DelTLSConfig(name)
		return true
	})
}

// GetUpdaters returns all the updaters.
func (m *Manager) GetUpdaters() map[string]Updater {
	updaters := make(map[string]Updater, 32)
	m.updaters.Range(func(key, value interface{}) bool {
		updaters[key.(string)] = value.(Updater)
		return true
	})
	return updaters
}

// GetUpdater returns the updater by the name.
//
// Return nil if the updater does not exist.
func (m *Manager) GetUpdater(name string) Updater {
	if name == "" {
		panic("the tls config updater name is empty")
	}

	if value, ok := m.updaters.Load(name); ok {
		return value.(Updater)
	}
	return nil
}

// AddUpdater adds the updater with the name.
func (m *Manager) AddUpdater(name string, updater Updater) (err error) {
	if name == "" {
		panic("the tls config updater name is empty")
	}
	if updater == nil {
		panic("the tls config updater is nil")
	}

	if _, loaded := m.updaters.LoadOrStore(name, updater); loaded {
		err = fmt.Errorf("the tls config updater named '%s' has been added", name)
	} else {
		m.configs.Range(func(key, value interface{}) bool {
			updater.AddTLSConfig(key.(string), value.(*tls.Config))
			return true
		})
	}

	return
}

// DelUpdater deletes the updater by the name.
func (m *Manager) DelUpdater(name string) {
	if name == "" {
		panic("the tls config updater name is empty")
	}
	m.updaters.Delete(name)
}
