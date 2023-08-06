// Copyright 2022~2023 xgfone
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
	CodeBadRequest                     = "BadRequest"
	CodeBadRequestInvalidAction        = "BadRequest.InvalidAction"
	CodeBadRequestInvalidParams        = "BadRequest.InvalidParams"
	CodeBadRequestInvalidVersion       = "BadRequest.InvalidVersion"
	CodeBadRequestInvalidCaptcha       = "BadRequest.InvalidCaptcha"
	CodeBadRequestMissingContentType   = "BadRequest.MissingContentType"
	CodeBadRequestUnsupportedProtocol  = "BadRequest.UnsupportedProtocol"
	CodeBadRequestUnsupportedOperation = "BadRequest.UnsupportedOperation"
	CodeBadRequestUnsupportedMediaType = "BadRequest.UnsupportedMediaType"

	CodeNotFound         = "NotFound"
	CodeNotFoundInstance = "NotFound.Instance"
	CodeNotFoundResource = "NotFound.Resource"

	CodeAuthFailure                 = "AuthFailure"
	CodeAuthFailureMissing          = "AuthFailure.Missing"
	CodeAuthFailureInvalid          = "AuthFailure.Invalid"
	CodeAuthFailureInvalidToken     = "AuthFailure.Invalid.Token"
	CodeAuthFailureInvalidCookie    = "AuthFailure.Invalid.Cookie"
	CodeAuthFailureInvalidAPIKey    = "AuthFailure.Invalid.ApiKey"
	CodeAuthFailureInvalidSignature = "AuthFailure.Invalid.Signature"
	CodeAuthFailureDisabled         = "AuthFailure.Disabled"

	CodeUnallowed                     = "Unallowed"              // Operation is prevented because a condition is not satisfied.
	CodeUnallowedInUse                = "Unallowed.InUse"        // Something is in use, and the operation is exclusive.
	CodeUnallowedExist                = "Unallowed.Exist"        // Something exists, and the operation is exclusive.
	CodeUnallowedNotExist             = "Unallowed.NotExist"     // Something does not exist, and the operation is exclusive.
	CodeUnallowedInOperation          = "Unallowed.InOperation"  // Another is exeucted, and the current is exclusive.
	CodeUnallowedUnavailable          = "Unallowed.Unavailable"  // Something is unavailable for the moment, and maybe recover later.
	CodeUnallowedUnauthorized         = "Unallowed.Unauthorized" // The operator has no permission of the operation.
	CodeUnallowedInconsistent         = "Unallowed.Inconsistent" // The status may be inconsistent.
	CodeUnallowedInsufficient         = "Unallowed.Insufficient"
	CodeUnallowedInsufficientBalance  = "Unallowed.Insufficient.Balance"
	CodeUnallowedInsufficientResource = "Unallowed.Insufficient.Resource"
	CodeUnallowedExceedLimit          = "Unallowed.ExceedLimit"
	CodeUnallowedExceedLimitRate      = "Unallowed.ExceedLimit.Rate"
	CodeUnallowedExceedLimitQuota     = "Unallowed.ExceedLimit.Quota"
	CodeUnallowedNotRegistered        = "Unallowed.NotRegistered"

	CodeInternalServerError            = "InternalServerError"         // All errors from the server, such an unknown exception.
	CodeInternalServerErrorTimeout     = "InternalServerError.Timeout" // Timeout when the gateway forwards the request.
	CodeInternalServerErrorBadGateway  = "InternalServerError.BadGateway"
	CodeInternalServerErrorUnavailable = "InternalServerError.Unavailable" // No available backends for the gateway.
)

