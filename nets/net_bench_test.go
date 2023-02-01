// Copyright 2022 xgfone
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

package nets

import (
	"net"
	"testing"
)

func BenchmarkIPChecker_IPv4(b *testing.B) {
	checker, err := NewIPChecker("1.2.3.4")
	if err != nil {
		panic(err)
	}

	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			checker.CheckIP(net.ParseIP("1.2.3.4"))
		}
	})
}

func BenchmarkIPChecker_CIDRv4(b *testing.B) {
	checker, err := NewIPChecker("10.1.3.4/8")
	if err != nil {
		panic(err)
	}

	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			checker.CheckIP(net.ParseIP("10.2.3.4"))
		}
	})
}

func BenchmarkIPChecker_IPv6(b *testing.B) {
	checker, err := NewIPChecker("1.2.3.4")
	if err != nil {
		panic(err)
	}

	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			checker.CheckIP(net.ParseIP("'fe80::215:5dff:fe8c:6de7"))
		}
	})
}

func BenchmarkIPChecker_CIDRv6(b *testing.B) {
	checker, err := NewIPChecker("fe80::/16")
	if err != nil {
		panic(err)
	}

	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			checker.CheckIP(net.ParseIP("'fe80::215:5dff:fe8c:6de7"))
		}
	})
}
