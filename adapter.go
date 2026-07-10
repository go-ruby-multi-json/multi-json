// SPDX-License-Identifier: BSD-3-Clause
//
// Copyright (c) 2026, the go-ruby-multi-json/multi-json authors

package multijson

import (
	"bytes"
	"encoding/json"
	"strings"
)

// Adapter names accepted by the facade. Each mirrors an adapter shipped by the
// multi_json gem; on this CGO=0 target they all resolve to the single
// encoding/json backend (see the package doc). Their observable behaviour is
// identical; only the name reported by [AdapterName] differs.
const (
	JSONGem             = "json_gem"            // Ruby's stdlib json gem
	JSONPure            = "json_pure"           // json gem's pure-Ruby variant
	Oj                  = "oj"                  // the oj C extension
	Yajl                = "yajl"                // the yajl-ruby C extension
	OkJSON              = "ok_json"             // the bundled fallback parser
	NSJSONSerialization = "nsjsonserialization" // RubyMotion / Cocoa
	Gson                = "gson"                // JRuby, Google gson
	JrJackson           = "jr_jackson"          // JRuby, Jackson
)

// DefaultAdapter is the adapter selected before any [SetAdapter] call. The gem
// probes for the fastest available backend; with a unified engine we pick the
// stdlib json gem name, which is the closest match to encoding/json.
const DefaultAdapter = JSONGem

// knownAdapters is the registry of accepted adapter names.
var knownAdapters = map[string]struct{}{
	JSONGem:             {},
	JSONPure:            {},
	Oj:                  {},
	Yajl:                {},
	OkJSON:              {},
	NSJSONSerialization: {},
	Gson:                {},
	JrJackson:           {},
}

// resolveAdapter validates and normalises an adapter name. An empty name selects
// [DefaultAdapter], mirroring the gem's use(nil). An unknown name yields a
// [*AdapterError], as MultiJson.use raises MultiJson::AdapterError.
func resolveAdapter(name string) (Adapter, error) {
	n := strings.ToLower(strings.TrimSpace(name))
	if n == "" {
		n = DefaultAdapter
	}
	if _, ok := knownAdapters[n]; !ok {
		return Adapter{}, &AdapterError{Name: name}
	}
	return Adapter{name: n}, nil
}

// Adapter is a resolved JSON adapter. Every adapter shares the encoding/json
// engine; the name is retained for [Name] and [AdapterName] reporting.
type Adapter struct {
	name string
}

// Name reports the adapter's registered name.
func (a Adapter) Name() string { return a.name }

// Load parses a JSON string into a Go value, honouring the symbolize_keys
// option. A parse failure is wrapped in a [*ParseError] carrying the input and
// the underlying cause, exactly as the gem wraps adapter errors.
func (a Adapter) Load(s string, opts map[string]any) (any, error) {
	dec := json.NewDecoder(strings.NewReader(s))
	dec.UseNumber()
	var raw any
	if err := dec.Decode(&raw); err != nil {
		return nil, &ParseError{Data: s, Cause: err}
	}
	// Reject trailing tokens after the first value (e.g. "1 2"), which the
	// Ruby JSON parser also rejects.
	if dec.More() {
		return nil, &ParseError{Data: s, Cause: errTrailingData}
	}
	return convert(raw, boolOpt(opts, OptSymbolizeKeys)), nil
}

// Dump serialises a Go value to a JSON string, honouring the pretty option. A
// value the engine cannot serialise makes Dump return that engine error
// directly, as the gem lets its generator error propagate.
func (a Adapter) Dump(obj any, opts map[string]any) (string, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	// Ruby's JSON generator does not HTML-escape <, > or &.
	enc.SetEscapeHTML(false)
	if boolOpt(opts, OptPretty) {
		enc.SetIndent("", "  ")
	}
	if err := enc.Encode(obj); err != nil {
		return "", err
	}
	// Encoder.Encode appends a newline the gem's generators do not.
	return strings.TrimSuffix(buf.String(), "\n"), nil
}

// errTrailingData is the cause reported when input carries tokens after the
// first parsed JSON value.
var errTrailingData = trailingDataError{}

type trailingDataError struct{}

func (trailingDataError) Error() string { return "unexpected token after JSON value" }
