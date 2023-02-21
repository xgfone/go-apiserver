// Copyright 2021~2023 xgfone
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
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/xgfone/go-apiserver/log"
	"github.com/xgfone/go-apiserver/nets"
	"github.com/xgfone/go-apiserver/upstream"
	"github.com/xgfone/go-apiserver/upstream/balancer"
)

var _ upstream.Server = new(testServer)

type testServer struct {
	ip string
}

func (s *testServer) ID() string                           { return s.ip }
func (s *testServer) Type() string                         { return "" }
func (s *testServer) Info() interface{}                    { return nil }
func (s *testServer) Check(context.Context) error          { return nil }
func (s *testServer) Update(interface{}) error             { return nil }
func (s *testServer) Status() upstream.ServerStatus        { return upstream.ServerStatusOnline }
func (s *testServer) RuntimeState() (rs nets.RuntimeState) { return rs }
func (s *testServer) Serve(c context.Context, r interface{}) error {
	return nil
}

func newTestServer(ip string) *testServer { return &testServer{ip: ip} }

func BenchmarkLoadBalancer(b *testing.B) {
	log.SetDefault(nil, log.NewJSONHandler(io.Discard, nil))
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
