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

// Package nets provides some convenient net functions.
package nets

import (
	"errors"
	"net"
	"strings"
)

type timeoutError interface {
	Timeout() bool // Is the error a timeout?
	error
}

// IsTimeout reports whether the error is timeout.
func IsTimeout(err error) bool {
	var timeoutErr timeoutError
	return errors.As(err, &timeoutErr) && timeoutErr.Timeout()
}

// IsClosed reports whether the error is closed.
func IsClosed(err error) bool { return errors.Is(err, net.ErrClosed) }

// NormalizeMac normalizes the mac, which is the convenient function
// of net.ParseMAC, but only supports the 48-bit format and outputs
// the string like "xx:xx:xx:xx:xx:xx".
//
// Return "" if the mac is an invalid mac.
func NormalizeMac(mac string) string {
	if ha, err := net.ParseMAC(mac); err == nil || len(ha) == 6 {
		return ha.String()
	}
	return ""
}

// IPIsOnInterface reports whether the ip is on the given network interface
// named ifaceName.
//
// If ip is empty or invalid, return false.
// If ifaceName is empty, it checks all the network interfaces.
func IPIsOnInterface(ip, ifaceName string) (on bool, err error) {
	netip := net.ParseIP(strings.TrimSpace(ip))
	if netip == nil {
		return false, nil
	}

	var addrs []net.Addr
	var iface *net.Interface
	if ifaceName == "" {
		addrs, err = net.InterfaceAddrs()
	} else if iface, err = net.InterfaceByName(ifaceName); err == nil {
		addrs, err = iface.Addrs()
	}

	if err != nil {
		return
	}

	ip = netip.String()
	for _, addr := range addrs {
		if strings.Split(addr.String(), "/")[0] == ip {
			return true, nil
		}
	}

	return false, nil
}

// SplitHostPort separates host and port. If the port is not valid, it returns
// the entire input as host, and it doesn't check the validity of the host.
// Unlike net.SplitHostPort, but per RFC 3986, it requires ports to be numeric.
func SplitHostPort(hostport string) (host, port string) {
	host = hostport

	colon := strings.LastIndexByte(host, ':')
	if colon != -1 && validOptionalPort(host[colon:]) {
		host, port = host[:colon], host[colon+1:]
	}

	if strings.HasPrefix(host, "[") && strings.HasSuffix(host, "]") {
		host = host[1 : len(host)-1]
	}

	return
}

// validOptionalPort reports whether port is either an empty string
// or matches /^:\d*$/
func validOptionalPort(port string) bool {
	if port == "" {
		return true
	}
	if port[0] != ':' {
		return false
	}
	for _, b := range port[1:] {
		if b < '0' || b > '9' {
			return false
		}
	}
	return true
}
