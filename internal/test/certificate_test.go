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

package test

import (
	"crypto/tls"
	"crypto/x509"
	"testing"

	"github.com/xgfone/go-apiserver/tools/slices"
)

func TestCertificate(t *testing.T) {
	tlsCert, err := tls.X509KeyPair([]byte(Cert), []byte(Key))
	if err != nil {
		t.Fatal(err)
	}

	if tlsCert.Leaf == nil {
		tlsCert.Leaf, err = x509.ParseCertificate(tlsCert.Certificate[0])
		if err != nil {
			t.Fatal(err)
		}
	}

	cert := tlsCert.Leaf

	if cn := cert.Subject.CommonName; cn != CertCN {
		t.Errorf("expect CN '%s', but got '%s'", CertCN, cn)
	}

	if !slices.Equal(cert.DNSNames, CertDNSNames) {
		t.Errorf("expect DNS '%v', but got '%v'", CertDNSNames, cert.DNSNames)
	}

	if len1, len2 := len(CertIPAddresses), len(cert.IPAddresses); len1 != len2 {
		t.Errorf("expect %d ips, but got %d: %v", len1, len2, cert.IPAddresses)
	} else {
		for _, ip := range cert.IPAddresses {
			if !slices.Contains(CertIPAddresses, ip.String()) {
				t.Errorf("unexpected ip '%s'", ip.String())
			}
		}
	}
}
