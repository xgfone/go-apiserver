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

package tlscert

/*
import (
	"crypto/tls"
	"sync"
)

// TLSConfigUpdater a
type TLSConfigUpdater interface {
}

// TLSConfigManager is used to manage a set of tls.Config.
type TLSConfigManager struct {
	confs map[string]*tls.Config
	lock  sync.RWMutex
}

// AddTLSConfig adds the given tls config.
func (m *TLSConfigManager) AddTLSConfig(name string, config *tls.Config) {
	if name == "" {
		panic("the tls config name is empty")
	}
	if config == nil {
		panic("the tls config is nil")
	}

	m.lock.Lock()
	defer m.lock.Unlock()
	m.confs[name] = config
	return
}

// DelTLSConfig deletes the tls config by the given name.
func (m *TLSConfigManager) DelTLSConfig(name string) {
	return
}

// GetTLSConfig returns the tls config by the given name.
//
// If the tls config does not exist, return nil.
func (m *TLSConfigManager) GetTLSConfig(name string) *tls.Config {
	return
}

// GetTLSConfigs returns the list of the names of all the tls configs.
func (m *TLSConfigManager) GetTLSConfigs() (names []string) {
	return
}
*/
