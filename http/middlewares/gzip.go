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

package middlewares

import (
	"compress/gzip"
	"net/http"
	"strings"
	"sync"

	"github.com/xgfone/go-apiserver/http/header"
	"github.com/xgfone/go-apiserver/http/reqresp"
	"github.com/xgfone/go-apiserver/middleware"
	"github.com/xgfone/go-apiserver/nets"
)

// Gzip returns a middleware to compress the response body by GZIP.
//
// level is the compression level, range [-1, 9].
//
// domains is the host domains enabling the gzip compression.
// which supports the exact, prefix and suffix match. For example,
//   - Exact:  www.example.com
//   - Prefix: www.example.*
//   - Suffix: *.example.com
//
// If empty, compress all the requests to all the host domains.
//
// Notice:
//  1. the returned gzip middleware will always compress it,
//     no matter whether the response body is empty or not.
//  2. the gzip middleware must be the last to handle the response.
//     If returning an error stands for the failure result, therefore,
//     it should be handled before compressing the response body,
//     that's, the error handler middleware must be appended
//     after the GZip middleware.
func Gzip(priority, level int, domains ...string) middleware.Middleware {
	if _, err := gzip.NewWriterLevel(nil, level); err != nil {
		panic(err)
	}

	gpool := sync.Pool{New: func() interface{} {
		w, _ := gzip.NewWriterLevel(nil, level)
		return w
	}}

	releaseGzipResponse := func(w *gzip.Writer) { w.Close(); gpool.Put(w) }
	acquireGzipResponse := func(w http.ResponseWriter) (gw *gzip.Writer) {
		gw = gpool.Get().(*gzip.Writer)
		gw.Reset(w)
		return
	}

	var exactDomains []string
	var prefixDomains []string
	var suffixDomains []string
	for _, domain := range domains {
		if domain == "" {
			panic("GZip: empty domain")
		} else if strings.HasPrefix(domain, "*.") {
			suffixDomains = append(suffixDomains, domain[1:])
		} else if strings.HasSuffix(domain, ".*") {
			prefixDomains = append(prefixDomains, domain[:len(domain)-1])
		} else {
			exactDomains = append(exactDomains, domain)
		}
	}

	noDomain := len(domains) == 0
	matchDomain := func(host string) bool {
		for i, _len := 0, len(exactDomains); i < _len; i++ {
			if exactDomains[i] == host {
				return true
			}
		}
		for i, _len := 0, len(prefixDomains); i < _len; i++ {
			if strings.HasPrefix(host, prefixDomains[i]) {
				return true
			}
		}
		for i, _len := 0, len(suffixDomains); i < _len; i++ {
			if strings.HasSuffix(host, suffixDomains[i]) {
				return true
			}
		}
		return false
	}

	return middleware.NewMiddleware("gzip", priority, func(h interface{}) interface{} {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.Contains(r.Header.Get(header.HeaderAcceptEncoding), "gzip") {
				respHeader := w.Header()
				if noDomain || matchDomain(splitHost(r.Host)) {
					respHeader.Add(header.HeaderVary, header.HeaderAcceptEncoding)
					respHeader.Set(header.HeaderContentEncoding, "gzip")

					gw := acquireGzipResponse(w)
					defer releaseGzipResponse(gw)
					w = reqresp.NewResponseWriter(w, reqresp.Write(gw.Write))
				}
			}

			h.(http.Handler).ServeHTTP(w, r)
		})
	})
}

func splitHost(hostport string) (host string) {
	host, _ = nets.SplitHostPort(hostport)
	return
}
