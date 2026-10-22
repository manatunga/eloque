# Eloque Benchmarks

This directory contains reproducible frontend performance benchmarks, historical optimization results, token-delivery architecture experiments, and cross-language comparisons against Go and CPython.

**Current production baseline:** H2 direct-streaming parser with an arena-backed flat AST.

**Historical baseline:** M1 direct-streaming parser with an interface-based AST.

The benchmarks measure lexical analysis, parsing, and AST construction. They do not measure Eloque program execution, static type checking, bytecode generation, or VM performance.

## Directory structure

```text
benchmarks/
├── README.md
├── cpython/
├── go/
└── results/
    ├── current/
    ├── experimental/
    ├── historical/
    └── official/
```

- `results/official/`: Finalized benchmark evidence for the H2 checkpoint.
- `results/current/`: Previously maintained production baseline measurements.
- `results/experimental/`: Exploratory optimization, allocation, and architecture measurements.
- `results/historical/`: Earlier implementation measurements retained for comparison.

The current Eloque lexer, parser, and AST benchmarks reside under:

```text
internal/frontend/lexer/
internal/frontend/parser/
internal/frontend/ast/
```

Cross-language Go benchmarks reside under `benchmarks/go/`, and CPython benchmarks reside under `benchmarks/cpython/`.

Historical result files and their original contents should be preserved.

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

## Historical M1 Benchmark Environment

The following environment describes the earlier M1 measurements. For the official H2 environment, refer to results/official/h2_controlled_20261020.meta.txt.

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

### Eloque lexer

Run from the repository root:

```bash
go test ./internal/frontend/lexer \
  -run '^$' -bench . -benchmem \
  -benchtime=1s -count=3
```

### Eloque parser and complete frontend

```bash
go test ./internal/frontend/parser \
  -run '^$' -bench . -benchmem \
  -benchtime=1s -count=3
```

The canonical `parser.Parse(sourceText)` entry point performs direct pull-based lexical streaming and constructs an arena-backed `ast.FlatProgram`.

The retained M1 parser and buffered frontend are reference implementations for regression and comparative benchmarks.

These commands execute the benchmark functions registered in the respective Go packages. Reproducing an individual official H2 result requires the exact benchmark selector, iteration settings, and environment recorded for that result.

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

## H2 — Arena-backed flat AST and streaming parser

**Checkpoint:** H2, October 20, 2026.

H2 replaced the production interface-based AST with an arena-backed flat representation while retaining direct pull-based token streaming.

The previous M1 implementation remains available as a comparison baseline.

The optimization was intended to reduce AST allocation overhead without changing the implemented grammar or diagnostic contract.

### Official benchmark evidence

The following files contain the finalized H2 measurements:

| File | Purpose |
|---|---|
| `h2_eloque_controlled_20261020.txt` | Eloque H2 controlled arithmetic measurements |
| `h2_go_compiler_controlled_20261020.txt` | Go compiler syntax parser measurements |
| `h2_go_parser_controlled_20261020.txt` | Go standard-library parser measurements |
| `h2_cpython_controlled_20261020.txt` | CPython controlled arithmetic measurements |
| `h2_controlled_20261020.meta.txt` | Recorded benchmark environment and methodology |
| `h2_m1_workloads_20261020.txt` | M1 versus H2 representative workload measurements |

All six files are stored under `benchmarks/results/official/`.

### M1 versus H2 performance

The workload comparison covers declarations, expressions, strings, comments, and mixed programs.

The official suite used ten repeated measurements.

At 1,000 workload units:

| Workload | M1 median | H2 median | H2 speedup |
|---|---:|---:|---:|
| Declarations | 238.1 µs | 209.0 µs | 1.14× |
| Expressions | 1,181.5 µs | 667.4 µs | 1.77× |
| Strings | 400.1 µs | 394.2 µs | 1.01× |
| Comments | 407.8 µs | 346.4 µs | 1.18× |
| Mixed | 1,683.8 µs | 1,191.1 µs | 1.41× |

