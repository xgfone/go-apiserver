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

// Predefine some error codes.
const (
	CodeInvalidAction  = "InvalidAction"
	CodeInvalidVersion = "InvalidVersion"
	CodeInvalidParams  = "InvalidParams"
	CodeInvalidAuth    = "InvalidAuth"

	CodeUnsupportedProtocol  = "UnsupportedProtocol"
	CodeUnsupportedOperation = "UnsupportedOperation"
	CodeUnsupportedMediaType = "UnsupportedMediaType"
	CodeMissingContentType   = "MissingContentType"

	CodeUnauthorizedOperation = "UnauthorizedOperation"
	CodeUnallowedOperation    = "UnallowedOperation"
	CodeFailedOperation       = "FailedOperation"

	CodeAuthFailureTokenFailure     = "AuthFailure.TokenFailure"
	CodeAuthFailureSignatureFailure = "AuthFailure.SignatureFailure"
	CodeAuthFailureSignatureExpire  = "AuthFailure.SignatureExpire"

	CodeInternalServerError = "InternalServerError"
	CodeServiceUnavailable  = "ServiceUnavailable"
	CodeGatewayTimeout      = "GatewayTimeout"

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
	ErrInvalidAction    = NewError(CodeInvalidAction, "invalid action")
	ErrInvalidVersion   = NewError(CodeInvalidVersion, "invalid version")
	ErrInvalidParameter = NewError(CodeInvalidParams, "invalid parameter")
	ErrInvalidAuth      = NewError(CodeInvalidAuth, "invalid auth")

	ErrUnsupportedProtocol  = NewError(CodeUnsupportedProtocol, "protocol is unsupported")
	ErrUnsupportedOperation = NewError(CodeUnsupportedOperation, "operation is unsupported")
	ErrUnsupportedMediaType = NewError(CodeUnsupportedMediaType, "media type is unsupported")
	ErrMissingContentType   = NewError(CodeMissingContentType, "missing the header Content-Type")

	ErrUnauthorizedOperation = NewError(CodeUnauthorizedOperation, "operation is unauthorized")
	ErrUnallowedOperation    = NewError(CodeUnallowedOperation, "operation is not allowed")
	ErrFailedOperation       = NewError(CodeFailedOperation, "operation failed")

	ErrAuthFailureTokenFailure     = NewError(CodeAuthFailureTokenFailure, "token verification failed")
	ErrAuthFailureSignatureFailure = NewError(CodeAuthFailureSignatureFailure, "signature verification failed")
	ErrAuthFailureSignatureExpire  = NewError(CodeAuthFailureSignatureExpire, "signature is expired")

	ErrInternalServerError = NewError(CodeInternalServerError, "internal server error")
	ErrServiceUnavailable  = NewError(CodeServiceUnavailable, "service is unavailable")
	ErrGatewayTimeout      = NewError(CodeGatewayTimeout, "gateway timeout")

	ErrQuotaLimitExceeded   = NewError(CodeQuotaLimitExceeded, "exceed the quota limit")
	ErrRequestLimitExceeded = NewError(CodeRequestLimitExceeded, "exceed the request limit")

	ErrInstanceInUse        = NewError(CodeInstanceInUse, "instance is in use")
	ErrInstanceNotFound     = NewError(CodeInstanceNotFound, "instance is not found")
	ErrInstanceUnavailable  = NewError(CodeInstanceUnavailable, "instance is unavailable")
	ErrInstanceInconsistent = NewError(CodeInstanceInconsistent, "instance is inconsistent")
	ErrResourceInsufficient = NewError(CodeResourceInsufficient, "resource is insufficient")
	ErrBalanceInsufficient  = NewError(CodeBalanceInsufficient, "balance is insufficient")
)

// CodeGetter is an interface used to get the error code.
type CodeGetter interface {
	GetCode() string
}

// GetCode gets the error code from the error if it implements
// the interface CodeGetter. Or returns "".
func GetCode(err error) string {
	if c, ok := err.(CodeGetter); ok {
		return c.GetCode()
	}
	return ""
}

// IsCoder is used to reports whether the error is the target code.
type IsCoder interface {
	IsCode(target string) bool
}

// IsCode reports whether the code is target or child of that.
//
// Example
//
//	IsCode("InstanceNotFound", "")                    // => true
//	IsCode("InstanceNotFound", "InstanceNotFound")    // => true
//	IsCode("InstanceNotFound", "InstanceUnavailable") // => false
//	IsCode("AuthFailure.TokenFailure", "AuthFailure") // => true
//	IsCode("AuthFailure", "AuthFailure.TokenFailure") // => false
func IsCode(code, target string) bool {
	mlen := len(target)
	return mlen == 0 || code == target ||
		(len(code) > mlen && code[mlen] == '.' && code[:mlen] == target)
}

// ErrIsCode reports whether the code of the error is the target code
// or the child of that.
//
// err need to implement the interfaces IsCoder and CodeGetter. Or, return false.
func ErrIsCode(err error, targetCode string) bool {
	switch e := err.(type) {
	case IsCoder:
		return e.IsCode(targetCode)

	case CodeGetter:
		return IsCode(e.GetCode(), targetCode)

	default:
		return false
	}
}
