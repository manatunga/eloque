<p align="center">
  <img src=".github/branding/readme-banner.png" alt="Eloque Programming Language Banner" width="100%">
</p>

# The Eloque Programming Language

**Eloque** is a proposed general-purpose programming language with structured, conversational-English-inspired syntax and precise, statically checked semantics. It is intended to be approachable for beginners while providing predictable diagnostics and tooling for AI-assisted development.

> **Status:** Early development — Milestone 1 and the post-M1 H2 frontend optimization are complete. Eloque has a working Go lexer and parser using direct pull-based token streaming and an arena-backed flat AST. The canonical `Parse(sourceText)` entry point supports the implemented arithmetic, literal, declaration, assignment, and output syntax, with source-aware fail-fast diagnostics. Name resolution, static type checking, bytecode compilation, and VM execution are not yet implemented. The examples below are **illustrative, not executable**.

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
- **Measure before optimizing:** the first runtime is a Go-based stack bytecode VM; future optimization and backend choices depend on evidence.

## Documentation

- [Language specification](docs/LANGUAGE_SPEC.md) — intended syntax and semantics
- [Design decisions](docs/DESIGN_DECISIONS.md) — current agreements and superseded syntax
- [Architecture](docs/ARCHITECTURE.md) — proposed compiler, checker, runtime and VM
- [Roadmap](docs/ROADMAP.md) — incremental milestones and provisional releases
- [Technical backlog](docs/TECHNICAL_BACKLOG.md) — deferred engineering decisions and review triggers

## Implementation architecture

Eloque is implemented in **Go**.

Its production frontend is operational. The semantic analysis and execution stages are planned.

```text
. elq source
    |
    v
Lexer (direct token streaming)
    |
    v
Parser (Pratt expressions)
    |
    v
Arena-backed flat AST
    |
    v
Name resolution             [planned]
    |
    v
Static type checker         [planned]
    |
    v
Bytecode compiler           [planned]
    |
    v
Stack-based VM              [planned]
    |
    v
Program execution
```

The frontend packages are organized under `internal/frontend/`, with shared source-location utilities in `internal/source/`.

The production entry point is `parser.Parse(sourceText)`, returning `(*ast.FlatProgram, error)`.

The original M1 parser and interface-based AST remain available internally for regression testing and performance comparisons.

The next development phase, **M2 — First execution**, will implement the smallest statically checked executable subset, including name resolution, basic type checking, bytecode compilation, and VM execution.

More advanced language features will be introduced incrementally.

There are no installation instructions for executing Eloque programs yet because the runtime has not been implemented.

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

The production frontend processes source text through `Parse(sourceText)`, incrementally requesting lexer tokens and constructing an arena-backed `FlatProgram`. The original M1 parser and buffered token-delivery implementation remain available as reference implementations.

**Current status:** M1 and H2 are complete. M2 — First execution — is the next implementation milestone.

**Next milestone:** M2, as defined in the project roadmap.

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

The official results, historical measurements, methodology, and limitations are documented in [benchmarks/README.md](benchmarks/README.md).

### Implementation limitations

Eloque's M1 frontend and H2 optimization are complete, but the language cannot yet execute source programs. Name resolution, static type checking, bytecode compilation, and VM execution remain unimplemented. The production frontend supports only the selected initial grammar, not the full proposed language.

Name resolution, static type checking, bytecode compilation, and virtual-machine execution have not yet been implemented. The current frontend supports only the documented M1 syntax subset, not the full proposed language.

The [language specification](docs/LANGUAGE_SPEC.md) describes the intended language, not the currently executable feature set. Syntax, release boundaries, and technical details may evolve before version 1.0.
