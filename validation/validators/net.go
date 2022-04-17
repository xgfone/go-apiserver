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

package validators

import (
	"errors"
	"fmt"
	"net"

	"github.com/xgfone/go-apiserver/nets"
	"github.com/xgfone/go-apiserver/validation"
)

var (
	errInvalidMac = errors.New("the string is not a valid mac")

	errInvalidIP       = errors.New("invalid ip")
	errInvalidStringIP = errors.New("the string is not a valid ip")

	errInvalidCidr       = errors.New("invalid cidr")
	errInvalidStringCidr = errors.New("the string is not a valid cidr")
)

// Mac returns a new Validator to chech whether the string value is a valid MAC.
func Mac() validation.Validator {
	return validation.NewValidator("mac", func(i interface{}) error {
		if s, ok := i.(string); ok {
			if nets.NormalizeMac(s) == "" {
				return errInvalidMac
			}
		}
		return fmt.Errorf("expect a string, but got %T", i)
	})
}

// IP returns a new Validator to chech whether the value is a valid IP.
//
// Support the types: string or net.IP.
func IP() validation.Validator {
	return validation.NewValidator("ip", func(i interface{}) error {
		switch v := i.(type) {
		case string:
			if net.ParseIP(v) == nil {
				return errInvalidStringIP
			}

		case net.IP:
			switch len(v) {
			case net.IPv4len, net.IPv6len:
			default:
				return errInvalidIP
			}

		default:
			return fmt.Errorf("unsupported type %T", i)
		}

		return nil
	})
}

// Cidr returns a new Validator to chech whether the value is a valid cidr.
//
// Support the types: string or net.IPNet.
func Cidr() validation.Validator {
	return validation.NewValidator("cidr", func(i interface{}) error {
		switch v := i.(type) {
		case string:
			if _, _, err := net.ParseCIDR(v); err != nil {
				return errInvalidStringCidr
			}

		case *net.IPNet:
			if v == nil {
				return errInvalidCidr
			}

		default:
			return fmt.Errorf("unsupported type %T", i)
		}

		return nil
	})
}
