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
	"fmt"
	"strings"
	"sync"
)

var _ CertUpdater = &CertManagerGroup{}

// CertManagerGroup is used to manage a group of the certificate managers.
// which implements the interface CertUpdater.
type CertManagerGroup struct {
	lock sync.RWMutex
	cms  map[string]*CertManager
}

// NewCertManagerGroup returns a new certificate manager group.
func NewCertManagerGroup() *CertManagerGroup {
	return &CertManagerGroup{cms: make(map[string]*CertManager, 4)}
}

// AddCertManager adds the certificate manager.
func (g *CertManagerGroup) AddCertManager(cm *CertManager) (err error) {
	name := cm.Name()
	g.lock.Lock()
	if _, ok := g.cms[name]; ok {
		err = fmt.Errorf("the cert manager named '%s' has existed", name)
	} else {
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

// AddGroupCertificate adds the named certificate into the certificate manager.
//
// Notice: A certificate manager is a group, so group is the name of the manager,
// and name is the name of the certificate.
func (g *CertManagerGroup) AddGroupCertificate(group, name string, cert Certificate) (err error) {
	if cm := g.GetCertManager(group); cm != nil {
		cm.AddCertificate(name, cert)
	} else {
		err = fmt.Errorf("no the certificate manager named '%s'", group)
	}
	return
}

// DelGroupCertificate deletes the named certificate from the certificate manager.
func (g *CertManagerGroup) DelGroupCertificate(group, name string) {
	if cm := g.GetCertManager(group); cm != nil {
		cm.DelCertificate(name)
	}
}

func (g *CertManagerGroup) getCertManager(name string) (*CertManager, string) {
	certName := name
	var cm *CertManager
	if index := strings.IndexByte(name, '@'); index > -1 {
		certName = name[index+1:]
		cm = g.GetCertManager(name[:index])
	}
	return cm, certName
}

// AddCertificate implements the interface CertUpdater.
//
// If the name contains the character '@', the front part is the name of
// the certifcate manager and the back part is the real name of the certificate.
// Now, the certificate is only been added into the specific certificate manager.
// Or, it will be added into all the certificate managers.
func (g *CertManagerGroup) AddCertificate(name string, cert Certificate) {
	cm, name := g.getCertManager(name)
	if cm != nil {
		cm.AddCertificate(name, cert)
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
// If the name contains the character '@', the front part is the name of
// the certifcate manager and the back part is the real name of the certificate.
// Now, the certificate is only been deleted from the specific certificate manager.
// Or, it will be deleted from all the certificate managers.
func (g *CertManagerGroup) DelCertificate(name string) {
	cm, name := g.getCertManager(name)
	if cm != nil {
		cm.DelCertificate(name)
		return
	}

	g.lock.RLock()
	defer g.lock.RUnlock()
	for _, cm := range g.cms {
		cm.DelCertificate(name)
	}
}
