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

package cert

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"time"

	"github.com/xgfone/go-apiserver/log"
)

type fileInfo struct {
	File string
	Data []byte
	Last time.Time
	Info log.Field
}

type fileCertInfo struct {
	Name string
	Last time.Time

	CA   fileInfo
	Key  fileInfo
	Cert fileInfo
}

// FileProvider is the certificate provider based on the files,
// which will watch the change of the certificate files and update
// the certificate to the new one.
type FileProvider struct {
	name     string
	interval time.Duration

	lock  sync.RWMutex
	certs map[string]*fileCertInfo
	delch chan string
}

var _ Provider = &FileProvider{}

// NewFileProvider returns a new file certificate provider with the name
// and the interval duration to check the certificate files.
//
// If interval is ZERO, it is time.Minute by default.
func NewFileProvider(name string, interval time.Duration) *FileProvider {
	if name == "" {
		panic("the file provider name must not be empty")
	}
	if interval <= 0 {
		interval = time.Minute
	}

	return &FileProvider{
		name:     name,
		interval: interval,
		delch:    make(chan string, 8),
		certs:    make(map[string]*fileCertInfo, 4),
	}
}

// GetCertNames returns the names of all the file certificates.
func (p *FileProvider) GetCertNames() []string {
	p.lock.RLock()
	names := make([]string, 0, len(p.certs))
	for name := range p.certs {
		names = append(names, name)
	}
	p.lock.RUnlock()
	return names
}

// GetCertFile returns the certificate file by the name.
//
// If the name does not exist, return ("", "", "", false).
func (p *FileProvider) GetCertFile(name string) (ca, key, cert string, ok bool) {
	p.lock.RLock()
	info, ok := p.certs[name]
	if ok {
		ca, key, cert = info.CA.File, info.Key.File, info.Cert.File
	}
	p.lock.RUnlock()
	return
}

// AddCertFile adds the files associated with the certificate.
func (p *FileProvider) AddCertFile(name, caFile, keyFile, certFile string) error {
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
		CA:   fileInfo{File: caFile, Info: log.F("cafile", caFile)},
		Key:  fileInfo{File: keyFile, Info: log.F("keyfile", keyFile)},
		Cert: fileInfo{File: certFile, Info: log.F("certfile", certFile)},
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

	if _, ok := p.certs[name]; ok {
		delete(p.certs, name)
		select {
		case p.delch <- name:
		default:
		}
	}
}

// Name implements the interface Provider.
func (p *FileProvider) Name() string { return p.name }

// OnChanged implements the interface Provider.
func (p *FileProvider) OnChanged(ctx context.Context, updater CertUpdater) {
	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

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

func (p *FileProvider) update(updater CertUpdater) {
	p.lock.Lock()
	defer p.lock.Unlock()

	for _, cert := range p.certs {
		p.checkAndUpdate(cert, updater)
	}
}

func (p *FileProvider) checkAndUpdate(info *fileCertInfo, updater CertUpdater) {
	info.Cert = readCertificateFile(info.Cert)
	minTime := info.Cert.Last

	info.Key = readCertificateFile(info.Key)
	if minTime.After(info.Key.Last) {
		minTime = info.Key.Last
	}

	if info.CA.File != "" {
		info.CA = readCertificateFile(info.CA)
		if minTime.After(info.CA.Last) {
			minTime = info.CA.Last
		}
	}

	if !minTime.After(info.Last) { // No Change
		return
	}

	cert, err := NewCertificate(info.CA.Data, info.Key.Data, info.Cert.Data)
	if err != nil {
		log.Error("fail to create certificate", info.CA.Info, info.Key.Info,
			info.Cert.Info, log.E(err))
		return
	}

	updater.AddCertificate(info.Name, cert)
	info.Last = minTime
}

func readCertificateFile(cert fileInfo) fileInfo {
	data, modTime, err := readChangedFile(cert.File, cert.Last)
	if err != nil {
		if os.IsNotExist(err) {
			log.Warn("the certificate file does not exist", cert.Info)
		} else {
			log.Error("fail to read the certificate file", cert.Info)
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
