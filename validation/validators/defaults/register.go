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

// Package defaults is used to register a set of the default validator building
// functions.
package defaults

import (
	"github.com/xgfone/go-apiserver/validation"
	"github.com/xgfone/go-apiserver/validation/validators"
)

func init() { RegisterDefaults(validation.DefaultBuilder) }

// RegisterDefaults registers the default validator building functions
// into the builder.
//
// When importing the package, it will register the default validator function
// into the default validation builder, that's validation.DefaultBuilder.
func RegisterDefaults(b *validation.Builder) {
	validation.RegisterFunction(validation.NewFunctionWithoutArgs("zero", validators.Zero))
	validation.RegisterFunction(validation.NewFunctionWithoutArgs("required", validators.Required))

	validation.RegisterFunction(validation.NewFunctionWithoutArgs("ip", validators.IP))
	validation.RegisterFunction(validation.NewFunctionWithoutArgs("mac", validators.Mac))
	validation.RegisterFunction(validation.NewFunctionWithoutArgs("cidr", validators.Cidr))
	validation.RegisterFunction(validation.NewFunctionWithoutArgs("addr", validators.Addr))

	validation.RegisterFunction(validation.NewFunctionWithOneFloat("min", validators.Min))
	validation.RegisterFunction(validation.NewFunctionWithOneFloat("max", validators.Max))
	validation.RegisterFunction(validation.NewFunctionWithTwoFloats("ranger", validators.Ranger))
	validation.RegisterFunction(validation.NewFunctionWithThreeInts("exp", validators.Exp))

	validation.RegisterFunction(validation.NewFunctionWithStrings("oneof", validators.OneOf))
	validation.RegisterFunction(validation.NewFunctionWithValidators("array", validation.Array))
	validation.RegisterFunction(validation.NewFunctionWithValidators("mapk", validation.MapK))
	validation.RegisterFunction(validation.NewFunctionWithValidators("mapv", validation.MapV))
	validation.RegisterFunction(validation.NewFunctionWithValidators("mapkv", validation.MapKV))
}
