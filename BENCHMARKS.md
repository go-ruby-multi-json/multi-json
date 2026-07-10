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
gem.

## Measured results — real hardware (2026-07-10)

Both sides were run on the **same host** over the **identical** `benchJSON`
payload, MRI pinned to `MultiJson.use :json_gem`, `Benchmark.realtime` over
300 000 iterations after a 50 000-iteration warm-up (steady state). Measured on
two real, non-x86 architectures.

| Arch | Host | CPU | OS | Go | Ruby (MRI) | gem |
|------|------|-----|----|----|-----------|-----|
| `s390x` | LinuxONE | IBM z15 (8561), 2 vCPU | Ubuntu 24.04 | go1.26.4 | 3.2.3 | multi_json 1.21.1 |
| `riscv64` | cfarm95 (GCC farm) | SpacemiT X60 (rv64gcv, RVV 1.0), 8 core | Debian | go1.26.4 | 3.3.8 | multi_json 1.21.1 |

### s390x (IBM z15, LinuxONE)

| Op | Go ns/op | MRI `multi_json` (json_gem) ns/op | ratio (MRI ÷ Go) |
|----|---------:|----------------------------------:|-----------------:|
| Load          | 8 542 | 10 353 | **1.21× faster** |
| LoadSymbolize | 8 931 | 11 740 | **1.31× faster** |
| Dump          | 5 723 |  7 307 | **1.28× faster** |
| DumpPretty    | 8 235 | 10 444 | **1.27× faster** |

### riscv64 (SpacemiT X60, cfarm95)

| Op | Go ns/op | MRI `multi_json` (json_gem) ns/op | ratio (MRI ÷ Go) |
|----|---------:|----------------------------------:|-----------------:|
| Load          | 86 624 | 219 491 | **2.53× faster** |
| LoadSymbolize | 89 939 | 285 892 | **3.18× faster** |
| Dump          | 52 475 | 150 137 | **2.86× faster** |
| DumpPretty    | 87 401 | 166 690 | **1.91× faster** |

The pure-Go port is **at least as fast as — and on both real arches strictly
faster than — the reference** on every load/dump path, so the standing rule is
satisfied. The margin is wider on `riscv64`, where MRI's per-call Ruby method
resolution and managed-heap object construction cost relatively more than the
compiled `encoding/json` path.
