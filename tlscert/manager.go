// Copyright 2021~2022 xgfone
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
	"strings"
	"sync"
	"sync/atomic"
)

// CertFilter is used to filter the certificate by the name,
// which is added or deleted only when the filter returns true.
type CertFilter func(name string) (ok bool)

func defaultCertFilter(name string) bool { return true }

// CertPrefixFilter returns a certificate filter that only allows
// the certificates whose name has the given prefix.
func CertPrefixFilter(prefix string) CertFilter {
	return func(s string) bool { return strings.HasPrefix(s, prefix) }
}

var _ CertUpdater = &CertManager{}

type tlsCertsWrapper struct{ Certs []Certificate }

// CertManager is used to manage a set of the certificates,
// which implements the interface CertUpdater.
type CertManager struct {
	name    string
	lock    sync.RWMutex
	certs   map[string]Certificate
	filter  CertFilter
	updater CertUpdater

	tlsCerts atomic.Value
}

// NewCertManager returns a new central certificate manager.
func NewCertManager(name string) *CertManager {
	cm := &CertManager{name: name, certs: make(map[string]Certificate, 8)}
	cm.tlsCerts.Store(tlsCertsWrapper{})
	cm.SetCertFilter(nil)
	cm.OnChanged(nil)
	return cm
}

// Name returns the name of the manager.
func (m *CertManager) Name() string { return m.name }

// OnChanged sets the certificate updater, and calls back the updater
// when adding or deleting a certain certificate.
//
// If the updater is nil, unset it.
func (m *CertManager) OnChanged(updater CertUpdater) {
	m.lock.Lock()
	if updater == nil {
		m.updater = noopCertUpdater{}
	} else {
		m.updater = updater
	}
	m.lock.Unlock()
}

// SetCertFilter sets the certificate filter to filter the added or deleted
// certificates.
//
// If filter is nil, it uses the default, which is equal to return the original name.
func (m *CertManager) SetCertFilter(filter CertFilter) {
	m.lock.Lock()
	if filter == nil {
		m.filter = defaultCertFilter
	} else {
		m.filter = filter
	}
	m.lock.Unlock()
}

// GetCertificates returns all the certificates.
func (m *CertManager) GetCertificates() map[string]Certificate {
	m.lock.RLock()
	certs := make(map[string]Certificate, len(m.certs))
	for name, cert := range m.certs {
		certs[name] = cert
	}
	m.lock.RUnlock()
	return certs
}

// GetCertificate returns the certificate information by the name.
func (m *CertManager) GetCertificate(name string) (cert Certificate, ok bool) {
	m.lock.RLock()
	cert, ok = m.certs[name]
	m.lock.Unlock()
	return
}

// AddCertificate adds the certificate with the name.
//
// If the certificate has existed, update it.
func (m *CertManager) AddCertificate(name string, cert Certificate) {
	if name == "" {
		panic("no the certificate name")
	} else if len(cert.TLSCert.Certificate) == 0 {
		panic("invalid certificate")
	}

	m.lock.Lock()
	defer m.lock.Unlock()
	if m.filter(name) {
		if old, ok := m.certs[name]; ok && old.IsEqual(cert) {
			return
		}

		m.certs[name] = cert
		m.updateCertificates()
		m.updater.AddCertificate(name, cert)
	}
	return
}

// DelCertificate deletes the certificate by the name, which does nothing
// if the certificate does not exist.
func (m *CertManager) DelCertificate(name string) {
	m.lock.Lock()
	defer m.lock.Unlock()
	if m.filter(name) {
		if _, ok := m.certs[name]; ok {
			delete(m.certs, name)
			m.updateCertificates()
			m.updater.DelCertificate(name)
		}
	}
}

func (m *CertManager) updateCertificates() {
	certs := make([]Certificate, 0, len(m.certs))
	for _, cert := range m.certs {
		certs = append(certs, cert)
	}
	m.tlsCerts.Store(tlsCertsWrapper{certs})
}

// FindCertificate traverses all the certificates and finds the matched certificate.
func (m *CertManager) FindCertificate(chi *tls.ClientHelloInfo) (cert Certificate, ok bool) {
	certs := m.tlsCerts.Load().(tlsCertsWrapper).Certs
	for i, _len := 0, len(certs); i < _len; i++ {
		if err := chi.SupportsCertificate(certs[i].TLSCert); err == nil {
			return certs[i], true
		}
	}
	return
}

// GetTLSCertificate finds the matched TLS certificate, which is used as
// tls.Config.GetCertificate.
func (m *CertManager) GetTLSCertificate(chi *tls.ClientHelloInfo) (*tls.Certificate, error) {
	if cert, ok := m.FindCertificate(chi); ok {
		return cert.TLSCert, nil
	}
	return nil, fmt.Errorf("tls: no proper certificate is configured for '%s'", chi.ServerName)
}
