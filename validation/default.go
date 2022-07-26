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

package validation

import "github.com/xgfone/go-apiserver/validation/validator"

func init() { RegisterDefaults(DefaultBuilder) }

// RegisterDefaults registers the default validator building functions
// into the builder.
func RegisterDefaults(b *Builder) {
	b.RegisterFunction(NewFunctionWithoutArgs("zero", validator.Zero))
	b.RegisterFunction(NewFunctionWithoutArgs("required", validator.Required))

	b.RegisterFunction(NewFunctionWithoutArgs("ip", validator.IP))
	b.RegisterFunction(NewFunctionWithoutArgs("mac", validator.Mac))
	b.RegisterFunction(NewFunctionWithoutArgs("cidr", validator.Cidr))
	b.RegisterFunction(NewFunctionWithoutArgs("addr", validator.Addr))

	b.RegisterFunction(NewFunctionWithOneFloat("min", validator.Min))
	b.RegisterFunction(NewFunctionWithOneFloat("max", validator.Max))
	b.RegisterFunction(NewFunctionWithTwoFloats("ranger", validator.Ranger))
	b.RegisterFunction(NewFunctionWithThreeInts("exp", validator.Exp))

	b.RegisterFunction(NewFunctionWithStrings("oneof", validator.OneOf))
	b.RegisterFunction(NewFunctionWithValidators("array", validator.Array))
	b.RegisterFunction(NewFunctionWithValidators("mapk", validator.MapK))
	b.RegisterFunction(NewFunctionWithValidators("mapv", validator.MapV))
	b.RegisterFunction(NewFunctionWithValidators("mapkv", validator.MapKV))

	// We use "structure" instead of "struct" because "struct" is the keyword in Go.
	b.RegisterValidatorFunc("structure", b.ValidateStruct)
}
