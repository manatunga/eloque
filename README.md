<p align="center">
  <img src=".github/branding/readme-banner.png" alt="Eloque Programming Language Banner" width="100%">
</p>

# The Eloque Programming Language

**Eloque** is a proposed general-purpose programming language with structured, conversational-English-inspired syntax and precise, statically checked semantics. It is intended to be approachable for beginners while providing predictable diagnostics and tooling for AI-assisted development.

> **Status:** Early development — M1, the post-M1 H2 frontend optimization, and pre-M2 structured diagnostic recovery are implemented. Eloque has a working Go lexer and parser using direct pull-based token streaming and an arena-backed flat AST. Both fail-fast and recovery-capable parsing are supported, with structured, source-aware diagnostics.
>
> **Next milestone:** M2 — First Execution. Eloque will use **Monowire**, its selected register-based bytecode VM, with Zig provisionally preferred as the runtime implementation language. Name resolution, mandatory static type checking, production bytecode compilation, and VM execution are not yet implemented.
>
> The examples below are **illustrative, not executable**.

## A glimpse of Eloque

```text
let greeting be "Hello"
let name be "Alex"

to greet with (person as String) giving String:
    give "{greeting}, {person}!"

say greet with (name)

let numbers be [1, 2, 3]
add 4 to numbers

for each number in numbers:
    say number

define Person:
    name as String
    age as Int

    to introduce with ():
        say "I'm {self.name} and I'm {self.age}."

let person be Person with (name be "Alex", age be 20)
person.introduce with ()
```

This example illustrates the intended language, including features that will arrive at different implementation milestones. It is **not** a statement that these features currently run.

## Design principles

- **Readable, not ambiguous:** conversational keywords follow a deterministic grammar; Eloque does not attempt to interpret unrestricted English.
- **Static safety:** types are checked before execution, with inference for eligible local bindings and explicit function signatures.
- **Predictable behavior:** strict booleans, checked integer arithmetic, explicit mutation and defined value semantics.
- **Useful diagnostics:** errors should be clear to beginners and available in structured form for editors and AI agents.
- **Concurrency by design:** safe native concurrency is a longer-term goal, not a feature of the initial implementation.
- **Evidence-driven execution:** Monowire is the selected register-based bytecode VM, using overlapping register windows for function calls. Runtime implementation and optimization decisions are guided by reproducible benchmarks, correctness, and maintainability.

## Documentation

- [Language specification](docs/LANGUAGE_SPEC.md) — intended syntax and semantics
- [Design decisions](docs/DESIGN_DECISIONS.md) — current agreements and superseded syntax
- [Architecture](docs/ARCHITECTURE.md) — proposed compiler, checker, runtime and VM
- [Roadmap](docs/ROADMAP.md) — incremental milestones and provisional releases
- [Technical backlog](docs/TECHNICAL_BACKLOG.md) — deferred engineering decisions and review triggers

## Compiler & Runtime Architecture

Eloque combines a **Go-based compiler frontend** with **Monowire**, its selected register-based bytecode execution architecture.

The implemented frontend uses pull-based token streaming, Pratt expression parsing, and an arena-backed flat AST. Monowire is designed around explicit register operands and overlapping call-frame register windows.

The frontend is operational, including structured diagnostic recovery. Name resolution, mandatory static checking, bytecode compilation, and runtime execution are planned for **M2 — First Execution**.

```text
        Eloque source (.elq)
                │
                ▼
┌──────────────────────────────┐
│ Pull-based streaming lexer   │  ✓ Implemented
└──────────────┬───────────────┘
               ▼
┌──────────────────────────────┐
│ Pratt expression parser      │  ✓ Implemented
│ + statement parsing          │
└──────────────┬───────────────┘
               ▼
┌──────────────────────────────┐
│ Arena-backed flat AST        │  ✓ Implemented
└──────────────┬───────────────┘
               ▼
┌──────────────────────────────┐
│ Name resolution              │  Planned — M2
└──────────────┬───────────────┘
               ▼
┌──────────────────────────────┐
│ Mandatory static checker     │  Planned — M2
└──────────────┬───────────────┘
               ▼
┌──────────────────────────────┐
│ Monowire register            │  Planned — M2
│ bytecode compiler            │
└──────────────┬───────────────┘
               ▼
┌──────────────────────────────┐
│ Monowire execution engine    │  Planned — M2
│ Overlapping register windows │
└──────────────┬───────────────┘
               ▼
        Program execution
```

### Compiler frontend

The frontend packages are organized under `internal/frontend/`, with shared source-location utilities in `internal/source/`.

Three architectural choices define the implemented frontend:

- **Pull-based streaming lexer:** The parser requests tokens on demand, avoiding the need to materialize the entire token stream before parsing.
- **Pratt expression parser:** Expression parsing is precedence-aware and integrated with Eloque's statement grammar.
- **Arena-backed flat AST:** Syntax nodes use a compact, index-addressed representation instead of individually allocated interface-backed AST objects.

