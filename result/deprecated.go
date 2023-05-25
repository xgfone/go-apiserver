// Copyright 2023 xgfone
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

// Deprecated Codes
const (
	CodeInvalidAction  = "InvalidAction"
	CodeInvalidParams  = "InvalidParams"
	CodeInvalidVersion = "InvalidVersion"
	CodeInvalidCaptcha = "InvalidCaptcha"

	CodeMissingContentType   = "MissingContentType"
	CodeUnsupportedProtocol  = "UnsupportedProtocol"
	CodeUnsupportedOperation = "UnsupportedOperation"
	CodeUnsupportedMediaType = "UnsupportedMediaType"

	CodeUnauthorizedOperation = "UnauthorizedOperation"
	CodeUnallowedOperation    = "UnallowedOperation"
	CodeFailedOperation       = "FailedOperation"

	CodeQuotaLimitExceeded   = "QuotaLimitExceeded"
	CodeRequestLimitExceeded = "RequestLimitExceeded"

	CodeInstanceInUse        = "InstanceInUse"
	CodeInstanceNotFound     = "InstanceNotFound"
	CodeInstanceUnavailable  = "InstanceUnavailable"
	CodeInstanceInconsistent = "InstanceInconsistent"
	CodeResourceInsufficient = "ResourceInsufficient"
	CodeBalanceInsufficient  = "BalanceInsufficient"

	CodeServiceUnavailable = "ServiceUnavailable"
	CodeGatewayTimeout     = "GatewayTimeout"
)

// Deprecated Errors
var (
	ErrInvalidAction    = NewError(CodeInvalidAction, "invalid action")
	ErrInvalidVersion   = NewError(CodeInvalidVersion, "invalid version")
	ErrInvalidParameter = NewError(CodeInvalidParams, "invalid parameter")
	ErrInvalidCaptcha   = NewError(CodeInvalidCaptcha, "invalid captcha")

	ErrUnsupportedProtocol  = NewError(CodeUnsupportedProtocol, "protocol is unsupported")
	ErrUnsupportedOperation = NewError(CodeUnsupportedOperation, "operation is unsupported")
	ErrUnsupportedMediaType = NewError(CodeUnsupportedMediaType, "media type is unsupported")
	ErrMissingContentType   = NewError(CodeMissingContentType, "missing the header Content-Type")

	ErrUnauthorizedOperation = NewError(CodeUnauthorizedOperation, "operation is unauthorized")
	ErrUnallowedOperation    = NewError(CodeUnallowedOperation, "operation is not allowed")
	ErrFailedOperation       = NewError(CodeFailedOperation, "operation failed")

	ErrServiceUnavailable = NewError(CodeServiceUnavailable, "service is unavailable")
	ErrGatewayTimeout     = NewError(CodeGatewayTimeout, "gateway timeout")

	ErrQuotaLimitExceeded   = NewError(CodeQuotaLimitExceeded, "exceed the quota limit")
	ErrRequestLimitExceeded = NewError(CodeRequestLimitExceeded, "exceed the request limit")

	ErrInstanceInUse        = NewError(CodeInstanceInUse, "instance is in use")
	ErrInstanceNotFound     = NewError(CodeInstanceNotFound, "instance is not found")
	ErrInstanceUnavailable  = NewError(CodeInstanceUnavailable, "instance is unavailable")
	ErrInstanceInconsistent = NewError(CodeInstanceInconsistent, "instance is inconsistent")
	ErrResourceInsufficient = NewError(CodeResourceInsufficient, "resource is insufficient")
	ErrBalanceInsufficient  = NewError(CodeBalanceInsufficient, "balance is insufficient")
)
