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
	"net"
	"testing"

	"github.com/xgfone/go-apiserver/internal/test"
)

func TestCertificate(t *testing.T) {
	cert, err := NewCertificate([]byte(test.Ca), []byte(test.Key), []byte(test.Cert))
	if err != nil {
		t.Fatal(err)
	}

	if cert.StartTime.IsZero() || cert.EndTime.IsZero() {
		t.Error("no start or end time of the certificate")
	}

	if len(cert.DNSNames) != len(test.CertDNSNames) {
		t.Errorf("expect '%d' DNS names, but got '%d': %v",
			len(test.CertDNSNames), len(cert.DNSNames), cert.DNSNames)
	} else {
		for _, dnsname := range cert.DNSNames {
			if !inStrings(test.CertDNSNames, dnsname) {
				t.Errorf("unexpected dns name '%s'", dnsname)
			}
		}
	}

	if len(cert.IPAddresses) != len(test.CertIPAddresses) {
		t.Errorf("expect '%d' IPs, but got '%d': %v",
			len(test.CertIPAddresses), len(cert.IPAddresses), cert.IPAddresses)
	} else {
		for _, ip := range cert.IPAddresses {
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
