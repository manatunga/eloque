# Eloque / Go compiler / CPython controlled frontend benchmark

Each engine: 10 repetitions per size. Lower is better.

| Engine | N=10 (µs) | N=100 (µs) | N=1,000 (ms) | N=10,000 (ms) |
|---|---:|---:|---:|---:|
| Eloque H2 fail-fast | 6.92 | 64.43 | 0.750 | 7.207 |
| Eloque H2 recovering | 6.88 | 68.18 | 0.822 | 8.470 |
| Go compiler parser | 11.75 | 108.40 | 1.251 | 11.462 |
| CPython ast.parse | 104.31 | 1049.87 | 12.081 | 160.605 |

## Scope

All measurements exclude subprocess startup from the timed operation. The Go compiler benchmark uses cmd/compile/internal/syntax via a generated test overlay. Eloque benchmarks reuse the existing production parser benchmark functions. CPython uses ast.parse. Grammars, wrappers, ASTs, and diagnostic responsibilities differ. This is not a full compiler or runtime performance ranking.

Raw samples: samples.csv; Go logs and Python log are retained; environment: manifest.json.
