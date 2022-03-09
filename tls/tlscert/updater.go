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
