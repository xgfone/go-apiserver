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

package tlsconfig

import (
	"crypto/tls"
	"fmt"
	"sync"
)

// Getter is used to get the tls config.
type Getter interface {
	GetTLSConfig() *tls.Config
}

// Setter is used to set the tls config.
type Setter interface {
	SetTLSConfig(*tls.Config)
}

// GetterFunc is the function to get the tls config.
type GetterFunc func() *tls.Config

// GetTLSConfig implements the interface Getter.
func (f GetterFunc) GetTLSConfig() *tls.Config { return f() }

// SetterFunc is the function to set the tls config.
type SetterFunc func(*tls.Config)

// SetTLSConfig implements the interface Setter.
func (f SetterFunc) SetTLSConfig(c *tls.Config) { f(c) }

// Updater is used to add or delete the TLS config.
type Updater interface {
	AddTLSConfig(name string, config *tls.Config)
	DelTLSConfig(name string)
}

// Updaters is a set of the updaters.
type Updaters []Updater

// AddTLSConfig implements the interface Updater#AddTLSConfig.
func (us Updaters) AddTLSConfig(name string, config *tls.Config) {
	for _, updater := range us {
		updater.AddTLSConfig(name, config)
	}
}

// DelTLSConfig implements the interface Updater#DelTLSConfig.
func (us Updaters) DelTLSConfig(name string) {
	for _, updater := range us {
		updater.DelTLSConfig(name)
	}
}

// NamedUpdaters is a set of the named updaters.
type NamedUpdaters struct{ updaters sync.Map }

// NewNamedUpdaters returns a new NamedUpdaters.
func NewNamedUpdaters() *NamedUpdaters { return &NamedUpdaters{} }

// AddTLSConfig implements the interface Updater#AddTLSConfig.
func (us *NamedUpdaters) AddTLSConfig(name string, config *tls.Config) {
	us.updaters.Range(func(_, value interface{}) bool {
		value.(Updater).AddTLSConfig(name, config)
		return true
	})
}

// DelTLSConfig implements the interface Updater#DelTLSConfig.
func (us *NamedUpdaters) DelTLSConfig(name string) {
	us.updaters.Range(func(_, value interface{}) bool {
		value.(Updater).DelTLSConfig(name)
		return true
	})
}

// AddUpdater adds the tls config updater with the name.
//
// The updater will be called when to add or delete the tls config.
func (us *NamedUpdaters) AddUpdater(name string, updater Updater) (err error) {
	if name == "" {
		panic("the tls config updater name is empty")
	} else if updater == nil {
		panic("the tls config updater is nil")
	}

	if _, loaded := us.updaters.LoadOrStore(name, updater); loaded {
		err = fmt.Errorf("the tls config updater named '%s' has been added", name)
	}

	return
}

// DelUpdater deletes the tls config updater by the name.
func (us *NamedUpdaters) DelUpdater(name string) {
	if name == "" {
		panic("the tls config updater name is empty")
	}
	us.updaters.Delete(name)
}

// GetUpdater returns the tls config updater by the name.
//
// Return nil if the tls config updater does not exist.
func (us *NamedUpdaters) GetUpdater(name string) Updater {
	if name == "" {
		panic("the tls config updater name is empty")
	}

	if value, ok := us.updaters.Load(name); ok {
		return value.(Updater)
	}
	return nil
}

// GetUpdaters returns all the tls config updaters.
func (us *NamedUpdaters) GetUpdaters() map[string]Updater {
	updaters := make(map[string]Updater)
	us.updaters.Range(func(key, value interface{}) bool {
		updaters[key.(string)] = value.(Updater)
		return true
	})
	return updaters
}

// FilterUpdater returns a new tls config updater to add or delete
// the tls config only if filter returns true.
func FilterUpdater(updater Updater, filter func(name string) bool) Updater {
	if updater == nil {
		panic("the tls config updater must not be nil")
	}
	if filter == nil {
		panic("the tls config updation filter must not be nil")
	}
	return filterUpdater{updater: updater, filter: filter}
}

type filterUpdater struct {
	updater Updater
	filter  func(string) bool
}

func (u filterUpdater) AddTLSConfig(name string, config *tls.Config) {
	if u.filter(name) {
		u.updater.AddTLSConfig(name, config)
	}
}

func (u filterUpdater) DelTLSConfig(name string) {
	if u.filter(name) {
		u.updater.DelTLSConfig(name)
	}
}
