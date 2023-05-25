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
	"errors"
	"fmt"
	"strings"
)

var _ CodeGetter = Error{}

// CodeError is used to convert itself to Error.
type CodeError interface {
	CodeError() Error
}

// Error represents an error.
type Error struct {
	Code      string  `json:",omitempty" yaml:",omitempty" xml:",omitempty"`
	Message   string  `json:",omitempty" yaml:",omitempty" xml:",omitempty"`
	Component string  `json:",omitempty" yaml:",omitempty" xml:",omitempty"`
	Causes    []error `json:",omitempty" yaml:",omitempty" xml:",omitempty"`
	Err       error   `json:"-" yaml:"-" xml:"-"`
}

// NewError returns a new Error.
func NewError(code, msg string) Error { return Error{Code: code, Message: msg} }

// Clone clones itself to a new one.
func (e Error) Clone() Error {
	if len(e.Causes) > 0 {
		e.Causes = append([]error{}, e.Causes...)
	}
	return e
}

// As finds the first error in the error that matches target,
// and if one is found, sets target to that error value and returns true.
// Otherwise, it returns false.
func (e Error) As(target interface{}) bool {
	if e.Err != nil && errors.As(e.Err, target) {
		return true
	}

	for _, err := range e.Causes {
		if errors.As(err, target) {
			return true
		}
	}

	return false
}

// Is reports whether the error matches target.
func (e Error) Is(target error) bool {
	if e.Err != nil && errors.Is(e.Err, target) {
		return true
	}

	for _, err := range e.Causes {
		if errors.Is(err, target) {
			return true
		}
	}

	return false
}

// Unwrap unwraps the inner error.
func (e Error) Unwrap() error { return e.Err }

// IsCode is equal to IsCode(e.Code, target).
func (e Error) IsCode(target string) bool { return IsCode(e.Code, target) }

// GetCode returns the error code.
func (e Error) GetCode() string { return e.Code }

// GetMessage returns the error message.
func (e Error) GetMessage() string {
	if e.Message != "" {
		return e.Message
	} else if e.Err != nil {
		return e.Err.Error()
	}
	return ""
}

// Error implements the interface error.
func (e Error) Error() string {
	if msg := e.GetMessage(); len(msg) > 0 {
		return fmt.Sprintf("%s: %s", e.Code, msg)
	}

	return e.Code
}

// String implements the interface fmt.Stringer.
func (e Error) String() string {
	msg := e.GetMessage()
	_len := len(e.Causes)
	if _len == 0 {
		if e.Component == "" {
			return fmt.Sprintf("code=%s, msg=%s", e.Code, msg)
		}
		return fmt.Sprintf("component=%s, code=%s, msg=%s",
			e.Component, e.Code, msg)
	}

	_causes := make([]string, _len)
	for _len--; _len >= 0; _len-- {
		_causes[_len] = e.Causes[_len].Error()
	}
	causes := strings.Join(_causes, " |> ")

	if e.Component == "" {
		return fmt.Sprintf("code=%s, msg=%s, causes=[%s]", e.Code, msg, causes)
	}
	return fmt.Sprintf("component=%s, code=%s, msg=%s, causes=[%s]",
		e.Component, e.Code, msg, causes)
}

// WithCode clones itself and returns a new Error with the code.
func (e Error) WithCode(code string) Error {
	ne := e.Clone()
	ne.Code = code
	return ne
}

// WithError clones itself and returns the new one, which will inspect
// the error code and message from the error.
func (e Error) WithError(err error) Error {
	ne := e.Clone()
	ne.Err = err
	switch ce := err.(type) {
	case nil:
	case Error:
		ne = ce

	case interface {
		CodeGetter
		GetMessage() string
	}:
		ne.Code = ce.GetCode()
		ne.Message = ce.GetMessage()

	case CodeGetter:
		ne.Code = ce.GetCode()
		ne.Message = err.Error()

	default:
		ne.Message = err.Error()
	}

	return ne
}

// WithMessage clones itself and returns a new Error with the message.
func (e Error) WithMessage(msgfmt string, msgargs ...interface{}) Error {
	ne := e.Clone()
	if len(msgargs) == 0 {
		ne.Message = msgfmt
	} else {
		ne.Message = fmt.Sprintf(msgfmt, msgargs...)
	}
	return ne
}

// WithComponent clones itself and returns a new Error with the component.
func (e Error) WithComponent(component string) Error {
	ne := e.Clone()
	ne.Component = component
	return ne
}

// WithCauses clones itself and returns a new Error appending the errors.
func (e Error) WithCauses(errs ...error) Error {
	if len(errs) == 0 {
		return e
	}

	ne := e.Clone()
	ne.Causes = append(ne.Causes, errs...)
	return ne
}

// AppendCauses appends the error causes into the original causes,
// and returns itself.
func (e Error) AppendCauses(errs ...error) Error {
	if len(errs) > 0 {
		e.Causes = append(e.Causes, errs...)
	}
	return e
}
