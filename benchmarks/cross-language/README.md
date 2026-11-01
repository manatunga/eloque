
## M2 — Cross-Language Name-Resolution Benchmarks

### Purpose

M2 measures Eloque's name-resolution performance and provides contextual
measurements from six other language implementations:

- Eloque (Go implementation)
- CPython
- C (Clang)
- C++ (Clang++)
- Rust (rustc)
- Zig (zig)
- Go (Go compiler)

The benchmark uses five workload families:

| Workload | Description |
|---|---|
| `declarations` | Introduce N identifiers and reference each |
| `references` | Resolve N references to one declared identifier |
| `mixed` | Interleave N declarations with N references |
| `sparse-undefined` | Introduce an unresolved identifier every 100 references |
| `undefined` | Reference N unresolved identifiers |

Sizes: N = 10, 100, 1000, 3000.

### Measurement boundaries

**These measurements are not equivalent compiler phases.**

| Engine | Measured operation |
|---|---|
| Eloque | Resolver on an already constructed AST |
| CPython `ast.parse` | Python source parsing |
| CPython `symtable` | Symbol-table construction, including internal parsing |
| C | Clang subprocess, syntax and semantic checking |
| C++ | Clang++ subprocess, syntax and semantic checking |
| Rust | rustc metadata compilation subprocess |
| Zig | Zig AST checking subprocess |
| Go | Go compiler subprocess |

Native compiler measurements include process startup and compiler work beyond
identifier resolution.

Undefined-name workloads also differ across implementations. CPython permits
unresolved names during parsing and symbol-table construction, whereas native
compilers may reject them.

Consequently, the results must not be interpreted as a direct ranking of
name-resolver implementations.

### Files

- `m2_cross_language.py`: Main benchmark collector and report generator.
- `m2_native.py`: Native compiler source generators and subprocess adapters.
- `test_m2_native.py`: Source-generation, command, and compiler integration tests.

### Prerequisites

Run commands from the Eloque repository root.

Required:
- Python 3
- Go
- Eloque's existing Go resolver benchmarks

For all native engines:
- `clang`
- `clang++`
- `rustc`
- `zig`
- `go`

Check compiler availability:

```bash
command -v python3 go clang clang++ rustc zig
```

### Validate the adapters

```bash
python3 -m unittest discover \
  -s benchmarks/cross-language \
  -p 'test_m2_native.py' -v
```

### Run a smoke benchmark

```bash
python3 benchmarks/cross-language/m2_cross_language.py \
  --repeats 1 \
  --benchtime 100ms \
  --python-duration 0.01 \
  --out benchmarks/results/cross-language/m2-semantic-smoke
```

This runs all seven language implementations.

To restrict native compiler measurements:

```bash
python3 benchmarks/cross-language/m2_cross_language.py \
  --repeats 1 \
  --benchtime 100ms \
  --python-duration 0.01 \
  --native-languages c cpp go \
  --out benchmarks/results/cross-language/m2-semantic-smoke
```

To run only Eloque and CPython, use `--skip-native`.

### Full benchmark

```bash
python3 benchmarks/cross-language/m2_cross_language.py \
  --repeats 10 \
  --benchtime 1s \
  --python-duration 0.1 \
  --out benchmarks/results/cross-language/m2-semantic-v1
```

### Output

The collector writes:

- `samples.csv`: Individual timing measurements.
- `REPORT.md`: Median timings, grouped by measurement boundary.
- `manifest.json`: Run configuration and native workload metadata.
- `eloque-resolver.log`: Raw Go benchmark output, when Eloque is enabled.

Native source generation and initial validation occur outside the timed region.

### Interpretation and limitations

1. Do not directly compare native compiler subprocess latency with Eloque's
   isolated resolver latency.
2. Compare scaling trends within the same engine and workload.
3. Consider source size and the number of declarations and references.
4. Error-path measurements may be affected by diagnostic generation and
   compiler error limits.
5. Eloque's complete parse-to-resolve pipeline is not measured by this collector.
6. Compiler versions, hardware, system load, and run configuration should be
   recorded when publishing benchmark results.
