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
	"sync/atomic"

	"github.com/xgfone/go-apiserver/tls/tlscert"
)

type clientUpdater struct{ *Config }

func (u clientUpdater) AddCertificate(string, tlscert.Certificate) { u.sync() }
func (u clientUpdater) DelCertificate(string)                      { u.sync() }
func (u clientUpdater) sync() {
	u.Config.SetTLSConfig(u.Config.tlsConfig.Load().(*tls.Config))
}

// Config is used to maintain the tls config with the certificates.
type Config struct {
	certManager *tlscert.Manager
	certFilter  *tlscert.NameFilterUpdater

	configName atomic.Value
	tlsConfig  atomic.Value

	callback        func(*tls.Config)
	updateTLSConfig func(*Config, *tls.Config)
}

// NewClientConfig returns a new config that is used by the TLS client.
//
// If tlsConfig is nil, use new(tls.Config) instead.
func NewClientConfig(tlsConfig *tls.Config, certNames ...string) *Config {
	if tlsConfig == nil {
		tlsConfig = new(tls.Config)
	}

	c := new(Config)
	c.certManager = tlscert.NewManager()
	c.certManager.AddUpdater("clientconfig", clientUpdater{c})
	c.certFilter = tlscert.NewNameFilterUpdater(c.certManager, certNames...)
	c.updateTLSConfig = updateClientConfig
	c.SetTLSConfig(tlsConfig)
	c.SetConfigName("")
	return c
}

// NewServerConfig returns a new config that is used by the TLS server.
//
// If tlsConfig is nil, use new(tls.Config) instead.
func NewServerConfig(tlsConfig *tls.Config, certNames ...string) *Config {
	if tlsConfig == nil {
		tlsConfig = new(tls.Config)
	}

	c := new(Config)
	c.certManager = tlscert.NewManager()
	c.certFilter = tlscert.NewNameFilterUpdater(c.certManager, certNames...)
	c.updateTLSConfig = updateServerConfig
	c.SetTLSConfig(tlsConfig)
	c.SetConfigName("")
	return c
}

func updateClientConfig(c *Config, tlsConfig *tls.Config) {
	tlsConfig = tlsConfig.Clone()
	tlsConfig.Certificates = c.certManager.GetTLSCertificates()
	c.tlsConfig.Store(tlsConfig)
}

func updateServerConfig(c *Config, tlsConfig *tls.Config) {
	tlsConfig = tlsConfig.Clone()
	tlsConfig.GetCertificate = c.certManager.GetTLSCertificate
	c.tlsConfig.Store(tlsConfig)
}

// SetConfigName resets the name of tls.Config.
func (c *Config) SetConfigName(tlsConfigName string) {
	c.configName.Store(tlsConfigName)
}

// OnChangedTLSConfig sets the callback function when the tls.Config is updated.
func (c *Config) OnChangedTLSConfig(callback func(newTLSConfig *tls.Config)) {
	c.callback = callback
}

// GetTLSConfig returns the tls.Config, which must not be modified.
func (c *Config) GetTLSConfig() *tls.Config {
	return c.tlsConfig.Load().(*tls.Config)
}

// SetTLSConfig implements the interface tlsconfig.Setter to update tls.Config.
func (c *Config) SetTLSConfig(tlsConfig *tls.Config) {
	c.updateTLSConfig(c, tlsConfig)
	if c.callback != nil {
		c.callback(c.GetTLSConfig())
	}
}

// AddTLSConfig implements the interface tlsconfig.Updater to update tls.Config
// only if the tls.Config name is empty or equal to the configured name.
func (c *Config) AddTLSConfig(name string, config *tls.Config) {
	if cname := c.configName.Load().(string); len(cname) == 0 || cname == name {
		c.SetTLSConfig(config)
	}
}

// DelTLSConfig implements the interface tlsconfig.Updater, which does nothing.
func (c *Config) DelTLSConfig(name string) {}

// AddCertificate implements the interface tlscert.Updater to add the certificate.
func (c *Config) AddCertificate(name string, cert tlscert.Certificate) {
	c.certFilter.AddCertificate(name, cert)
}

// DelCertificate implements the interface tlscert.Updater to delete the certificate.
func (c *Config) DelCertificate(name string) {
	c.certFilter.DelCertificate(name)
}

// AddCertNames adds the names into the supported certificates.
func (c *Config) AddCertNames(names ...string) {
	c.certFilter.AddNames(names...)
}

// DelCertNames removes the names from the supported certificates.
func (c *Config) DelCertNames(names ...string) {
	c.certFilter.DelNames(names...)
}

// GetCertNames returns the names of all the supported certificates.
func (c *Config) GetCertNames() []string {
	return c.certFilter.Names()
}
