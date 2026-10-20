# Eloque Architecture

**Status:** Milestone 1 completed. H2 frontend optimization completed and validated. Semantic analysis, bytecode compilation, and VM execution have not yet been implemented.

Eloque is a proposed general-purpose programming language with structured, conversational-English-inspired syntax, strict static typing, and an eventual native concurrency model.

The compiler is being implemented in Go. Its production frontend now uses **direct pull-based lexical streaming, Pratt expression parsing, and an arena-backed flat abstract syntax tree (AST)**.

The previous M1 parser and interface-based AST remain available internally for regression testing and performance comparisons.

Eloque can parse its implemented syntax subset but **cannot yet execute Eloque programs**.

## 1. Architectural principles

The implementation follows these principles:

1. **Correctness before optimization.** Language semantics and diagnostic behavior must remain stable through internal representation changes.
2. **Static checking before execution.** Programs containing unresolved type errors must not proceed to normal bytecode execution.
3. **Explicit implementation boundaries.** The frontend, semantic analysis, and runtime have separate responsibilities.
4. **Evidence-driven optimization.** Performance decisions require reproducible measurements and correctness validation.
5. **Minimal premature abstraction.** New intermediate representations, backends, and optimization machinery should be introduced only when justified.
6. **Backend-independent language semantics.** Observable Eloque behavior must not depend on a particular VM implementation.

The [language specification](LANGUAGE_SPEC.md) defines intended syntax and semantics. The [design decisions](DESIGN_DECISIONS.md) records accepted and superseded decisions.

## 2. Compilation and execution pipeline

The intended complete pipeline is:

```text
Eloque source
    |
    v
Lexer
    |
    v
Parser
    |
    v
Arena-backed flat AST
    |
    v
Name resolution
    |
    v
Static type checking
    |
    v
Bytecode compilation
    |
    v
Stack-based virtual machine
    |
    v
Program execution
```

**Implemented:** Lexical analysis, parsing, flat AST construction, source-aware diagnostics, and frontend integration.

**Not implemented:** Name resolution, static type checking, bytecode generation, virtual-machine execution, and runtime values.

The current implementation accepts only a subset of the proposed language. Unsupported constructs must be rejected rather than assigned speculative semantics.

## 3. Repository organization

The frontend packages have been grouped under `internal/frontend/`.

```text
internal/
├── frontend/
│   ├── ast/
│   ├── lexer/
│   └── parser/
└── source/
```

### Frontend packages

**`internal/frontend/lexer`**

Responsible for lexical scanning, token generation, indentation handling, literal recognition, and lexical errors.

**`internal/frontend/parser`**

Responsible for consuming lexer tokens, recognizing the implemented grammar, constructing the AST, and reporting syntax diagnostics.

The package contains the production H2 implementation and retained legacy parsing components.

**`internal/frontend/ast`**

Defines the production flat AST representation, arena storage, node payloads, and program representation.

It also retains legacy AST definitions for compatibility and comparative testing.

**`internal/source`**

Contains shared source-position and line-mapping functionality. It remains outside `internal/frontend` because source-location facilities may also be required by semantic analysis and later compiler stages.

### Planned package boundaries

Semantic analysis, bytecode compilation, VM execution, and runtime support will be organized when their implementations begin.

Potential package boundaries include semantic resolution and checking, bytecode generation, runtime values, and the VM.

These are architectural directions rather than declarations of existing packages. Empty package trees should not be created solely to represent future plans.

## 4. Production frontend

### 4.1 Canonical entry point

The canonical source-parsing entry point is:

```go
parser.Parse(sourceText)
```

It returns a production flat-program representation and an error:

```go
(*ast.FlatProgram, error)
```

The frontend processes source text through lexical scanning, parsing, and AST construction.

Its responsibility ends at syntactic validation and AST production. It does not resolve identifiers, check types, or execute statements.

### 4.2 Direct pull-based token streaming

The production parser requests tokens incrementally from the lexer.

This avoids constructing a complete token slice before parsing.

