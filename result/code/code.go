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

// Predefine some codes.
const (
	BadRequest                     = "BadRequest"
	BadRequestInvalidAction        = "BadRequest.InvalidAction"
	BadRequestInvalidParams        = "BadRequest.InvalidParams"
	BadRequestInvalidVersion       = "BadRequest.InvalidVersion"
	BadRequestInvalidCaptcha       = "BadRequest.InvalidCaptcha"
	BadRequestMissingContentType   = "BadRequest.MissingContentType"
	BadRequestUnsupportedProtocol  = "BadRequest.UnsupportedProtocol"
	BadRequestUnsupportedOperation = "BadRequest.UnsupportedOperation"
	BadRequestUnsupportedMediaType = "BadRequest.UnsupportedMediaType"

	NotFound         = "NotFound"
	NotFoundInstance = "NotFound.Instance"
	NotFoundResource = "NotFound.Resource"

	AuthFailure                 = "AuthFailure"
	AuthFailureMissing          = "AuthFailure.Missing"
	AuthFailureInvalid          = "AuthFailure.Invalid"
	AuthFailureInvalidToken     = "AuthFailure.Invalid.Token"
	AuthFailureInvalidCookie    = "AuthFailure.Invalid.Cookie"
	AuthFailureInvalidAPIKey    = "AuthFailure.Invalid.ApiKey"
	AuthFailureInvalidSignature = "AuthFailure.Invalid.Signature"
	AuthFailureDisabled         = "AuthFailure.Disabled"

	Unallowed                     = "Unallowed"              // Operation is prevented because a condition is not satisfied.
	UnallowedInUse                = "Unallowed.InUse"        // Something is in use, and the operation is exclusive.
	UnallowedExist                = "Unallowed.Exist"        // Something exists, and the operation is exclusive.
	UnallowedNotExist             = "Unallowed.NotExist"     // Something does not exist, and the operation is exclusive.
	UnallowedInOperation          = "Unallowed.InOperation"  // Another is exeucted, and the current is exclusive.
	UnallowedUnavailable          = "Unallowed.Unavailable"  // Something is unavailable for the moment, and maybe recover later.
	UnallowedUnauthorized         = "Unallowed.Unauthorized" // The operator has no permission of the operation.
	UnallowedInconsistent         = "Unallowed.Inconsistent" // The status may be inconsistent.
	UnallowedInsufficient         = "Unallowed.Insufficient"
	UnallowedInsufficientBalance  = "Unallowed.Insufficient.Balance"
	UnallowedInsufficientResource = "Unallowed.Insufficient.Resource"
	UnallowedExceedLimit          = "Unallowed.ExceedLimit"
	UnallowedExceedLimitRate      = "Unallowed.ExceedLimit.Rate"
	UnallowedExceedLimitQuota     = "Unallowed.ExceedLimit.Quota"
	UnallowedNotRegistered        = "Unallowed.NotRegistered"

	InternalServerError            = "InternalServerError"         // All errors from the server, such an unknown exception.
	InternalServerErrorTimeout     = "InternalServerError.Timeout" // Timeout when the gateway forwards the request.
	InternalServerErrorBadGateway  = "InternalServerError.BadGateway"
	InternalServerErrorUnavailable = "InternalServerError.Unavailable" // No available backends for the gateway.
)

