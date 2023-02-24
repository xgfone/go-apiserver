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

package logger

import (
	"context"
	"testing"
)

func TestConfigOption(t *testing.T) {
	opts := []Option{
		SetLogReqBodyLen(func(context.Context) int { return 123 }),
		SetLogRespBodyLen(func(context.Context) int { return 456 }),
		SetLogRespHeaders(func(context.Context) bool { return true }),
		SetLogReqHeaders(func(context.Context) bool { return true }),
	}

	var config Config
	for _, opt := range opts {
		opt(&config)
	}

	if _len := config.GetLogReqBodyLen(context.Background()); _len != 123 {
		t.Errorf("expect %d, but got %d", 123, _len)
	}
	if _len := config.GetLogRespBodyLen(context.Background()); _len != 456 {
		t.Errorf("expect %d, but got %d", 456, _len)
	}
	if !config.GetLogReqHeaders(context.Background()) {
		t.Errorf("expect %v, but got %v", true, false)
	}
	if !config.GetLogRespHeaders(context.Background()) {
		t.Errorf("expect %v, but got %v", true, false)
	}
}
