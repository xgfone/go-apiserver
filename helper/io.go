// Copyright 2021 xgfone
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

package helper

import "io"

// Close closes the closer if it has implemented the interface io.Closer
// or interface{ Close() }.
func Close(closer interface{}) (err error) {
	switch v := closer.(type) {
	case io.Closer:
		err = v.Close()

	case interface{ Close() }:
		v.Close()

	case interface{ Unwrap() error }:
		err = Close(v.Unwrap())
	}

	return
}
