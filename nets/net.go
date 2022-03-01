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
	"fmt"
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

// IPChecker is used to check whether the ip is legal or allowed.
type IPChecker interface {
	CheckIPString(ip string) (ok bool)
	fmt.Stringer
}

// IPCheckers is a set of IPChecker.
type IPCheckers []IPChecker

func (cs IPCheckers) String() string {
	_len := len(cs)
	if _len == 0 {
		return ""
	}

	var buf strings.Builder
	buf.Grow(128)
	for i := 0; i < _len; i++ {
		buf.WriteString(cs[i].String())
	}
	return buf.String()
}

// CheckIPString implements the interface IPChecker, which returns true
// if any ip checker return true.
func (cs IPCheckers) CheckIPString(ip string) bool {
	for i, _len := 0, len(cs); i < _len; i++ {
		if cs[i].CheckIPString(ip) {
			return true
		}
	}
	return false
}

// NewIPChecker returns a new IPChecker based on an IP or CIDR.
func NewIPChecker(ipOrCidr string) (IPChecker, error) {
	if strings.IndexByte(ipOrCidr, '/') > -1 { // For CIDR
		_, ipnet, err := net.ParseCIDR(ipOrCidr)
		if err != nil {
			return nil, fmt.Errorf("invalid cidr network address '%s'", ipOrCidr)
		}
		return ipChecker{ipnet}, nil
	}

	cidr := ipOrCidr
	if strings.IndexByte(cidr, '.') == -1 { // For IPv6
		cidr += "/128"
	} else { // For IPv4
		cidr += "/32"
	}

	_, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, fmt.Errorf("invalid ip address '%s'", ipOrCidr)
	}
	return ipChecker{ipnet}, nil
}

type ipChecker struct{ *net.IPNet }

func (c ipChecker) String() string { return c.IPNet.String() }

func (c ipChecker) CheckIPString(ip string) bool {
	_ip := net.ParseIP(ip)
	if _ip == nil {
		return false
	}
	return c.Contains(_ip)
}
