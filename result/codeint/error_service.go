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

var (
	ErrInvalid      = ErrConflict.WithCode(400000).WithMessage("invalid")
	ErrNothing      = ErrConflict.WithCode(400001).WithMessage("nothing")
	ErrUnavailable  = ErrConflict.WithCode(400001).WithMessage("unavailable")
	ErrInconsistent = ErrConflict.WithCode(400002).WithMessage("inconsistent")

	ErrExist    = ErrConflict.WithCode(400003).WithMessage("has existed")
	ErrNotExist = ErrConflict.WithCode(400004).WithMessage("not exist")
	ErrFull     = ErrConflict.WithCode(400005).WithMessage("full")
	ErrNotFull  = ErrConflict.WithCode(400006).WithMessage("not full")
	ErrUsed     = ErrConflict.WithCode(400007).WithMessage("has used")
	ErrNotUsed  = ErrConflict.WithCode(400008).WithMessage("not used")
	ErrDone     = ErrConflict.WithCode(400009).WithMessage("has done")
	ErrUndone   = ErrConflict.WithCode(400010).WithMessage("has not done")
	ErrPaid     = ErrConflict.WithCode(400011).WithMessage("has paid")
	ErrNotPaid  = ErrConflict.WithCode(400012).WithMessage("has not paid")
	ErrRefunded = ErrConflict.WithCode(400013).WithMessage("has refunded")
	ErrReturned = ErrConflict.WithCode(400013).WithMessage("has returned")

	ErrUnallowed   = ErrConflict.WithCode(400030).WithMessage("unallowed")
	ErrUnsupported = ErrConflict.WithCode(400031).WithMessage("unsupported")

	ErrInUse      = ErrConflict.WithCode(400032).WithMessage("in use")
	ErrProcessing = ErrConflict.WithCode(400033).WithMessage("processing")  // Doing
	ErrInProgress = ErrConflict.WithCode(400033).WithMessage("in progress") // Doing
	ErrNotStarted = ErrConflict.WithCode(400034).WithMessage("not started")
	ErrHasEnded   = ErrConflict.WithCode(400035).WithMessage("has ended")

	ErrIllegal      = ErrConflict.WithCode(400040).WithMessage("illegal")
	ErrIllegalText  = ErrConflict.WithCode(400041).WithMessage("illegal text")
	ErrIllegalImage = ErrConflict.WithCode(400042).WithMessage("illegal image")
	ErrIllegalVideo = ErrConflict.WithCode(400043).WithMessage("illegal video")

	ErrInsufficient         = ErrConflict.WithCode(400050).WithMessage("insufficient")
	ErrInsufficientBalance  = ErrConflict.WithCode(400051).WithMessage("balance is insufficient")
	ErrInsufficientResource = ErrConflict.WithCode(400052).WithMessage("resource is insufficient")
	ErrInsufficientNumber   = ErrConflict.WithCode(400053).WithMessage("number is insufficient")
	ErrInsufficientToken    = ErrConflict.WithCode(400054).WithMessage("token is insufficient")
	ErrInsufficientPrize    = ErrConflict.WithCode(400055).WithMessage("prize is insufficient")
	ErrInsufficientPaper    = ErrConflict.WithCode(400056).WithMessage("paper is insufficient")
	ErrInsufficientPoint    = ErrConflict.WithCode(400057).WithMessage("point is insufficient")
)

var (
	ErrNotRegistered = ErrConflict.WithCode(401001).WithMessage("not registered")
	ErrUserDisabled  = ErrConflict.WithCode(401002).WithMessage("user is disabled")

	ErrAuthMissing = ErrConflict.WithCode(401010).WithMessage("auth is missing")
	ErrAuthInvalid = ErrConflict.WithCode(401011).WithMessage("auth is invalid")
	ErrAuthExpired = ErrConflict.WithCode(401012).WithMessage("auth is expired")
)
