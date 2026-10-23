# Monowire VM — Implementation-Language Selection Experiment

**Status:** Initial experimental benchmark phase complete (23 October 2026).  
**Architecture:** Monowire, Eloque's selected register-based bytecode virtual machine.  
**Implementation language:** Zig provisionally preferred; Rust retained as a fallback.  
**Scope:** Independent experimental harness, not the production Eloque compiler or runtime.

## 1. Purpose

This experiment evaluates candidate implementation languages and instruction-dispatch strategies for Monowire, Eloque's register-based bytecode virtual machine.

Monowire replaces the previously planned stack-based VM architecture. Its initial design uses register operands, function call frames, and overlapping caller-outgoing/callee-incoming register windows.

The experiment compares:

- Go
- Rust
- Zig with conventional switch dispatch
- Zig with labeled-switch dispatch

The objective is to collect reproducible preliminary performance evidence and assess implementation feasibility before developing the production VM during M2 — First Execution.

These prototypes are not complete language runtimes and do not execute Eloque source files.

## 2. Experiment structure

The benchmark suite contains four workloads:

| Workload | Purpose |
|---|---|
| `call_loop` | Repeated VM function calls and register-window operations |
| `arithmetic` | Integer arithmetic and register operations without function calls |
| `branch` | Conditional branching and arithmetic |
| `recursion` | Nested function calls, frame management, register windows, and return propagation |

Each workload contains implementations in Go, Rust, and both Zig dispatch variants.

The suite also includes:

- `SPEC.md` — Experimental bytecode semantics and comparability requirements.
- `build.sh` — Builds the benchmark implementations.
- `compare.py` — Executes correctness checks, collects timing samples, and exports results.

The original `call_loop/zig/main.zig` correctness prototype is retained separately from the measured Zig dispatch variants.

## 3. Build and execution

From `benchmarks/monowire/`:

```sh
bash build.sh
uv run --no-project compare.py
```

To benchmark recursion independently:

```sh
uv run --no-project compare.py --workload recursion
```

The recorded comparison used:

| Parameter | Value |
|---|---|
| Workload size (`n`) | 10,000 |
| Repetitions (`reps`) | 100 |
| Measurement runs | 15 per implementation |
| Timing method | External subprocess wall-clock time |
| Build configuration | Optimized compiler builds |
| Measurement date | 23 October 2026 |

The comparison runner performs an initial warmup, varies execution order between measurement rounds, and records median, mean, standard deviation, minimum, maximum, and individual timing samples.

Measurements include process startup and shutdown overhead. They must not be interpreted as isolated bytecode dispatch or VM execution times.

Compiler versions, machine details, source revision, and execution metadata should be taken from the corresponding saved result files rather than inferred from this summary.

## 4. Correctness verification

Each workload verifies the expected numerical result before performance measurements are accepted.

For `call_loop`, `arithmetic`, and `recursion`:

```text
expected = n * (n + 1) / 2
```

For `branch`:

```text
expected = (n // 2) * ((n + 1) // 2)
```

The repeated-result checksum uses XOR. With 100 repetitions of an unchanged result, the expected checksum is zero.

All 16 benchmark configurations completed the recorded comparison with their expected output checks passing.

These checks establish limited correctness for the selected inputs. They do not establish complete bytecode semantics, general memory safety, stack-overflow handling, or correctness for arbitrary programs.

## 5. Recorded results

**Measurement:** Median subprocess wall-clock time in milliseconds. Lower is better.

| Implementation | Call loop | Arithmetic | Branch | Recursion |
|---|---:|---:|---:|---:|
| Zig labeled | **18.802** | **9.268** | **15.919** | **14.033** |
| Zig switch | 19.794 | 9.280 | 15.985 | 14.413 |
| Rust | 33.049 | 11.892 | 21.395 | 23.859 |
| Go | 53.362 | 18.262 | 29.022 | 39.246 |

All values are milliseconds.

Zig ranked first in every measured workload, followed by Rust and Go.

Relative to Rust, Zig labeled dispatch recorded approximately:

- 1.76× lower elapsed time ratio for `call_loop`.
- 1.28× for `arithmetic`.
- 1.34× for `branch`.
- 1.70× for `recursion`.

The geometric mean of these four speedup ratios is approximately 1.50×.

Relative to Go, the corresponding geometric-mean speedup is approximately 2.38×.

These ratios characterize the measured prototypes and workloads. They do not establish general performance superiority between the implementation languages.

### Zig dispatch comparison

Zig labeled-switch dispatch produced slightly lower median times than conventional switch dispatch in all four workloads.

However, the differences were small, especially for arithmetic and branching, relative to run-to-run variability.

The current evidence does not establish a statistically reliable or generally applicable advantage for labeled-switch dispatch.

