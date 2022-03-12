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

import "crypto/tls"

// Setter is used to set the tls config.
type Setter interface {
	SetTLSConfig(*tls.Config)
}

// SetterUpdater is a tls.Config updater to set tls.Config when it has changed.
type SetterUpdater struct {
	// TLSConfigName is used to filter the specific named tls.Config if not empty.
	TLSConfigName string

	// TLSConfigSetter is used to set and update the tls config to the new.
	TLSConfigSetter Setter
}

// NewSetterUpdater returns a new SetterUpdater.
func NewSetterUpdater(tlsConfigName string, tlsConfigSetter Setter) SetterUpdater {
	return SetterUpdater{tlsConfigName, tlsConfigSetter}
}

// AddTLSConfig implements the interface Updater, which will call the setter
// to set the tls config to the new.
func (u SetterUpdater) AddTLSConfig(name string, config *tls.Config) {
	if name == "" || name == u.TLSConfigName {
		u.TLSConfigSetter.SetTLSConfig(config)
	}
}

// DelTLSConfig implements the interface Updater, which does nothing.
func (u SetterUpdater) DelTLSConfig(name string) {}
