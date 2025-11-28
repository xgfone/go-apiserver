// Copyright 2025 xgfone
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

package codeint

import "github.com/xgfone/go-toolkit/codeint"

var (
	ErrInvalid      = codeint.ErrInvalid
	ErrNothing      = codeint.ErrNothing
	ErrUnavailable  = codeint.ErrUnavailable
	ErrInconsistent = codeint.ErrInconsistent

	ErrExist    = codeint.ErrExist
	ErrNotExist = codeint.ErrNotExist
	ErrFull     = codeint.ErrFull
	ErrNotFull  = codeint.ErrNotFull
	ErrUsed     = codeint.ErrUsed
	ErrNotUsed  = codeint.ErrNotUsed
	ErrDone     = codeint.ErrDone
	ErrUndone   = codeint.ErrUndone
	ErrPaid     = codeint.ErrPaid
	ErrNotPaid  = codeint.ErrNotPaid
	ErrRefunded = codeint.ErrRefunded
	ErrReturned = codeint.ErrReturned

	ErrUnallowed   = codeint.ErrUnallowed
	ErrUnsupported = codeint.ErrUnsupported

	ErrInUse      = codeint.ErrInUse
	ErrProcessing = codeint.ErrProcessing
	ErrInProgress = codeint.ErrInProgress
	ErrNotStarted = codeint.ErrNotStarted
	ErrHasEnded   = codeint.ErrHasEnded

	ErrIllegal      = codeint.ErrIllegal
	ErrIllegalText  = codeint.ErrIllegalText
	ErrIllegalImage = codeint.ErrIllegalImage
	ErrIllegalVideo = codeint.ErrIllegalVideo

	ErrInsufficient         = codeint.ErrInsufficient
	ErrInsufficientBalance  = codeint.ErrInsufficientBalance
	ErrInsufficientResource = codeint.ErrInsufficientResource
	ErrInsufficientNumber   = codeint.ErrInsufficientNumber
	ErrInsufficientToken    = codeint.ErrInsufficientToken
	ErrInsufficientPrize    = codeint.ErrInsufficientPrize
	ErrInsufficientPaper    = codeint.ErrInsufficientPaper
	ErrInsufficientPoint    = codeint.ErrInsufficientPoint
)

var (
	ErrNotRegistered = codeint.ErrNotRegistered
	ErrUserDisabled  = codeint.ErrUserDisabled

	ErrAuthMissing = codeint.ErrAuthMissing
	ErrAuthInvalid = codeint.ErrAuthInvalid
	ErrAuthExpired = codeint.ErrAuthExpired
)
