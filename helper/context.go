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

package helper

import "context"

type contextKeyType string

// NewContextWithValue is the same as context.WithValue(c, key, value),
// but use the string as the key.
func NewContextWithValue(c context.Context, key string, value interface{}) context.Context {
	return context.WithValue(c, contextKeyType(key), value)
}

// GetContextValue is the same as c.Value(key), but use the string as the key.
func GetContextValue(c context.Context, key string) (value interface{}) {
	return c.Value(contextKeyType(key))
}

// ContextCancelWithReason is the same context.WithCancel, but the cancel
// function receives an error reason.
func ContextCancelWithReason(parent context.Context) (context.Context, func(reason error)) {
	c := new(cancelContext)
	c.Context, c.CancelFunc = context.WithCancel(parent)
	return c, c.cancel
}