H2 reduced allocation counts across the tested workloads.

The largest measured latency improvement was in expression-heavy parsing. String-heavy parsing improved only marginally.

These figures describe specific synthetic inputs and should not be generalized to every Eloque source program.

### Controlled cross-language frontend comparison

The official H2 comparison measured frontend processing for arithmetic workloads across Eloque, two Go parsers, and CPython 3.14.4.

At 10,000 workload units:

| Implementation | Median latency |
|---|---:|
| Eloque H2 | 7.16 ms |
| Go compiler syntax parser | 14.12 ms |
| Go `go/parser` | 23.92 ms |
| CPython 3.14.4 `ast.parse` | 147.86 ms |

All official H2 suites used ten repeated measurements.

These implementations do not recognize identical language grammars or produce identical AST structures. The source wrappers and benchmark harnesses also differ.

The comparison demonstrates measured frontend latency for the selected workloads, not equivalent compiler capabilities or general language-performance superiority.

### H2 environment

The official H2 environment and methodology are recorded in:

`results/official/h2_controlled_20261020.meta.txt`

Use this metadata file as the authoritative record when reproducing the official comparison.

**CPython version:** 3.14.4.

This is distinct from the historical M1-P5-C comparison, which used CPython 3.15.0rc2.

Do not combine results from these different interpreter versions into a single supposedly uniform benchmark run.

### H2 correctness and optimization experiments

H2 validation included differential parser testing, arena-invariant testing, and regression checks.

Experimental work investigated arena capacity, allocation behavior, and growth strategies.

Experimental outputs are retained under `results/experimental/`, separately from the official results.

Additional arena-capacity tuning is deferred pending representative profiling evidence. The decision is tracked in `docs/TECHNICAL_BACKLOG.md` under TB-025.

### H2 outcome

The H2 parser and arena-backed flat AST are the canonical production frontend.

The retained M1 parser remains a historical reference and a regression-testing implementation.

Future frontend optimization should preserve the official H2 baseline and report correctness, latency, allocation counts, allocated bytes, and memory trade-offs.

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

## Historical M1 Benchmark Conclusions

On the tested synthetic workloads, Eloque's optimized frontend consistently completed processing faster than CPython 3.15.0rc2's `ast.parse()`.

The measured advantage ranged from approximately 5× to 11× for the 1,000-unit workloads.

These results demonstrate promising frontend performance, but do not establish that Eloque is generally faster than Python or that its future compiler and runtime will outperform CPython.

Further benchmarking should include representative real-world source files, more controlled repeated measurements, and eventually bytecode compilation and execution benchmarks.

## Current Benchmark Conclusions — H2

H2 is the current production frontend performance baseline.

The official M1/H2 comparison demonstrates reduced allocation counts and improved median parsing latency across the five tested workload categories.

The controlled arithmetic comparison also records lower median frontend latency for Eloque H2 than the two measured Go parser implementations and CPython 3.14.4 on the selected inputs.

These findings do not establish general compiler superiority, language execution performance, or CPython runtime parity.

The next major performance evaluation should occur after M2 provides a working statically checked bytecode VM. At that stage, runtime comparisons must use correctness-equivalent programs, clearly defined measurement boundaries, and reproducible environments.

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

## H2 post-recovery frontend comparison — Experimental (October 22–23, 2026)

This experiment evaluates Eloque's frontend after the introduction of structured diagnostic recovery, statement-level AST preservation, and expression-depth protection.

The experiment is separate from the official October 20 H2 baseline, which remains unchanged.

### Results

Ten measurements were collected for each Go benchmark configuration. Reported latencies are medians.

