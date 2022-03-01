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
	"sync"
	"sync/atomic"
)

var _ CertUpdater = &CertManager{}

type tlsCertsWrapper struct{ Certs []Certificate }

// CertManager is used to manage a set of the certificates,
// which implements the interface CertUpdater.
type CertManager struct {
	// CertMatchHost is used to check whether the certificate matches the host.
	//
	// Default: cert.MatchHost(host) || cert.MatchIP(host)
	CertMatchHost func(cert Certificate, host string) bool

	name    string
	lock    sync.RWMutex
	certs   map[string]Certificate
	updater CertUpdater

	tlsCerts  atomic.Value
	tlsConfig atomic.Value
}

// NewCertManager returns a new central certificate manager.
func NewCertManager(name string) *CertManager {
	cm := &CertManager{name: name, certs: make(map[string]Certificate, 8)}
	cm.tlsCerts.Store(tlsCertsWrapper{})
	cm.CertMatchHost = cm.certMatchHost
	cm.SetTLSConfig(&tls.Config{})
	cm.OnChanged(nil)
	return cm
}

func (m *CertManager) certMatchHost(cert Certificate, host string) bool {
	return cert.MatchHost(host) || cert.MatchIP(host)
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
	m.certs[name] = cert
	m.updateCertificates()
	m.updater.AddCertificate(name, cert)
	return
}

// DelCertificate deletes the certificate by the name, which does nothing
// if the certificate does not exist.
func (m *CertManager) DelCertificate(name string) {
	m.lock.Lock()
	defer m.lock.Unlock()
	if _, ok := m.certs[name]; ok {
		delete(m.certs, name)
		m.updateCertificates()
		m.updater.DelCertificate(name)
	}
}

func (m *CertManager) updateCertificates() {
	certs := make([]Certificate, 0, len(m.certs))
	for _, cert := range m.certs {
		certs = append(certs, cert)
	}
	m.tlsCerts.Store(tlsCertsWrapper{certs})
}

// FindCertificate traverses all the certificates and finds the certificate
// whose SANs matches the host name that may be an ip if the certificate
// supports it.
func (m *CertManager) FindCertificate(host string) (cert Certificate, ok bool) {
	certs := m.tlsCerts.Load().(tlsCertsWrapper).Certs
	for _, cert := range certs {
		if m.CertMatchHost(cert, host) {
			return cert, true
		}
	}
	return
}

// GetConfigForClient is used as tls.Config.GetConfigForClient to look up
// the TLS config by SNI.
func (m *CertManager) GetConfigForClient(chi *tls.ClientHelloInfo) (config *tls.Config, err error) {
	if cert, ok := m.FindCertificate(chi.ServerName); ok {
		config = m.TLSConfig().Clone()
		config.GetConfigForClient = nil
		cert.UpdateTLSConfig(config)
	} else {
		err = fmt.Errorf("no certificate for the server name '%s'", chi.ServerName)
	}
	return
}

// SetTLSConfig resets the TLS config template.
func (m *CertManager) SetTLSConfig(config *tls.Config) {
	if config == nil {
		panic("TLS config is nil")
	}

	config = config.Clone()
	config.GetConfigForClient = m.GetConfigForClient
	m.tlsConfig.Store(config)
}

// TLSConfig returns the TLS config, which looks up the certificate by SNI.
func (m *CertManager) TLSConfig() *tls.Config {
	return m.tlsConfig.Load().(*tls.Config)
}
