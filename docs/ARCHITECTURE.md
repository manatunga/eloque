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

---

## M2 implementation blueprint: diagnostics and name resolution

**Status:** Approved architecture — implementation pending.

**Last reviewed:** October 2026.

This section records the architectural decisions agreed before the M2 implementation phase. It is intended to remain the primary implementation reference when development resumes.

### A. Compiler pipeline and stage boundaries

The target pipeline is:

```text
Source registration
       |
       v
Streaming lexer
       |
       v
Recovering parser
       |
       v
Validated flat AST
       |
       v
Name resolver
       |
       v
Mandatory static checker
       |
       v
Bytecode compiler
       |
       v
Stack VM
```

The diagnostic system spans the frontend and semantic stages, but does not eliminate their individual responsibilities.

Name resolution must not be embedded into parsing or bytecode generation.

The resolver receives a syntactically valid `ast.FlatProgram` and returns resolution information plus diagnostics.

Recovered or incomplete frontend output must not enter the normal semantic compilation pipeline.

### B. Shared diagnostic contract

The conceptual diagnostic model is:

```go
type Severity uint8

const (
    SeverityError Severity = iota
    SeverityWarning
    SeverityInfo
)

type Code string

type Diagnostic struct {
    Code     Code
    Severity Severity
    Message  string
    Primary  Location
    Related  []RelatedLocation
    Notes    []string
}

type Location struct {
    File FileID
    Span source.ByteSpan
}

type RelatedLocation struct {
    Location Location
    Message  string
}
```

These types describe the approved contract, not mandatory final Go declarations.

The concrete package boundaries and names may be refined without changing the agreed behavior.

#### Source ownership

The compiler should maintain a minimal compilation-local source registry.

It must be capable of associating each `FileID` with the corresponding source text, source name/path, and position mapping.

`FileID` is compilation-local and must not be interpreted as a persistent file identity across recompilations.

Byte spans are half-open, using inclusive start and exclusive end offsets.

Line and column positions are calculated when diagnostics are rendered.

The existing frontend position-mapping implementation should be reused or consolidated where appropriate, rather than duplicated unnecessarily.

A full multi-file workspace or incremental-compilation framework is not required for M2.

#### Diagnostic codes

Codes must be stable and machine-readable.

Proposed code families:

- `LEX_*`
- `PARSE_*`
- `RESOLVE_*`
- `TYPE_*`

Examples:

- `LEX_UNEXPECTED_CHARACTER`
- `LEX_INVALID_INTEGER`
- `LEX_INVALID_INDENTATION`
- `PARSE_EXPECTED_TOKEN`
- `RESOLVE_UNDEFINED_NAME`
- `RESOLVE_DUPLICATE_DECLARATION`
- `RESOLVE_SHADOWED_DECLARATION`
- `TYPE_UNINITIALIZED_READ`
- `TYPE_MISMATCH`

Exact identifiers should be finalized alongside the corresponding implementation.

Published diagnostic codes should retain their meanings even when messages are improved.

#### Parser-specific metadata

The parser currently retains expected and actual token information.

That metadata must not be discarded during migration.

It should remain in a parser-specific representation and be convertible into generic structured diagnostic output.

The shared diagnostic package must not import the lexer merely to store token kinds.

The exact machine-readable serialization schema is deferred until tooling requires it.

### C. Diagnostic collection policy

Each stage owns its own diagnostic slice.

A compiler driver coordinates:

- A compilation-wide diagnostic budget.
- Deterministic diagnostic ordering.
- Blocking-error detection.
- Stage eligibility.
- Final diagnostic presentation.

A default budget of 100 errors is an initial candidate.

The total budget must not be accidentally multiplied by the number of compiler stages.

Independent errors should be reported when safe. Cascading diagnostics arising solely from earlier failures should be suppressed where practical.

A diagnostic collection must not require global mutable state.

Compiler stages should receive lightweight policy/configuration values when necessary.

#### Stage outcomes

Three conceptual outcomes are required:

**Complete**

The stage finished without blocking errors. Warnings and informational diagnostics may be present.

**Recovered**

The stage continued after blocking errors and may have produced partial results. Those results are not eligible for normal executable compilation.

