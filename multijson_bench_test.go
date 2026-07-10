// SPDX-License-Identifier: BSD-3-Clause
//
// Copyright (c) 2026, the go-ruby-multi-json/multi-json authors

package multijson

import "testing"

// benchJSON is a representative mixed-type payload used by the load/dump
// benchmarks. See BENCHMARKS.md for the methodology comparing these numbers
// against the MRI multi_json gem.
const benchJSON = `{"id":12345,"name":"widget","price":19.99,"active":true,` +
	`"tags":["a","b","c"],"meta":{"w":100,"h":200,"nested":{"x":1,"y":2}},` +
	`"items":[{"k":1},{"k":2},{"k":3},{"k":4},{"k":5}]}`

func BenchmarkLoad(b *testing.B) {
	_ = SetAdapter(DefaultAdapter)
	SetDefaultOptions(nil)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := Load(benchJSON); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkLoadSymbolize(b *testing.B) {
	opts := map[string]any{OptSymbolizeKeys: true}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := Load(benchJSON, opts); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDump(b *testing.B) {
	v, err := Load(benchJSON)
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := Dump(v); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDumpPretty(b *testing.B) {
	v, err := Load(benchJSON)
	if err != nil {
		b.Fatal(err)
	}
	opts := map[string]any{OptPretty: true}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := Dump(v, opts); err != nil {
			b.Fatal(err)
		}
	}
}
