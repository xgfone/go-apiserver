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

package io2

import (
	"errors"
	"io"
	"testing"
)

type syncWriter struct{ err error }

func newSyncWriter(err error) io.Writer          { return syncWriter{err: err} }
func (w syncWriter) Write(b []byte) (int, error) { return len(b), nil }
func (w syncWriter) Sync() error                 { return w.err }

func TestSyncWriter(t *testing.T) {
	err1 := SyncWriter(NewSwitchWriter(NewSafeWriter(io.Discard)))
	if err1 != nil {
		t.Errorf("expect nil, but got an error: %v", err1)
	}

	err2 := SyncWriter(NewSwitchWriter(NewSafeWriter(newSyncWriter(nil))))
	if err2 != nil {
		t.Errorf("expect nil, but got an error: %v", err2)
	}

	err3 := SyncWriter(NewSwitchWriter(NewSafeWriter(newSyncWriter(errors.New("test")))))
	if err3 == nil {
		t.Errorf("expect an error, but got nil")
	} else if e := err3.Error(); e != "test" {
		t.Errorf("expect '%s', but got '%s'", "test", e)
	}
}
