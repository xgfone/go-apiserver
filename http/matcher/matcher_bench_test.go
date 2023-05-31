// Copyright 2021~2022 xgfone
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

package matcher

import (
	"context"
	"net/http"
	"net/url"
	"testing"

	"github.com/xgfone/go-apiserver/http/reqresp"
)

func BenchmarkPath(b *testing.B) {
	matcher := Must(Path("/path"))
	path := "/path"
	req := &http.Request{URL: &url.URL{}}
	req.URL.Path = path

	c := reqresp.DefaultContextAllocator.Acquire()
	c.Request = req

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		if !matcher.Match(context.Background(), req) {
			panic("not match")
		}
	}
}

func BenchmarkPathParameter(b *testing.B) {
	matcher := Must(Path("/prefix/{id}"))
	path := "/prefix/123"
	req := &http.Request{URL: &url.URL{}}
	req.URL.Path = path

	c := reqresp.DefaultContextAllocator.Acquire()
	c.Request = req

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		if !matcher.Match(context.Background(), req) {
			panic("not match")
		}
	}
}

func BenchmarkClientIPMatcher(b *testing.B) {
	matcher := Must(ClientIP("10.0.0.0/8"))
	req := &http.Request{RemoteAddr: "10.1.2.3:80"}

	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			matcher.Match(nil, req)
		}
	})
}
