// Copyright 2024 xgfone
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

package code

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/xgfone/go-apiserver/result"
)

var (
	_ CodeGetter[int]    = Error[int]{}
	_ CodeGetter[string] = Error[string]{}
)

// Error represents a code error.
type Error[T any] struct {
	Code    T      `json:"code,omitempty" yaml:"code,omitempty" xml:"code,omitempty"`
	Message string `json:"message,omitempty" yaml:"message,omitempty" xml:"message,omitempty"`

	Err error `json:"-" yaml:"-" xml:"-"`
	Ctx any   `json:"-" yaml:"-" xml:"-"`
}

// NewError returns a new Error.
func NewError[T any](code T, msg string) Error[T] { return Error[T]{Code: code, Message: msg} }

// Unwrap unwraps the inner error.
func (e Error[T]) Unwrap() error { return e.Err }

// Error implements the interface error.
func (e Error[T]) Error() string {
	if e.Message == "" {
		return fmt.Sprint(e.Code)
	}
	return fmt.Sprintf("%v: %s", e.Code, e.Message)
}

// String implements the interface fmt.Stringer.
func (e Error[T]) String() string {
	return fmt.Sprintf("code=%v, msg=%s", e.Code, e.Message)
}

// GetCode returns the error code.
func (e Error[T]) GetCode() T { return e.Code }

// WithCtx returns a new Error with the context.
func (e Error[T]) WithCtx(ctx any) Error[T] {
	e.Ctx = ctx
	return e
}

// WithError returns a new error, which inspects the error code and message from err.
func (e Error[T]) WithError(err error) error {
	switch _e := err.(type) {
	case nil:
		return nil

	case Error[T]:
		return _e

	case CodeGetter[T]:
		e.Err = err
		e.Code = _e.GetCode()
		e.Message = err.Error()

	default:
		e.Err = err
		e.Message = err.Error()
	}

	return e
}

// WithMessage returns a new Error with the message.
func (e Error[T]) WithMessage(msgfmt string, msgargs ...interface{}) Error[T] {
	if len(msgargs) == 0 {
		e.Message = msgfmt
	} else {
		e.Message = fmt.Sprintf(msgfmt, msgargs...)
	}
	return e
}

// Respond sends the error as result.Response by the responder.
func (e Error[T]) Respond(responder any) {
	result.Err(e).Respond(responder)
}

// Decode uses the decode function to decode the result to the error.
func (e *Error[T]) Decode(decode func(interface{}) error) error {
	return decode(e)
}

// DecodeJSON uses json decoder to decode from the reader into the error.
func (e *Error[T]) DecodeJSON(reader io.Reader) error {
	return json.NewDecoder(reader).Decode(e)
}

// DecodeJSONBytes uses json decoder to decode the []byte data into the error.
func (e *Error[T]) DecodeJSONBytes(data []byte) error {
	return json.Unmarshal(data, e)
}
