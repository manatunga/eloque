# Eloque Benchmarks

This directory contains performance benchmarks for Eloque's lexer
and parser, historical optimization results, token-delivery
architecture experiments, and cross-language frontend comparisons
against Go and CPython.

## Directory Structure

```text
benchmarks/
├── README.md
├── cpython/
│   ├── frontend.py
│   └── frontend_controlled.py
├── go/
│   ├── frontend_test.go
│   └── frontend_controlled_test.go
└── results/
    ├── current/
    ├── experimental/
    └── historical/
```

- `current/`: Current performance baselines.
- `experimental/`: Measurements from architecture and optimization experiments.
- `historical/`: Preserved measurements from earlier implementations.

Eloque's lexer and parser benchmarks reside in `internal/lexer/`
and `internal/parser/`, respectively.

Cross-language Go frontend benchmarks reside in `benchmarks/go/`.
CPython frontend benchmarks reside in `benchmarks/cpython/`.

## Benchmark Result Naming Convention

Benchmark result files use:

`<checkpoint>_<experiment>_<variant>_<YYYYMMDD>.<extension>`

- `checkpoint`: Roadmap milestone or checkpoint, such as `m1-p5a`.
- `experiment`: Subject being measured, such as `frontend` or `token-delivery`.
- `variant`: Implementation or comparison, such as `streaming`, `buffered`,
  `baseline`, or `comparison`.
- `YYYYMMDD`: Actual execution date.
- `extension`: Output format, usually `.txt`.

Use lowercase kebab-case within components and underscores between
components.

For multiple results with otherwise identical filenames, append a run
suffix such as `_r02` before the extension.

Never invent historical execution dates or checkpoint assignments.
Preserve the original result contents when renaming files.

Benchmark comparisons should record the environment, command,
methodology, and implementation revision where available.

## Benchmark Environment

The recorded measurements were collected using:

| Component | Environment |
|---|---|
| CPU | Intel Core Ultra 7 155H |
| Architecture | AMD64 |
| Operating system | Linux under WSL2 |
| Go | 1.27.2 |
| Python implementation | CPython |
| Python version | 3.15.0rc2 |
| Python build compiler | Clang 22.1.3 |

Results are machine-dependent and may vary with system load, thermal conditions, compiler versions, and runtime configuration.

## Reproducing the Benchmarks

Run commands from the repository root.

### Eloque Lexer

```bash
go test ./internal/lexer \
  -run '^$' -bench . -benchmem \
  -benchtime=1s -count=3
```

### Eloque Parser and Complete Frontend

```bash
go test ./internal/parser \
  -run '^$' -bench . -benchmem \
  -benchtime=1s -count=3
```

The production Eloque frontend incrementally requests tokens from the lexer while parsing and constructing its AST. Buffered frontend benchmarks are retained as a reference implementation. Both measure lexing, parsing, and AST construction, but differ in token-delivery strategy.

### Go Cross-Language Frontend

The Go comparison benchmarks use the standard library's
`go/parser.ParseFile` to measure source parsing and AST construction.

Run from the repository root:

```bash
go test ./benchmarks/go \
  -run '^$' -bench . -benchmem \
  -benchtime=1s -count=5
```

The controlled arithmetic and language-native benchmark cases
are implemented in `benchmarks/go/`.

### CPython Frontend

The Python benchmarks use CPython 3.15.0rc2 through uv.

```bash
uv python install 3.15
uv run --no-project python benchmarks/cpython/frontend.py
```

For the M1-P5-C controlled arithmetic comparison, run:

```bash
uv run --no-project python benchmarks/cpython/frontend_controlled.py
```

This benchmark measures CPython's `ast.parse()` using
arithmetic expressions corresponding to the controlled Go
and Eloque workloads.

Record the exact Python interpreter version and benchmark
configuration when reproducing results.

Verify the exact interpreter version before comparing results. A Python 3.15 release candidate and a later stable release may produce different measurements.

## Workloads

Both frontends are evaluated using five categories:

| Workload | Description |
|---|---|
| Declarations | Variable initialization |
| Expressions | Arithmetic expressions with grouping, precedence, and exponentiation |
| Strings | Output statements containing string literals |
| Comments | Comments interspersed with variable declarations |
| Mixed | Comments, declarations, reassignment, arithmetic, and output statements |

