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

// Predefine some error codes.
const (
	CodeInvalidAction        = "InvalidAction"
	CodeInvalidVersion       = "InvalidVersion"
	CodeInvalidParams        = "InvalidParams"
	CodeUnsupportedProtocol  = "UnsupportedProtocol"
	CodeUnsupportedOperation = "UnsupportedOperation"

	CodeAuthFailureTokenFailure     = "AuthFailure.TokenFailure"
	CodeAuthFailureSignatureFailure = "AuthFailure.SignatureFailure"
	CodeAuthFailureSignatureExpire  = "AuthFailure.SignatureExpire"
	CodeUnauthorizedOperation       = "UnauthorizedOperation"

	CodeFailedOperation     = "FailedOperation"
	CodeInternalServerError = "InternalServerError"
	CodeGatewayTimeout      = "GatewayTimeout"
	CodeServiceUnavailable  = "ServiceUnavailable"

	CodeQuotaLimitExceeded   = "QuotaLimitExceeded"
	CodeRequestLimitExceeded = "RequestLimitExceeded"

	CodeInstanceInUse        = "InstanceInUse"
	CodeInstanceNotFound     = "InstanceNotFound"
	CodeInstanceUnavailable  = "InstanceUnavailable"
	CodeInstanceInconsistent = "InstanceInconsistent"
	CodeResourceInsufficient = "ResourceInsufficient"
	CodeBalanceInsufficient  = "BalanceInsufficient"
)

// Predefine some errors.
var (
	ErrInvalidAction        = NewError(CodeInvalidAction, "invalid action")
	ErrInvalidVersion       = NewError(CodeInvalidVersion, "invalid version")
	ErrInvalidParameter     = NewError(CodeInvalidParams, "invalid parameter")
	ErrUnsupportedProtocol  = NewError(CodeUnsupportedProtocol, "protocol is unsupported")
	ErrUnsupportedOperation = NewError(CodeUnsupportedOperation, "operation is unsupported")

	ErrAuthFailureTokenFailure     = NewError(CodeAuthFailureTokenFailure, "token verification failed")
	ErrAuthFailureSignatureFailure = NewError(CodeAuthFailureSignatureFailure, "signature verification failed")
	ErrAuthFailureSignatureExpire  = NewError(CodeAuthFailureSignatureExpire, "signature is expired")
	ErrUnauthorizedOperation       = NewError(CodeUnauthorizedOperation, "operation is unauthorized")

	ErrFailedOperation     = NewError(CodeFailedOperation, "operation failed")
	ErrInternalServerError = NewError(CodeInternalServerError, "internal server error")
	ErrGatewayTimeout      = NewError(CodeGatewayTimeout, "gateway timeout")
	ErrServiceUnavailable  = NewError(CodeServiceUnavailable, "service is unavailable")

	ErrQuotaLimitExceeded   = NewError(CodeQuotaLimitExceeded, "exceed the quota limit")
	ErrRequestLimitExceeded = NewError(CodeRequestLimitExceeded, "exceed the request limit")

	ErrInstanceInUse        = NewError(CodeInstanceInUse, "instance is in use")
	ErrInstanceNotFound     = NewError(CodeInstanceNotFound, "instance is not found")
	ErrInstanceUnavailable  = NewError(CodeInstanceUnavailable, "instance is unavailable")
	ErrInstanceInconsistent = NewError(CodeInstanceInconsistent, "instance is inconsistent")
	ErrResourceInsufficient = NewError(CodeResourceInsufficient, "resource is insufficient")
	ErrBalanceInsufficient  = NewError(CodeBalanceInsufficient, "balance is insufficient")
)

// IsCode reports whether the code is equal to or the child-code of target.
//
// Example
//
//   IsCode("InstanceNotFound", "InstanceNotFound")    // => true
//   IsCode("InstanceNotFound", "InstanceUnavailable") // => false
//   IsCode("AuthFailure.TokenFailure", "AuthFailure") // => true
//   IsCode("AuthFailure", "AuthFailure.TokenFailure") // => false
//
func IsCode(code, target string) bool {
	if code == target {
		return true
	}

	minlen := len(target)
	return len(code) > minlen && code[minlen] == '.' && code[:minlen] == target
}

// ErrorIsCode reports whether the code of the error is equal to or the child-code
// of the target code.
//
// If err has not implemented the interface CodeGetter, return false.
func ErrorIsCode(err error, targetCode string) bool {
	if c, ok := err.(CodeGetter); ok {
		return IsCode(c.GetCode(), targetCode)
	}
	return false
}

// GetCode gets the error code from the error if it implements
// the interface CodeGetter. Or returns "".
func GetCode(err error) string {
	if c, ok := err.(CodeGetter); ok {
		return c.GetCode()
	}
	return ""
}

// CodeGetter is an interface used to get the error code.
type CodeGetter interface {
	GetCode() string
}

var _ CodeGetter = Error{}

// Error represents an error.
type Error struct {
	Code      string  `json:",omitempty" yaml:",omitempty" xml:",omitempty"`
	Message   string  `json:",omitempty" yaml:",omitempty" xml:",omitempty"`
	Component string  `json:",omitempty" yaml:",omitempty" xml:",omitempty"`
	Causes    []error `json:",omitempty" yaml:",omitempty" xml:",omitempty"`
}

// NewError returns a new Error.
func NewError(code, msg string) Error { return Error{Code: code, Message: msg} }

// Respond sends the response by the context as JSON.
func (e Error) Respond(c *Context) { c.Failure(e) }

// Clone clones itself to a new one.
func (e Error) Clone() Error {
	if len(e.Causes) > 0 {
		e.Causes = append([]error{}, e.Causes...)
	}
	return e
}

// IsCode is equal to IsCode(e.Code, target).
func (e Error) IsCode(target string) bool { return IsCode(e.Code, target) }

// GetCode returns the error code.
func (e Error) GetCode() string { return e.Code }

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

// WithError is equal to e.WithMessage(err.Error()).
func (e Error) WithError(err error) Error {
	return e.WithMessage(err.Error())
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
	if len(errs) > 0 {
		e.Causes = append(e.Causes, errs...)
	}
	return e
}

// WithData returns a Response with the error and data.
func (e Error) WithData(data interface{}) Response {
	return Response{Data: data, Error: e}
}

// WithRequestID returns a Response with the error and request id.
func (e Error) WithRequestID(requestID string) Response {
	return Response{RequestID: requestID, Error: e}
}
