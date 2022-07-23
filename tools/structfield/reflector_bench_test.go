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

package structfield

import (
	"reflect"
	"testing"
)

func BenchmarkReflector_Field0(b *testing.B) {
	sf := NewReflector()
	sf.RegisterSimpleFunc("noop", func(reflect.Value, interface{}) error { return nil })
	type S struct{}

	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			var s S
			sf.Reflect(nil, &s)
		}
	})
}

func BenchmarkReflector_Field1(b *testing.B) {
	sf := NewReflector()
	sf.RegisterSimpleFunc("noop", func(reflect.Value, interface{}) error { return nil })
	type S struct {
		F1 int `noop:"noop"`
	}

	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			var s S
			sf.Reflect(nil, &s)
		}
	})
}

func BenchmarkReflector_Field2(b *testing.B) {
	sf := NewReflector()
	sf.RegisterSimpleFunc("noop", func(reflect.Value, interface{}) error { return nil })
	type S struct {
		F1 int    `noop:"noop"`
		F2 string `noop:"noop"`
	}

	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			var s S
			sf.Reflect(nil, &s)
		}
	})
}
