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

package herrors

// NewCT is alias of WithCT.
//
// Deprecated! Please use WithCT.
func (e Error) NewCT(ct string) Error { return e.WithCT(ct) }

// New is alias of WithErr.
//
// Deprecated! Please use WithErr.
func (e Error) New(err error) Error { return e.WithErr(err) }

// Newf is alias of WithMsg.
//
// Deprecated! Please use WithMsg.
func (e Error) Newf(msg string, args ...interface{}) Error {
	return e.WithMsg(msg, args...)
}
