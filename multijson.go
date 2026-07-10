// SPDX-License-Identifier: BSD-3-Clause
//
// Copyright (c) 2026, the go-ruby-multi-json/multi-json authors

package multijson

import (
	"encoding/json"
	"strings"
	"sync"
)

// Option keys accepted by [Load] and [Dump], matching the multi_json gem's
// option symbols.
const (
	// OptSymbolizeKeys, when true, keys decoded objects with [Symbol] instead
	// of string (recursively). Mirrors the gem's :symbolize_keys.
	OptSymbolizeKeys = "symbolize_keys"
	// OptPretty, when true, dumps indented JSON. Mirrors the gem's :pretty.
	OptPretty = "pretty"
	// OptAdapter selects an adapter for a single call. Mirrors the gem's
	// :adapter.
	OptAdapter = "adapter"
)

// Symbol is the key type used for decoded objects when [OptSymbolizeKeys] is
// set, modelling a Ruby Symbol so a binding layer can distinguish a
// symbol-keyed hash from a string-keyed one.
type Symbol string

// facade holds the mutable global state (current adapter and default options),
// guarded for use under the race detector.
var facade = struct {
	mu             sync.RWMutex
	adapter        string
	defaultOptions map[string]any
}{
	adapter:        DefaultAdapter,
	defaultOptions: map[string]any{},
}

// AdapterName reports the name of the currently selected adapter (the gem's
// MultiJson.adapter, reduced to its name).
func AdapterName() string {
	facade.mu.RLock()
	defer facade.mu.RUnlock()
	return facade.adapter
}

// SetAdapter selects the adapter used by subsequent [Load]/[Dump] calls,
// mirroring MultiJson.use / MultiJson.adapter=. An empty name selects
// [DefaultAdapter]; an unknown name returns a [*AdapterError].
func SetAdapter(name string) error {
	a, err := resolveAdapter(name)
	if err != nil {
		return err
	}
	facade.mu.Lock()
	facade.adapter = a.name
	facade.mu.Unlock()
	return nil
}

// WithAdapter runs fn with name selected as the current adapter, restoring the
// previous adapter afterwards, mirroring MultiJson.with_adapter. If name cannot
// be resolved, fn is not run and the [*AdapterError] is returned.
func WithAdapter(name string, fn func() error) error {
	if _, err := resolveAdapter(name); err != nil {
		return err
	}
	prev := AdapterName()
	_ = SetAdapter(name)
	defer func() { _ = SetAdapter(prev) }()
	return fn()
}

// CurrentAdapter resolves the adapter for a set of call options: the
// [OptAdapter] key overrides the current adapter, mirroring
// MultiJson.current_adapter. An unknown override returns a [*AdapterError].
func CurrentAdapter(opts ...map[string]any) (Adapter, error) {
	o := firstOpt(opts)
	if v, ok := o[OptAdapter]; ok {
		if name, ok := v.(string); ok {
			return resolveAdapter(name)
		}
	}
	return resolveAdapter(AdapterName())
}

// DefaultOptions returns a copy of the default options merged into every
// [Load]/[Dump] call, mirroring MultiJson.default_options.
func DefaultOptions() map[string]any {
	facade.mu.RLock()
	defer facade.mu.RUnlock()
	return cloneOpts(facade.defaultOptions)
}

// SetDefaultOptions replaces the default options merged into every [Load]/[Dump]
// call, mirroring MultiJson.default_options=. A nil map clears them.
func SetDefaultOptions(opts map[string]any) {
	facade.mu.Lock()
	facade.defaultOptions = cloneOpts(opts)
	facade.mu.Unlock()
}

// Load parses a JSON string into a Go value. Options are merged over
// [DefaultOptions] with per-call keys winning. Mirrors MultiJson.load.
func Load(s string, opts ...map[string]any) (any, error) {
	o := firstOpt(opts)
	ad, err := CurrentAdapter(o)
	if err != nil {
		return nil, err
	}
	return ad.Load(s, mergeOpts(DefaultOptions(), o))
}

// Decode is an alias for [Load], mirroring MultiJson.decode.
func Decode(s string, opts ...map[string]any) (any, error) { return Load(s, opts...) }

// Dump serialises a Go value to a JSON string. Options are merged over
// [DefaultOptions] with per-call keys winning. Mirrors MultiJson.dump.
func Dump(obj any, opts ...map[string]any) (string, error) {
	o := firstOpt(opts)
	ad, err := CurrentAdapter(o)
	if err != nil {
		return "", err
	}
	return ad.Dump(obj, mergeOpts(DefaultOptions(), o))
}

// Encode is an alias for [Dump], mirroring MultiJson.encode.
func Encode(obj any, opts ...map[string]any) (string, error) { return Dump(obj, opts...) }

// convert post-processes a value decoded by encoding/json: JSON numbers become
// int64 or float64, and object keys become [Symbol] when symbolize is set.
func convert(v any, symbolize bool) any {
	switch x := v.(type) {
	case map[string]any:
		if symbolize {
			m := make(map[Symbol]any, len(x))
			for k, val := range x {
				m[Symbol(k)] = convert(val, symbolize)
			}
			return m
		}
		m := make(map[string]any, len(x))
		for k, val := range x {
			m[k] = convert(val, symbolize)
		}
		return m
	case []any:
		for i := range x {
			x[i] = convert(x[i], symbolize)
		}
		return x
	case json.Number:
		return convertNumber(x)
	default:
		return v
	}
}

// convertNumber turns a json.Number into an int64 when it is an integer in
// range, otherwise a float64, mirroring Ruby's Integer vs Float distinction.
func convertNumber(n json.Number) any {
	s := string(n)
	if !strings.ContainsAny(s, ".eE") {
		if i, err := n.Int64(); err == nil {
			return i
		}
	}
	f, _ := n.Float64()
	return f
}

// boolOpt reads a bool option, treating a missing or non-bool value as false.
func boolOpt(opts map[string]any, key string) bool {
	if v, ok := opts[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}

// firstOpt returns the first options map from a variadic argument, or an empty
// map when none was passed.
func firstOpt(opts []map[string]any) map[string]any {
	if len(opts) > 0 && opts[0] != nil {
		return opts[0]
	}
	return map[string]any{}
}

// mergeOpts returns base overlaid with over; keys in over win. Neither input is
// mutated.
func mergeOpts(base, over map[string]any) map[string]any {
	out := cloneOpts(base)
	for k, v := range over {
		out[k] = v
	}
	return out
}

// cloneOpts returns a shallow copy of an options map (nil becomes empty).
func cloneOpts(m map[string]any) map[string]any {
	out := make(map[string]any, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}
