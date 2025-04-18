// Copyright 2024~2025 xgfone
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
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/xgfone/go-apiserver/internal/pools"
	"github.com/xgfone/go-apiserver/result"
	"github.com/xgfone/go-toolkit/jsonx"
)

var _ error = Error{}

// Error is used to stand for an error based the integer code.
type Error struct {
	Data    any    `json:",omitempty"`
	Code    int    `json:",omitempty"`
	Reason  string `json:",omitempty"`
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
	switch {
	case e.Err != nil:
		return e.Err.Error()

	case e.Reason != "":
		return e.Reason

	case e.Message != "":
		return e.Message

	default:
		if e.Status > 0 {
			return fmt.Sprintf("status=%d, code=%d", e.Status, e.Code)
		}
		return fmt.Sprintf("code=%d", e.Code)
	}
}

// String implements the interface fmt.Stringer.
func (e Error) String() string {
	pool, buf := pools.GetBuffer(256)
	defer pools.PutBuffer(pool, buf)

	_, _ = fmt.Fprintf(buf, "code=%d", e.Code)

	if e.Message != "" {
		_, _ = fmt.Fprintf(buf, ", msg=%s", e.Message)
	}
	if e.Reason != "" {
		_, _ = fmt.Fprintf(buf, ", reason=%s", e.Reason)
	}
	if e.Data != nil {
		_, _ = fmt.Fprintf(buf, ", data=%v", e.Data)
	}

	return buf.String()
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
//
// If err is nil, it will clear Reason and Err to ZERO.
func (e Error) WithError(err error) Error {
	if err == nil {
		e.Reason = ""
		e.Err = nil
	} else {
		e.Reason = err.Error()
		e.Err = err
	}
	return e
}

// WithReason returns a new Error with the reason.
func (e Error) WithReason(reason string) Error {
	e.Reason = reason
	return e
}

// WithReasonf returns a new Error with the reason formatted by fmt.Sprintf(reason, args...).
func (e Error) WithReasonf(format string, args ...any) Error {
	return e.WithReason(fmt.Sprintf(format, args...))
}

// WithMessage returns a new Error with the message.
func (e Error) WithMessage(msg string) Error {
	e.Message = msg
	return e
}

// WithMessagef returns a new Error with the message formatted by fmt.Sprintf(msg, args...).
func (e Error) WithMessagef(msg string, args ...any) Error {
	return e.WithMessage(fmt.Sprintf(msg, args...))
}

// WithStatus returns a new Error with the status.
func (e Error) WithStatus(status int) Error {
	if status < 600 {
		e.Status = status
	} else {
		e.Status = 500
	}

	if e.Message == "" {
		e.Message = http.StatusText(status)
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
func (e *Error) Decode(decode func(any) error) error {
	return decode(e)
}

// DecodeJSON uses json decoder to decode from the reader into the error.
func (e *Error) DecodeJSON(reader io.Reader) error {
	return jsonx.UnmarshalReader(e, reader)
}

// DecodeJSONBytes uses json decoder to decode the []byte data into the error.
func (e *Error) DecodeJSONBytes(data []byte) error {
	return jsonx.UnmarshalReader(e, bytes.NewReader(data))
}
