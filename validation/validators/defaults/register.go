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
	"github.com/xgfone/go-apiserver/validation/helper"
	"github.com/xgfone/go-apiserver/validation/validators"
)

func init() { RegisterDefaults(validation.DefaultBuilder) }

// RegisterDefaults registers the default validator building functions
// into the builder.
func RegisterDefaults(b *validation.Builder) {
	helper.RegisterFuncNoArg(validation.DefaultBuilder, "zero", validators.Zero)
	helper.RegisterFuncNoArg(validation.DefaultBuilder, "required", validators.Required)

	helper.RegisterFuncNoArg(validation.DefaultBuilder, "ip", validators.IP)
	helper.RegisterFuncNoArg(validation.DefaultBuilder, "mac", validators.Mac)
	helper.RegisterFuncNoArg(validation.DefaultBuilder, "cidr", validators.Cidr)
	helper.RegisterFuncNoArg(validation.DefaultBuilder, "addr", validators.Addr)

	helper.RegisterFuncOneFloat(validation.DefaultBuilder, "min", validators.Min)
	helper.RegisterFuncOneFloat(validation.DefaultBuilder, "max", validators.Max)
	helper.RegisterFuncStrings(validation.DefaultBuilder, "oneof", validators.OneOf)
	helper.RegisterFuncValidators(validation.DefaultBuilder, "array", validation.Array)
	helper.RegisterFuncValidators(validation.DefaultBuilder, "mapk", validation.MapK)
	helper.RegisterFuncValidators(validation.DefaultBuilder, "mapv", validation.MapV)
	helper.RegisterFuncValidators(validation.DefaultBuilder, "mapkv", validation.MapKV)
}
