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

package tls

import (
	"crypto/tls"
	"sync"
	"sync/atomic"

	"github.com/xgfone/go-apiserver/tls/tlscert"
	"github.com/xgfone/go-apiserver/tls/tlsconfig"
)

type clientUpdater struct{ *Config }

func (u clientUpdater) AddCertificate(string, tlscert.Certificate) { u.sync() }
func (u clientUpdater) DelCertificate(string)                      { u.sync() }
func (u clientUpdater) sync() {
	u.Config.SetTLSConfig(u.Config.tlsConfig.Load().(*tls.Config))
}

// Config is used to maintain the tls config with the certificates.
type Config struct {
	tlsConfig   atomic.Value
	certManager *tlscert.Manager

	updateTLSCert   func(*Config)
	updateTLSConfig func(*Config, *tls.Config)

	lock     sync.RWMutex
	callback func(*tls.Config)
}

// NewClientConfig returns a new config that is used by the TLS client.
//
// If tlsConfig is nil, use new(tls.Config) instead.
func NewClientConfig(tlsConfig *tls.Config) *Config {
	return newConfig(tlsConfig, updateClientConfig, updateClientCert)
}

// NewServerConfig returns a new config that is used by the TLS server.
//
// If tlsConfig is nil, use new(tls.Config) instead.
func NewServerConfig(tlsConfig *tls.Config) *Config {
	return newConfig(tlsConfig, updateServerConfig, updateServerCert)
}

func newConfig(config *tls.Config,
	updateConfig func(*Config, *tls.Config),
	updateCert func(*Config)) *Config {
	if config == nil {
		config = new(tls.Config)
	}

	c := &Config{
		certManager:     tlscert.NewManager(nil),
		updateTLSCert:   updateCert,
		updateTLSConfig: updateConfig,
	}
	c.SetTLSConfig(config)
	return c
}

func updateServerCert(*Config) {}
func updateClientCert(c *Config) {
	c.lock.Lock()
	defer c.lock.Unlock()

	tlsConfig := c.GetTLSConfig()
	tlsConfig.Certificates = c.certManager.GetTLSCertificates()
	c.tlsConfig.Store(tlsConfig)
	if c.callback != nil {
		c.callback(tlsConfig)
	}
}

func updateClientConfig(c *Config, tlsConfig *tls.Config) {
	c.lock.Lock()
	defer c.lock.Unlock()

	tlsConfig = tlsConfig.Clone()
	tlsConfig.Certificates = c.certManager.GetTLSCertificates()
	c.tlsConfig.Store(tlsConfig)
	if c.callback != nil {
		c.callback(tlsConfig)
	}
}

func updateServerConfig(c *Config, tlsConfig *tls.Config) {
	c.lock.Lock()
	defer c.lock.Unlock()

	tlsConfig = tlsConfig.Clone()
	tlsConfig.GetCertificate = c.certManager.GetTLSCertificate
	c.tlsConfig.Store(tlsConfig)
	if c.callback != nil {
		c.callback(tlsConfig)
	}
}

// OnChangedTLSConfig sets the callback function when tls.Config is changed.
func (c *Config) OnChangedTLSConfig(callback func(*tls.Config)) {
	c.lock.Lock()
	c.callback = callback
	c.lock.Unlock()
}

var (
	_ tlsconfig.Getter = new(Config)
	_ tlsconfig.Setter = new(Config)
	_ tlscert.Updater  = new(Config)
)

// GetTLSConfig implements the interface tlsconfig.Getter to get the tls config.
func (c *Config) GetTLSConfig() *tls.Config {
	return c.tlsConfig.Load().(*tls.Config)
}

// SetTLSConfig implements the interface tlsconfig.Setter to update tls.Config.
func (c *Config) SetTLSConfig(tlsConfig *tls.Config) {
	if tlsConfig == nil {
		panic("tls.Config must not be nil")
	}
	c.updateTLSConfig(c, tlsConfig)
}

// AddCertificate implements the interface tlscert.Updater to add the certificate.
func (c *Config) AddCertificate(name string, cert tlscert.Certificate) {
	c.certManager.AddCertificate(name, cert)
	c.updateTLSCert(c)
}

// DelCertificate implements the interface tlscert.Updater to delete the certificate.
func (c *Config) DelCertificate(name string) {
	c.certManager.DelCertificate(name)
	c.updateTLSCert(c)
}
