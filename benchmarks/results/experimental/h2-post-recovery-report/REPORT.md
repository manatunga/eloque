# H2 post-recovery controlled frontend comparison

Report generated: 2026-10-23. Results are from separately executed benchmark runs; see source logs for run dates.

## Methodology

- Host: Linux amd64; Intel Core Ultra 7 155H (from Go benchmark output).
- Go version: 1.27.2 (reported by operator). CPython: 3.15.0rc2, Clang 22.1.3.
- Go benchmark measurements: 10 repetitions per frontend per size, medians reported.
- CPython: only four reported medians available; raw repeats and variability unavailable.
- Workloads: 10, 100, 1,000, and 10,000 arithmetic declarations. Languages have different syntax, wrappers, grammars, ASTs, and validation work.
- Go compiler-native benchmark: cmd/compile/internal/syntax via test overlay; Go standard library: go/parser; CPython: ast.parse.
- MB/s reflects bytes of each language-specific source input, not normalized work.
- B/op measures total allocated bytes, **not** peak memory; allocs/op counts allocations, not objects retained.

## Median frontend latency

| Frontend | N=10 (µs) | N=100 (µs) | N=1,000 (ms) | N=10,000 (ms) |
|---|---:|---:|---:|---:|
| Eloque fail-fast | 6.74 | 59.97 | 0.716 | 7.866 |
| Eloque recovering | 7.51 | 71.31 | 0.819 | 8.122 |
| Go compiler-native | 12.18 | 113.28 | 1.208 | 11.218 |
| Go go/parser | 18.49 | 170.93 | 1.865 | 18.208 |
| CPython 3.15.0rc2 | 87.46 | 908.59 | 11.106 | 141.661 |

## N=10,000 throughput and allocation

| Frontend | MB/s | Statements/s | Allocated MiB/op | Allocs/op |
|---|---:|---:|---:|---:|
| Eloque fail-fast | 68.22 | 1,271,280 | 14.62 | 98.0 |
| Eloque recovering | 66.08 | 1,231,240 | 14.62 | 97.5 |
| Go compiler-native | 44.28 | 891,397 | 14.42 | 200,010.0 |
| Go go/parser | 27.27 | 549,198 | 10.73 | 230,126.0 |
| CPython 3.15.0rc2 | Not reported | 70,591 | Not reported | Not reported |

## Eloque recovering overhead

| N | Recovery latency overhead | Allocation count difference |
|---|---:|---:|
| 10 | +11.48% | +0.0 |
| 100 | +18.91% | +0.0 |
| 1,000 | +14.36% | +0.0 |
| 10,000 | +3.25% | -0.5 |

## Relative Eloque fail-fast speedup at N=10,000

- Against Go compiler-native: **1.43×**.
- Against Go go/parser: **2.31×**.
- Against CPython 3.15.0rc2: **18.01×**.

## Reproducibility and caveats

The source logs, overlay harness, generator script, and exact invocation commands should be retained with this report. Benchmark source generation and parser input validation should be checked against the implementation to confirm identical measured operations. The compiler-native overlay benchmark measures parsing and AST construction, not the entire Go compiler pipeline. The Eloque benchmarks measure frontend parsing, not static type checking, bytecode generation, or runtime execution.

Cross-language comparisons are illustrative, **not apples-to-apples compiler-performance rankings**. Comparisons of MB/s, allocations, and latency are especially sensitive to source length and AST structure. Separate benchmark runs may be affected by thermal and scheduling drift; no statistical significance claim is made. Python sample-level variance cannot be computed from its current summary output.

Historical October 20 H2 official results must remain unchanged. These post-recovery measurements form a new experimental baseline.