Direct streaming was selected during M1-P5 after evaluating eager token buffering, direct streaming, and bounded lookahead.

The integrated experiments found direct streaming faster in 19 of 20 tested workload-size combinations and lower in allocated bytes in all 20.

H2 retains direct streaming as the production token-delivery strategy.

Buffered token delivery remains available for reference and comparative testing. Additional lookahead machinery is deferred until a concrete grammar requirement or measured regression justifies it.

### 4.3 Lexer responsibilities

The lexer supports the implemented lexical subset, including:

- Reserved keywords and identifiers.
- Integer and string literals.
- Arithmetic operators and punctuation.
- Significant newlines.
- Four-space indentation, `INDENT`, and `DEDENT`.
- Comments.
- Parenthesized line continuation.
- Source spans and positioned lexical errors.

Token kinds use a compact `uint8`-backed representation.

Token lexemes reference the original source string where applicable, avoiding unnecessary text copies.

Generated structural tokens use a reusable pending-token queue.

The lexer is designed for linear scanning time relative to source length, subject to the behavior of supported lexical constructs.

### 4.4 Parser responsibilities

The production parser recognizes the currently implemented arithmetic, declaration, assignment, and output grammar.

Supported statement forms include:

```eloque
let value be 10 + 20
change value to value * 2
say value
```

These examples are syntactically supported but are not yet executable.

The expression parser uses Pratt parsing for precedence and associativity.

Its arithmetic grammar includes unary `+` and `-`, binary arithmetic operators, and grouping.

Exponentiation is right-associative and binds more tightly than unary negation on its left.

Consequently:

```text
-2 ** 2     → -(2 ** 2)
2 ** -3     → 2 ** (-3)
2 ** 3 ** 2 → 2 ** (3 ** 2)
```

The parser constructs a complete program representation for valid input.

Malformed input must not produce a successfully parsed program.

### 4.5 Diagnostics

The production frontend preserves fail-fast lexical and syntax diagnostics.

Diagnostics retain source-location information and relevant expected/actual token metadata where applicable.

The first encountered lexical or syntax error terminates the operation.

Multi-error accumulation, synchronization, and recovery remain deferred.

H2 optimization must preserve the established diagnostic contract rather than trading diagnostic correctness for performance.

## 5. H2 arena-backed AST

### 5.1 Motivation

The original M1 parser constructed an interface-based AST with individually represented expression and statement nodes.

That design was straightforward and useful for establishing correctness, but its representation introduced allocation and traversal costs.

H2 introduced an arena-backed flat AST to reduce allocation overhead while retaining the required syntactic structure and source information.

### 5.2 Representation

The production AST uses flat nodes stored in arena-managed storage.

The implementation includes:

- `FlatNode` — the production flat-node representation.
- `FlatProgram` — the production parsed-program representation.
- Arena storage and allocation functionality.
- Node payload definitions.
- Source-span preservation.
- Tests covering arena behavior, invariants, and AST construction.

Rather than treating each AST node as an independently allocated interface-backed object, the flat representation uses arena-managed nodes and associated structural information.

The AST remains a syntactic representation. It does not encode resolved names, inferred types, runtime values, or executable bytecode.

### 5.3 Arena lifetime and ownership

The arena is associated with AST construction and the resulting program representation.

Its storage must remain valid for as long as the parsed program is used.

Future semantic analysis must consume the AST without assuming that arena-managed nodes have independent allocation lifetimes.

The implementation's precise node-reference and payload invariants are enforced by the corresponding AST tests.

### 5.4 Legacy representation

The M1 AST and parser remain in the frontend packages under `legacy_*` source filenames.

They serve as reference implementations for:

- Regression testing.
- Differential correctness checks.
- Performance comparisons.
- Investigating behavioral differences introduced by H2.

The legacy representation is not the canonical production output.

The current `FlatNode` and `FlatProgram` names remain in use. Any future naming simplification should be handled separately from this optimization milestone.

## 6. H2 correctness validation

H2 was validated using unit tests, integration tests, differential tests, and frontend benchmarks.

The validation includes AST arena invariants and parser behavior.

