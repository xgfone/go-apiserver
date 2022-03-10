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
	"strings"
	"sync"

	"github.com/xgfone/go-apiserver/log"
)

// Updater is used to update the certificates.
type Updater interface {
	AddCertificate(name string, cert Certificate)
	DelCertificate(name string)
}

// LogUpdater returns a new certificate updater to log the operation record
// to add or delete the certificates.
//
// If the wrapped updater is not nil, it will forward the adding or deleting
// to the updater.
func LogUpdater(updater Updater) Updater { return logUpdater{updater} }

type logUpdater struct{ updater Updater }

func (u logUpdater) AddCertificate(name string, cert Certificate) {
	log.Info("add the certificate", "name", name)
	if u.updater != nil {
		u.updater.AddCertificate(name, cert)
	}
}

func (u logUpdater) DelCertificate(name string) {
	log.Info("delete the certificate", "name", name)
	if u.updater != nil {
		u.updater.DelCertificate(name)
	}
}

// PrefixUpdater returns a new certificate updater to only add or delete
// the certificate whose name has the given prefix.
func PrefixUpdater(prefix string, updater Updater) Updater {
	if prefix == "" {
		panic("the certificate updater prefix is empty")
	} else if updater == nil {
		panic("the certificate updater is nil")
	}
	return prefixUpdater{prefix: prefix, updater: updater}
}

type prefixUpdater struct {
	updater Updater
	prefix  string
}

func (u prefixUpdater) AddCertificate(name string, cert Certificate) {
	if strings.HasPrefix(name, u.prefix) {
		u.updater.AddCertificate(name, cert)
	}
}

func (u prefixUpdater) DelCertificate(name string) {
	if strings.HasPrefix(name, u.prefix) {
		u.updater.DelCertificate(name)
	}
}

// NameFilterUpdater is a certificate updater proxy to add or delete the certificates
// that have the specific name.
type NameFilterUpdater struct {
	updater Updater
	names   sync.Map
}

// NewNameFilterUpdater returns a new NameFilterUpdater with the wrapped updater.
func NewNameFilterUpdater(updater Updater, names ...string) *NameFilterUpdater {
	u := &NameFilterUpdater{updater: updater}
	u.AddNames(names...)
	return u
}

// Names returns the name list of the supported certificates.
func (u *NameFilterUpdater) Names() []string {
	names := make([]string, 0, 4)
	u.names.Range(func(key, _ interface{}) bool {
		names = append(names, key.(string))
		return true
	})
	return names
}

// AddNames adds the names of the supported certificates.
func (u *NameFilterUpdater) AddNames(names ...string) {
	for i, _len := 0, len(names); i < _len; i++ {
		u.names.LoadOrStore(names[i], struct{}{})
	}
}

// DelNames adds the names of the no longer supported certificates.
func (u *NameFilterUpdater) DelNames(names ...string) {
	for i, _len := 0, len(names); i < _len; i++ {
		u.names.Delete(names[i])
	}
}

// AddCertificate implements the interface Updater, which only adds
// the certificates that have the specific name.
func (u *NameFilterUpdater) AddCertificate(name string, cert Certificate) {
	if _, ok := u.names.Load(name); ok {
		u.updater.AddCertificate(name, cert)
	}
}

// DelCertificate implements the interface Updater, which only deletes
// the certificates that have the specific name.
func (u *NameFilterUpdater) DelCertificate(name string) {
	if _, ok := u.names.Load(name); ok {
		u.updater.DelCertificate(name)
	}
}