The M1-P5-C cross-language validation uses four workload sizes:
10, 100, 1,000, and 10,000 units.

Earlier preliminary frontend comparisons used 10, 100,
and 1,000 units.

The programs express comparable operations using the syntax of their respective languages. They are not byte-for-byte identical and do not necessarily produce equivalent AST structures.

## M1-P5-A — Token-delivery architecture experiments

Three experiments evaluated the performance and correctness implications of eager token buffering, direct streaming, and bounded lookahead.

These are experimental measurements and must not be confused with the canonical frontend performance baselines or the preliminary CPython comparisons.

### Experiment 1 — Token delivery

Compared eager token buffering with direct streaming independently of AST construction.

Direct streaming outperformed buffering across all 20 tested workload-size combinations.

Because this experiment excluded AST construction, its results were treated as preliminary architectural evidence rather than sufficient grounds for the final decision.

### Experiment 2 — Complete frontend integration

Compared eager buffering and direct streaming while performing lexing, parsing, and AST construction.

Five workload categories were tested:

- Declarations
- Expressions
- Strings
- Comments
- Mixed

Each workload was measured at four input sizes: 10, 100, 1,000, and 10,000.

Five benchmark runs were collected, and median results were used for comparison.

Direct streaming was faster in 19 of 20 workload-size comparisons and allocated fewer bytes in all 20.

At the largest input size, allocated-byte reductions ranged from approximately 22% to 52%, depending on the workload.

The number of allocations per operation remained broadly comparable between implementations.

### Experiment 3 — Bounded lookahead comparison

Compared direct streaming, bounded two-token lookahead, and eager buffering.

Direct streaming outperformed the alternatives across the 20 tested token-delivery comparisons.

The result supports avoiding additional lookahead machinery until a concrete grammar requirement justifies it.

### Correctness validation

Experimental correctness testing included:

- AST structural equivalence between buffered and streaming frontends.
- Stable EOF handling.
- Lexical error propagation.
- Source-position preservation.
- Targeted lexical and syntax diagnostic regression cases.

These tests support the selected architecture for the currently implemented grammar. They do not establish correctness for future grammar features.

### Experimental results

Store experimental benchmark output under:

`benchmarks/results/experimental/`

Keep these results separate from:

- `results/current/` — canonical implementation measurements.
- `results/historical/` — historical implementation baselines.

The experimental results are evidence for an architectural decision, not a replacement for production performance-regression measurements.

### Architectural outcome

Direct pull-based streaming was selected for production integration.

Bounded lookahead remains deferred until required by a concrete grammar feature. Eager buffering remains useful for correctness comparisons and regression testing.

The decision is tracked in `docs/TECHNICAL_BACKLOG.md`, and implementation progress is tracked in `docs/ROADMAP.md`.

## M1-P5-C — Cross-language frontend validation

Eloque's production streaming frontend was benchmarked against
Go's `go/parser.ParseFile` and CPython 3.15.0rc2 `ast.parse()`.

Two benchmark suites were evaluated:

1. **Language-native workloads:** Declarations, expressions,
  strings, comments, and mixed statements, each at sizes of
  10, 100, 1,000, and 10,000.
2. **Controlled arithmetic:** Identical arithmetic expressions
  across all three languages, using language-specific declarations
  and program wrappers.

### Controlled arithmetic results

Median frontend latency at N=10,000:

| Frontend | Median latency |
|---|---:|
| Eloque | 9.774 ms |
| Go | 17.809 ms |
| CPython 3.15.0rc2 | 145.350 ms |

Eloque recorded approximately 1.82× lower latency relative to
Go and 14.87× relative to CPython for this workload.

### Methodology and limitations

- Eloque and Go used Go's benchmark framework with five runs.
- CPython used `timeit` with five repeated measurements.
- Each measurement included lexical analysis, parsing, and
  AST construction.
- The controlled arithmetic expressions were identical, but
  declarations, wrappers, and AST representations differed.
- Go's parser benchmark did not perform type checking.
- CPython's public `tokenize` benchmark was measured separately
  and must not be added to `ast.parse()` timings.
