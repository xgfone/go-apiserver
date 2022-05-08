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

	"github.com/xgfone/go-apiserver/validation/internal"
	"github.com/xgfone/predicate"
)

// StructFieldTag is the tag name to get the validation rule.
//
// If empty, use "validate" instead.
var StructFieldTag = "validate"

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

// RegisterSymbolNames is equal to DefaultBuilder.RegisterSymbolNames(names...).
func RegisterSymbolNames(names ...string) {
	DefaultBuilder.RegisterSymbolNames(names...)
}

// RegisterFunction is eqaul to DefaultBuilder.RegisterFunction(function).
func RegisterFunction(function Function) {
	DefaultBuilder.RegisterFunction(function)
}

// RegisterValidatorFunc is equal to
// DefaultBuilder.RegisterValidatorFunc(name, f).
func RegisterValidatorFunc(name string, f ValidatorFunc) {
	DefaultBuilder.RegisterValidatorFunc(name, f)
}

// RegisterValidatorFuncBool is equal to
// DefaultBuilder.RegisterValidatorFuncBool(name, f, err).
func RegisterValidatorFuncBool(name string, f func(interface{}) bool, err error) {
	DefaultBuilder.RegisterValidatorFuncBool(name, f, err)
}

// RegisterValidatorFuncBoolString is equal to
// DefaultBuilder.RegisterValidatorFuncBoolString(name, f, err).
func RegisterValidatorFuncBoolString(name string, f func(string) bool, err error) {
	DefaultBuilder.RegisterValidatorFuncBoolString(name, f, err)
}

// RegisterValidatorOneof is equal to
// DefaultBuilder.RegisterValidatorOneof(name,values...).
func RegisterValidatorOneof(name string, values ...string) {
	DefaultBuilder.RegisterValidatorOneof(name, values...)
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

var lookupStructFieldName = LookupStructFieldNameByTags("json", "query")

// Builder is used to build the validator based on the rule.
type Builder struct {
	// LookupStructFieldName is used to lookup the name of the struct field.
	//
	// If nil, use LookupStructFieldNameByTags("json", "query") instead.
	LookupStructFieldName func(reflect.StructField) string

	// Symbols is used to define the global symbols,
	// which is used by the default of GetIdentifier.
	Symbols map[string]interface{}

	// StructFieldTag is the tag name to get the validation rule.
	//
	// If empty, use the global variable StructFieldTag instead.
	StructFieldTag string

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

// RegisterSymbolNames registers a set of symbols with the names,
// the value of whose are equal to their name.
func (b *Builder) RegisterSymbolNames(names ...string) {
	for _, name := range names {
		b.RegisterSymbol(name, name)
	}
}

// RegisterFunction registers the builder function.
//
// If the function has existed, reset it to the new function.
func (b *Builder) RegisterFunction(function Function) {
	b.Builder.RegisterFunc(function.Name(), toBuilderFunction(function))
}

// RegisterValidatorFunc is a convenient method to treat the validation
// function with the name as a builder function to be registered, which
// is equal to
//
//   b.RegisterFunction(NewFunctionWithoutArgs(name, func() Validator {
//       return NewValidator(name, f)
//   }))
//
func (b *Builder) RegisterValidatorFunc(name string, f ValidatorFunc) {
	validator := NewValidator(name, f)
	b.RegisterFunction(NewFunctionWithoutArgs(name, func() Validator {
		return validator
	}))
}

// RegisterValidatorFuncBool is a convenient method to treat the bool
// validation function with the name as a builder function to be registered,
// which is equal to
//
//   b.RegisterValidatorFunc(name, BoolValidatorFunc(f, err))
//
func (b *Builder) RegisterValidatorFuncBool(name string, f func(interface{}) bool, err error) {
	b.RegisterValidatorFunc(name, BoolValidatorFunc(f, err))
}

// RegisterValidatorFuncBoolString is a convenient method to treat the string
// bool validation function with the name as a builder function to be registered,
// which is equal to
//
//   b.RegisterValidatorFunc(name, StringBoolValidatorFunc(f, err))
//
func (b *Builder) RegisterValidatorFuncBoolString(name string, f func(string) bool, err error) {
	b.RegisterValidatorFunc(name, StringBoolValidatorFunc(f, err))
}

// RegisterValidatorOneof is a convenient method to register a oneof validator
// as the builder Function, which is equal to
//
//   b.RegisterFunction(ValidatorFunction(name, validators.OneOfWithName(name, values...)))
//
func (b *Builder) RegisterValidatorOneof(name string, values ...string) {
	b.RegisterFunction(ValidatorFunction(name, internal.NewOneOf(name, values...)))
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
	tag := b.StructFieldTag
	if len(tag) == 0 {
		if tag = StructFieldTag; len(tag) == 0 {
			tag = "validate"
		}
	}

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

		rule := ft.Tag.Get(tag)
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