The production frontend exposes two parsing entry points:

- `parser.Parse(sourceText)` returns `(*ast.FlatProgram, error)` and provides fail-fast parsing.
- `parser.ParseRecovering(sourceText, options)` returns a `ParseResult` containing structured diagnostics, a termination status, and a potentially partial AST.

Recovery-capable parsing supports configurable error budgets and preserves successfully completed statements where possible. Its termination statuses are `ParseComplete`, `ParseRecovered`, and `ParseAborted`.

A recovered or partial AST is not semantically validated or executable.

The original M1 parser and interface-based AST remain available internally for regression testing and performance comparisons.

### Monowire execution engine

**Monowire** is Eloque's selected register-based bytecode virtual machine, designed around explicit register operands and overlapping caller-outgoing/callee-incoming register windows.

Rather than using a conventional operand stack, Monowire instructions operate on registers. Function-call transitions are designed around overlapping register windows, allowing the caller's outgoing argument registers to correspond to the callee's incoming parameter registers.

This architecture aims to reduce unnecessary operand movement and support efficient function execution. Its actual performance characteristics must be validated against the production implementation.

The previous Go-based stack-VM plan has been superseded.

The runtime implementation language is not yet finalized:

- **Zig:** Provisionally preferred following the pre-M2 implementation-language experiment.
- **Rust:** Retained as a fallback if integration, portability, ownership, or maintenance constraints make Zig unsuitable.

The pre-M2 experiment evaluated Go, Rust, Zig conventional-switch dispatch, and Zig labeled-switch dispatch across arithmetic, branching, call-loop, and recursion workloads.

Zig recorded the lowest median wall-clock latency across all four synthetic workloads. These results inform implementation-language selection but do not establish production VM performance.

Conventional switch-based dispatch is the initial implementation baseline. Specialized dispatch strategies remain subject to production profiling.

**Implementation status:** Monowire's architecture has been selected, but its production bytecode compiler, VM, and Go–runtime integration have not yet been implemented.

See [Monowire Benchmarks](benchmarks/monowire/README.md) for methodology and experimental evidence.

### Next implementation milestone

**M2 — First Execution** will introduce:

- Lexical name resolution and symbol binding.
- Mandatory basic static type checking and definite-assignment analysis.
- Monowire bytecode format and register-window execution contracts.
- Monowire register-based bytecode compilation.
- Sequential Monowire VM execution.
- End-to-end correctness tests and initial runtime performance baselines.

The initial executable subset will cover supported literals, arithmetic, `let`, `change`, and `say`.

Control flow, language-level functions, collections, custom types, modules, and native concurrency are planned for later milestones. Internal VM call machinery may be implemented before language-level function support.

There are no installation instructions for executing Eloque programs yet because the runtime has not been implemented.

See [Architecture](docs/ARCHITECTURE.md) for implementation details, [Design Decisions](docs/DESIGN_DECISIONS.md) for accepted architectural choices, and [Technical Backlog](docs/TECHNICAL_BACKLOG.md) for unresolved engineering decisions.

## Project status

Eloque has completed **Milestone 1 — Frontend foundations**
and is preparing for the next implementation milestone.

The following checkpoints are complete:

- **M0:** Repository foundations, Go module, minimal CLI skeleton, and initial testing infrastructure.
- **M1-P1:** AST foundations, shared source spans, parser infrastructure, and structured syntax diagnostics.
- **M1-P2:** Pratt arithmetic expression parsing, including precedence, associativity, grouping, and source-location tracking.
- **M1-P3:** `let`, `change`, and `say` statement parsing, complete-program AST construction, and fail-fast syntax diagnostics.
- **Post-M1-P3 optimization:** Canonical lexer/parser consolidation, performance improvements, and updated frontend benchmarks.
- **M1-P4:** Source-based syntax conformance testing, expanded diagnostic coverage, and validation of the implemented grammar.
- **M1-P5-A:** Token-delivery architecture experiments comparing eager buffering, direct streaming, and bounded lookahead. Direct streaming was selected.
- **M1-P5-B:** Production lexer–parser integration through `Parse(sourceText)`, with end-to-end regression validation.
- **M1-P5-C:** Cross-language frontend benchmarking against Go and CPython 3.15.0rc2, including controlled arithmetic workloads.
- **Post-M1 H2 optimization:** Replaced the production interface-based AST with an arena-backed flat AST, retained direct pull-based token streaming, preserved the legacy M1 parser for regression comparisons, and validated the new frontend using differential tests, arena-invariant tests, and controlled benchmarks.
- **Frontend package migration:** Reorganized the lexer, parser, and AST packages under `internal/frontend/`, retaining shared source-location utilities in `internal/source/`.
- **Pre-M2 structured diagnostics and recovery:** Introduced shared diagnostic infrastructure, lexer and parser recovery, configurable error budgets, statement-level AST preservation, expression-depth safeguards, and dedicated regression and stress tests. Retained the existing fail-fast parsing API.
- **Pre-M2 Monowire experiment:** Implemented equivalent synthetic register-VM workloads in Go, Rust, Zig with conventional switch dispatch, and Zig with labeled-switch dispatch. Benchmarked arithmetic, branching, repeated calls, and recursion. Selected Monowire as the production register-based VM architecture and provisionally preferred Zig as its implementation language, subject to integration validation.

