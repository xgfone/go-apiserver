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
	"encoding/binary"
	"net"
	"net/http"

	"github.com/xgfone/go-apiserver/http/upstream"
	"github.com/xgfone/go-apiserver/nets"
)

func init() {
	registerBuiltinBuidler("source_ip_hash", SourceIPHash)
}

// SourceIPHash returns a new balancer based on the source-ip hash.
//
// The policy name is "source_ip_hash".
func SourceIPHash() Balancer {
	random := newRandom()
	return NewBalancer("source_ip_hash",
		func(w http.ResponseWriter, r *http.Request, f func() upstream.Servers) error {
			ss := f()
			_len := len(ss)
			if _len == 1 {
				return ss[0].HandleHTTP(w, r)
			}

			var value uint64
			host, _ := nets.SplitHostPort(r.RemoteAddr)
			switch ip := net.ParseIP(host); len(ip) {
			case net.IPv4len:
				value = uint64(binary.BigEndian.Uint32(ip))
			case net.IPv6len:
				value = binary.BigEndian.Uint64(ip[8:16])
			default:
				value = uint64(random(_len))
			}

			return ss[value%uint64(_len)].HandleHTTP(w, r)
		})
}