Conventional switch dispatch remains an acceptable initial implementation strategy. Labeled-switch dispatch should be retained as an optimization candidate for later controlled evaluation.

## 6. Methodological limitations

The results are exploratory rather than a fully controlled comparison of equivalent production VM implementations.

### 6.1 Process-level timing

The comparison measures subprocess wall-clock time, including process startup, runtime initialization, allocation, execution, and teardown.

The measured differences cannot be attributed exclusively to instruction dispatch or interpreter execution.

### 6.2 Storage and allocation differences

The prototype implementations do not use fully equivalent storage-management strategies.

Go and Rust use dynamically sized storage, while Zig uses fixed-capacity register and frame arrays in the relevant prototypes.

In particular, the recursion workload initializes fixed-capacity Zig storage independently of the input-dependent allocation strategies used by Go and Rust.

These differences affect comparability and must be considered before attributing the observed timing gaps to implementation-language efficiency.

### 6.3 Compiler and optimization differences

The implementations use their respective language toolchains and optimized build configurations.

Compiler optimization strategies, bounds-check behavior, runtime facilities, and generated machine code may differ.

No instruction-level profiling or equivalence audit has yet established that all implementations perform comparable machine-level work.

### 6.4 Workload representativeness

The four workloads test selected interpreter mechanisms rather than complete Eloque programs.

They do not adequately represent:

- Heap allocation and garbage collection.
- String and collection operations.
- Mixed runtime value types.
- Realistic bytecode distributions.
- Concurrency and task scheduling.
- Host-language integration and foreign-function overhead.
- Full compilation-to-execution latency.

The `branch` workload uses a predictable alternating branch pattern, not unpredictable branching.

### 6.5 Limited correctness coverage

Successful expected-result checks do not substitute for dedicated tests of instruction semantics, invalid operands, register bounds, frame exhaustion, recursion limits, and runtime error propagation.

A zero XOR checksum with an even repetition count is not independently sufficient to establish correctness.

## 7. Experiment conclusion

The first benchmark phase is complete.

**Architectural decision:** Eloque will use Monowire, a register-based bytecode VM, rather than the previously planned stack-based VM.

**Implementation-language direction:** Zig is the provisional preferred implementation language for Monowire, with Rust retained as a fallback.

This preference is supported by Zig's consistent lead across the measured prototype workloads and its low-level memory-control capabilities.

However, the benchmark evidence alone does not establish the final production implementation's performance or integration feasibility.

The remaining decision is not whether to use a stack-based or register-based VM. Monowire's register-based architecture has already been selected.

The remaining engineering questions concern how to implement Monowire reliably and efficiently.

## 8. Production follow-through

The following work remains necessary during M2 and subsequent runtime development:

1. Validate the integration boundary between Eloque's Go compiler and a Zig-based Monowire runtime.
2. Define the production bytecode format, instruction encoding, and register allocation conventions.
3. Establish runtime value representation, ownership, allocation, and error-propagation rules.
4. Implement a correctness-tested minimal Monowire VM.
5. Integrate bytecode generation with the statically checked Eloque frontend.
6. Benchmark complete Eloque programs and isolate VM execution time from process startup.
7. Reevaluate dispatch optimizations using representative workloads and controlled profiling.

The implementation-language preference should be revisited if integration, portability, safety, or maintenance evidence materially undermines Zig's practical suitability.

A future change in implementation language would not, by itself, reverse the decision to use Monowire's register-based architecture.

## 9. Result preservation

The recorded benchmark outputs are stored under:

```text
benchmarks/results/monowire/lang-comparison/
├── arithmetic/
│   └── pre-m2_monowire-lang-comparison_arithmetic_20261023.txt
├── branch/
│   └── pre-m2_monowire-lang-comparison_branch_20261023.txt
├── call_loop/
│   └── pre-m2_monowire-lang-comparison_call-loop_20261023.txt
└── recursion/
    └── pre-m2_monowire-lang-comparison_recursion_20261023.txt
```

These result files preserve the measurement records and are the primary evidence for the figures summarized above.

The experiment sources and `SPEC.md` should be retained for reproducibility and historical comparison.

## 10. Historical development

The experiment evolved through four prototype stages:

| Stage | Development |
|---|---|
| v0.1 | Initial Go semantic reference and `call_loop` workload |
| v0.2 | Rust and Zig implementations; initial dispatch comparison work |
| v0.3 | Arithmetic and branching workloads; expanded comparison harness |
| v0.4 | Recursive-call workload; completed four-workload comparison |

These labels describe the experimental harness's development history, not Eloque language releases.

The historical prototypes remain useful for understanding the experiment's evolution, but their earlier verification-status notes are superseded by the completed October 23 comparison.
