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

package service

import (
	"context"

	"github.com/xgfone/go-apiserver/nets"
)

// NewVipChecker returns a new vip checker that checks whether the vip
// is bound to the given network interface named interfaceName.
//
// If interfaceName is empty, check all the network interfaces.
// If vip is empty, the checker always returns (true, nil).
func NewVipChecker(vip, interfaceName string) Checker {
	return vipChecker{vip: vip, iface: interfaceName}
}

type vipChecker struct {
	iface string
	vip   string
}

func (c vipChecker) Name() string { return "vip:" + c.vip }
func (c vipChecker) Check(context.Context) (ok bool, err error) {
	if c.vip == "" {
		return true, nil
	}
	return nets.IPIsOnInterface(c.vip, c.iface)
}
