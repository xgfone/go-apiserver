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

// RegisterSymbol is equal to DefaultBuilder.RegisterSymbol(name, value).
func RegisterSymbol(name string, value interface{}) {
	DefaultBuilder.RegisterSymbol(name, value)
}

// RegisterSymbols is equal to DefaultBuilder.RegisterSymbols(maps).
func RegisterSymbols(maps map[string]interface{}) {
	DefaultBuilder.RegisterSymbols(maps)
}

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

	// Symbols is used to define the global symbols,
	// which is used by the default of GetIdentifier.
	Symbols map[string]interface{}

	*predicate.Builder
	validators atomic.Value
	vcacheLock sync.Mutex
	vcacheMap  map[string]Validator
}

// NewBuilder returns a new validation rule builder.
func NewBuilder() *Builder {
	builder := &Builder{
		vcacheMap: make(map[string]Validator),
		Symbols:   make(map[string]interface{}),
	}

	builder.Builder = predicate.NewBuilder()
	builder.Builder.GetIdentifier = builder.getIdentifier
	builder.Builder.EQ = builder.eq

	builder.updateValidators()
	return builder
}

func (b *Builder) getIdentifier(selector []string) (interface{}, error) {
	// Support the format "zero" instead of "zero()"

	// First, lookup the function table.
	if f := b.GetFunc(selector[0]); f != nil {
		return f, nil
	}

	// Second, lookup the symbol table.
	if v, ok := b.Symbols[selector[0]]; ok {
		return v, nil
	}

	// We find no the identifier.
	return nil, fmt.Errorf("%s is not defined", selector[0])
}

func (b *Builder) eq(ctx predicate.BuilderContext, left, right interface{}) error {
	// Support the format "min == 123" or "123 == min"
	if f, ok := left.(predicate.BuilderFunction); ok {
		return f(ctx, right)
	}
	if f, ok := right.(predicate.BuilderFunction); ok {
		return f(ctx, left)
	}
	return fmt.Errorf("left or right is not BuilderFunction: %T, %T", left, right)
}

// RegisterSymbol registers the symbol with the name and value.
func (b *Builder) RegisterSymbol(name string, value interface{}) {
	if name == "" {
		panic("the symbol name must not be empty")
	}
	if value == nil {
		panic("the symbol value must not be nil")
	}
	b.Symbols[name] = value
}

// RegisterSymbols registers a set of symbols from a map.
func (b *Builder) RegisterSymbols(maps map[string]interface{}) {
	for name, value := range maps {
		b.RegisterSymbol(name, value)
	}
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

	errs := b.validateStruct("", v, nil)
	if len(errs) > 0 {
		return errs
	}
	return nil
}

var validatorImpl = reflect.TypeOf((*ValueValidator)(nil)).Elem()

func (b *Builder) validateStruct(prefix string, v reflect.Value, errs NamedErrors) NamedErrors {
	t := v.Type()
	for i, _len := 0, v.NumField(); i < _len; i++ {
		ft := t.Field(i)
		fv := v.Field(i)

		// Validate the fields of the sub-struct recursively.
		if ft.Type.Kind() == reflect.Struct {
			name := b.getStructFieldName(prefix, ft)
			errs = b.validateStruct(name, fv, errs)
			if errs == nil {
				if ft.Type.Implements(validatorImpl) {
					err := fv.Interface().(ValueValidator).Validate()
					errs = addError(errs, name, err)
				}
			}
			continue
		}

		rule := ft.Tag.Get("validate")
		if rule == "" {
			continue
		}

		err := b.Validate(fv.Interface(), rule)
		if err != nil {
			errs = addError(errs, b.getStructFieldName(prefix, ft), err)
		}
	}

	return errs
}

func (b *Builder) getStructFieldName(prefix string, ft reflect.StructField) (name string) {
	if b.LookupStructFieldName == nil {
		name = lookupStructFieldName(ft)
	} else {
		name = b.LookupStructFieldName(ft)
	}

	if name == "" {
		name = ft.Name
	}

	if prefix != "" {
		name = strings.Join([]string{prefix, name}, ".")
	}

	return
}

func addError(errs NamedErrors, name string, err error) NamedErrors {
	if err != nil {
		if errs == nil {
			errs = make(NamedErrors, 4)
		}
		errs[name] = err
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
