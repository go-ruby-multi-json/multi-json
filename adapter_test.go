// SPDX-License-Identifier: BSD-3-Clause
//
// Copyright (c) 2026, the go-ruby-multi-json/multi-json authors

package multijson

import (
	"errors"
	"testing"
)

func TestSetAndGetAdapter(t *testing.T) {
	reset(t)
	if AdapterName() != DefaultAdapter {
		t.Fatalf("default adapter = %q, want %q", AdapterName(), DefaultAdapter)
	}
	for _, name := range []string{JSONGem, JSONPure, Oj, Yajl, OkJSON, NSJSONSerialization, Gson, JrJackson} {
		if err := SetAdapter(name); err != nil {
			t.Fatalf("SetAdapter(%q): %v", name, err)
		}
		if AdapterName() != name {
			t.Errorf("AdapterName = %q, want %q", AdapterName(), name)
		}
	}
	reset(t)
}

func TestSetAdapterNormalisation(t *testing.T) {
	reset(t)
	if err := SetAdapter("  OJ "); err != nil {
		t.Fatalf("SetAdapter with padding/case: %v", err)
	}
	if AdapterName() != Oj {
		t.Errorf("normalised adapter = %q, want %q", AdapterName(), Oj)
	}
	// Empty selects the default.
	if err := SetAdapter(""); err != nil {
		t.Fatal(err)
	}
	if AdapterName() != DefaultAdapter {
		t.Errorf("empty adapter -> %q, want default", AdapterName())
	}
	reset(t)
}

func TestSetAdapterUnknown(t *testing.T) {
	reset(t)
	err := SetAdapter("bogus")
	var ae *AdapterError
	if !errors.As(err, &ae) {
		t.Fatalf("error = %v, want *AdapterError", err)
	}
	if ae.Name != "bogus" {
		t.Errorf("AdapterError.Name = %q, want %q", ae.Name, "bogus")
	}
	if AdapterName() != DefaultAdapter {
		t.Errorf("adapter changed on failure to %q", AdapterName())
	}
}

func TestWithAdapter(t *testing.T) {
	reset(t)
	if err := SetAdapter(JSONGem); err != nil {
		t.Fatal(err)
	}
	var seen string
	err := WithAdapter(Yajl, func() error {
		seen = AdapterName()
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if seen != Yajl {
		t.Errorf("inside WithAdapter = %q, want %q", seen, Yajl)
	}
	if AdapterName() != JSONGem {
		t.Errorf("after WithAdapter = %q, want restored %q", AdapterName(), JSONGem)
	}
}

func TestWithAdapterPropagatesFnError(t *testing.T) {
	reset(t)
	sentinel := errors.New("boom")
	err := WithAdapter(Oj, func() error { return sentinel })
	if !errors.Is(err, sentinel) {
		t.Errorf("err = %v, want sentinel", err)
	}
	if AdapterName() != DefaultAdapter {
		t.Errorf("adapter not restored: %q", AdapterName())
	}
}

func TestWithAdapterUnknown(t *testing.T) {
	reset(t)
	called := false
	err := WithAdapter("bogus", func() error { called = true; return nil })
	var ae *AdapterError
	if !errors.As(err, &ae) {
		t.Fatalf("error = %v, want *AdapterError", err)
	}
	if called {
		t.Error("fn should not run for an unknown adapter")
	}
}

func TestCurrentAdapter(t *testing.T) {
	reset(t)
	if err := SetAdapter(Yajl); err != nil {
		t.Fatal(err)
	}
	// No options: current adapter.
	a, err := CurrentAdapter()
	if err != nil || a.Name() != Yajl {
		t.Fatalf("CurrentAdapter() = %q, %v", a.Name(), err)
	}
	// Per-call string override.
	a, err = CurrentAdapter(map[string]any{OptAdapter: Oj})
	if err != nil || a.Name() != Oj {
		t.Fatalf("override = %q, %v", a.Name(), err)
	}
	// Non-string override value falls back to the current adapter.
	a, err = CurrentAdapter(map[string]any{OptAdapter: 123})
	if err != nil || a.Name() != Yajl {
		t.Fatalf("non-string override = %q, %v", a.Name(), err)
	}
	// Unknown override -> AdapterError.
	_, err = CurrentAdapter(map[string]any{OptAdapter: "bogus"})
	var ae *AdapterError
	if !errors.As(err, &ae) {
		t.Fatalf("unknown override error = %v, want *AdapterError", err)
	}
	reset(t)
}

func TestPerCallAdapterOverrideInLoadDump(t *testing.T) {
	reset(t)
	if _, err := Load(`1`, map[string]any{OptAdapter: Oj}); err != nil {
		t.Fatalf("Load with adapter override: %v", err)
	}
	if _, err := Load(`1`, map[string]any{OptAdapter: "bogus"}); err == nil {
		t.Error("Load with bogus adapter should error")
	}
	if _, err := Dump(1, map[string]any{OptAdapter: Oj}); err != nil {
		t.Fatalf("Dump with adapter override: %v", err)
	}
	if _, err := Dump(1, map[string]any{OptAdapter: "bogus"}); err == nil {
		t.Error("Dump with bogus adapter should error")
	}
}

func TestDefaultOptionsMerge(t *testing.T) {
	reset(t)
	SetDefaultOptions(map[string]any{OptSymbolizeKeys: true})
	// Default applies.
	got, err := Load(`{"a":1}`)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := got.(map[Symbol]any); !ok {
		t.Errorf("default symbolize not applied: %T", got)
	}
	// Per-call option overrides the default.
	got, err = Load(`{"a":1}`, map[string]any{OptSymbolizeKeys: false})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := got.(map[string]any); !ok {
		t.Errorf("per-call override failed: %T", got)
	}
	// Default pretty applies to Dump.
	SetDefaultOptions(map[string]any{OptPretty: true})
	out, err := Dump(map[string]any{"a": int64(1)})
	if err != nil {
		t.Fatal(err)
	}
	if out != "{\n  \"a\": 1\n}" {
		t.Errorf("default pretty not applied: %q", out)
	}
	// DefaultOptions returns a copy that cannot mutate internal state.
	d := DefaultOptions()
	d[OptPretty] = false
	if !DefaultOptions()[OptPretty].(bool) {
		t.Error("DefaultOptions returned an aliased map")
	}
	reset(t)
}

func TestFirstOptEmptyAndNil(t *testing.T) {
	reset(t)
	// Passing an explicit nil options map is treated as no options.
	if _, err := Load(`1`, nil); err != nil {
		t.Fatalf("Load with nil opts: %v", err)
	}
}