| Frontend | N=10 | N=100 | N=1,000 | N=10,000 |
|---|---:|---:|---:|---:|
| Eloque fail-fast | 6.74 µs | 59.97 µs | 0.716 ms | 7.866 ms |
| Eloque recovering | 7.51 µs | 71.31 µs | 0.819 ms | 8.122 ms |
| Go compiler-native | 12.18 µs | 113.28 µs | 1.208 ms | 11.218 ms |
| Go `go/parser` | 18.49 µs | 170.93 µs | 1.865 ms | 18.208 ms |
| CPython 3.15.0rc2 | 87.46 µs | 908.59 µs | 11.106 ms | 141.661 ms |

At 10,000 arithmetic declarations, recovery-enabled Eloque parsing incurred **3.25% median latency overhead** compared with fail-fast parsing, with effectively unchanged allocation counts and allocated bytes.

Eloque fail-fast parsing recorded approximately 1.43× lower latency than the Go compiler-native parser, 2.31× lower latency than `go/parser`, and 18.01× lower latency than CPython `ast.parse()` on the selected workload.

These are frontend-only measurements and are not general compiler or execution-performance rankings.

### Environment

- CPU: Intel Core Ultra 7 155H
- Operating environment: Linux amd64 under WSL2
- Go: 1.27.2
- CPython: 3.15.0rc2, built with Clang 22.1.3
- Workload sizes: 10, 100, 1,000, and 10,000 arithmetic declarations
- Go measurements: 10 repetitions per configuration
- CPython: 10 timing repetitions per size, with only the resulting medians retained in the available summary

### Reproduction

Run the commands from the repository root.

**Eloque:**

```bash
go test ./internal/frontend/parser \
  -run '^$' \
  -bench '^BenchmarkControlledEloque(Frontend|RecoveringFrontend)$' \
  -benchmem -benchtime=1s -count=10
```

**Go standard library:**

```bash
go test ./benchmarks/go \
  -run '^$' \
  -bench '^BenchmarkControlledGoFrontend$' \
  -benchmem -benchtime=1s -count=10
```

**CPython:**

```bash
uv run --no-project --python 3.15 \
  python benchmarks/cpython/frontend_controlled.py
```

**Go compiler-native:**

The compiler-native benchmark uses `cmd/compile/internal/syntax` with a Go test overlay. Its overlay JSON must map the compiler package's `frontend_controlled_test.go` path to the local benchmark source file. The overlay contains absolute paths and must be generated or adjusted for the local checkout.

Once the overlay is configured:

```bash
go test \
  -overlay="$(pwd)/benchmarks/tmp/go-compiler/overlay.json" \
  cmd/compile/internal/syntax \
  -run '^$' \
  -bench '^BenchmarkControlledGoCompilerFrontend$' \
  -benchmem -benchtime=1s -count=10
```

### Saved evidence

The experimental results are stored under `results/experimental/`.

The generated report directory, `h2-post-recovery-report/`, contains:

- `REPORT.md` — methodology, results, and limitations
- `summary.csv` — per-frontend median measurements
- `raw_go_samples.csv` — 160 individual Go benchmark measurements
- `manifest.json` — source filenames and recorded language versions

The report generator is `benchmarks/scripts/generate_h2_frontend_report.py`.

### Interpretation and limitations

The implementations use different language grammars, source wrappers, AST representations, and validation responsibilities. Consequently, their measured latencies and memory usage do not represent identical parsing work.

Allocated bytes per operation are not equivalent to peak memory consumption. The reported throughput uses each implementation's language-specific source length.

The CPython summary does not retain individual timing samples, so its sample-level variability cannot be calculated from the saved evidence.

All benchmark runs were performed separately and may be affected by system load, scheduling, and thermal variation. No statistical-significance claim is made.

The recovery-enabled benchmark uses valid source programs and measures the overhead of enabling recovery, not the cost of processing malformed source. Separate recovery stress and robustness tests cover malformed-input behavior.

Further representative mixed-program benchmarks and runtime comparisons are deferred to appropriate future milestones.