// Predefine some errors.
var (
	ErrBadRequest                     = NewError(BadRequest, "bad request")
	ErrBadRequestInvalidAction        = NewError(BadRequestInvalidAction, "invalid version")
	ErrBadRequestInvalidParams        = NewError(BadRequestInvalidParams, "invalid parameters")
	ErrBadRequestInvalidVersion       = NewError(BadRequestInvalidVersion, "invalid version")
	ErrBadRequestInvalidCaptcha       = NewError(BadRequestInvalidCaptcha, "invalid captcha")
	ErrBadRequestMissingContentType   = NewError(BadRequestMissingContentType, "missing the header Content-Type")
	ErrBadRequestUnsupportedProtocol  = NewError(BadRequestUnsupportedProtocol, "protocol is unsupported")
	ErrBadRequestUnsupportedOperation = NewError(BadRequestUnsupportedOperation, "operation is unsupported")
	ErrBadRequestUnsupportedMediaType = NewError(BadRequestUnsupportedMediaType, "media type is unsupported")

	ErrNotFound         = NewError(NotFound, "not found")
	ErrNotFoundInstance = NewError(NotFoundInstance, "instance is not found")
	ErrNotFoundResource = NewError(NotFoundResource, "resource is not found")

	ErrAuthFailure                 = NewError(AuthFailure, "authentication failure")
	ErrAuthFailureMissing          = NewError(AuthFailureMissing, "missing authentication")
	ErrAuthFailureInvalid          = NewError(AuthFailureInvalid, "invalid authentication")
	ErrAuthFailureInvalidToken     = NewError(AuthFailureInvalidToken, "invalid authentication token")
	ErrAuthFailureInvalidCookie    = NewError(AuthFailureInvalidCookie, "invalid authentication cookie")
	ErrAuthFailureInvalidAPIKey    = NewError(AuthFailureInvalidAPIKey, "invalid authentication apikey")
	ErrAuthFailureInvalidSignature = NewError(AuthFailureInvalidSignature, "invalid authentication signature")
	ErrAuthFailureDisabled         = NewError(AuthFailureDisabled, "the user is disabled")

	ErrUnallowed                     = NewError(Unallowed, "operation is not allowed")
	ErrUnallowedInUse                = NewError(UnallowedInUse, "in use")
	ErrUnallowedExist                = NewError(UnallowedExist, "exist")
	ErrUnallowedNotExist             = NewError(UnallowedNotExist, "not exist")
	ErrUnallowedInOperation          = NewError(UnallowedInOperation, "in operation")
	ErrUnallowedUnavailable          = NewError(UnallowedUnavailable, "unavailable")
	ErrUnallowedUnauthorized         = NewError(UnallowedUnauthorized, "operation is unauthorized")
	ErrUnallowedInconsistent         = NewError(UnallowedInconsistent, "inconsistent")
	ErrUnallowedInsufficient         = NewError(UnallowedInsufficient, "insufficient")
	ErrUnallowedInsufficientBalance  = NewError(UnallowedInsufficientBalance, "balance is insufficient")
	ErrUnallowedInsufficientResource = NewError(UnallowedInsufficientResource, "resource is insufficient")
	ErrUnallowedExceedLimit          = NewError(UnallowedExceedLimit, "exceed the limit")
	ErrUnallowedExceedLimitRate      = NewError(UnallowedExceedLimitRate, "exceed the rate limit")
	ErrUnallowedExceedLimitQuota     = NewError(UnallowedExceedLimitQuota, "exceed the quota limit")
	ErrUnallowedNotRegistered        = NewError(UnallowedNotRegistered, "not registered")

	ErrInternalServerError            = NewError(InternalServerError, "internal server error")
	ErrInternalServerErrorTimeout     = NewError(InternalServerErrorTimeout, "gateway timeout")
	ErrInternalServerErrorBadGateway  = NewError(InternalServerErrorBadGateway, "bad gateway")
	ErrInternalServerErrorUnavailable = NewError(InternalServerErrorUnavailable, "service is unavailable")
)

// Equal is used by Is to compare whether code is target.
var Equal func(code, target any) bool

// CodeGetter is an interface used to get the error code.
type CodeGetter[T any] interface {
	GetCode() T
}

// IsCoder is used to reports whether the error is the target code.
type IsCoder[T any] interface {
	IsCode(target T) bool
}

func stris(code, target string) bool {
	mlen := len(target)
	return mlen == 0 || code == target ||
		(len(code) > mlen && code[mlen] == '.' && code[:mlen] == target)
}

// Is reports whether the code is target or child of that.
//
// For the string type, it supports the prefix parent, for example,
//
//	Is("InstanceNotFound", "")                    // => true
//	Is("InstanceNotFound", "InstanceNotFound")    // => true
//	Is("InstanceNotFound", "InstanceUnavailable") // => false
//	Is("AuthFailure.TokenFailure", "AuthFailure") // => true
//	Is("AuthFailure", "AuthFailure.TokenFailure") // => false
//
// For other types, it just compares whether both of them are equal or not.
func Is(code, target any) bool {
	if Equal != nil {
		return Equal(code, target)
	}

	switch src := code.(type) {
	case string:
		return stris(src, target.(string))

	default:
		return code == target
	}
}

// ErrIs reports whether the code of the error is the target code
// or the child of that.
//
// err need to implement the interfaces IsCoder and CodeGetter[string].
// Or, return false.
func ErrIs[T any](err error, targetCode T) bool {
	switch e := err.(type) {
	case IsCoder[T]:
		return e.IsCode(targetCode)

	case CodeGetter[T]:
		return Is(e.GetCode(), targetCode)

	default:
		return false
	}
}

// IsCode is equal to Is(e.Code, target).
func (e Error[T]) IsCode(target T) bool { return Is(e.Code, target) }

// IsBadRequest reports whether the error is BadRequest.
func (e Error[T]) IsBadRequest() bool { return Is(e.Code, BadRequest) }

// IsNotFound reports whether the error is NotFound.
func (e Error[T]) IsNotFound() bool { return Is(e.Code, NotFound) }

// IsUnallowed reports whether the error is Unallowed.
func (e Error[T]) IsUnallowed() bool { return Is(e.Code, Unallowed) }

// IsAuthFailure reports whether the error is AuthFailure.
func (e Error[T]) IsAuthFailure() bool { return Is(e.Code, AuthFailure) }

// IsUnauthorized reports whether the error is UnallowedUnauthorized.
func (e Error[T]) IsUnauthorized() bool { return Is(e.Code, UnallowedUnauthorized) }

// IsInternalServerError reports whether the error is InternalServerError.
func (e Error[T]) IsInternalServerError() bool { return Is(e.Code, InternalServerError) }
