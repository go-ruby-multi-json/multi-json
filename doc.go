// SPDX-License-Identifier: BSD-3-Clause
//
// Copyright (c) 2026, the go-ruby-multi-json/multi-json authors

// Package multijson is a pure-Go (no cgo) port of the Ruby multi_json gem: a
// swappable JSON facade that lets a program pick a JSON adapter at runtime and
// speak one uniform load/dump API against it.
//
// # Backend
//
// The Ruby gem is a thin dispatcher in front of many concrete JSON libraries
// (oj, yajl, the json gem, ok_json, ...). On a CGO=0 Go target there is exactly
// one JSON engine worth shipping — the standard library encoding/json — so all
// of the gem's adapter names ([JSONGem], [Oj], [Yajl], ...) resolve to that
// single engine. This is faithful: the observable behaviour (parsing, dumping,
// option handling, error types) matches the gem; only the invisible backend is
// unified. Calls such as the Ruby MultiJson.use :oj are accepted and honoured
// (the name is remembered and reported by [AdapterName]) but behave identically
// to every other adapter.
//
// # Data model
//
// The gem operates on Ruby values; this package operates on the Go values that
// model them: map[string]any for objects, []any for arrays, string, bool,
// int64, float64 and nil for scalars. JSON numbers decode to int64 when they
// are integral and in range, otherwise to float64 (mirroring Ruby's Integer vs
// Float). The symbolize_keys option is modelled with the distinct key type
// [Symbol]: when set, decoded objects are keyed by [Symbol] instead of string,
// recursively, which lets a Ruby binding layer (rbgo) tell a symbol-keyed hash
// from a string-keyed one.
//
// # API surface
//
// [Load]/[Decode] parse a JSON string; [Dump]/[Encode] serialise a Go value.
// Both accept an optional options map (keys [OptSymbolizeKeys], [OptPretty],
// [OptAdapter]) that is merged over [DefaultOptions] with per-call keys winning,
// exactly as the gem merges its default options. Adapter selection is managed
// with [AdapterName], [SetAdapter], [WithAdapter], [CurrentAdapter],
// [DefaultOptions] and [SetDefaultOptions].
//
// # Errors
//
// Invalid JSON yields a [*ParseError] carrying the offending Data and the
// underlying Cause (mirroring MultiJson::ParseError, aka MultiJson::LoadError,
// whose #data and #cause the gem exposes). Selecting an unregistered adapter
// yields a [*AdapterError] (the gem's MultiJson::AdapterError, an ArgumentError).
// A value encoding/json cannot serialise makes [Dump] return that engine error
// directly, as the gem lets its adapter's generator error propagate.
//
// # Limitations
//
// Because a Ruby Hash is modelled as a Go map, the insertion order of object
// keys is not preserved on [Dump]; keys are emitted in the deterministic order
// encoding/json uses (sorted). The gem preserves Hash insertion order. An
// order-preserving Ruby hash representation is a concern for the binding layer,
// not this library.
//
// The package has no dependency on any Ruby runtime: the surface is Go-typed,
// and a Ruby binding layer marshals Ruby values onto these Go types.
package multijson
