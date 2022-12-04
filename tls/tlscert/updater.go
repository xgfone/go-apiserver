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
	"fmt"
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

// FilterUpdater returns a new certificate updater to add or delete
// the certificate only if filter returns true.
func FilterUpdater(updater Updater, filter func(name string) bool) Updater {
	if updater == nil {
		panic("the certificate updater must not be nil")
	}
	if filter == nil {
		panic("the certificate updation filter must not be nil")
	}
	return filterUpdater{updater: updater, filter: filter}
}

type filterUpdater struct {
	updater Updater
	filter  func(string) bool
}

func (u filterUpdater) AddCertificate(name string, cert Certificate) {
	if u.filter(name) {
		u.updater.AddCertificate(name, cert)
	}
}

func (u filterUpdater) DelCertificate(name string) {
	if u.filter(name) {
		u.updater.DelCertificate(name)
	}
}

// Updaters is a set of the updaters.
type Updaters []Updater

// AddCertificate implements the interface Updater#AddCertificate.
func (us Updaters) AddCertificate(name string, cert Certificate) {
	for _, updater := range us {
		updater.AddCertificate(name, cert)
	}
}

// DelCertificate implements the interface Updater#DelCertificate.
func (us Updaters) DelCertificate(name string) {
	for _, updater := range us {
		updater.DelCertificate(name)
	}
}

// NamedUpdaters is a set of the named updaters.
type NamedUpdaters struct{ updaters sync.Map }

// NewNamedUpdaters returns a new NamedUpdaters.
func NewNamedUpdaters() *NamedUpdaters { return &NamedUpdaters{} }

// AddCertificate implements the interface Updater#AddCertificate.
func (us *NamedUpdaters) AddCertificate(name string, cert Certificate) {
	us.updaters.Range(func(_, value interface{}) bool {
		value.(Updater).AddCertificate(name, cert)
		return true
	})
}

// DelCertificate implements the interface Updater#DelCertificate.
func (us *NamedUpdaters) DelCertificate(name string) {
	us.updaters.Range(func(_, value interface{}) bool {
		value.(Updater).DelCertificate(name)
		return true
	})
}

// AddUpdater adds the certificate updater with the name.
//
// The updater will be called when to add or delete the certificate.
func (us *NamedUpdaters) AddUpdater(name string, updater Updater) (err error) {
	if name == "" {
		panic("the certificate updater name is empty")
	} else if updater == nil {
		panic("the certificate updater is nil")
	}

	if _, loaded := us.updaters.LoadOrStore(name, updater); loaded {
		err = fmt.Errorf("the certificate updater named '%s' has been added", name)
	}

	return
}

// DelUpdater deletes the certificate updater by the name.
func (us *NamedUpdaters) DelUpdater(name string) {
	if name == "" {
		panic("the certificate updater name is empty")
	}
	us.updaters.Delete(name)
}

// GetUpdater returns the certificate updater by the name.
//
// Return nil if the certificate updater does not exist.
func (us *NamedUpdaters) GetUpdater(name string) Updater {
	if name == "" {
		panic("the certificate updater name is empty")
	}

	if value, ok := us.updaters.Load(name); ok {
		return value.(Updater)
	}
	return nil
}

// GetUpdaters returns all the certificate updaters.
func (us *NamedUpdaters) GetUpdaters() map[string]Updater {
	updaters := make(map[string]Updater)
	us.updaters.Range(func(key, value interface{}) bool {
		updaters[key.(string)] = value.(Updater)
		return true
	})
	return updaters
}
