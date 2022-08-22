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

import "os"

func ExampleLogger() {
	// Set the default global logger.
	DefaultLogger = NewLogger(os.Stdout, "myapp: ", 0, LvlInfo)

	// Log the message by DefaultLogger
	DefaultLogger.Log(LvlDebug, 0, "log msg")
	DefaultLogger.Log(LvlInfo, 0, "log msg", "key1", "value1")
	DefaultLogger.Log(LvlWarn, 0, "log msg", "key1", "value1", "key2", "value2")

	// Log the message by the key-value log functions.
	Debug("log msg")
	Info("log msg", "key1", "value1")
	Warn("log msg", "key1", "value1", "key2", "value2")

	// Log the message by the format log functions.
	Debugf("log msg")
	Infof("log msg: %s=%s", "key1", "value1")
	Warnf("log msg: %s=%s, %s=%s", "key1", "value1", "key2", "value2")

	// Output:
	// myapp: log msg; level=info; key1=value1
	// myapp: log msg; level=warn; key1=value1; key2=value2
	// myapp: log msg; level=info; key1=value1
	// myapp: log msg; level=warn; key1=value1; key2=value2
	// myapp: log msg: key1=value1; level=info
	// myapp: log msg: key1=value1, key2=value2; level=warn
}
