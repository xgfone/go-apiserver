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

package middleware

// DefaultBuilderManager is the default global middleware buidler manager.
var DefaultBuilderManager = NewBuilderManager()

// Build is equal to DefaultBuilderManager.Build(typ, name, conf).
func Build(typ, name string, conf map[string]interface{}) (Middleware, error) {
	return DefaultBuilderManager.Build(typ, name, conf)
}

// RegisterBuilder is equal to DefaultBuilderManager.RegisterBuilder(typ, builder).
func RegisterBuilder(typ string, builder Builder) (err error) {
	return DefaultBuilderManager.RegisterBuilder(typ, builder)
}

// UnregisterBuilder is equal to DefaultBuilderManager.UnregisterBuilder(typ).
func UnregisterBuilder(typ string) {
	DefaultBuilderManager.UnregisterBuilder(typ)
}

// GetBuilder is equal to DefaultBuilderManager.GetBuilder(typ).
func GetBuilder(typ string) Builder {
	return DefaultBuilderManager.GetBuilder(typ)
}

// GetBuilders is equal to DefaultBuilderManager.GetBuilders().
func GetBuilders() map[string]Builder {
	return DefaultBuilderManager.GetBuilders()
}
