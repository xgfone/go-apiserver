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

// Package auth provides the common auth interface.
package auth

// Pre-define some permssion actions.
const (
	ActionCreate = "Create"
	ActionDelete = "Delete"
	ActionUpdate = "Update"
	ActionRead   = "Read"
)

// Subject represents a user subject.
type Subject interface {
	ID() string
}

// Object represents an object.
type Object interface {
	Name() string
}

type wrapper string

func (w wrapper) ID() string   { return string(w) }
func (w wrapper) Name() string { return string(w) }

// NewSubject returns a Subject with the id.
func NewSubject(id string) Subject { return wrapper(id) }

// NewObject returns an Object with the name.
func NewObject(name string) Object { return wrapper(name) }
