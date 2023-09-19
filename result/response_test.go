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

package result

import (
	"errors"
	"strings"
	"testing"
)

type respondfunc func(Response)

func (f respondfunc) Respond(r Response) { f(r) }

func TestResponse(t *testing.T) {
	resp := NewResponse(0, nil).WithData(123).WithError(errors.New("test"))
	resp.Respond(respondfunc(func(Response) {}))

	resp = Response{}
	err := resp.Decode(func(i interface{}) error {
		r := i.(*Response)
		r.Data = 123
		return nil
	})
	if err != nil {
		t.Error(err)
	} else if v, ok := resp.Data.(int); !ok {
		t.Errorf("expect an int, but got %T", resp.Data)
	} else if v != 123 {
		t.Errorf("expect %d, but got %d", 123, v)
	}

	err = resp.DecodeJSON(strings.NewReader(`{"data": 456}`))
	if err != nil {
		t.Error(err)
	} else if v, ok := resp.Data.(float64); !ok {
		t.Errorf("expect an float64, but got %T", resp.Data)
	} else if v != 456 {
		t.Errorf("expect %d, but got %v", 456, v)
	}

	err = resp.DecodeJSONBytes([]byte(`{"data": 789}`))
	if err != nil {
		t.Error(err)
	} else if v, ok := resp.Data.(float64); !ok {
		t.Errorf("expect an int, but got %T", resp.Data)
	} else if v != 789 {
		t.Errorf("expect %d, but got %v", 789, v)
	}
}
