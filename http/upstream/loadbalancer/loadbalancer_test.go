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
	"github.com/xgfone/go-apiserver/http/upstream/balancer"
)

func testHandler(key string) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(rw, key)
	})
}

func TestLoadBalancer(t *testing.T) {
	lb := NewLoadBalancer("test", balancer.RoundRobin())
	lb.SwapBalancer(balancer.Retry(balancer.RoundRobin()))

	go func() {
		server := http.Server{Addr: "127.0.0.1:8101", Handler: testHandler("8101")}
		server.ListenAndServe()
	}()

	go func() {
		server := http.Server{Addr: "127.0.0.1:8102", Handler: testHandler("8102")}
		server.ListenAndServe()
	}()

	time.Sleep(time.Millisecond * 100)

	server1, err := upstream.NewServer(upstream.ServerConfig{
		URL:          upstream.URL{Domain: "www.example.com", IP: "127.0.0.1", Port: 8101},
		StaticWeight: 1,
	})
	if err != nil {
		t.Fatal(err)
	}

	server2, err := upstream.NewServer(upstream.ServerConfig{
		URL:          upstream.URL{Domain: "www.example.com", IP: "127.0.0.1", Port: 8102},
		StaticWeight: 2,
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

	lb.ResetServers(server1, server2)

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "http://127.0.0.1", nil)
	lb.ServeHTTP(rec, req)
	lb.ServeHTTP(rec, req)
	lb.ServeHTTP(rec, req)
	lb.ServeHTTP(rec, req)

	expects := []string{
		"8101",
		"8102",
		"8101",
		"8102",
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

	state := server1.State()
	if state.Total != 2 {
		t.Errorf("expect %d total requests, but got %d", 4, state.Total)
	}
	if state.Success != 2 {
		t.Errorf("expect %d success requests, but got %d", 4, state.Total)
	}

	/// ------------------------------------------------------------------ ///

	lb.SetServerOnline(server1.ID(), false)
	if online, ok := lb.ServerIsOnline(server1.ID()); !ok || online {
		t.Errorf("invalid the server1 online status: online=%v, ok=%v", online, ok)
	}
	if online, ok := lb.ServerIsOnline(server2.ID()); !ok || !online {
		t.Errorf("invalid the server2 online status: online=%v, ok=%v", online, ok)
	}

	rec.Body.Reset()
	lb.ServeHTTP(rec, req)
	lb.ServeHTTP(rec, req)
	expects = []string{
		"8102",
		"8102",
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

	/// ------------------------------------------------------------------ ///

	lb.SetServerOnlines(map[string]bool{server2.ID(): false})
	if online, ok := lb.ServerIsOnline(server2.ID()); !ok || online {
		t.Errorf("invalid the server2 online status: online=%v, ok=%v", online, ok)
	}

	servers := lb.GetServers()
	if len(servers) != 2 {
		t.Errorf("expect %d servers, but got %d", 2, len(servers))
	}
	for server, online := range servers {
		id := server.ID()
		switch id {
		case server1.ID(), server2.ID():
		default:
			t.Errorf("unknown server id '%s'", id)
		}

		if online {
			t.Errorf("expect server '%s' online is false, but got true", id)
		}
	}

	rec = httptest.NewRecorder()
	lb.ServeHTTP(rec, req)
	if rec.Code != 503 {
		t.Errorf("unexpected response: statuscode=%d, body=%s", rec.Code, rec.Body.String())
	}
}
