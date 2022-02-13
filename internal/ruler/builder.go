// The MIT License (MIT)
//
// Copyright (c) 2016-2020 Containous SAS; 2020-2021 Traefik Labs; 2022 xgfone
//
// Permission is hereby granted, free of charge, to any person obtaining
// a copy of this software and associated documentation files (the "Software"),
// to deal in the Software without restriction, including without limitation
// the rights to use, copy, modify, merge, publish, distribute, sublicense,
// and/or sell copies of the Software, and to permit persons to whom the
// Software is furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
// THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

// Package ruler provides the matcher rule parser and builder.
package ruler

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/xgfone/go-apiserver/http/matcher"
	"github.com/xgfone/predicate"
)

var funcs = map[string]MatcherBuidler{
	"ClientIP":      clientIP,
	"Method":        methods,
	"Query":         query,
	"Path":          path,
	"PathPrefix":    pathPrefix,
	"Host":          host,
	"HostRegexp":    hostRegexp,
	"Headers":       headers,
	"HeadersRegexp": headersRegexp,
}

// MatcherBuidler is used to build the http route matcher.
type MatcherBuidler func(ctx *BuilderContext, args ...string) error

// RegisterMatcherBuilder registers the http route matcher builder.
//
// If the named matcher builder has been registered, reset it to the new.
func RegisterMatcherBuilder(name string, builder MatcherBuidler) {
	funcs[name] = builder
}

// BuilderContext is the context to build the http route matcher.
type BuilderContext struct {
	Matchers matcher.Matchers
}

// New returns a new builder context without the original matchers.
func (c *BuilderContext) New() *BuilderContext {
	return &BuilderContext{}
}

// AppendMatchers appends the route matchers.
func (c *BuilderContext) AppendMatchers(matcher ...matcher.Matcher) {
	c.Matchers = append(c.Matchers, matcher...)
}

// DefaultBuilder is the default builder of the http route matcher rule,
// which is used by the function Build to build the matcher rule.
var DefaultBuilder = NewBuilder()

// Build is used to build the http route matcher rule as a matcher.
func Build(matcherRule string) (matcher.Matcher, error) {
	return DefaultBuilder.Parse(matcherRule)
}

// Builder parses the rule string and build it as the http route matcher.
type Builder struct {
	parser predicate.Parser
}

// NewBuilder returns a new route builder.
func NewBuilder() *Builder {
	parser, _ := newParser()
	return &Builder{parser: parser}
}

// Parse parses the rule string and builds it as the http route matcher.
func (r *Builder) Parse(matchRule string) (matcher.Matcher, error) {
	parse, err := r.parser.Parse(matchRule)
	if err != nil {
		return nil, fmt.Errorf("fail to parse route match rule '%s': %w", matchRule, err)
	}

	buildTree, ok := parse.(treeBuilder)
	if !ok {
		return nil, fmt.Errorf("fail to parse route match rule '%s'", matchRule)
	}

	ctx := &BuilderContext{}
	err = _addRuleOnBuilder(ctx, buildTree())
	if err != nil {
		return nil, err
	}

	return matcher.And(ctx.Matchers...), nil
}

func clientIP(ctx *BuilderContext, clientIPs ...string) (err error) {
	matchers := make([]matcher.Matcher, len(clientIPs))
	for i, clientIP := range clientIPs {
		matchers[i], err = matcher.ClientIP(clientIP)
		if err != nil {
			return
		}
	}
	ctx.AppendMatchers(matcher.Or(matchers...))
	return
}

func methods(ctx *BuilderContext, methods ...string) (err error) {
	matchers := make([]matcher.Matcher, len(methods))
	for i, method := range methods {
		matchers[i], err = matcher.Method(method)
		if err != nil {
			return
		}
	}
	ctx.AppendMatchers(matcher.Or(matchers...))
	return
}

func path(ctx *BuilderContext, paths ...string) (err error) {
	matchers := make([]matcher.Matcher, len(paths))
	for i, path := range paths {
		matchers[i], err = matcher.Path(path)
		if err != nil {
			return
		}
	}
	ctx.AppendMatchers(matcher.Or(matchers...))
	return
}

func pathPrefix(ctx *BuilderContext, paths ...string) (err error) {
	matchers := make([]matcher.Matcher, len(paths))
	for i, path := range paths {
		matchers[i], err = matcher.PathPrefix(path)
		if err != nil {
			return
		}
	}
	ctx.AppendMatchers(matcher.Or(matchers...))
	return
}

func query(ctx *BuilderContext, query ...string) (err error) {
	matchers := make([]matcher.Matcher, len(query))
	for i, q := range query {
		if index := strings.IndexByte(q, '='); index > -1 {
			matchers[i], err = matcher.Query(q[:index], q[index+1:])
		} else {
			matchers[i], err = matcher.Query(q, "")
		}
		if err != nil {
			return
		}
	}
	ctx.AppendMatchers(matcher.And(matchers...))
	return
}

