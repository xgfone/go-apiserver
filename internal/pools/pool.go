// Copyright 2025 xgfone
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

package pools

import (
	"fmt"
	"strings"
	"sync"
)

func newBuilder(cap int) *strings.Builder {
	b := new(strings.Builder)
	b.Grow(cap)
	return b
}

var buildpool256 = sync.Pool{New: func() any { return newBuilder(256) }}

func GetBuilder(cap int) (pool *sync.Pool, builder *strings.Builder) {
	switch cap {
	case 256:
		pool = &buildpool256

	default:
		panic(fmt.Errorf("GetBuilder: unsupported cap %d", cap))
	}

	builder = buildpool256.Get().(*strings.Builder)
	builder.Reset()
	return
}

func PutBuilder(pool *sync.Pool, builder *strings.Builder) {
	builder.Reset()
	pool.Put(builder)
}
