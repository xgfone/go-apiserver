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
	"errors"

	mw "github.com/xgfone/go-apiserver/middleware"
	"github.com/xgfone/go-apiserver/nets"
	"github.com/xgfone/go-apiserver/tcp"
)

// IPWhitelist returns a tcp middleware to filter the connections
// that the client ip is not in the given ip or cidr list.
func IPWhitelist(priority int, ipOrCidrs ...string) (mw.Middleware, error) {
	if len(ipOrCidrs) == 0 {
		return nil, errors.New("TCP ClientIP middleware: no ips or cidrs")
	}

	checker, err := nets.NewIPCheckers(ipOrCidrs...)
	if err != nil {
		return nil, err
	}

	return mw.NewMiddleware("ip_whitelist", priority, func(h interface{}) interface{} {
		return tcp.NewIPWhitelistHandler(h.(tcp.Handler), checker)
	}), nil
}
