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
	"sync"
	"sync/atomic"
)

// DefaultManager is the default tls.Config manager.
var DefaultManager = NewManager(nil)

type updaterWraper struct{ Updater }

// Manager is used to manage a set of tls.Config.
type Manager struct {
	configs sync.Map
	updater atomic.Value
}

// NewManager returns a new TLS config manager.
func NewManager(updater Updater) *Manager {
	m := &Manager{}
	m.updater.Store(updaterWraper{updater})
	return m
}

func (m *Manager) updaterAddTLSConfig(name string, config *tls.Config) {
	if updater := m.GetUpdater(); updater != nil {
		updater.AddTLSConfig(name, config)
	}
}

func (m *Manager) updaterDelCertificate(name string) {
	if updater := m.GetUpdater(); updater != nil {
		updater.DelTLSConfig(name)
	}
}

// SetUpdater resets the tls config updater.
func (m *Manager) SetUpdater(updater Updater) {
	m.updater.Store(updaterWraper{updater})
}

// GetUpdater returns the tls config updater.
func (m *Manager) GetUpdater() Updater {
	return m.updater.Load().(updaterWraper).Updater
}

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
	m.updaterAddTLSConfig(name, config)
}

// DelTLSConfig deletes the tls config by the given name.
func (m *Manager) DelTLSConfig(name string) {
	if name == "" {
		panic("the tls config name is empty")
	}

	m.configs.Delete(name)
	m.updaterDelCertificate(name)
}
