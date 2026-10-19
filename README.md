<p align="center">
  <img src=".github/branding/readme-banner.png" alt="Eloque Programming Language Banner" width="100%">
</p>

# The Eloque Programming Language

**Eloque** is a proposed general-purpose programming language with structured, conversational-English-inspired syntax and precise, statically checked semantics. It is intended to be approachable for beginners while providing predictable diagnostics and tooling for AI-assisted development.

> **Status:** Early development — Milestone 1 (Frontend foundations completed. Eloque has a working lexer and parser with a canonical `Parse(sourceText)` entry point, direct pull-based token streaming complete-program AST construction, source spans, and structured fail-fast diagnostics. The implemented syntax includes arithmetic expressions, identifiers, literals, `let` declarations, `change` assignments, and `say` statements. Frontend integration, syntax conformance testing, and performance validation are complete. Eloque cannot yet execute source programs. The examples below are **illustrative, not executable**.

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

## Planned implementation

Eloque is being implemented in **Go**. Its lexer, parser, and AST construction stages are operational; the remaining stages are planned:

```text
.elq source
    ↓
Lexer → Parser → AST
    ↓
Name resolution → Static type checker
    ↓
Bytecode compiler → Stack-based VM
    ↓
Program execution
```

The initial milestones established the Go project, lexer, parser, AST construction, and source diagnostics.

The next development phase will introduce the components needed for a **small statically checked executable subset**, including name resolution, type checking, bytecode compilation, and virtual-machine execution.

More advanced features such as collections, custom types, modules, exceptions, and concurrency will be introduced incrementally.

No installation instructions are provided yet because there is no working Eloque interpreter or compiler to install.

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

The production frontend processes source text directly through `Parse(sourceText)`, incrementally requesting lexer tokens while constructing a complete-program AST.

Buffered token delivery remains available as a reference implementation for regression testing and performance comparisons.

**Current status:** Milestone 1 is complete. The production frontend supports lexing, parsing, and AST construction for the implemented syntax subset, with validated diagnostics and performance benchmarks.

**Next milestone:** M2, as defined in the project roadmap.

### Performance benchmarks

The production streaming frontend has reproducible lexer, parser, and complete-frontend benchmarks.

M1-P5 validated the token-delivery architecture and compared Eloque's frontend against Go and CPython 3.15.0rc2 across synthetic language-native and controlled arithmetic workloads.

Benchmark methodology, results, reproduction instructions, and limitations are documented in [benchmarks/README.md](benchmarks/README.md).

These measurements concern frontend processing only. They do not demonstrate runtime execution performance or general superiority over Python.

### Implementation limitations

Eloque's Milestone 1 frontend is complete, but the language cannot yet execute source programs.

Name resolution, static type checking, bytecode compilation, and virtual-machine execution have not yet been implemented. The current frontend supports only the documented M1 syntax subset, not the full proposed language.

The [language specification](docs/LANGUAGE_SPEC.md) describes the intended language, not the currently executable feature set. Syntax, release boundaries, and technical details may evolve before version 1.0.
