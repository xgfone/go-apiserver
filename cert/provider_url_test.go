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
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/xgfone/go-apiserver/internal/test"
)

func TestURLProvider(t *testing.T) {
	httpserver := http.Server{
		Addr: "127.0.0.1:8888",
		Handler: http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			rw.WriteHeader(200)
			json.NewEncoder(rw).Encode(urlCert{
				CA:   test.Ca,
				Key:  test.Key,
				Cert: test.Cert,
			})
		}),
	}
	go httpserver.ListenAndServe()
	defer httpserver.Shutdown(context.Background())
	time.Sleep(time.Millisecond * 50)

	certmanager := NewCertManager("")
	urlProvider := NewURLProvider("url", time.Millisecond*100)
	err := urlProvider.AddCertURL("test", "http://127.0.0.1:8888")
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	go urlProvider.OnChanged(ctx, certmanager)

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

	urlProvider.DelCertURL("test")
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
