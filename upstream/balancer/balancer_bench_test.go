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

package balancer

import (
	"context"
	"net/http"
	"testing"

	"github.com/xgfone/go-apiserver/upstream"
)

func benchmarkBalancer(b *testing.B, balancer Balancer) {
	servers := upstream.Servers{
		newTestServer("127.0.0.1", 1),
		newTestServer("127.0.0.2", 2),
		newTestServer("127.0.0.3", 3),
		newTestServer("127.0.0.4", 4),
		newTestServer("127.0.0.5", 5),
		newTestServer("127.0.0.6", 6),
		newTestServer("127.0.0.7", 7),
		newTestServer("127.0.0.8", 8),
	}

	req, _ := http.NewRequest(http.MethodGet, "http://127.0.0.1", nil)
	req.RemoteAddr = "127.0.0.1"

	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			balancer.Forward(context.Background(), req, servers)
		}
	})
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

func BenchmarkSourceIPHash(b *testing.B) {
	benchmarkBalancer(b, SourceIPHash())
}

func BenchmarkLeastConn(b *testing.B) {
	benchmarkBalancer(b, LeastConn())
}
