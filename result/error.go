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

package result

import (
	"encoding/json"
	"fmt"
	"io"
)

var _ CodeGetter = Error{}

// CodeError is used to convert itself to Error.
type CodeError interface {
	CodeError() Error
}

// Error represents an error.
type Error struct {
	Code    string `json:"code,omitempty" yaml:"code,omitempty" xml:"code,omitempty"`
	Message string `json:"message,omitempty" yaml:"message,omitempty" xml:"message,omitempty"`
	Data    any    `json:"data,omitempty" yaml:"data,omitempty" xml:"data,omitempty"`
	Err     error  `json:"-" yaml:"-" xml:"-"`
}

// NewError returns a new Error.
func NewError(code, msg string) Error { return Error{Code: code, Message: msg} }

// Unwrap unwraps the inner error.
func (e Error) Unwrap() error { return e.Err }

// Error implements the interface error.
func (e Error) Error() string {
	if e.Message == "" {
		return e.Code
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// String implements the interface fmt.Stringer.
func (e Error) String() string {
	return fmt.Sprintf("code=%s, msg=%s", e.Code, e.Message)
}

// WithMessage returns a new Error with the data.
func (e Error) WithData(data any) Error {
	e.Data = data
	return e
}

// WithError returns a new Error, which inspects the error code and message from err.
func (e Error) WithError(err error) Error {
	e.Err = err
	switch _e := err.(type) {
	case nil:
	case Error:
		e = _e

	case CodeGetter:
		e.Code = _e.GetCode()
		e.Message = err.Error()

	default:
		e.Message = err.Error()
	}

	return e
}

// WithMessage returns a new Error with the message.
func (e Error) WithMessage(msgfmt string, msgargs ...interface{}) Error {
	if len(msgargs) == 0 {
		e.Message = msgfmt
	} else {
		e.Message = fmt.Sprintf(msgfmt, msgargs...)
	}
	return e
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
