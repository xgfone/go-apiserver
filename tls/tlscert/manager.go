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

import (
	"crypto/tls"
	"fmt"
	"sync"
	"sync/atomic"
)

// DefaultManager is the default certificate manager.
var DefaultManager = NewManager()

var _ Updater = &Manager{}

type certsWrapper struct{ Certs []Certificate }

// Manager is used to manage a set of the certificates,
// which implements the interface Updater.
type Manager struct {
	clock sync.RWMutex
	cmaps map[string]Certificate
	certs atomic.Value

	updaters sync.Map
}

// NewManager returns a new certificate manager.
func NewManager() *Manager {
	cm := &Manager{cmaps: make(map[string]Certificate, 8)}
	cm.certs.Store(certsWrapper{})
	return cm
}

// GetCertificates returns all the certificates.
func (m *Manager) GetCertificates() map[string]Certificate {
	m.clock.RLock()
	certs := make(map[string]Certificate, len(m.cmaps))
	for name, cert := range m.cmaps {
		certs[name] = cert
	}
	m.clock.RUnlock()
	return certs
}

// GetCertificate returns the certificate information by the name.
func (m *Manager) GetCertificate(name string) (cert Certificate, ok bool) {
	m.clock.RLock()
	cert, ok = m.cmaps[name]
	m.clock.RUnlock()
	return
}

// AddCertificate adds the certificate with the name.
//
// If the certificate has existed, update it.
func (m *Manager) AddCertificate(name string, cert Certificate) {
	if name == "" {
		panic("the certificate name is empty")
	} else if len(cert.TLSCert.Certificate) == 0 {
		panic("invalid certificate")
	}

	m.clock.Lock()
	if old, ok := m.cmaps[name]; !ok || !old.IsEqual(cert) {
		m.cmaps[name] = cert
		m.updateCertificates()
		m.clock.Unlock()

		m.updaters.Range(func(_, value interface{}) bool {
			value.(Updater).AddCertificate(name, cert)
			return true
		})
	} else {
		m.clock.Unlock()
	}
}

// DelCertificate deletes the certificate by the name,
// which does nothing if the certificate does not exist.
func (m *Manager) DelCertificate(name string) {
	if name == "" {
		panic("the certificate name is empty")
	}

	m.clock.Lock()
	if _, ok := m.cmaps[name]; ok {
		delete(m.cmaps, name)
		m.updateCertificates()
		m.clock.Unlock()

		m.updaters.Range(func(_, value interface{}) bool {
			value.(Updater).DelCertificate(name)
			return true
		})
	} else {
		m.clock.Unlock()
	}
}

func (m *Manager) updateCertificates() {
	certs := make([]Certificate, 0, len(m.cmaps))
	for _, cert := range m.cmaps {
		certs = append(certs, cert)
	}
	m.certs.Store(certsWrapper{certs})
}

// FindCertificate traverses all the certificates and finds the matched certificate.
func (m *Manager) FindCertificate(chi *tls.ClientHelloInfo) (cert Certificate, ok bool) {
	certs := m.certs.Load().(certsWrapper).Certs
	for i, _len := 0, len(certs); i < _len; i++ {
		if err := chi.SupportsCertificate(certs[i].TLSCert); err == nil {
			return certs[i], true
		}
	}
	return
}

// GetTLSCertificate finds the matched TLS certificate, which is used as
// tls.Config.GetCertificate.
func (m *Manager) GetTLSCertificate(chi *tls.ClientHelloInfo) (*tls.Certificate, error) {
	if cert, ok := m.FindCertificate(chi); ok {
		return cert.TLSCert, nil
	}
	return nil, fmt.Errorf("tls: no proper certificate is configured for '%s'", chi.ServerName)
}

// AddUpdater adds the certificate updater with the name.
//
// The updater will be called when to add or delete the certificate.
func (m *Manager) AddUpdater(name string, updater Updater) error {
	if name == "" {
		panic("the certificate updater name is empty")
	} else if updater == nil {
		panic("the certificate updater is nil")
	}

	if _, loaded := m.updaters.LoadOrStore(name, updater); loaded {
		return fmt.Errorf("the certificate updater named '%s' has been added", name)
	}

	m.clock.RLock()
	defer m.clock.RUnlock()
	for name, cert := range m.cmaps {
		updater.AddCertificate(name, cert)
	}

	return nil
}

// DelUpdater deletes the certificate updater by the name.
func (m *Manager) DelUpdater(name string) {
	if name == "" {
		panic("the certificate updater name is empty")
	}
	m.updaters.Delete(name)
}

// GetUpdater returns the certificate updater by the name.
//
// Return nil if the certificate updater does not exist.
func (m *Manager) GetUpdater(name string) Updater {
	if name == "" {
		panic("the certificate updater name is empty")
	}

	if value, ok := m.updaters.Load(name); ok {
		return value.(Updater)
	}
	return nil
}

// GetUpdaters returns all the certificate updaters.
func (m *Manager) GetUpdaters() map[string]Updater {
	updaters := make(map[string]Updater)
	m.updaters.Range(func(key, value interface{}) bool {
		updaters[key.(string)] = value.(Updater)
		return true
	})
	return updaters
}