The frontend reorganization was subsequently checked using:

```bash
go test ./...
go vet ./...
```

Both passed after removal of the temporary compiler benchmark workspace from the Go module's package discovery path.

The migration did not intentionally change Eloque's accepted grammar or observable parsing semantics.

## 7. H2 performance evaluation

H2 was benchmarked against the retained M1 frontend and external parser implementations.

The official results are preserved in:

```text
benchmarks/results/official/
```

Experimental arena-capacity and allocation investigations are preserved separately under:

```text
benchmarks/results/experimental/
```

### 7.1 M1 versus H2

The official M1/H2 workload suite evaluated declarations, expressions, strings, comments, and mixed programs.

At 1,000 workload units, the recorded median results were:

| Workload | M1 | H2 | Speedup |
|---|---:|---:|---:|
| Declarations | 238.1 µs | 209.0 µs | 1.14× |
| Expressions | 1,181.5 µs | 667.4 µs | 1.77× |
| Strings | 400.1 µs | 394.2 µs | 1.01× |
| Comments | 407.8 µs | 346.4 µs | 1.18× |
| Mixed | 1,683.8 µs | 1,191.1 µs | 1.41× |

H2 reduced allocation counts across these workloads.

The largest measured improvement among these examples occurred in the expression-heavy workload. String-heavy performance was approximately unchanged.

These results support H2 as a beneficial implementation optimization for the tested workloads, not a universal speed guarantee.

### 7.2 Controlled cross-language comparison

A separate controlled arithmetic suite compared Eloque H2, the Go compiler syntax parser, Go's standard-library parser, and CPython 3.14.4.

At 10,000 workload units:

| Frontend | Median latency |
|---|---:|
| Eloque H2 | 7.16 ms |
| Go compiler syntax parser | 14.12 ms |
| Go `go/parser` | 23.92 ms |
| CPython 3.14.4 `ast.parse` | 147.86 ms |

These measurements are specific to the benchmark inputs and harnesses.

The implementations process different language grammars and produce different AST structures. The results do not establish equivalent compiler responsibilities or general superiority over Go or Python.

They also do not measure Eloque execution performance, since Eloque has no executable runtime yet.

For methodology, environment details, reproduction commands, and historical comparisons, see [benchmarks/README.md](../benchmarks/README.md).

## 8. Deferred frontend optimizations

H2 experiments investigated arena capacity, growth behavior, and allocation patterns.

These experiments do not automatically justify additional production optimizations.

Further arena-capacity tuning should be reconsidered when representative workloads reveal meaningful allocation or growth overhead.

A proposed optimization must demonstrate:

- Correctness preservation.
- Improved performance on representative inputs.
- Acceptable memory consumption across different program shapes.
- No unjustified regression for smaller workloads.
- Benefits that justify additional implementation complexity.

The review trigger and supporting evidence belong in [TECHNICAL_BACKLOG.md](TECHNICAL_BACKLOG.md).

## 9. Semantic analysis

The next compiler phase will introduce name resolution and static type checking.

### Name resolution

The resolver will bind identifier uses to declarations and enforce the language's scope rules.

It must reject undeclared identifiers and prohibited shadowing.

### Static type checking

The checker will validate the implemented language's type rules before execution.

Planned responsibilities include:

- Type inference for eligible local bindings.
- Assignment compatibility.
- Arithmetic operand validation.
- Explicit function signatures when functions are implemented.
- Return-path checking.
- Optional-value handling.
- Source-aware type diagnostics.

Static checking is mandatory for normal compilation.

The frontend's AST representation must not be confused with the semantic information produced by these future stages.

## 10. Bytecode compiler

The initial backend will compile statically validated programs into bytecode.

Its intended responsibilities include:

- Instruction generation.
- Constant-pool management.
- Local-variable representation.
- Branch and call support as language features are added.
- Source-location and debugging metadata.

The initial design may compile from the validated AST without a general-purpose intermediate representation.

A separate backend-neutral IR should be introduced only when a concrete optimization or additional backend justifies its complexity.

