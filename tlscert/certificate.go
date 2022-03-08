// Copyright 2021~2022 xgfone
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
	"encoding/pem"
	"errors"
	"time"
)

// Certificate represents the information of a certificate.
type Certificate struct {
	// The original PEM data of the certificate.
	KeyPEM, CertPEM []byte

	// The parsed TLS certificate.
	X509Cert *x509.Certificate
	TLSCert  *tls.Certificate
}

// NewCACertificate only parses the CA certificate, which is equal to
//   NewCertificate(caPEM, nil)
func NewCACertificate(caPEM []byte) (c Certificate, err error) {
	return NewCertificate(caPEM, nil)
}

// NewCertificate parses the given PEM certificate and returns a new Certificate.
//
// Notice: keyPEM may be empty, which will only parse the certificate PEM block.
func NewCertificate(certPEM, keyPEM []byte) (c Certificate, err error) {
	c.CertPEM = bytes.TrimSpace(certPEM)
	if len(keyPEM) > 0 {
		c.KeyPEM = bytes.TrimSpace(keyPEM)
	}

	if len(c.KeyPEM) == 0 {
		var skippedBlockTypes []string

		c.TLSCert = new(tls.Certificate)
		certPEMBlock := c.CertPEM
		for {
			var certDERBlock *pem.Block
			certDERBlock, certPEMBlock = pem.Decode(certPEMBlock)
			if certDERBlock == nil {
				break
			}

			if certDERBlock.Type == "CERTIFICATE" {
				c.TLSCert.Certificate = append(c.TLSCert.Certificate, certDERBlock.Bytes)
			} else {
				skippedBlockTypes = append(skippedBlockTypes, certDERBlock.Type)
			}
		}

		if len(c.TLSCert.Certificate) == 0 {
			err = errors.New(`tls: failed to find "CERTIFICATE" PEM block in certificate pem`)
			return
		}

		c.TLSCert.Leaf, err = x509.ParseCertificate(c.TLSCert.Certificate[0])
		if err != nil {
			return
		}

	} else {
		var tlsCert tls.Certificate
		tlsCert, err = tls.X509KeyPair(c.CertPEM, c.KeyPEM)
		if err != nil {
			return c, err
		}

		c.TLSCert = &tlsCert
		if c.TLSCert.Leaf == nil {
			c.TLSCert.Leaf, err = x509.ParseCertificate(c.TLSCert.Certificate[0])
			if err != nil {
				return
			}
		}
	}

	c.X509Cert = c.TLSCert.Leaf
	return
}

// IsCA is a simplified function to report whether the certificate is a CA.
func (c Certificate) IsCA() bool { return c.X509Cert.IsCA }

// IsExpired reports whether the certificate is expired.
//
// if now is ZERO, use time.Now() instead.
func (c Certificate) IsExpired(now time.Time) bool {
	if now.IsZero() {
		now = time.Now().UTC()
	}
	return now.After(c.X509Cert.NotAfter) || c.X509Cert.NotBefore.After(now)
}

// IsEqual reports whether the current certificate is equal to o.
func (c Certificate) IsEqual(o Certificate) bool {
	return bytes.Equal(c.CertPEM, o.CertPEM) && bytes.Equal(c.KeyPEM, o.KeyPEM)
}

// UpdateCertificates fills the TLS certificate of the TLS config.
func (c Certificate) UpdateCertificates(config *tls.Config) (err error) {
	if c.X509Cert.IsCA {
		err = errors.New("the certificate is a CA certificate")
	} else {
		config.Certificates = append(config.Certificates, *c.TLSCert)
	}
	return
}

// UpdateRootCAs fills the root CAs of the TLS config, which is used
// by the client to verify a server certificate.
func (c Certificate) UpdateRootCAs(config *tls.Config) (err error) {
	if c.X509Cert.IsCA {
		if config.RootCAs == nil {
			config.RootCAs = x509.NewCertPool()
			config.RootCAs.AppendCertsFromPEM(c.CertPEM)
		}
	} else {
		err = errors.New("the certificate is not a CA certificate")
	}
	return
}

// UpdateClientCAs fills the client CAs of the TLS config, which is used
// by the server to verify a client certificate by the policy in ClientAuth.
func (c Certificate) UpdateClientCAs(config *tls.Config) (err error) {
	if c.X509Cert.IsCA {
		if config.ClientCAs == nil {
			config.ClientCAs = x509.NewCertPool()
			config.ClientCAs.AppendCertsFromPEM(c.CertPEM)
		}
	} else {
		err = errors.New("the certificate is not a CA certificate")
	}
	return
}
