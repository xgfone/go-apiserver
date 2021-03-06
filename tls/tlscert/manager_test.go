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
	"testing"

	"github.com/xgfone/go-apiserver/helper"
	"github.com/xgfone/go-apiserver/internal/test"
)

func TestManager(t *testing.T) {
	cert, _ := NewCertificate([]byte(test.Cert), []byte(test.Key))

	cm1 := NewManager()
	cm2 := NewManager()

	m := NewManager()
	m.AddUpdater("cm1", PrefixUpdater("cm1@", cm1))
	m.AddCertificate("cm1@name", cert)
	m.AddCertificate("cm2@name", cert)
	m.AddUpdater("cm2", PrefixUpdater("cm2@", cm2))

	checkCerts(t, "t1", cm1.GetCertificates(), []string{"cm1@name"})
	checkCerts(t, "t2", cm2.GetCertificates(), []string{"cm2@name"})

	m.DelCertificate("cm1@name")
	checkCerts(t, "t3", cm1.GetCertificates(), []string{})
	checkCerts(t, "t4", cm2.GetCertificates(), []string{"cm2@name"})
}

func checkCerts(t *testing.T, prefix string, certs map[string]Certificate, names []string) {
	if len(certs) != len(names) {
		t.Errorf("%s: expect %d certificates, but got %d", prefix, len(names), len(certs))
	} else {
		for name := range certs {
			if !helper.InStrings(name, names) {
				t.Errorf("%s: unexpected certificate named '%s'", prefix, name)
			}
		}
	}
}
