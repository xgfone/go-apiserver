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

package balancer

import (
	"fmt"
)

// Builder is used to build a new Balancer with the config.
type Builder func(config interface{}) (Balancer, error)

var builders = make(map[string]Builder, 16)

func registerBuiltinBuidler(t string, f func() Balancer) {
	RegisterBuidler(t, func(interface{}) (Balancer, error) { return f(), nil })
}

// RegisterBuidler registers the given balancer builder.
//
// If the balancer builder typed "typ" has existed, override it to the new.
//
// For the builtin builders as following, they ignore the config parameter
// and never return an error.
//
//   - random
//   - round_robin
//   - weight_random
//   - weight_round_robin
//   - source_ip_hash
//   - least_conn
func RegisterBuidler(typ string, builder Builder) { builders[typ] = builder }

// GetBuilder returns the registered balancer builder by the type.
//
// If the balancer builder typed "typ" does not exist, return nil.
func GetBuilder(typ string) Builder { return builders[typ] }

// Build is a convenient function to build a new balancer typed "typ".
func Build(typ string, config interface{}) (balancer Balancer, err error) {
	if builder := GetBuilder(typ); builder != nil {
		balancer, err = builder(config)
	} else {
		err = fmt.Errorf("no the balancer builder typed '%s'", typ)
	}
	return
}