**Aborted**

The stage could not safely continue. Downstream stages must not consume its output as valid.

The final concrete result types may differ between lexer, parser, and semantic analysis.

A common diagnostic model does not require a universal stage-result struct.

### D. Frontend recovery migration

The current implementation is first-error terminating.

The lexer returns `(Token, error)`.

The streaming token source caches the first lexical error and substitutes `TokenInvalid`, preventing further scanning.

The parser's `expect()` returns a diagnostic without consuming the unexpected token.

These behaviors are suitable for the original fail-fast implementation but require deliberate changes for recovery.

#### Lexer recovery requirements

Unexpected characters may be skipped when doing so preserves reliable scanning.

Malformed numeric literals should be consumed as a coherent malformed unit where possible, preventing misleading follow-on tokenization.

Invalid escapes and unterminated strings require recovery at a reliable string or line boundary.

Invalid indentation must not leave the indentation stack in an inconsistent state.

Unmatched or unclosed parentheses require special treatment because they influence logical newline recognition.

When structural state cannot be restored safely, recovery may terminate the affected region or stage.

Repeated calls must never reproduce the same error indefinitely without advancing or terminating.

#### Parser recovery requirements

Parser recovery must explicitly guarantee progress.

Potential synchronization tokens include:

- `NEWLINE`
- `DEDENT`
- `EOF`

Their suitability depends on grammatical context, nesting, and future block constructs.

A newline inside parentheses must not automatically be treated as a safe statement boundary.

When a statement cannot be completed safely, the parser should abandon that statement and synchronize before attempting another.

The arena may contain expression nodes constructed before a failure. Those nodes must not be included in the program's valid statement range unless the containing statement was successfully constructed and validated.

Recovered AST output must be clearly distinguished from a complete, valid AST.

#### Migration acceptance criteria

- Multiple independent errors can be reported in one compilation.
- Lexical recovery does not corrupt structural tokenization.
- Parser recovery cannot enter an infinite loop.
- Diagnostic positions remain accurate.
- Errors are reported in deterministic order.
- Cascading diagnostics are appropriately suppressed.
- Recovered programs cannot be executed.
- Existing valid M1/H2 programs continue parsing correctly.
- Streaming token production remains supported.
- Arena invariants remain valid.
- Representative benchmark regressions are measured and reviewed.

### E. Name-resolution responsibilities

The resolver determines which declaration each name refers to.

It is not responsible for general type compatibility, bytecode emission, or runtime storage allocation.

The resolver must:

1. Create symbol identities for declarations.
2. Manage lexical scopes.
3. Resolve identifier expressions.
4. Resolve assignment targets.
5. Detect undefined names.
6. Detect duplicate declarations.
7. Report shadowing warnings.
8. Preserve binding information for later compiler stages.
9. Produce deterministic diagnostics.
10. Leave the AST unchanged.

Definite-assignment analysis is a separate responsibility.

### F. Symbol storage

Use a contiguous append-only symbol arena, conceptually:

```go
type SymbolID uint32

type Resolution struct {
    symbols []Symbol
}
```

An explicit invalid symbol sentinel is required.

Symbol identities are stable within one compilation unit but not across recompilations.

A symbol ID is not a VM register, stack slot, or global runtime address.

Future module systems may require qualified identities, but that should not be introduced prematurely.

Each symbol must retain sufficient information to identify its declaration and support diagnostics.

Exact `Symbol` fields should be finalized when implementation begins.

### G. Lexical scope representation

Use parent-linked scopes.

Each scope contains its locally declared symbols, conceptually:

```go
type Scope struct {
    parent *Scope
    names  map[string]SymbolID
}
```

Lookup searches:

1. Current scope.
2. Parent scope.
3. Successive enclosing scopes.
4. Stop when found or no parent remains.

Do not copy inherited scope maps into nested scopes.

A duplicate name in the same scope is an error.

A matching name in an enclosing scope is permitted but generates a warning.

The shadowing warning must reference the earlier declaration's source location.

The original symbol must remain authoritative after a duplicate-declaration error.

### H. Source-backed identifiers

The H2 AST retains original source text and byte spans.

