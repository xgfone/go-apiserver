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

package action

import (
	"fmt"
	"strings"
)

// Predefine some errors.
var (
	ErrInvalidAction        = NewError("InvalidAction", "invalid action")
	ErrInvalidVersion       = NewError("InvalidVersion", "invalid version")
	ErrInvalidParameter     = NewError("InvalidParams", "invalid parameter")
	ErrUnsupportedProtocol  = NewError("UnsupportedProtocol", "protocol is unsupported")
	ErrUnsupportedOperation = NewError("UnsupportedOperation", "operation is unsupported")

	ErrAuthFailureTokenFailure     = NewError("AuthFailure.TokenFailure", "token verification failed")
	ErrAuthFailureSignatureFailure = NewError("AuthFailure.SignatureFailure", "signature verification failed")
	ErrAuthFailureSignatureExpire  = NewError("AuthFailure.SignatureExpire", "signature is expired")
	ErrUnauthorizedOperation       = NewError("UnauthorizedOperation", "operation is unauthorized")

	ErrFailedOperation    = NewError("FailedOperation", "operation failed")
	ErrServerError        = NewError("ServerError", "server error")
	ErrGatewayTimeout     = NewError("GatewayTimeout", "gateway timeout")
	ErrServiceUnavailable = NewError("ServiceUnavailable", "service is unavailable")

	ErrQuotaLimitExceeded   = NewError("QuotaLimitExceeded", "exceed the quota limit")
	ErrRequestLimitExceeded = NewError("RequestLimitExceeded", "exceed the request limit")

	ErrInstanceInUse        = NewError("InstanceInUse", "instance is in use")
	ErrInstanceNotFound     = NewError("InstanceNotFound", "instance is not found")
	ErrInstanceUnavailable  = NewError("InstanceUnavailable", "instance is unavailable")
	ErrResourceInsufficient = NewError("ResourceInsufficient", "resource is insufficient")
)

// Error represents an error.
type Error struct {
	Code      string  `json:",omitempty" xml:",omitempty"`
	Message   string  `json:",omitempty" xml:",omitempty"`
	Component string  `json:",omitempty" xml:",omitempty"`
	Causes    []error `json:",omitempty" xml:",omitempty"`
}

// NewError returns a new Error.
func NewError(code, msg string) Error { return Error{Code: code, Message: msg} }

// Clone clones itself to a new one.
func (e Error) Clone() Error {
	ne := e
	if len(e.Causes) == 0 {
		ne.Causes = append([]error{}, e.Causes...)
	}
	return ne
}

// Error implements the interface error.
func (e Error) Error() string {
	if len(e.Message) == 0 {
		return e.Code
	}

	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// String implements the interface fmt.Stringer.
func (e Error) String() string {
	_len := len(e.Causes)
	if _len == 0 {
		if e.Component == "" {
			return fmt.Sprintf("code=%s, msg=%s", e.Code, e.Message)
		}
		return fmt.Sprintf("component=%s, code=%s, msg=%s",
			e.Component, e.Code, e.Message)
	}

	_causes := make([]string, _len)
	for _len--; _len >= 0; _len-- {
		_causes[_len] = e.Causes[_len].Error()
	}
	causes := strings.Join(_causes, " |> ")

	if e.Component == "" {
		return fmt.Sprintf("code=%s, msg=%s, causes=[%s]", e.Code, e.Message, causes)
	}
	return fmt.Sprintf("component=%s, code=%s, msg=%s, causes=[%s]",
		e.Component, e.Code, e.Message, causes)
}

// WithCode clones itself and returns a new Error with the code.
func (e Error) WithCode(code string) Error {
	ne := e.Clone()
	ne.Code = code
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
	ne := e.Clone()
	ne.Causes = append(ne.Causes, errs...)
	return ne
}

// AppendCauses appends the error causes into the original causes,
// and returns itself.
func (e Error) AppendCauses(errs ...error) Error {
	e.Causes = append(e.Causes, errs...)
	return e
}