func headers(ctx *BuilderContext, headers ...string) (err error) {
	var m matcher.Matcher
	switch _len := len(headers); _len {
	case 0:
	case 1:
		if m, err = matcher.Header(headers[0], ""); err == nil {
			ctx.AppendMatchers(m)
		}
	default:
		if m, err = matcher.Header(headers[0], headers[1]); err == nil {
			ctx.AppendMatchers(m)
		}
	}

	return
}

func headersRegexp(ctx *BuilderContext, headers ...string) (err error) {
	var m matcher.Matcher
	switch _len := len(headers); _len {
	case 0:
	case 1:
		m, err = matcher.HeaderRegexp(headers[0], "")
		if err == nil {
			ctx.AppendMatchers(m)
		}
	default:
		m, err = matcher.Header(headers[0], headers[1])
		if err == nil {
			ctx.AppendMatchers(m)
		}
	}

	return
}

func host(ctx *BuilderContext, hosts ...string) (err error) {
	matchers := make([]matcher.Matcher, len(hosts))
	for i, host := range hosts {
		if !IsASCII(host) {
			return fmt.Errorf("invalid regexp host '%s': non-ASCII characters are not allowed", host)
		}

		matchers[i], err = matcher.Host(strings.ToLower(host))
		if err != nil {
			return
		}
	}
	ctx.AppendMatchers(matcher.Or(matchers...))
	return
}

func hostRegexp(ctx *BuilderContext, hosts ...string) (err error) {
	matchers := make(matcher.Matchers, len(hosts))
	for i, host := range hosts {
		if !IsASCII(host) {
			return fmt.Errorf("invalid regexp host '%s': non-ASCII characters are not allowed", host)
		}

		matchers[i], err = matcher.HostRegexp(host)
		if err != nil {
			return
		}
	}
	ctx.AppendMatchers(matcher.Or(matchers...))
	return
}

func addRuleOnBuilder(ctx *BuilderContext, rule *tree) (err error) {
	switch rule.matcher {
	case "and":
		newctx := ctx.New()
		if err = _addRuleOnBuilder(newctx, rule.ruleLeft); err != nil {
			return
		}
		if err = _addRuleOnBuilder(newctx, rule.ruleRight); err != nil {
			return
		}

		if len(newctx.Matchers) > 0 {
			// ctx.AppendMatchers(newctx.Matchers...)
			ctx.AppendMatchers(matcher.And(newctx.Matchers...))
		}

		return

	case "or":
		if err = addRuleOnBuilder(ctx, rule.ruleLeft); err != nil {
			return
		}

		return addRuleOnBuilder(ctx, rule.ruleRight)

	default:
		if err = checkRule(rule); err != nil {
			return
		}

		newctx := ctx.New()
		if rule.not {
			err = not(funcs[rule.matcher])(newctx, rule.value...)
		} else {
			err = funcs[rule.matcher](newctx, rule.value...)
		}

		if err != nil {
			return
		}

		if len(newctx.Matchers) > 0 {
			ctx.AppendMatchers(newctx.Matchers...)
			// ctx.AppendMatchers(matcher.And(newctx.Matchers...))
		}

		return
	}
}

func not(b MatcherBuidler) MatcherBuidler {
	return func(ctx *BuilderContext, args ...string) (err error) {
		newctx := ctx.New()
		if err = b(newctx, args...); err != nil {
			return
		}

		switch len(newctx.Matchers) {
		case 0:
		case 1:
			ctx.AppendMatchers(matcher.Not(newctx.Matchers[0]))
		default:
			ctx.AppendMatchers(matcher.Not(matcher.And(newctx.Matchers...)))
		}

		return
	}
}

func _addRuleOnBuilder(ctx *BuilderContext, rule *tree) (err error) {
	switch rule.matcher {
	case "and":
		if err = _addRuleOnBuilder(ctx, rule.ruleLeft); err != nil {
			return
		}
		return _addRuleOnBuilder(ctx, rule.ruleRight)

	case "or":
		newctx := ctx.New()
		if err = addRuleOnBuilder(newctx, rule.ruleLeft); err != nil {
			return
		}
		if err = addRuleOnBuilder(newctx, rule.ruleRight); err != nil {
			return
		}

		if len(newctx.Matchers) > 0 {
			ctx.AppendMatchers(matcher.Or(newctx.Matchers...))
		}

		return

	default:
		if err = checkRule(rule); err != nil {
			return err
		}

		if rule.not {
			return not(funcs[rule.matcher])(ctx, rule.value...)
		}

		return funcs[rule.matcher](ctx, rule.value...)
	}
}

func checkRule(rule *tree) error {
	if len(rule.value) == 0 {
		return fmt.Errorf("no args for the route matcher rule '%s'", rule.matcher)
	}

	for _, v := range rule.value {
		if len(v) == 0 {
			return fmt.Errorf("empty args for the route rule matcher '%s': %v",
				rule.matcher, rule.value)
		}
	}

	return nil
}

// IsASCII checks if the given string contains only ASCII characters.
func IsASCII(s string) bool {
	for i, _len := 0, len(s); i < _len; i++ {
		if s[i] >= utf8.RuneSelf {
			return false
		}
	}
	return true
}
