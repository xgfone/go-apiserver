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
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"net"
	"net/url"
	"strings"
	"time"
)

// Certificate represents the information of a certificate.
type Certificate struct {
	// The original PEM data of the certificate.
	CA, Key, Cert []byte

	// The start and end time of the validity of the certificate.
	//
	// Notice: They are pared from the Cert PEM.
	StartTime, EndTime time.Time

	// The SAN information signed by the certificate.
	//
	// Notice: They are pared from the Cert PEM.
	DNSNames       []string
	EmailAddresses []string
	IPAddresses    []net.IP
	URIs           []*url.URL

	// The certificate information.
	CN      string
	IsCA    bool
	Version int

	// The parsed TLS certificate and Root CA.
	RootCAs *x509.CertPool
	TLSCert tls.Certificate

	// TLSConfig is the TLS config, which is generated with RootCAs and TLSCert.
	TLSConfig *tls.Config

	// It is a cache of the certificate list to avoid the memory allocation.
	tlsCerts []tls.Certificate
}

// NewCertificate returns the a new Certificate.
//
// Notice: Both key and cert are the PEM block.
func NewCertificate(ca, key, cert []byte) (c Certificate, err error) {
	ca = bytes.TrimSpace(ca)
	key = bytes.TrimSpace(key)
	cert = bytes.TrimSpace(cert)

	var capool *x509.CertPool
	if len(ca) != 0 {
		capool = x509.NewCertPool()
		capool.AppendCertsFromPEM(ca)
	}

	tlsCert, err := tls.X509KeyPair(cert, key)
	if err != nil {
		return
	}

	if tlsCert.Leaf == nil {
		tlsCert.Leaf, err = x509.ParseCertificate(tlsCert.Certificate[0])
		if err != nil {
			return
		}
	}

	c.CN = tlsCert.Leaf.Subject.CommonName
	c.IsCA = tlsCert.Leaf.IsCA
	c.Version = tlsCert.Leaf.Version

	c.URIs = tlsCert.Leaf.URIs
	c.DNSNames = tlsCert.Leaf.DNSNames
	c.IPAddresses = tlsCert.Leaf.IPAddresses
	c.EmailAddresses = tlsCert.Leaf.EmailAddresses

	c.StartTime = tlsCert.Leaf.NotBefore.UTC()
	c.EndTime = tlsCert.Leaf.NotAfter.UTC()
	c.TLSCert = tlsCert
	c.RootCAs = capool

	c.tlsCerts = []tls.Certificate{tlsCert}
	c.TLSConfig = &tls.Config{RootCAs: capool, Certificates: c.tlsCerts}
	c.CA, c.Key, c.Cert = ca, key, cert

	return
}

// IsExpired reports whether the certificate is expired.
//
// if now is ZERO, use time.Now() instead.
func (c Certificate) IsExpired(now time.Time) bool {
	if now.IsZero() {
		now = time.Now().UTC()
	}
	return now.After(c.EndTime) || c.StartTime.After(now)
}

// HasChanged reports whether the certificate has changed.
func (c Certificate) HasChanged(ca, key, cert []byte) bool {
	return !bytes.Equal(c.Cert, cert) || !bytes.Equal(c.Key, key) || !bytes.Equal(c.CA, ca)
}

// IsEqual reports whether the current certificate is equal to other.
func (c Certificate) IsEqual(other Certificate) bool {
	return !c.HasChanged(other.CA, other.Key, other.Cert)
}

// UpdateTLSConfig fills the TLS certificate into the TLS config.
func (c Certificate) UpdateTLSConfig(config *tls.Config) {
	config.Certificates = c.tlsCerts
	if c.RootCAs != nil {
		config.RootCAs = c.RootCAs
	}
}

// MatchIP checks whether there is one of the IP SANs to match the ip.
func (c Certificate) MatchIP(ip string) bool {
	netip := net.ParseIP(ip)
	if netip == nil {
		return false
	}

	for _, ipaddr := range c.IPAddresses {
		if ipaddr.Equal(netip) {
			return true
		}
	}

	return false
}

// MatchHost checks whether there is one of the DNSName SANs to match the host.
//
// Notice: It only supports the full domain host or the wild
func (c Certificate) MatchHost(host string) bool {
	for _, dnsname := range c.DNSNames {
		if strings.HasPrefix(dnsname, "*.") {
			dnsname = dnsname[2:]
		}

		if dnsname == host {
			return true
		}
	}
	return false
}