- Different grammar coverage, source sizes, runtime behavior,
  and measurement harnesses limit direct comparability.
- These results do not establish general execution speed or
  compiler performance superiority.

Experimental measurements are preserved under
`benchmarks/results/experimental/`.

## Preliminary Frontend Results

The following results represent the 1,000-unit workloads.

| Workload | Eloque | CPython 3.15.0rc2 | Relative speed |
|---|---:|---:|---:|
| Declarations | 0.251 ms | 2.199 ms | 8.8× |
| Expressions | 1.268 ms | 8.920 ms | 7.0× |
| Strings | 0.444 ms | 3.655 ms | 8.2× |
| Comments | 0.438 ms | 2.215 ms | 5.1× |
| Mixed | 1.704 ms | 19.495 ms | 11.4× |

Lower execution time is better.

These figures are derived from separately collected benchmark runs and are preliminary rather than controlled, statistically rigorous comparisons.

## Measurement Methodology

### Eloque

Go's testing benchmark framework measures operations per second, execution time, throughput where configured, and memory allocations.

The earlier Eloque benchmarks use three runs with a target duration
of one second per benchmark. M1-P5-C uses five runs for its
cross-language comparisons, with median results reported.

### CPython

The original CPython frontend benchmark uses
`timeit.Timer.autorange()` to select an iteration count and
reports the median of five timed batches.

The M1-P5-C controlled arithmetic benchmark uses five repeated
measurements, with the median reported for comparison against
Eloque and Go.

Two operations are measured independently:

- **Tokenization:** `tokenize.generate_tokens()` with `io.StringIO`, including materialization of the resulting tokens into a list.
- **Frontend parsing:** `ast.parse()`, including CPython's internal tokenization and AST construction.

The tokenization and parsing timings must not be added together.

## Interpretation and Limitations

1. **Different language grammars:** Eloque and Python have different syntax, grammar rules, and AST representations. The workloads are comparable in intent, not identical in implementation complexity.

2. **Different tokenizer implementations:** Python's public `tokenize` module is not the same implementation as CPython's internal tokenizer. Its results should not be presented as native CPython lexer performance.

3. **Different benchmark harnesses:** Go's testing framework and Python's `timeit` use different measurement and iteration-selection mechanisms.

4. **Different frontend responsibilities:** The benchmarks do not establish that Eloque and CPython perform identical validation or construct equally complex ASTs.

5. **Synthetic workloads:** These measurements do not necessarily represent performance on real-world applications or large, heterogeneous codebases.

6. **No runtime comparison:** These benchmarks do not measure program execution, bytecode compilation, virtual-machine performance, or Eloque's future static type checker.

7. **Prerelease CPython:** The comparison uses CPython 3.15.0rc2, not a final stable Python 3.15 release.

## Historical Results

The `results/historical/` directory preserves measurements from earlier lexer/parser implementations, including the previous Full versus Compact frontend comparison.

These results are useful for investigating the effects of optimization and architectural consolidation.

Historical and current measurements should only be compared directly when the benchmark workload and measured operations remain equivalent.

## Conclusions from Preliminary Benchmarks

On the tested synthetic workloads, Eloque's optimized frontend consistently completed processing faster than CPython 3.15.0rc2's `ast.parse()`.

The measured advantage ranged from approximately 5× to 11× for the 1,000-unit workloads.

These results demonstrate promising frontend performance, but do not establish that Eloque is generally faster than Python or that its future compiler and runtime will outperform CPython.

Further benchmarking should include representative real-world source files, more controlled repeated measurements, and eventually bytecode compilation and execution benchmarks.

### M1-P5-C Validation Outcome

The subsequent M1-P5-C benchmarks extended frontend comparisons
to Go and introduced controlled arithmetic workloads.

Eloque recorded the lowest median frontend latency across all
20 language-native workload-size comparisons.

In the controlled arithmetic benchmark at N=10,000, Eloque
completed frontend processing approximately 1.82× faster than
Go and 14.87× faster than CPython 3.15.0rc2.

These findings support the performance of Eloque's current
streaming frontend on the tested workloads. They do not establish
general language-performance superiority, equivalent parser
capabilities, or future compiler and runtime performance.
