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

package tlscert

import (
	"errors"
	"fmt"
	"sync"
)

var _ CertUpdater = &CertUpdaterManager{}

// CertUpdaterManager is used to manage a group of the certificate updaters.
// which implements the interface CertUpdater.
type CertUpdaterManager struct{ updaters sync.Map }

// NewCertUpdaterManager returns a new manager of the certificate updaters.
func NewCertUpdaterManager() *CertUpdaterManager { return &CertUpdaterManager{} }

// AddCertUpdater adds the certificate updater.
func (m *CertUpdaterManager) AddCertUpdater(name string, updater CertUpdater) (err error) {
	if name == "" {
		return errors.New("the certificate updater name is empty")
	}
	if updater == nil {
		return errors.New("the certificate updater is nil")
	}

	if _, loaded := m.updaters.LoadOrStore(name, updater); loaded {
		err = fmt.Errorf("the certificate updater named '%s' has been added", name)
	}
	return
}

// DelCertUpdater deletes the certificate updater by the name.
//
// If the certificate updater does not exist, do nothing.
func (m *CertUpdaterManager) DelCertUpdater(name string) {
	m.updaters.Delete(name)
}

// GetCertUpdater returns the certificate updater by the name.
//
// If the certificate updater does not exist, return nil.
func (m *CertUpdaterManager) GetCertUpdater(name string) (updater CertUpdater) {
	if value, ok := m.updaters.Load(name); ok {
		updater = value.(CertUpdater)
	}
	return
}

// GetCertUpdaters returns all the certificate updaters.
func (m *CertUpdaterManager) GetCertUpdaters() map[string]CertUpdater {
	updaters := make(map[string]CertUpdater, 32)
	m.updaters.Range(func(key, value interface{}) bool {
		updaters[key.(string)] = value.(CertUpdater)
		return true
	})
	return updaters
}

// AddCertificate implements the interface CertUpdater,
// which will call each certificate updater to add the given certificate.
func (m *CertUpdaterManager) AddCertificate(name string, cert Certificate) {
	m.updaters.Range(func(_, value interface{}) bool {
		value.(CertUpdater).AddCertificate(name, cert)
		return true
	})
}

// DelCertificate implements the interface CertUpdater,
// which will call each certificate updater to delete the given certificate.
func (m *CertUpdaterManager) DelCertificate(name string) {
	m.updaters.Range(func(_, value interface{}) bool {
		value.(CertUpdater).DelCertificate(name)
		return true
	})
}
