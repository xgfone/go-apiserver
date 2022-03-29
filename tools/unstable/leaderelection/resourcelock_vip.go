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

package leaderelection

import (
	"context"
	"errors"
	"net"
	"strings"
	"time"
)

// NewVipResourceLock a resource locker based on the vip.
func NewVipResourceLock(vip string) ResourceLock {
	return vipResourceLock{
		vip: vip,
		rsc: "vip:" + vip,
		hld: vip + "_fake",
	}
}

type vipResourceLock struct {
	vip string
	rsc string
	hld string
}

func (rl vipResourceLock) Resource() string { return rl.rsc }

func (rl vipResourceLock) Get(c context.Context) (r ElectionRecord, ok bool, err error) {
	switch err = vipIsOnHost(rl.vip); err {
	case nil:
		ok = true
		now := time.Now()
		r = ElectionRecord{
			HolderIdentity:       rl.vip,
			LeaseDurationSeconds: 10,
			AcquireTime:          now,
			RenewTime:            now,
		}

	case errNoVIP:
		err = nil
		ok = true
		now := time.Now()
		r = ElectionRecord{
			HolderIdentity:       rl.hld,
			LeaseDurationSeconds: 10,
			AcquireTime:          now,
			RenewTime:            now,
		}
	}

	return
}

func (rl vipResourceLock) Create(c context.Context, r ElectionRecord) error {
	return vipIsOnHost(rl.vip)
}

func (rl vipResourceLock) Update(c context.Context, r ElectionRecord) error {
	if r.HolderIdentity == "" {
		return nil
	}
	return vipIsOnHost(rl.vip)
}

func vipIsOnHost(vip string) error {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return err
	}

	for _, addr := range addrs {
		ip := addr.String()
		if index := strings.IndexByte(ip, '/'); index > -1 {
			ip = ip[:index]
		}

		if ip == vip {
			return nil
		}
	}

	return errNoVIP
}

var errNoVIP = errors.New("the vip is not on the host")
