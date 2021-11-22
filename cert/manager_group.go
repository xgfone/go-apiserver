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

package cert

import (
	"crypto/tls"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
)

var _ CertUpdater = &CertManagerGroup{}

// CertManagerGroup is used to manage a group of the certificate manager.
// which implements the interface CertUpdater.
type CertManagerGroup struct {
	conf atomic.Value
	lock sync.RWMutex
	cms  map[string]*CertManager
}

// NewCertManagerGroup returns a new certificate manager group.
func NewCertManagerGroup() *CertManagerGroup {
	cm := &CertManagerGroup{cms: make(map[string]*CertManager, 4)}
	cm.SetTLSConfig(&tls.Config{})
	return cm
}

// AddCertManager adds the certificate manager.
func (g *CertManagerGroup) AddCertManager(cm *CertManager) (err error) {
	name := cm.Name()
	g.lock.Lock()
	if _, ok := g.cms[name]; ok {
		err = fmt.Errorf("the cert manager named '%s' has existed", name)
	} else {
		cm.SetTLSConfig(g.TLSConfig())
		g.cms[name] = cm
	}
	g.lock.Unlock()
	return
}

// DelCertManager deletes the certificate manager by the name.
//
// If the cert manager does not exist, do nothing.
func (g *CertManagerGroup) DelCertManager(name string) {
	g.lock.Lock()
	delete(g.cms, name)
	g.lock.Unlock()
	return
}

// GetCertManager returns the certificate manager by the name.
//
// If the cert manager does not exist, return nil.
func (g *CertManagerGroup) GetCertManager(name string) *CertManager {
	g.lock.RLock()
	cm := g.cms[name]
	g.lock.RUnlock()
	return cm
}

// GetCertManagers returns all the certificate managers.
func (g *CertManagerGroup) GetCertManagers() []*CertManager {
	g.lock.RLock()
	cms := make([]*CertManager, 0, len(g.cms))
	for _, cm := range g.cms {
		cms = append(cms, cm)
	}
	g.lock.RUnlock()
	return cms
}

// AddCertificate implements the interface CertUpdater.
//
// If the name contains the character ':', the front part is the name of
// the certifcate manager and the back part is the real name of the certificate.
// Now, the certificate is only been added into the specific certificate manager.
// Or, it will be added into all the certificate managers.
func (g *CertManagerGroup) AddCertificate(name string, cert Certificate) {
	if index := strings.IndexByte(name, ':'); index > -1 {
		if cm := g.GetCertManager(name[:index]); cm != nil {
			cm.AddCertificate(name[index+1:], cert)
		}
		return
	}

	g.lock.RLock()
	defer g.lock.RUnlock()
	for _, cm := range g.cms {
		cm.AddCertificate(name, cert)
	}
}

// DelCertificate implements the interface CertUpdater.
//
// If the name contains the character ':', the front part is the name of
// the certifcate manager and the back part is the real name of the certificate.
// Now, the certificate is only been deleted from the specific certificate manager.
// Or, it will be deleted from all the certificate managers.
func (g *CertManagerGroup) DelCertificate(name string) {
	if index := strings.IndexByte(name, ':'); index > -1 {
		if cm := g.GetCertManager(name[:index]); cm != nil {
			cm.DelCertificate(name[index+1:])
		}
		return
	}

	g.lock.RLock()
	defer g.lock.RUnlock()
	for _, cm := range g.cms {
		cm.DelCertificate(name)
	}
}

// SetTLSConfig resets the TLS config template of all the certificate managers.
func (g *CertManagerGroup) SetTLSConfig(config *tls.Config) {
	config = config.Clone()

	g.lock.RLock()
	defer g.lock.RUnlock()

	g.conf.Store(config)
	for _, cm := range g.cms {
		cm.SetTLSConfig(config)
	}
}

// TLSConfig returns the TLS config template.
func (g *CertManagerGroup) TLSConfig() *tls.Config {
	return g.conf.Load().(*tls.Config)
}
