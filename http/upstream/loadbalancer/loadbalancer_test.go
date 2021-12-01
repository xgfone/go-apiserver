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

package loadbalancer

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/xgfone/go-apiserver/http/upstream"
)

func testHandler(key string) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(rw, key)
	})
}

func TestLoadBalancer(t *testing.T) {
	up := NewLoadBalancer("test", nil)
	up.SwapForwarder(Retry(roundRobin(0)))

	go func() {
		server := http.Server{Addr: "127.0.0.1:8101", Handler: testHandler("8101")}
		t.Cleanup(func() { server.Shutdown(context.Background()) })
		server.ListenAndServe()
	}()

	go func() {
		server := http.Server{Addr: "127.0.0.1:8102", Handler: testHandler("8102")}
		t.Cleanup(func() { server.Shutdown(context.Background()) })
		server.ListenAndServe()
	}()

	time.Sleep(time.Millisecond * 100)

	server1, err := upstream.NewServer(upstream.ServerConfig{
		URL: upstream.URL{Domain: "www.example.com", IP: "127.0.0.1", Port: 8101},
	})
	if err != nil {
		t.Fatal(err)
	}

	server2, err := upstream.NewServer(upstream.ServerConfig{
		URL: upstream.URL{Domain: "www.example.com", IP: "127.0.0.1", Port: 8102},
	})
	if err != nil {
		t.Fatal(err)
	}

	if url := server1.URL().String(); url != "http://127.0.0.1:8101" {
		t.Errorf("expect url '%s', but got '%s'", "http://127.0.0.1:8101", url)
	}
	if err := server1.Check(context.Background(), upstream.URL{}); err != nil {
		t.Errorf("health check failed: %s", err)
	}

	up.ResetServers(server1, server2)

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "http://127.0.0.1", nil)
	up.ServeHTTP(rec, req)
	up.ServeHTTP(rec, req)
	up.ServeHTTP(rec, req)
	up.ServeHTTP(rec, req)

	expects := []string{
		"8102",
		"8101",
		"8102",
		"8101",
		"",
	}
	results := strings.Split(rec.Body.String(), "\n")
	if len(expects) != len(results) {
		t.Errorf("expect %d lines, but got %d", len(expects), len(results))
	} else {
		for i, line := range results {
			if line != expects[i] {
				t.Errorf("%d line: expect '%s', but got '%s'", i, expects[i], line)
			}
		}
	}

	server3, err := upstream.NewServer(upstream.ServerConfig{
		URL: upstream.URL{Domain: "www.example.com", IP: "127.0.0.1", Port: 8103},
	})
	if err != nil {
		t.Fatal(err)
	}
	up.UpsertServers(server3)
	up.SetHealthCheck(HealthCheckInfo{
		Interval: time.Millisecond * 100,
		Timeout:  time.Millisecond * 100,
	})

	time.Sleep(time.Millisecond * 500)
	rec.Body.Reset()
	up.ServeHTTP(rec, req)
	up.ServeHTTP(rec, req)
	up.ServeHTTP(rec, req)
	up.ServeHTTP(rec, req)

	expects = []string{
		"8102",
		"8101",
		"8102",
		"8101",
		"",
	}
	results = strings.Split(rec.Body.String(), "\n")
	if len(expects) != len(results) {
		t.Errorf("expect %d lines, but got %d", len(expects), len(results))
	} else {
		for i, line := range results {
			if line != expects[i] {
				t.Errorf("%d line: expect '%s', but got '%s'", i, expects[i], line)
			}
		}
	}

	state := server1.State()
	if state.Total != 4 {
		t.Errorf("expect %d total requests, but got %d", 4, state.Total)
	}
	if state.Success != 4 {
		t.Errorf("expect %d success requests, but got %d", 4, state.Total)
	}
}
