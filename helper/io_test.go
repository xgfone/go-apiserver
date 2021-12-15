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

import (
	"bytes"
	"errors"
	"testing"
)

type errCloser struct{ err error }

func (c errCloser) Close() error { return c.err }

type bufCloser struct{ buf *bytes.Buffer }

func (c bufCloser) Close() { c.buf.WriteString("err") }

func TestClose(t *testing.T) {
	var err error
	if _err := Close(err); _err != nil {
		t.Error(_err)
	}

	err = errors.New("closer")
	e := errCloser{err: err}
	if _err := Close(e); err != _err {
		t.Errorf("expect error '%v', but got '%v'", err, _err)
	}

	buf := bytes.NewBuffer(nil)
	b := bufCloser{buf: buf}
	if _err := Close(b); _err != nil {
		t.Error(_err)
	} else if s := buf.String(); s != "err" {
		t.Errorf("expect '%s', but got '%s'", "err", s)
	}
}
