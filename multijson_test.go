// SPDX-License-Identifier: BSD-3-Clause
//
// Copyright (c) 2026, the go-ruby-multi-json/multi-json authors

package multijson

import (
	"errors"
	"reflect"
	"testing"
)

// reset restores the package globals to their defaults so state-mutating tests
// do not leak into one another.
func reset(t *testing.T) {
	t.Helper()
	if err := SetAdapter(DefaultAdapter); err != nil {
		t.Fatalf("reset adapter: %v", err)
	}
	SetDefaultOptions(nil)
}

func TestLoadScalarsAndContainers(t *testing.T) {
	reset(t)
	cases := []struct {
		in   string
		want any
	}{
		{`null`, nil},
		{`true`, true},
		{`false`, false},
		{`"hi"`, "hi"},
		{`1`, int64(1)},
		{`-42`, int64(-42)},
		{`1.5`, float64(1.5)},
		{`1e2`, float64(100)},
		{`1E2`, float64(100)},
		{`100000000000000000000000`, float64(1e23)}, // beyond int64 -> float
		{`[1,2,3]`, []any{int64(1), int64(2), int64(3)}},
		{`{"a":1}`, map[string]any{"a": int64(1)}},
		{`{"a":{"b":[true,null]}}`, map[string]any{"a": map[string]any{"b": []any{true, nil}}}},
	}
	for _, c := range cases {
		got, err := Load(c.in)
		if err != nil {
			t.Fatalf("Load(%q) error: %v", c.in, err)
		}
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("Load(%q) = %#v, want %#v", c.in, got, c.want)
		}
	}
}

func TestDecodeAlias(t *testing.T) {
	reset(t)
	got, err := Decode(`[1]`)
	if err != nil || !reflect.DeepEqual(got, []any{int64(1)}) {
		t.Fatalf("Decode = %#v, %v", got, err)
	}
}

func TestSymbolizeKeys(t *testing.T) {
	reset(t)
	got, err := Load(`{"a":{"b":1}}`, map[string]any{OptSymbolizeKeys: true})
	if err != nil {
		t.Fatal(err)
	}
	want := map[Symbol]any{Symbol("a"): map[Symbol]any{Symbol("b"): int64(1)}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("symbolized = %#v, want %#v", got, want)
	}
}

func TestSymbolizeKeysNonBoolIgnored(t *testing.T) {
	reset(t)
	// A non-bool option value is treated as false (string keys retained).
	got, err := Load(`{"a":1}`, map[string]any{OptSymbolizeKeys: "yes"})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := got.(map[string]any); !ok {
		t.Errorf("expected string-keyed map, got %T", got)
	}
}

func TestLoadParseErrors(t *testing.T) {
	reset(t)
	for _, in := range []string{`{`, ``, `not json`} {
		_, err := Load(in)
		var pe *ParseError
		if !errors.As(err, &pe) {
			t.Fatalf("Load(%q) error = %v, want *ParseError", in, err)
		}
		if pe.Data != in {
			t.Errorf("ParseError.Data = %q, want %q", pe.Data, in)
		}
		if pe.Cause == nil {
			t.Errorf("ParseError.Cause is nil for %q", in)
		}
		if pe.Unwrap() != pe.Cause {
			t.Errorf("Unwrap mismatch")
		}
	}
}

func TestLoadTrailingData(t *testing.T) {
	reset(t)
	_, err := Load(`1 2`)
	var pe *ParseError
	if !errors.As(err, &pe) {
		t.Fatalf("error = %v, want *ParseError", err)
	}
	if !errors.Is(err, errTrailingData) {
		t.Errorf("cause = %v, want trailing-data error", pe.Cause)
	}
	if errTrailingData.Error() == "" {
		t.Error("trailing-data error message empty")
	}
}

func TestDump(t *testing.T) {
	reset(t)
	cases := []struct {
		in   any
		want string
	}{
		{nil, `null`},
		{true, `true`},
		{int64(1), `1`},
		{"a<b>&c", `"a<b>&c"`}, // no HTML escaping, matching Ruby JSON
		{[]any{int64(1), "x"}, `[1,"x"]`},
		{map[string]any{"a": int64(1)}, `{"a":1}`},
		{map[Symbol]any{Symbol("a"): int64(1)}, `{"a":1}`}, // symbol keys round-trip
	}
	for _, c := range cases {
		got, err := Dump(c.in)
		if err != nil {
			t.Fatalf("Dump(%#v) error: %v", c.in, err)
		}
		if got != c.want {
			t.Errorf("Dump(%#v) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestDumpPretty(t *testing.T) {
	reset(t)
	got, err := Dump(map[string]any{"a": int64(1)}, map[string]any{OptPretty: true})
	if err != nil {
		t.Fatal(err)
	}
	want := "{\n  \"a\": 1\n}"
	if got != want {
		t.Errorf("pretty = %q, want %q", got, want)
	}
}

func TestEncodeAlias(t *testing.T) {
	reset(t)
	got, err := Encode(int64(7))
	if err != nil || got != "7" {
		t.Fatalf("Encode = %q, %v", got, err)
	}
}

func TestDumpUnencodable(t *testing.T) {
	reset(t)
	_, err := Dump(make(chan int))
	if err == nil {
		t.Fatal("expected error dumping a channel")
	}
	// The engine error propagates directly, not wrapped in a ParseError.
	var pe *ParseError
	if errors.As(err, &pe) {
		t.Errorf("dump error should not be a ParseError: %v", err)
	}
}

func TestRoundTrip(t *testing.T) {
	reset(t)
	orig := map[string]any{"n": int64(3), "f": 2.5, "s": "x", "b": true, "z": nil, "a": []any{int64(1)}}
	enc, err := Dump(orig)
	if err != nil {
		t.Fatal(err)
	}
	back, err := Load(enc)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(back, orig) {
		t.Errorf("round-trip = %#v, want %#v", back, orig)
	}
}
