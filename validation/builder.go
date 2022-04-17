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
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/xgfone/predicate"
)

// Context is a builder context to manage the built validators.
type Context struct {
	validators []Validator
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
		c.AppendValidators(And(validators...))
	}
}

// Or implements the interface predicate.BuidlerContext.
func (c *Context) Or(bc predicate.BuilderContext) {
	c.AppendValidators(Or(bc.(*Context).Validators()...))
}

// AppendValidators appends the new validators into the context.
//
// The method is used by the validator building function.
func (c *Context) AppendValidators(validators ...Validator) {
	c.validators = append(c.validators, validators...)
}

// Validators returns all the inner validators.
func (c *Context) Validators() []Validator { return c.validators }

// Validator returns the inner validators as And Validator.
func (c *Context) Validator() Validator { return And(c.validators...) }

// BuilderFunction is a function used to build the validation rule into Context.
type BuilderFunction func(c *Context, args ...interface{}) error

func (f BuilderFunction) fn(c predicate.BuilderContext, args ...interface{}) error {
	return f(c.(*Context), args...)
}

// DefaultBuilder is the global default validation rule builder.
var DefaultBuilder = NewBuilder()

// RegisterFunc is eqaul to DefaultBuilder.RegisterFunc(name, f).
func RegisterFunc(name string, f BuilderFunction) {
	DefaultBuilder.RegisterFunc(name, f)
}

// Build is equal to DefaultBuilder.Build(c, rule).
func Build(c *Context, rule string) error {
	return DefaultBuilder.Build(c, rule)
}

// BuildValidator is equal to DefaultBuilder.BuildValidator(rule).
func BuildValidator(rule string) (Validator, error) {
	return DefaultBuilder.BuildValidator(rule)
}

// Validate is equal to DefaultBuilder.Validate(v, rule).
func Validate(v interface{}, rule string) error {
	return DefaultBuilder.Validate(v, rule)
}

// ValidateStruct is equal to DefaultBuilder.ValidateStruct(s).
func ValidateStruct(s interface{}) error {
	return DefaultBuilder.ValidateStruct(s)
}

// LookupStructFieldNameByTags returns a function to lookup the field name
// from the given tags.
//
// If failing to lookup, use the original name of the struct field.
func LookupStructFieldNameByTags(tags ...string) func(reflect.StructField) string {
	return func(sf reflect.StructField) string {
		for _, tag := range tags {
			if v := strings.TrimSpace(sf.Tag.Get(tag)); v != "" {
				return v
			}
		}
		return sf.Name
	}
}

var lookupStructFieldName = LookupStructFieldNameByTags("json")

// Builder is used to build the validator based on the rule.
type Builder struct {
	// LookupStructFieldName is used to lookup the name of the struct field.
	//
	// If nil, use LookupStructFieldNameByTags("json") instead.
	LookupStructFieldName func(reflect.StructField) string

	*predicate.Builder
	validators atomic.Value
	vcacheLock sync.Mutex
	vcacheMap  map[string]Validator
}

// NewBuilder returns a new validation rule builder.
func NewBuilder() *Builder {
	builder := predicate.NewBuilder()

	builder.GetIdentifier = func(selector []string) (interface{}, error) {
		// Support the format "zero" instead of "zero()"
		if f := builder.GetFunc(selector[0]); f != nil {
			return f, nil
		}
		return nil, fmt.Errorf("%s is not defined", selector[0])
	}

	builder.EQ = func(ctx predicate.BuilderContext, left, right interface{}) error {
		// Support the format "min == 123" or "123 == min"
		if f, ok := left.(predicate.BuilderFunction); ok {
			return f(ctx, right)
		}
		if f, ok := right.(predicate.BuilderFunction); ok {
			return f(ctx, left)
		}
		return fmt.Errorf("left or right is not BuilderFunction: %T, %T", left, right)
	}

	b := &Builder{Builder: builder, vcacheMap: make(map[string]Validator)}
	b.updateValidators()
	return b
}

// RegisterFunc registers the builder function with the name.
//
// If the function name has existed, reset it to the new function.
func (b *Builder) RegisterFunc(name string, f BuilderFunction) {
	b.Builder.RegisterFunc(name, f.fn)
}

// Build parses and builds the validation rule into the context.
func (b *Builder) Build(c *Context, rule string) error {
	return b.Builder.Build(c, rule)
}

// BuildValidator builds a validator from the validation rule.
//
// If the rule has been built, returns it from the caches.
func (b *Builder) BuildValidator(rule string) (Validator, error) {
	if rule == "" {
		return nil, errors.New("the validation rule must not be empty")
	}

	if validator, ok := b.loadValidator(rule); ok {
		return validator, nil
	}

	b.vcacheLock.Lock()
	defer b.vcacheLock.Unlock()

	if validator, ok := b.loadValidator(rule); ok {
		return validator, nil
	}

	c := NewContext()
	if err := b.Build(c, rule); err != nil {
		return nil, err
	}

	validator := c.Validator()
	b.vcacheMap[rule] = validator
	b.updateValidators()

	return validator, nil
}

func (b *Builder) loadValidator(rule string) (v Validator, ok bool) {
	v, ok = b.validators.Load().(map[string]Validator)[rule]
	return
}

func (b *Builder) updateValidators() {
	validators := make(map[string]Validator, len(b.vcacheMap))
	for rule, validator := range b.vcacheMap {
		validators[rule] = validator
	}
	b.validators.Store(validators)
}

type validatorWrapper struct{ Validator }

// Validate validates whether the value v is valid by the rule.
//
// If failing to build the rule to the validator, panic with the error.
func (b *Builder) Validate(v interface{}, rule string) (err error) {
	if rule == "" {
		return nil
	}

	validator, err := b.BuildValidator(rule)
	if err != nil {
		panic(err)
	}
	return validator.Validate(v)
}

// ValidateStruct validates whether the struct value is valid,
// which extracts the validation rule from the field tag "validate"
// validate the field value with the extracted rule.
func (b *Builder) ValidateStruct(s interface{}) error {
	v := reflect.ValueOf(s)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		panic(fmt.Errorf("the value is %T, not a struct", v.Interface()))
	}

	var errs NamedErrors

	t := v.Type()
	for i, _len := 0, v.NumField(); i < _len; i++ {
		ft := t.Field(i)

		rule := ft.Tag.Get("validate")
		if rule == "" {
			continue
		}

		err := b.Validate(v.Field(i).Interface(), rule)
		if err != nil {
			var name string
			if b.LookupStructFieldName == nil {
				name = lookupStructFieldName(ft)
			} else {
				name = b.LookupStructFieldName(ft)
			}

			if errs == nil {
				errs = make(NamedErrors, _len)
				errs.Add(name, err)
			}
		}
	}

	if errs == nil {
		return nil
	}
	return errs
}

// NamedErrors represents a set of errors with the names.
type NamedErrors map[string]error

// Error implements the interface error.
func (es NamedErrors) Error() string {
	var b strings.Builder
	b.Grow(len(es) * 64)

	var count int
	for name, err := range es {
		if count > 0 {
			b.WriteString("; ")
			count++
		}

		b.WriteString(name)
		b.WriteString(": ")
		b.WriteString(err.Error())
	}

	return b.String()
}

// Add adds the error with the name.
func (es NamedErrors) Add(name string, err error) { es[name] = err }

// MarshalJSON implements the interface json.Marshaler.
func (es NamedErrors) MarshalJSON() ([]byte, error) {
	maps := make(map[string]string, len(es))
	for name, err := range es {
		maps[name] = err.Error()
	}
	return json.Marshal(maps)
}
