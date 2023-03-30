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
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"time"

	"github.com/xgfone/go-apiserver/log"
	"github.com/xgfone/go-apiserver/tls/tlscert"
	"github.com/xgfone/go-generics/maps"
)

type fileInfo struct {
	File string
	Data []byte
	Last time.Time

	logk string
	logv string
}

type fileCertInfo struct {
	Name string
	Last time.Time

	Key  fileInfo
	Cert fileInfo
}

// FileProvider is the certificate provider based on the files,
// which will watch the change of the certificate files and update
// the certificate to the new one.
type FileProvider struct {
	interval time.Duration

	lock  sync.RWMutex
	certs map[string]*fileCertInfo
	delch chan string
}

var _ Provider = &FileProvider{}

// NewFileProvider returns a new file certificate provider
// and the interval duration to check the certificate files.
//
// If interval is ZERO, it is time.Minute by default.
func NewFileProvider(interval time.Duration) *FileProvider {
	if interval <= 0 {
		interval = time.Minute
	}

	return &FileProvider{
		interval: interval,
		delch:    make(chan string, 8),
		certs:    make(map[string]*fileCertInfo, 4),
	}
}

// GetCertNames returns the names of all the file certificates.
func (p *FileProvider) GetCertNames() []string {
	p.lock.RLock()
	names := maps.Keys(p.certs)
	p.lock.RUnlock()
	return names
}

// GetCertFile returns the certificate file by the name.
//
// If the name does not exist, return ("", "", "", false).
func (p *FileProvider) GetCertFile(name string) (key, cert string, ok bool) {
	p.lock.RLock()
	info, ok := p.certs[name]
	if ok {
		key, cert = info.Key.File, info.Cert.File
	}
	p.lock.RUnlock()
	return
}

// AddCertFile adds the files associated with the certificate.
func (p *FileProvider) AddCertFile(name, keyFile, certFile string) error {
	if name == "" {
		return fmt.Errorf("the file certificate name is empty")
	} else if keyFile == "" {
		return fmt.Errorf("the file certificate keyfile is empty")
	} else if certFile == "" {
		return fmt.Errorf("the file certificate certfile is empty")
	}

	p.lock.Lock()
	if _, ok := p.certs[name]; ok {
		p.lock.Unlock()
		return fmt.Errorf("the file certificate named '%s' has existed", name)
	}

	p.certs[name] = &fileCertInfo{
		Name: name,
		Key:  fileInfo{File: keyFile, logk: "keyfile", logv: keyFile},
		Cert: fileInfo{File: certFile, logk: "certfile", logv: certFile},
	}
	p.lock.Unlock()

	return nil
}

// DelCertFile deletes the certificate file by the name.
//
// If the name does not exist, do nothing.
func (p *FileProvider) DelCertFile(name string) {
	p.lock.Lock()
	defer p.lock.Unlock()

	if maps.Delete(p.certs, name) {
		select {
		case p.delch <- name:
		default:
		}
	}
}

// OnChangedCertificate implements the interface Provider.
func (p *FileProvider) OnChangedCertificate(ctx context.Context, updater tlscert.Updater) {
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

func (p *FileProvider) update(updater tlscert.Updater) {
	p.lock.Lock()
	defer p.lock.Unlock()

	for _, cert := range p.certs {
		p.checkAndUpdate(cert, updater)
	}
}

func (p *FileProvider) checkAndUpdate(info *fileCertInfo, updater tlscert.Updater) {
	info.Cert = readCertificateFile(info.Cert)
	minTime := info.Cert.Last

	info.Key = readCertificateFile(info.Key)
	if minTime.After(info.Key.Last) {
		minTime = info.Key.Last
	}

	if !minTime.After(info.Last) { // No Change
		return
	}

	cert, err := tlscert.NewCertificate(info.Cert.Data, info.Key.Data)
	if err != nil {
		log.Error("fail to create certificate",
			info.Key.logk, info.Key.logv,
			info.Cert.logk, info.Cert.logv,
			"err", err)
		return
	}

	updater.AddCertificate(info.Name, cert)
	info.Last = minTime
}

func readCertificateFile(cert fileInfo) fileInfo {
	data, modTime, err := readChangedFile(cert.File, cert.Last)
	if err != nil {
		if os.IsNotExist(err) {
			log.Warn("the certificate file does not exist", cert.logk, cert.logv)
		} else {
			log.Error("fail to read the certificate file", cert.logk, cert.logv)
		}

		return cert
	}

	if len(data) != 0 {
		cert.Data = data
		cert.Last = modTime
	}

	return cert
}

func readChangedFile(filename string, last time.Time) ([]byte, time.Time, error) {
	fi, err := os.Stat(filename)
	if err != nil {
		return nil, time.Time{}, err
	}

	modTime := fi.ModTime()
	if !modTime.After(last) {
		return nil, time.Time{}, nil
	}

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, time.Time{}, err
	}

	return data, modTime, nil
}
