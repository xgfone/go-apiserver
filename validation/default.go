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

// RegisterDefaults registers the default symbols and validators
// building functions into the builder.
//
// The registered default symbols:
//
//	timelayout: 15:04:05
//	datelayout: 2006-01-02
//	datetimelayout: 2006-01-02 15:04:05
//
// The Signature of the registered validator functions as follow:
//
//	ip() or ip
//	mac() or mac
//	addr() or addr
//	cidr() or cidr
//	zero() or zero
//	isinteger() or isinteger
//	isnumber() or isnumber
//	duration() or duration
//	required() or required
//	structure() or structure
//	exp(base, startExp, endExp int)
//	min(float64)
//	max(float64)
//	ranger(min, max float64)
//	time(formatLayout string)
//	oneof(...string)
//	array(...Validator)
//	mapkv(...Validator)
//	mapk(...Validator)
//	mapv(...Validator)
//	timeformat() or timeformat => time(timelayout)
//	dateformat() or dateformat => time(datelayout)
//	datetimeformat() or datetimeformat => time(datetimelayout)
func RegisterDefaults(b *Builder) {
	b.RegisterSymbol("timelayout", "15:04:05")
	b.RegisterSymbol("datelayout", "2006-01-02")
	b.RegisterSymbol("datetimelayout", "2006-01-02 15:04:05")
	registerTimeValidator(b, "timeformat", "15:04:05")
	registerTimeValidator(b, "dateformat", "2006-01-02")
	registerTimeValidator(b, "datetimeformat", "2006-01-02 15:04:05")

	b.RegisterFunction(NewFunctionWithoutArgs("zero", validator.Zero))
	b.RegisterFunction(NewFunctionWithoutArgs("required", validator.Required))
	b.RegisterFunction(NewFunctionWithoutArgs("isnumber", validator.IsNumber))
	b.RegisterFunction(NewFunctionWithoutArgs("isinteger", validator.IsInteger))

	b.RegisterFunction(NewFunctionWithoutArgs("ip", validator.IP))
	b.RegisterFunction(NewFunctionWithoutArgs("mac", validator.Mac))
	b.RegisterFunction(NewFunctionWithoutArgs("cidr", validator.Cidr))
	b.RegisterFunction(NewFunctionWithoutArgs("addr", validator.Addr))

	b.RegisterFunction(NewFunctionWithOneFloat("min", validator.Min))
	b.RegisterFunction(NewFunctionWithOneFloat("max", validator.Max))
	b.RegisterFunction(NewFunctionWithTwoFloats("ranger", validator.Ranger))
	b.RegisterFunction(NewFunctionWithThreeInts("exp", validator.Exp))

	b.RegisterFunction(NewFunctionWithOneString("time", validator.Time))
	b.RegisterFunction(NewFunctionWithoutArgs("duration", validator.Duration))

	b.RegisterFunction(NewFunctionWithStrings("oneof", validator.OneOf))
	b.RegisterFunction(NewFunctionWithValidators("array", validator.Array))
	b.RegisterFunction(NewFunctionWithValidators("mapk", validator.MapK))
	b.RegisterFunction(NewFunctionWithValidators("mapv", validator.MapV))
	b.RegisterFunction(NewFunctionWithValidators("mapkv", validator.MapKV))

	// We use "structure" instead of "struct" because "struct" is the keyword in Go.
	b.RegisterValidatorFunc("structure", b.ValidateStruct)
	b.RegisterFunction(NewFunctionWithOneString("ltf", validator.StructFieldLess))
	b.RegisterFunction(NewFunctionWithOneString("lef", validator.StructFieldLessEqual))
	b.RegisterFunction(NewFunctionWithOneString("gtf", validator.StructFieldGreater))
	b.RegisterFunction(NewFunctionWithOneString("gef", validator.StructFieldGreaterEqual))
}

func registerTimeValidator(b *Builder, name, layout string) {
	b.RegisterFunction(NewFunctionWithoutArgs(name, func() validator.Validator {
		return validator.Time(layout)
	}))
}
