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

// Package codeint provides an error based on the integer code.
package codeint

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"

	"github.com/xgfone/go-apiserver/result"
)

var _ error = Error{}

// Error is used to stand for an error based the integer code.
type Error struct {
	Data    any    `json:",omitempty"`
	Code    int    `json:",omitempty"`
	Message string `json:",omitempty"`

	Err error `json:"-"`
	Ctx any   `json:"-"`

	Status int `json:"-"`
}

// NewError returns a new Error with the code.
func NewError(code int) Error { return Error{Code: code}.WithStatus(code) }

// IsZero reports whether e is ZERO.
func (e Error) IsZero() bool {
	return e.Code == 0 && e.Message == ""
}

// Unwrap returns the wrapped error.
func (e Error) Unwrap() error {
	return e.Err
}

// Error implements the interface error.
func (e Error) Error() string {
	if e.Message == "" {
		return strconv.FormatInt(int64(e.Code), 10)
	}
	return fmt.Sprintf("%d: %s", e.Code, e.Message)
}

// String implements the interface fmt.Stringer.
func (e Error) String() string {
	if e.Data == nil {
		return fmt.Sprintf("code=%d, msg=%s", e.Code, e.Message)
	}
	return fmt.Sprintf("code=%d, msg=%s, data=%v", e.Code, e.Message, e.Data)
}

// WithCtx returns a new Error with the context information.
func (e Error) WithCtx(ctx any) Error {
	e.Ctx = ctx
	return e
}

// WithCode returns a new code with the status.
func (e Error) WithCode(code int) Error {
	e.Code = code
	return e
}

// WithData returns a new Error with the data.
func (e Error) WithData(data any) Error {
	e.Data = data
	return e
}

// WithError returns a new Error with the error.
func (e Error) WithError(err error) Error {
	e.Message = err.Error()
	e.Err = err
	return e
}

// WithMessage returns a new Error with the message.
func (e Error) WithMessage(msg string, args ...interface{}) Error {
	if len(args) == 0 {
		e.Message = msg
	} else {
		e.Message = fmt.Sprintf(msg, args...)
	}
	return e
}

// WithStatus returns a new Error with the status.
func (e Error) WithStatus(status int) Error {
	if status < 600 {
		e.Status = status
	} else {
		e.Status = 500
	}
	return e
}

// TryError tries to assert err to Error and return it.
// Or, wrap it and return a new Error.
func (e Error) TryError(err error) Error {
	if _err, ok := err.(Error); ok {
		return _err
	}
	return e.WithError(err)
}

// ToError is the same as TryError if err != nil. Or, return nil.
func (e Error) ToError(err error) error {
	if err != nil {
		err = e.TryError(err)
	}
	return err
}

// Respond sends the error as result.Response by the responder.
func (e Error) Respond(responder any) {
	result.Err(e).Respond(responder)
}

// Decode uses the decode function to decode the result to the error.
func (e *Error) Decode(decode func(interface{}) error) error {
	return decode(e)
}

// DecodeJSON uses json decoder to decode from the reader into the error.
func (e *Error) DecodeJSON(reader io.Reader) error {
	return json.NewDecoder(reader).Decode(e)
}

// DecodeJSONBytes uses json decoder to decode the []byte data into the error.
func (e *Error) DecodeJSONBytes(data []byte) error {
	return json.Unmarshal(data, e)
}
