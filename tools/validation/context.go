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

package validation

import (
	"github.com/xgfone/go-apiserver/tools/validation/validator"
	"github.com/xgfone/predicate"
)

// Context is a builder context to manage the built validators.
type Context struct {
	validators []validator.Validator
}

// NewContext returns a new builder context.
func NewContext() *Context { return &Context{} }

// New implements the interface predicate.BuidlerContext.
func (c *Context) New() predicate.BuilderContext { return NewContext() }

// Not implements the interface predicate.BuidlerContext.
func (c *Context) Not(predicate.BuilderContext) {
	panic("unsupport the NOT validation rule")
}

// And implements the interface predicate.BuidlerContext.
func (c *Context) And(bc predicate.BuilderContext) {
	if validators := bc.(*Context).Validators(); len(validators) > 0 {
		c.AppendValidators(validator.And(validators...))
	}
}

// Or implements the interface predicate.BuidlerContext.
func (c *Context) Or(bc predicate.BuilderContext) {
	c.AppendValidators(validator.Or(bc.(*Context).Validators()...))
}

// AppendValidators appends the new validators into the context.
//
// The method is used by the validator building function.
func (c *Context) AppendValidators(validators ...validator.Validator) {
	c.validators = append(c.validators, validators...)
}

// Validators returns all the inner validators.
func (c *Context) Validators() []validator.Validator { return c.validators }

// Validator returns the inner validators as And Validator.
func (c *Context) Validator() validator.Validator { return validator.And(c.validators...) }
