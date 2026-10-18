# Eloque Benchmarks

This directory contains performance benchmarks for Eloque's lexer and parser, historical optimization results, and an exploratory frontend comparison against CPython.

## Directory Structure

```text
benchmarks/
├── README.md
├── cpython/
│   └── frontend.py
└── results/
    ├── current/
    │   ├── lexer.txt
    │   ├── parser.txt
    │   └── cpython-frontend.txt
    └── historical/
        ├── environment.txt
        ├── lexer-baseline.txt
        ├── parser-expressions-baseline.txt
        ├── parser-program-baseline.txt
        └── frontend-full-vs-compact.txt
```

The Go benchmark implementations reside in `internal/lexer/` and `internal/parser/`.

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

The complete frontend benchmarks include lexing, token collection, parsing, and AST construction.

### CPython Frontend

The Python benchmarks use CPython 3.15.0rc2 through uv.

```bash
uv python install 3.15
uv run --no-project python benchmarks/cpython/frontend.py
```

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

Each workload is generated at sizes of 10, 100, and 1,000 units.

The programs express comparable operations using the syntax of their respective languages. They are not byte-for-byte identical and do not necessarily produce equivalent AST structures.

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

The Go measurements use three benchmark runs with a target duration of one second per benchmark.

### CPython

The Python benchmark script uses `timeit.Timer.autorange()` to select an iteration count and reports the median of five timed batches.

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

## Conclusions

On the tested synthetic workloads, Eloque's optimized frontend consistently completed processing faster than CPython 3.15.0rc2's `ast.parse()`.

The measured advantage ranged from approximately 5× to 11× for the 1,000-unit workloads.

These results demonstrate promising frontend performance, but do not establish that Eloque is generally faster than Python or that its future compiler and runtime will outperform CPython.

Further benchmarking should include representative real-world source files, more controlled repeated measurements, and eventually bytecode compilation and execution benchmarks.