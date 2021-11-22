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
	"net/http"
	"testing"
	"time"

	"github.com/xgfone/go-apiserver/cert"
	"github.com/xgfone/go-apiserver/internal/test"
	"github.com/xgfone/go-apiserver/log"
)

func init() { log.SetNothingWriter() }

func TestServer(t *testing.T) {
	cert, err := cert.NewCertificate([]byte(test.Ca), []byte(test.Key), []byte(test.Cert))
	if err != nil {
		t.Fatal(err)
	}

	ln, err := Listen("127.0.0.1:8001")
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
		rw.WriteHeader(200)
	})

	handler := NewHTTPServerHandler(ln.Addr(), httpHandler)
	server := NewServer(ln, handler, cert.TLSConfig)

	go server.Start()
	go handler.Start()

	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = cert.TLSConfig
	transport.ForceAttemptHTTP2 = false
	client := &http.Client{Transport: transport}
	waitDuration := time.Second

	// Test HTTP
	go func() {
		// request the block http to test the graceful shutdown.
		url := "http://127.0.0.1:8001/?sleep=" + waitDuration.String()
		testHTTPReq(t, client, url)
	}()

	// Test HTTPS: first time
	resp, err := client.Get("https://127.0.0.1:8001")
	if resp != nil {
		resp.Body.Close()
		if resp.StatusCode != 200 {
			t.Errorf("unexpected the status code '%d'", resp.StatusCode)
		}
	}
	if err != nil {
		t.Error(err)
	}

	// Test HTTPS for two times
	testHTTPReq(t, client, "https://127.0.0.1:8001")
	testHTTPReq(t, client, "https://127.0.0.1:8001")

	start := time.Now()
	server.Stop()
	if time.Since(start) < waitDuration {
		t.Error("fail to shutdown the server gracefully")
	}
}

func testHTTPReq(t *testing.T, client *http.Client, url string) {
	resp, err := client.Get(url)
	if resp != nil {
		resp.Body.Close()
		if resp.StatusCode != 200 {
			t.Errorf("unexpected the status code '%d'", resp.StatusCode)
		}
	}
	if err != nil {
		t.Error(err)
	}
}