The production frontend processes source text through `Parse(sourceText)`, incrementally requesting lexer tokens and constructing an arena-backed `FlatProgram`. The original M1 parser and buffered token-delivery implementation remain available as reference implementations.

**Current status:** M1, H2, structured diagnostic recovery, and the pre-M2 Monowire implementation-language experiment are complete. Monowire's register-based architecture is selected, but its production implementation has not begun.

**Next milestone:** M2 — First Execution, beginning with lexical name resolution and mandatory static type checking, followed by Monowire bytecode compilation and VM execution.

### Performance benchmarks

H2 is the current production frontend performance baseline.

The official M1/H2 comparison measured five workload categories over ten benchmark runs. At 1,000 workload units, H2 recorded median speedups ranging from **1.01× to 1.77×**, with reduced allocation counts across the tested workloads.

A separate controlled arithmetic benchmark at 10,000 units recorded:

| Implementation | Median latency |
|---|---:|
| Eloque H2 | 7.16 ms |
| Go compiler syntax parser | 14.12 ms |
| Go `go/parser` | 23.92 ms |
| CPython 3.14.4 | 147.86 ms |

These results describe synthetic frontend workloads. Different language grammars, AST representations, and benchmark harnesses limit direct comparability.

They do not establish general compiler superiority or Eloque runtime performance.

An additional **experimental post-recovery comparison (October 22–23, 2026)** measured fail-fast and recovery-enabled Eloque parsing alongside Go compiler-native parsing, Go `go/parser`, and CPython 3.15.0rc2.

At 10,000 arithmetic declarations, Eloque recorded median frontend latencies of **7.87 ms (fail-fast)** and **8.12 ms (recovery-enabled)**. The recovery-enabled path incurred approximately **3.25% median overhead** on this valid-source workload.

These experimental measurements do not replace the October 20 official H2 baseline. The complete report, methodology, raw Go measurements, and limitations are available in [the post-recovery benchmark report](benchmarks/results/experimental/h2-post-recovery-report/REPORT.md).

### Monowire implementation-language experiment

On October 23, 2026, Eloque completed a pre-M2 experiment comparing four implementations of a synthetic register-based VM:

- Go.
- Rust.
- Zig with conventional switch dispatch.
- Zig with labeled-switch dispatch.

The experiment covered arithmetic, conditional branching, repeated function calls, and recursion.

**Median subprocess wall-clock times (milliseconds):**

| Implementation | Call loop | Arithmetic | Branch | Recursion |
|---|---:|---:|---:|---:|
| Zig (labeled switch) | 18.802 | 9.268 | 15.919 | 14.033 |
| Zig (switch) | 19.794 | 9.280 | 15.985 | 14.413 |
| Rust | 33.049 | 11.892 | 21.395 | 23.859 |
| Go | 53.362 | 18.262 | 29.022 | 39.246 |

Zig recorded the lowest median time across all four workloads. The labeled-switch variant showed a smaller additional advantage over conventional switch dispatch.

**Limitations:** These are synthetic prototype measurements, not production Eloque execution benchmarks. They include subprocess startup overhead, and the implementations differ in memory-allocation strategies. The experiment does not establish complete runtime performance or prove that Zig will remain the fastest choice in production.

**Outcome:** Zig is provisionally preferred for Monowire, with Rust retained as a fallback. The production implementation choice remains subject to integration, correctness, portability, and maintainability validation.

The experimental specification, build instructions, benchmark harness, and results are documented in [`benchmarks/monowire/`](benchmarks/monowire/).

The official results, historical measurements, methodology, and limitations are documented in [benchmarks/README.md](benchmarks/README.md).

### Implementation limitations

Eloque currently supports lexical analysis, syntax parsing, structured diagnostics, and recovery-capable AST construction for its initial grammar.

It does **not yet** support:

- Lexical name resolution.
- Mandatory static type checking.
- Production Monowire bytecode compilation.
- Monowire VM execution.
- Running Eloque programs.

The Monowire architecture is selected, but the implementation-language choice remains provisional and the production VM is not yet operational.

The [language specification](docs/LANGUAGE_SPEC.md) describes the intended language rather than a fully implemented feature set. Syntax, release boundaries, and technical details may evolve before version 1.0.
