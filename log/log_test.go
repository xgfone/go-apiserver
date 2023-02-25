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

package log

import (
	"fmt"
	"os"

	"golang.org/x/exp/slog"
)

func ExampleLogger() {
	disableTime = true

	// Set the default global logger.
	SetDefault(NewJSONHandler(os.Stdout, nil), slog.String("ctxkey", "ctxvalue"))

	// Log the message by the key-value log functions.
	Debug("log msg")
	Info("log msg", "key1", "value1")
	Warn("log msg", "key1", "value1", "key2", "value2")

	fmt.Println()

	// Log the message by the format log functions.
	Debugf("log msg")
	Infof("log msg: %s=%s", "key1", "value1")
	Warnf("log msg: %s=%s, %s=%s", "key1", "value1", "key2", "value2")

	fmt.Println()

	Log(0, LevelInfo, "log msg", "key", "value")

	// Output:
	//
	// {"level":"INFO","source":"github.com/xgfone/go-apiserver/log/log_test.go:32","msg":"log msg","ctxkey":"ctxvalue","key1":"value1"}
	// {"level":"WARN","source":"github.com/xgfone/go-apiserver/log/log_test.go:33","msg":"log msg","ctxkey":"ctxvalue","key1":"value1","key2":"value2"}
	//
	// {"level":"INFO","source":"github.com/xgfone/go-apiserver/log/log_test.go:39","msg":"log msg: key1=value1","ctxkey":"ctxvalue"}
	// {"level":"WARN","source":"github.com/xgfone/go-apiserver/log/log_test.go:40","msg":"log msg: key1=value1, key2=value2","ctxkey":"ctxvalue"}
	//
	// {"level":"INFO","source":"github.com/xgfone/go-apiserver/log/log_test.go:44","msg":"log msg","ctxkey":"ctxvalue","key":"value"}
}
