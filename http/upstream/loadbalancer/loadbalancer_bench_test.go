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
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/xgfone/go-apiserver/http/upstream"
	"github.com/xgfone/go-apiserver/http/upstream/balancer"
	"github.com/xgfone/go-apiserver/log"
	"github.com/xgfone/go-apiserver/nets"
)

type discardWriter struct{}

func (w discardWriter) Write(p []byte) (int, error) { return len(p), nil }

type testServer struct{ url upstream.URL }

func (s testServer) ID() string                                          { return s.url.IP }
func (s testServer) URL() upstream.URL                                   { return s.url }
func (s testServer) Check(context.Context, upstream.URL) error           { return nil }
func (s testServer) State() (rs nets.RuntimeState)                       { return rs }
func (s testServer) HandleHTTP(http.ResponseWriter, *http.Request) error { return nil }

func newTestServer(ip string) testServer { return testServer{url: upstream.URL{IP: ip}} }

func BenchmarkLoadBalancer(b *testing.B) {
	log.DefaultLogger = log.NewLogger(discardWriter{}, "", 0, log.LvlAlert)
	lb := NewLoadBalancer("test", balancer.Random())
	lb.ResetServers(newTestServer("127.0.0.1"), newTestServer("127.0.0.2"))

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "http://127.0.0.1", nil)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		lb.ServeHTTP(rec, req)
	}
}