// Predefine some errors.
var (
	ErrBadRequest                     = NewError(CodeBadRequest, "bad request")
	ErrBadRequestInvalidAction        = NewError(CodeBadRequestInvalidAction, "invalid version")
	ErrBadRequestInvalidParams        = NewError(CodeBadRequestInvalidParams, "invalid parameters")
	ErrBadRequestInvalidVersion       = NewError(CodeBadRequestInvalidVersion, "invalid version")
	ErrBadRequestInvalidCaptcha       = NewError(CodeBadRequestInvalidCaptcha, "invalid captcha")
	ErrBadRequestMissingContentType   = NewError(CodeBadRequestMissingContentType, "missing the header Content-Type")
	ErrBadRequestUnsupportedProtocol  = NewError(CodeBadRequestUnsupportedProtocol, "protocol is unsupported")
	ErrBadRequestUnsupportedOperation = NewError(CodeBadRequestUnsupportedOperation, "operation is unsupported")
	ErrBadRequestUnsupportedMediaType = NewError(CodeBadRequestUnsupportedMediaType, "media type is unsupported")

	ErrNotFound         = NewError(CodeNotFound, "not found")
	ErrNotFoundInstance = NewError(CodeNotFoundInstance, "instance is not found")
	ErrNotFoundResource = NewError(CodeNotFoundResource, "resource is not found")

	ErrAuthFailure                 = NewError(CodeAuthFailure, "authentication failure")
	ErrAuthFailureMissing          = NewError(CodeAuthFailureMissing, "missing authentication")
	ErrAuthFailureInvalid          = NewError(CodeAuthFailureInvalid, "invalid authentication")
	ErrAuthFailureInvalidToken     = NewError(CodeAuthFailureInvalidToken, "invalid authentication token")
	ErrAuthFailureInvalidCookie    = NewError(CodeAuthFailureInvalidCookie, "invalid authentication cookie")
	ErrAuthFailureInvalidAPIKey    = NewError(CodeAuthFailureInvalidAPIKey, "invalid authentication apikey")
	ErrAuthFailureInvalidSignature = NewError(CodeAuthFailureInvalidSignature, "invalid authentication signature")
	ErrAuthFailureDisabled         = NewError(CodeAuthFailureDisabled, "the user is disabled")

	ErrUnallowed                     = NewError(CodeUnallowed, "operation is not allowed")
	ErrUnallowedInUse                = NewError(CodeUnallowedInUse, "in use")
	ErrUnallowedExist                = NewError(CodeUnallowedExist, "exist")
	ErrUnallowedNotExist             = NewError(CodeUnallowedNotExist, "not exist")
	ErrUnallowedInOperation          = NewError(CodeUnallowedInOperation, "in operation")
	ErrUnallowedUnavailable          = NewError(CodeUnallowedUnavailable, "unavailable")
	ErrUnallowedUnauthorized         = NewError(CodeUnallowedUnauthorized, "operation is unauthorized")
	ErrUnallowedInconsistent         = NewError(CodeUnallowedInconsistent, "inconsistent")
	ErrUnallowedInsufficient         = NewError(CodeUnallowedInsufficient, "insufficient")
	ErrUnallowedInsufficientBalance  = NewError(CodeUnallowedInsufficientBalance, "balance is insufficient")
	ErrUnallowedInsufficientResource = NewError(CodeUnallowedInsufficientResource, "resource is insufficient")
	ErrUnallowedExceedLimit          = NewError(CodeUnallowedExceedLimit, "exceed the limit")
	ErrUnallowedExceedLimitRate      = NewError(CodeUnallowedExceedLimitRate, "exceed the rate limit")
	ErrUnallowedExceedLimitQuota     = NewError(CodeUnallowedExceedLimitQuota, "exceed the quota limit")
	ErrUnallowedNotRegistered        = NewError(CodeUnallowedNotRegistered, "not registered")

	ErrInternalServerError            = NewError(CodeInternalServerError, "internal server error")
	ErrInternalServerErrorTimeout     = NewError(CodeInternalServerErrorTimeout, "gateway timeout")
	ErrInternalServerErrorBadGateway  = NewError(CodeInternalServerErrorBadGateway, "bad gateway")
	ErrInternalServerErrorUnavailable = NewError(CodeInternalServerErrorUnavailable, "service is unavailable")
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

// IsBadRequest reports whether the error is CodeBadRequest.
func (e Error) IsBadRequest() bool { return IsCode(e.Code, CodeBadRequest) }

// IsNotFound reports whether the error is CodeNotFound.
func (e Error) IsNotFound() bool { return IsCode(e.Code, CodeNotFound) }

// IsUnallowed reports whether the error is CodeUnallowed.
func (e Error) IsUnallowed() bool { return IsCode(e.Code, CodeUnallowed) }

// IsAuthFailure reports whether the error is CodeAuthFailure.
func (e Error) IsAuthFailure() bool { return IsCode(e.Code, CodeAuthFailure) }

// IsUnauthorized reports whether the error is CodeUnallowedUnauthorized.
func (e Error) IsUnauthorized() bool { return IsCode(e.Code, CodeUnallowedUnauthorized) }

// IsInternalServerError reports whether the error is CodeInternalServerError.
func (e Error) IsInternalServerError() bool { return IsCode(e.Code, CodeInternalServerError) }
