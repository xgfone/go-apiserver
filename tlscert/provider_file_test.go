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
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/xgfone/go-apiserver/internal/test"
)

func TestFileProvider(t *testing.T) {
	const (
		keyfile  = "keyfile.pem"
		certfile = "certfile.pem"
	)

	defer func() { os.Remove(keyfile); os.Remove(certfile) }()
	createFile(t, keyfile, test.Key)
	createFile(t, certfile, test.Cert)

	certmanager := NewCertManager("")
	fileProvider := NewFileProvider("file", time.Millisecond*100)
	err := fileProvider.AddCertFile("test", keyfile, certfile)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	go fileProvider.OnChanged(ctx, certmanager)

	time.Sleep(time.Millisecond * 500)
	switch certs := certmanager.GetCertificates(); len(certs) {
	case 0:
		t.Error("not found any certificates")

	case 1:
		for name := range certs {
			if name != "test" {
				t.Errorf("expect the certificate '%s', but got '%s'", "test", name)
			}
		}

	default:
		t.Errorf("too many certificates: %d", len(certs))
	}

	fileProvider.DelCertFile("test")
	time.Sleep(time.Millisecond * 500)
	if certs := certmanager.GetCertificates(); len(certs) > 0 {
		names := make([]string, 0, len(certs))
		for name := range certs {
			names = append(names, name)
		}
		t.Errorf("unexpected certificates: %v", names)
	}

	cancel()
	time.Sleep(time.Millisecond * 10)
}

func createFile(t *testing.T, filename, filedata string) {
	if err := ioutil.WriteFile(filename, []byte(filedata), 0600); err != nil {
		t.Fatal(err)
	}
}
