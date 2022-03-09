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
	"net"
	"testing"

	"github.com/xgfone/go-apiserver/internal/test"
)

func testCACertificate(t *testing.T, certPEM, keyPEM string) {
	cert, err := NewCertificate([]byte(certPEM), []byte(keyPEM))
	if err != nil {
		t.Fatal(err)
	}

	if !cert.X509Cert.IsCA {
		t.Error("the certificate is not a CA")
	}

	if cn := cert.X509Cert.Subject.CommonName; cn != "test-ca" {
		t.Errorf("expect CN '%s', but got '%s'", "test-ca", cn)
	}

	if cert.X509Cert.NotBefore.IsZero() || cert.X509Cert.NotAfter.IsZero() {
		t.Error("no start or end time of the certificate")
	}
}

func TestCACertificate(t *testing.T) {
	testCACertificate(t, test.Ca, "")
	testCACertificate(t, test.Ca, test.CaKey)
}

func TestCertificate(t *testing.T) {
	cert, err := NewCertificate([]byte(test.Cert), []byte(test.Key))
	if err != nil {
		t.Fatal(err)
	}

	if cert.X509Cert.IsCA {
		t.Error("the certificate is a CA")
	}

	if cert.X509Cert.NotBefore.IsZero() || cert.X509Cert.NotAfter.IsZero() {
		t.Error("no start or end time of the certificate")
	}

	if len(cert.X509Cert.DNSNames) != len(test.CertDNSNames) {
		t.Errorf("expect '%d' DNS names, but got '%d': %v",
			len(test.CertDNSNames), len(cert.X509Cert.DNSNames), cert.X509Cert.DNSNames)
	} else {
		for _, dnsname := range cert.X509Cert.DNSNames {
			if !inStrings(test.CertDNSNames, dnsname) {
				t.Errorf("unexpected dns name '%s'", dnsname)
			}
		}
	}

	if len(cert.X509Cert.IPAddresses) != len(test.CertIPAddresses) {
		t.Errorf("expect '%d' IPs, but got '%d': %v",
			len(test.CertIPAddresses), len(cert.X509Cert.IPAddresses), cert.X509Cert.IPAddresses)
	} else {
		for _, ip := range cert.X509Cert.IPAddresses {
			if !inIPs(test.CertIPAddresses, ip) {
				t.Errorf("unexpected dns name '%s'", ip.String())
			}
		}
	}
}

func inStrings(ss []string, s string) bool {
	for _, _s := range ss {
		if s == _s {
			return true
		}
	}
	return false
}

func inIPs(ips []string, ip net.IP) bool {
	for _, _ip := range ips {
		if _ip == ip.String() {
			return true
		}
	}
	return false
}
