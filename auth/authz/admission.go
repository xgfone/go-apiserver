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

package authz

import "github.com/xgfone/go-apiserver/auth"

// DefaultAdmission is the default global admission.
var DefaultAdmission Admission

// Admission is used to validate whether a subject has the given permission
// to finish the action on the specific object.
type Admission interface {
	Allow(s auth.Subject, action string, o auth.Object) (ok bool, err error)
}

// AdmissionFunc is an Admission function to convert a function to Admission.
type AdmissionFunc func(s auth.Subject, action string, o auth.Object) (bool, error)

// Allow implements the interface Admission.
func (f AdmissionFunc) Allow(s auth.Subject, action string, o auth.Object) (bool, error) {
	return f(s, action, o)
}
