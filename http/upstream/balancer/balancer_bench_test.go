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

package balancer

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/xgfone/go-apiserver/http/upstream"
)

func benchmarkBalancer(b *testing.B, balancer Balancer) {
	servers := upstream.Servers{
		newTestServer("127.0.0.1", 1),
		newTestServer("127.0.0.2", 1),
		newTestServer("127.0.0.3", 1),
		newTestServer("127.0.0.4", 1),
		newTestServer("127.0.0.5", 1),
		newTestServer("127.0.0.6", 1),
		newTestServer("127.0.0.7", 1),
		newTestServer("127.0.0.8", 1),
	}

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "http://127.0.0.1", nil)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		balancer.Forward(rec, req, servers)
	}
}

func BenchmarkSourceIPHash(b *testing.B) {
	benchmarkBalancer(b, SourceIPHash())
}

func BenchmarkRandom(b *testing.B) {
	benchmarkBalancer(b, Random())
}

func BenchmarkRoundRobin(b *testing.B) {
	benchmarkBalancer(b, RoundRobin())
}

func BenchmarkWeightedRandom(b *testing.B) {
	benchmarkBalancer(b, WeightedRandom())
}

func BenchmarkWeightedRoundRobin(b *testing.B) {
	benchmarkBalancer(b, WeightedRoundRobin())
}

func BenchmarkLeastConn(b *testing.B) {
	benchmarkBalancer(b, LeastConn())
}
