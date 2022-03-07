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
	"context"
	"os"
	"testing"
	"time"

	"github.com/xgfone/go-apiserver/internal/test"
)

func TestProviderManager(t *testing.T) {
	const (
		keyfile  = "provider_keyfile.pem"
		certfile = "provider_certfile.pem"
	)

	defer func() { os.Remove(keyfile); os.Remove(certfile) }()
	createFile(t, keyfile, test.Key)
	createFile(t, certfile, test.Cert)

	fileProvider1 := NewFileProvider("file1", time.Millisecond*100)
	fileProvider2 := NewFileProvider("file2", time.Millisecond*100)

	certmanager := NewCertManager("")
	pm := NewProviderManager(certmanager)
	pm.AddProvider(fileProvider1)
	pm.Start(context.Background())
	pm.Start(context.Background())
	pm.AddProvider(fileProvider2)

	err := fileProvider1.AddCertFile("cert1", keyfile, certfile)
	if err != nil {
		t.Fatal(err)
	}

	err = fileProvider2.AddCertFile("cert2", keyfile, certfile)
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(time.Millisecond * 500)
	switch certs := certmanager.GetCertificates(); len(certs) {
	case 0:
		t.Error("not found any certificates")

	case 1:
		names := make([]string, 0, 1)
		for name := range certs {
			names = append(names, name)
		}
		t.Errorf("too few cerrtificates: %v", names)

	case 2:
		for name := range certs {
			switch name {
			case "cert1", "cert2":
			default:
				t.Errorf("unexpected certificate '%s'", name)
			}
		}

	default:
		t.Errorf("too many certificates: %d", len(certs))
	}

	time.Sleep(time.Millisecond * 500)
	pm.DelProvider("file2")
	time.Sleep(time.Millisecond * 10)
	switch certs := certmanager.GetCertificates(); len(certs) {
	case 0:
		t.Error("not found any certificates")

	case 1:
		for name := range certs {
			if name != "cert1" {
				t.Errorf("expect the certificate '%s', but got '%s'", "cert1", name)
			}
		}

	default:
		t.Errorf("too many certificates: %d", len(certs))
	}

	if _, ok := certmanager.FindCertificate("127.0.0.1"); !ok {
		t.Errorf("not found the certificate for %s", "127.0.0.1")
	}
	if _, ok := certmanager.FindCertificate("localhost"); !ok {
		t.Errorf("not found the certificate for %s", "localhost")
	}

	pm.Stop()
	pm.Stop()
	time.Sleep(time.Millisecond * 10)
}
