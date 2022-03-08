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

package tcp

import (
	"crypto/tls"
	"net/http"
	"testing"
	"time"

	"github.com/xgfone/go-apiserver/internal/test"
	"github.com/xgfone/go-apiserver/tlscert"
)

func TestServer(t *testing.T) {
	caCert, err := tlscert.NewCACertificate([]byte(test.Ca))
	if err != nil {
		t.Fatal(err)
	}

	cert, err := tlscert.NewCertificate([]byte(test.Cert), []byte(test.Key))
	if err != nil {
		t.Fatal(err)
	}

	serverTLSConfig := new(tls.Config)
	cert.UpdateCertificates(serverTLSConfig)
	// caCert.UpdateClientCAs(serverTLSConfig)
	// serverTLSConfig.ClientAuth = tls.RequireAndVerifyClientCert

	clientTLSConfig := new(tls.Config)
	caCert.UpdateRootCAs(clientTLSConfig)
	// cert.UpdateCertificates(clientTLSConfig)

	ln, err := Listen("127.0.0.1:8301")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	var httpHandler http.Handler
	httpHandler = http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if sleep := r.URL.Query().Get("sleep"); sleep != "" {
			if interval, _ := time.ParseDuration(sleep); interval > 0 {
				time.Sleep(interval)
			}
		}
		if r.TLS == nil {
			rw.WriteHeader(201)
		} else {
			rw.WriteHeader(202)
		}
	})

	handler := NewHTTPServerHandler(ln.Addr(), httpHandler)
	server := NewServer(ln, handler)
	server.SetTLSConfig(serverTLSConfig, false)

	go server.Start()
	go handler.Start()

	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = clientTLSConfig
	transport.ForceAttemptHTTP2 = false
	client := &http.Client{Transport: transport}
	waitDuration := time.Second

	// Test HTTP
	go func() {
		// request the block http to test the graceful shutdown.
		url := "http://127.0.0.1:8301/?sleep=" + waitDuration.String()
		testHTTPReq(t, client, url, 201)
	}()

	// Test HTTPS: first time
	testHTTPReq(t, client, "https://127.0.0.1:8301", 202)

	// Test HTTPS for two times
	testHTTPReq(t, client, "https://127.0.0.1:8301", 202)
	testHTTPReq(t, client, "https://127.0.0.1:8301", 202)

	start := time.Now()
	server.Stop()
	if time.Since(start) < waitDuration {
		t.Error("fail to shutdown the server gracefully")
	}
}

func testHTTPReq(t *testing.T, client *http.Client, url string, code int) {
	resp, err := client.Get(url)
	if resp != nil {
		resp.Body.Close()
		if resp.StatusCode != code {
			t.Errorf("expected statuscode %d, but got %d", code, resp.StatusCode)
		}
	}
	if err != nil {
		t.Error(err)
	}
}
