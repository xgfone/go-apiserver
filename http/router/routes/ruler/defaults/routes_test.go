// Copyright 2023 xgfone
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

package defaults

import (
	"bytes"
	"strings"
	"testing"

	"github.com/xgfone/go-apiserver/http/router/routes/ruler"
)

func TestDefaultRoutes(t *testing.T) {
	expect := strings.Join([]string{
		"(Path(`/debug/router/action/actions`) && Method(`GET`))",
		"(Path(`/debug/router/ruler/routes`) && Method(`GET`))",
		"(Path(`/debug/pprof/threadcreate`) && Method(`GET`))",
		"(Path(`/debug/pprof/goroutine`) && Method(`GET`))",
		"(Path(`/debug/pprof/profile`) && Method(`GET`))",
		"(Path(`/debug/pprof/cmdline`) && Method(`GET`))",
		"(Path(`/debug/pprof/symbol`) && Method(`GET`))",
		"(Path(`/debug/pprof/allocs`) && Method(`GET`))",
		"(Path(`/debug/pprof/trace`) && Method(`GET`))",
		"(Path(`/debug/pprof/mutex`) && Method(`GET`))",
		"(Path(`/debug/pprof/block`) && Method(`GET`))",
		"(Path(`/debug/pprof/heap`) && Method(`GET`))",
		"(Path(`/debug/pprof/`) && Method(`GET`))",
		"(Path(`/debug/pprof`) && Method(`GET`))",
		"(Path(`/debug/vars`) && Method(`GET`))",
		"",
	}, "\n")

	buf := bytes.NewBuffer(nil)
	for _, route := range ruler.DefaultRouter.GetRoutes() {
		buf.WriteString(route.Name)
		buf.WriteByte('\n')
	}

	if s := buf.String(); s != expect {
		t.Errorf("expect '%s', but got '%s'", expect, s)
	}
}
