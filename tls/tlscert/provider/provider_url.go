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

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/xgfone/go-apiserver/log"
	"github.com/xgfone/go-apiserver/tls/tlscert"
)

var _ Provider = &URLProvider{}

type urlCert struct {
	KeyPEM  string `json:"keyPEM"`
	CertPEM string `json:"certPEM"`
}

type urlCertInfo struct {
	URL  string
	Name string
	Cert urlCert
}

// URLProvider is the certificate provider based on the urls,
// which will watch the change of the certificates from the urls
// and update the certificate to the new one.
//
// Notice: it only uses http.Get to access the url, and gets the certificate
// information from the response body, which is a JSON data with the three keys,
// "keyPEM" and "certPEM", the values of which is the PEM string, for example,
//
//     {
//         "keyPEM": "-----BEGIN RSA PRIVATE KEY-----......-----END RSA PRIVATE KEY-----",
//         "certPEM": "-----BEGIN CERTIFICATE-----......-----END CERTIFICATE-----"
//     }
//
type URLProvider struct {
	interval time.Duration

	lock  sync.RWMutex
	certs map[string]*urlCertInfo
	delch chan string
}

// NewURLProvider returns a new url certificate provider
// and the interval duration to check the certificate files.
//
// If interval is ZERO, it is time.Minute by default.
func NewURLProvider(interval time.Duration) *URLProvider {
	if interval <= 0 {
		interval = time.Minute
	}

	return &URLProvider{
		interval: interval,
		delch:    make(chan string, 8),
		certs:    make(map[string]*urlCertInfo, 4),
	}
}

// GetCertURLs returns the information of all the certificates.
//
// Notice: The key of the returned result is the name, and the value of that
// is the url.
func (p *URLProvider) GetCertURLs() map[string]string {
	p.lock.RLock()
	certs := make(map[string]string)
	for _, cert := range p.certs {
		certs[cert.Name] = cert.URL
	}
	p.lock.RUnlock()
	return certs
}

// GetCertURL returns the certificate url by the name.
//
// If the name does not exist, return "".
func (p *URLProvider) GetCertURL(name string) (url string) {
	p.lock.RLock()
	if cert, ok := p.certs[name]; ok {
		url = cert.URL
	}
	p.lock.RUnlock()
	return
}

// AddCertURL adds the certificate url.
func (p *URLProvider) AddCertURL(name, rawurl string) (err error) {
	if name == "" {
		return fmt.Errorf("the url certificate name is empty")
	} else if rawurl == "" {
		return fmt.Errorf("the url certificate keyfile is empty")
	}

	p.lock.Lock()
	if _, ok := p.certs[name]; ok {
		err = fmt.Errorf("the url certificate named '%s' has existed", name)
	} else {
		p.certs[name] = &urlCertInfo{Name: name, URL: rawurl}
	}
	p.lock.Unlock()

	return
}

// DelCertURL deletes the certificate url by the name.
//
// If the name does not exist, do nothing.
func (p *URLProvider) DelCertURL(name string) {
	p.lock.Lock()
	defer p.lock.Unlock()

	if _, ok := p.certs[name]; ok {
		delete(p.certs, name)
		select {
		case p.delch <- name:
		default:
		}
	}
}

// OnChangedCertificate implements the interface Provider.
func (p *URLProvider) OnChangedCertificate(ctx context.Context, updater tlscert.Updater) {
	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	p.update(updater)
	for {
		select {
		case <-ctx.Done():
			p.lock.RLock()
			for name := range p.certs {
				updater.DelCertificate(name)
			}
			p.lock.RUnlock()
			return

		case name := <-p.delch:
			updater.DelCertificate(name)

		case <-ticker.C:
			p.update(updater)
		}
	}
}

func (p *URLProvider) update(updater tlscert.Updater) {
	p.lock.Lock()
	defer p.lock.Unlock()

	for _, cert := range p.certs {
		p.checkAndUpdate(cert, updater)
	}
}

func (p *URLProvider) checkAndUpdate(info *urlCertInfo, updater tlscert.Updater) {
	resp, err := http.Get(info.URL)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		log.Error("fail to get the certificate",
			"name", info.Name, "url", info.URL, "err", err)
		return
	}

	var r urlCert
	if err = json.NewDecoder(resp.Body).Decode(&r); err != nil {
		log.Error("fail to decode the response body with json",
			"name", info.Name, "url", info.URL, "err", err)
		return
	}

	if r.KeyPEM == info.Cert.KeyPEM && r.CertPEM == info.Cert.CertPEM {
		return // No Change
	}

	cert, err := tlscert.NewCertificate([]byte(r.CertPEM), []byte(r.KeyPEM))
	if err != nil {
		log.Error("fail to create certificate",
			"name", info.Name, "url", info.URL, "err", err)
		return
	}

	updater.AddCertificate(info.Name, cert)
	info.Cert = r
}