The resolver should initially obtain identifier names through source-backed string slices rather than eagerly allocating copied strings.

The lifetime of the original source must be retained for as long as these names are referenced.

Do not introduce global string interning without benchmark evidence.

Identifier comparison must remain exact and case-sensitive.

### I. Reference binding representation

The initial representation is a dense array of `SymbolID` values indexed by AST `NodeID`.

Entries without bindings use the invalid symbol sentinel.

The representation must remain private behind accessors so it can be replaced with a sparse mapping if benchmarks justify doing so.

Conceptual operations:

```go
Symbol(SymbolID)
Reference(ast.NodeID)
Declaration(ast.NodeID)
```

The exact accessor signatures are not yet fixed.

#### H2 AST binding conventions

The current AST does not represent every name as an identifier-expression node.

**Identifier expressions**

`NodeIdentifier` has its own `NodeID`.

The resolver records the resolved symbol for that node.

**Declarations**

`NodeLetStatement` stores the declaration name as a byte span in its binding payload.

The resolver records the declaration symbol using the statement's `NodeID`.

**Assignments**

`NodeChangeStatement` stores its target name as a byte span in its binding payload.

The resolver records the assignment-target reference using the statement's `NodeID`.

The compiler must not assume assignment targets are separate `NodeIdentifier` nodes.

### J. Resolver traversal algorithm

For the currently supported grammar, process top-level statements in source order.

#### `let`

1. Read the declaration name from its source span.
2. Check the current scope for an existing declaration.
3. If duplicated, emit an error without replacing the original binding.
4. Otherwise, check enclosing scopes for a shadowed declaration.
5. Emit a warning when shadowing occurs.
6. Allocate and register the new symbol.
7. Resolve the initializer expression.
8. Record the declaration binding.

The declaration must be visible during initializer resolution.

#### `change`

1. Read the assignment target name.
2. Resolve it through the scope chain.
3. Report an undefined target when lookup fails.
4. Record the target binding when successful.
5. Resolve the assigned expression.

Mutability and type compatibility are checked separately.

#### `say`

Resolve the contained expression.

#### Expressions

- Integer and string literals require no name lookup.
- Identifier expressions resolve through the current scope.
- Unary expressions resolve their operand.
- Binary expressions resolve both operands.

Traversal must preserve deterministic diagnostic order.

Future nested blocks, functions, and named types extend this algorithm rather than changing the underlying lexical binding rules.

### K. Forward declarations

Variables require declaration before use.

Future functions and named types may be referenced before their textual definitions.

When those constructs are implemented, use an appropriate scope-local declaration prepass to register eligible symbols before resolving bodies.

Do not implement speculative function/type AST nodes or module-resolution behavior during the initial M2 resolver.

### L. Initializer self-reference

Consider:

```elq
let x be x + 1
```

The resolver creates the symbol for `x` before resolving the initializer.

The initializer's `x` resolves to that new symbol.

The definite-assignment checker subsequently rejects the uninitialized read.

For nested scopes:

```elq
let x be 10

if ready:
    let x be x + 1
```

The inner initializer refers to the inner declaration, not the outer one.

The inner declaration also produces a shadowing warning.

The syntax example is illustrative of future block-scope behavior; `if` is not yet part of the implemented M1/H2 parser.

### M. Behavioral test matrix

The following cases form the acceptance matrix for the resolver and diagnostic migration.

**Valid name resolution**

- R01: Empty program.
- R02: Single declaration.
- R03: Multiple declarations.
- R04: Identifier read after declaration.
- R05: Repeated references resolve to the same symbol.
- R06: Assignment target resolves correctly.
- R07: Assignment RHS resolves correctly.
- R08: Nested unary/binary expressions resolve all identifier references.
- R09: Independent declarations receive distinct stable symbol IDs.
- R10: Resolver output is deterministic for identical input.

**Invalid name resolution**

- R11: Undefined identifier expression.
- R12: Undefined assignment target.
- R13: Variable used before declaration.
- R14: Duplicate declaration in the same scope.
- R15: Duplicate declaration does not replace the original symbol.
- R16: Multiple independent undefined names produce independent diagnostics.
- R17: Malformed syntax is not passed to normal resolution.