## 11. Virtual machine and runtime

The first executable Eloque implementation is planned as a Go-hosted stack-based bytecode VM.

Its initial scope will cover a small statically checked subset, including constants, arithmetic, bindings, and output.

Future capabilities include control flow, functions, collections, custom types, exceptions, and concurrency.

### Runtime value model

The language's intended observable value semantics are independent of the internal Go representation.

Planned primitive types include:

- `Int`: checked signed 64-bit integers.
- `Float`: IEEE 754 binary64.
- `Bool`: strict Boolean values.
- `String`: immutable Unicode text.
- Optional values represented through `T?`.

Lists, maps, and structured values are intended to preserve logical value independence under assignment.

The runtime representation—Go interfaces, tagged values, pointer-backed objects, or another approach—remains undecided.

It must be reviewed before implementing the first VM value type.

### VM architecture review

The stack VM is the baseline, not an irreversible commitment.

Once a working VM and reproducible execution benchmarks exist, a register-based design may be evaluated against stack-machine optimizations.

The decision should consider dispatch cost, instruction count, memory traffic, bytecode size, and implementation complexity.

## 12. Modules, exceptions, and concurrency

### Modules

Each Eloque source file is intended to represent a module.

The proposed module system includes imports, controlled visibility, deterministic initialization, and rejection of circular imports in the initial design.

Module resolution and caching remain future work.

### Exceptions

The intended exception system uses typed failures and structured handling.

Exception matching, propagation, cleanup, and unwinding semantics require implementation-stage review.

### Concurrency

The initial VM is sequential.

Native concurrency is a longer-term language goal.

Go goroutines may provide an implementation mechanism, but Eloque's observable concurrency contract must be designed independently.

Task ownership, cancellation, shared mutable state, race safety, and error propagation must be specified before implementation.

The VM should avoid unnecessary global mutable state that would obstruct future independent execution contexts.

## 13. Future execution backends

Possible future improvements include:

- Bytecode specialization.
- Register-based VM evaluation.
- JIT compilation.
- AOT/native compilation.
- WebAssembly.
- Additional compiler optimizations.

These are research or review items, not features of the completed frontend.

Performance comparisons with CPython must eventually use equivalent, correctness-tested executable programs once Eloque has a working runtime.

Frontend parsing measurements must not be used as evidence of runtime performance parity.

## 14. Compatibility and evolution

Before version 1.0, syntax, semantics, and implementation details may evolve with corresponding specification and conformance-test updates.

At version 1.0, published source-level semantics will form the compatibility commitment.

Internal AST representations and bytecode encodings are not public compatibility guarantees unless explicitly standardized.

Changes to internal implementation should preserve observable behavior unless a deliberate language revision is documented.

## 15. Architecture review gates

| Decision | Review trigger |
|---|---|
| Arena capacity and allocation tuning | Representative frontend profiling demonstrates a meaningful opportunity |
| Additional parser lookahead | A grammar feature requires it or a measured regression justifies it |
| Runtime value representation | Before first VM value implementation |
| Numeric edge cases | Arithmetic checker and runtime implementation |
| Stack versus register VM | After working VM and reproducible execution benchmarks |
| Backend-neutral IR | Concrete second-backend or optimization requirement |
| Collection value independence | Before mutable collections |
| Exception taxonomy and unwinding | Exception implementation milestone |
| Module resolution and caching | Module implementation milestone |
| Native concurrency and ownership | Dedicated concurrency design milestone |
| JIT, AOT, and WebAssembly | Demonstrated performance or deployment requirement |
| Structured diagnostics and sandbox controls | Tooling and agent-execution milestones |

All deferred engineering decisions must retain a rationale, review trigger, and measurable acceptance criteria in [TECHNICAL_BACKLOG.md](TECHNICAL_BACKLOG.md).

---

**Current architectural baseline:** H2 direct-streaming frontend with arena-backed flat AST, retained M1 reference implementation, and verified frontend correctness.

**Next development direction:** A small statically checked executable subset, progressing through semantic analysis, bytecode compilation, and VM execution.