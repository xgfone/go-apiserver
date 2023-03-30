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

	"github.com/xgfone/go-generics/maps"
)

// DefaultManager is the default certificate manager.
var DefaultManager = NewManager()

var _ Updater = &Manager{}

type certsWrapper struct{ Certs []Certificate }
type updaterWraper struct{ Updater }

// Manager is used to manage a set of the certificates,
// which implements the interface Updater.
type Manager struct {
	clock   sync.RWMutex
	cmaps   map[string]Certificate
	certs   atomic.Value
	updater atomic.Value
}

// NewManager returns a new certificate manager.
func NewManager() *Manager {
	m := &Manager{cmaps: make(map[string]Certificate, 8)}
	m.certs.Store(certsWrapper{})
	m.SetUpdater(nil)
	return m
}

func (m *Manager) updaterAddCertificate(name string, cert Certificate) {
	if updater := m.GetUpdater(); updater != nil {
		updater.AddCertificate(name, cert)
	}
}

func (m *Manager) updaterDelCertificate(name string) {
	if updater := m.GetUpdater(); updater != nil {
		updater.DelCertificate(name)
	}
}

// SetUpdater resets the certificate updater.
func (m *Manager) SetUpdater(updater Updater) {
	m.updater.Store(updaterWraper{updater})
}

// GetUpdater returns the certificate updater.
func (m *Manager) GetUpdater() Updater {
	return m.updater.Load().(updaterWraper).Updater
}

// GetCertificates returns all the certificates.
func (m *Manager) GetCertificates() map[string]Certificate {
	m.clock.RLock()
	certs := maps.Clone(m.cmaps)
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
	defer m.clock.Unlock()
	if old, ok := m.cmaps[name]; !ok || !old.IsEqual(cert) {
		m.cmaps[name] = cert
		m.updateCertificates()
		m.updaterAddCertificate(name, cert)
	}
}

// DelCertificate deletes the certificate by the name,
// which does nothing if the certificate does not exist.
func (m *Manager) DelCertificate(name string) {
	if name == "" {
		panic("the certificate name is empty")
	}

	m.clock.Lock()
	defer m.clock.Unlock()
	if maps.Delete(m.cmaps, name) {
		m.updateCertificates()
		m.updaterDelCertificate(name)
	}
}

func (m *Manager) updateCertificates() {
	m.certs.Store(certsWrapper{maps.Values(m.cmaps)})
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

// GetTLSCertificates returns all the certificates as []tls.Certificate.
func (m *Manager) GetTLSCertificates() (certificates []tls.Certificate) {
	certs := m.certs.Load().(certsWrapper).Certs
	certificates = make([]tls.Certificate, len(certs))
	for i, _len := 0, len(certs); i < _len; i++ {
		certificates[i] = *certs[i].TLSCert
	}
	return
}