**Scope behavior**

- R18: Nearest lexical declaration wins.
- R19: Nested shadowing produces a warning.
- R20: Shadowing warning references the earlier declaration.
- R21: Unrelated scopes may reuse names.
- R22: Scope exit restores visibility of the enclosing declaration.

**Initializer and assignment analysis**

- R23: Initializer self-reference binds to the newly declared symbol.
- R24: Definite-assignment analysis rejects an uninitialized self-reference.
- R25: Shadowing initializer resolves to the inner symbol.

**Integrity and diagnostics**

- R26: Identifier, declaration, and assignment-target bindings use correct AST node identities.
- R27: Resolver does not mutate the AST; source spans and retained source remain valid.
- R28: Diagnostic spans, ordering, symbol sentinels, and result accessors behave deterministically.

Scope-related syntax-level tests must wait until nested scopes are supported. Resolver scope behavior may be exercised earlier through appropriately isolated fixtures.

### N. Additional diagnostic test requirements

Beyond R01–R28, validate:

- Exact half-open byte spans.
- Compilation-local file identity.
- Stable diagnostic codes.
- Severity classification.
- Related declaration locations.
- Warning-only compilation eligibility.
- Recovered AST ineligibility.
- Aborted-stage behavior.
- Global diagnostic budget enforcement.
- Deterministic ordering.
- No infinite lexer/parser recovery loops.
- No misleading duplicate cascades.
- No unnecessary diagnostic allocations on valid input.
- Safe behavior under deeply nested expressions.
- Preservation of expected/actual token metadata.

### O. Performance and deferred architecture

The following decisions must be revisited only when their stated trigger is reached.

**Dense versus sparse reference bindings**

Trigger: representative resolver memory and lookup benchmarks.

Criteria: retained memory, allocation count, lookup cost, implementation complexity.

**Source-backed strings versus interning**

Trigger: workloads with repeated identifiers and eventually multi-file compilation.

Criteria: allocation savings, retained source lifetime, lookup performance, memory overhead.

**Recursive versus iterative expression traversal**

Trigger: deeply nested expression stress tests.

Criteria: stack safety, correctness, traversal overhead, implementation complexity.

**Diagnostic budget and public serialization**

Trigger: practical CLI/editor/agent tooling.

Criteria: actionable output, deterministic serialization, bounded resource use, stable compatibility.

**Multi-file source manager**

Trigger: module imports and multi-file compilation, expected around M7.

Criteria: file identity correctness, source ownership, diagnostics across modules, incremental-compilation requirements if justified.

**Stack VM versus register VM**

Trigger: a functioning executable stack VM with representative runtime benchmarks.

Criteria: execution speed, memory use, bytecode size, compiler complexity, maintainability.

**JIT, AOT, WebAssembly and alternative backends**

Trigger: measured limitations of the baseline VM and concrete deployment requirements.

Criteria: demonstrable improvements against reproducible workloads, implementation cost, portability, maintenance burden.

**CPython performance comparisons**

Trigger: working runtime with comparable language workloads.

Frontend parsing benchmarks must not be presented as evidence of runtime parity.

### P. Implementation readiness checklist

Before beginning resolver implementation:

- [ ] Shared diagnostic contract recorded.
- [ ] Frontend recovery strategy implemented.
- [ ] Streaming token source supports safe diagnostic continuation.
- [ ] Parser recovery guarantees forward progress.
- [ ] Stage outcomes and compilation eligibility enforced.
- [ ] Diagnostic regression tests pass.
- [ ] Existing frontend correctness tests pass.
- [ ] H2 performance impact measured.
- [ ] Resolver data structures and APIs reviewed against actual current AST files.

Before considering M2 complete:

- [ ] Name resolution implemented and tested.
- [ ] Mandatory static checker implemented and tested.
- [ ] Definite-assignment behavior validated.
- [ ] Bytecode compiler implemented.
- [ ] Stack VM executes the supported language subset.
- [ ] Blocking diagnostics prevent execution.
- [ ] Runtime correctness and representative benchmarks recorded.

This blueprint is an approved design reference. It does not imply any listed implementation is complete.