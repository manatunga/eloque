# Eloque Design Decisions

> **Status:** Active design decision register, reviewed November 2026. Includes completed M1/H2 frontend work, post-H2 structured recovery, and M2-P1 lexical name resolution.

**Normative language specification:** [LANGUAGE_SPEC.md](LANGUAGE_SPEC.md).

This register records accepted design choices, provisional directions, deferred work, and selected implementation decisions. **Decided** means agreed, not necessarily implemented. **Provisional** requires further specification or validation. **Deferred** is deliberately postponed. **Implemented (M1)** identifies behavior or architecture established during Milestone 1. **Implemented (H2)** identifies the post-M1 frontend optimization and architectural migration. Neither status implies that Eloque programs can execute. **Implemented (post-H2)** identifies frontend functionality added after the H2 optimization, including structured diagnostic recovery. It does not imply that semantic analysis or execution is available.

**Implemented (M2-P1)** identifies the completed lexical name-resolution stage, including symbol binding and resolver diagnostics for the supported language subset. It does not imply that static type checking, bytecode compilation, or execution is available.

## Current implementation baseline — M1, H2, and M2-P1

Milestone 1 established Eloque's Go lexer, parser, AST construction, structured fail-fast lexical and syntax diagnostics, and canonical `Parse(sourceText)` entry point.

### Post-H2 structured diagnostics and recovery

The production frontend additionally supports `ParseRecovering(sourceText, ParseOptions)`, which returns a `ParseResult` containing a potentially partial arena-backed AST, collected diagnostics, and a termination status.

The possible statuses are `ParseComplete`, `ParseRecovered`, and `ParseAborted`. Recovery is bounded by a configurable error budget: zero selects the default of 20 errors, positive values specify a limit, and negative values disable the limit.

The recovery architecture preserves successfully completed statements where possible, while avoiding retention of incomplete AST construction. The frontend also guards against excessive expression nesting.

The original `Parse(sourceText)` API remains available for fail-fast parsing. Recovery-capable parsing is intended for diagnostics and tooling; a partial AST is not evidence of semantic validity or executability.

These capabilities are implemented frontend behavior, not changes to the language's normative grammar.

The supported syntax subset includes literals, identifiers, arithmetic expressions, and `let`, `change`, and `say` statements.

The post-M1 H2 optimization introduced an arena-backed flat AST and a new production parser implementation while preserving direct pull-based token streaming.

The production frontend exposes `Parse(sourceText) (*ast.FlatProgram, *diagnostic.Diagnostic)` for fail-fast parsing and `ParseRecovering(sourceText, ParseOptions) ParseResult` for structured recovery. Recovery-enabled parsing is the approved default for user-facing compilation and tooling; integration into the compiler driver remains pending.

The original M1 interface-based AST and parser remain available internally as legacy reference implementations for regression and performance comparisons.

The frontend packages are organized under `internal/frontend/{ast,lexer,parser}`, while shared source-location utilities remain in `internal/source`.

### M2-P1 lexical name resolution

M2-P1 introduced a dedicated lexical name-resolution stage following parsing.

The resolver associates identifier references and assignment targets with declaration symbols, detects undefined names and duplicate declarations, and preserves binding information for subsequent compiler stages.

The implemented resolver is separate from parsing and does not perform static type checking, definite-assignment analysis, bytecode generation, or execution.

Resolver correctness tests and dedicated benchmarks have been established. The M2 cross-language v2 comparison includes fail-fast and recovery-capable parse-and-resolve workloads.

M2 — First Execution remains in progress.

H2 was validated through correctness and differential testing, arena-invariant tests, and representative performance benchmarks.

At 1,000 workload units, H2 recorded median improvements over M1 ranging from approximately 1.01× for strings to 1.77× for expressions, with lower allocation counts across the tested workloads.

The controlled cross-language benchmarks concern frontend processing only and do not establish general compiler or runtime performance superiority.

**Static type checking, bytecode generation, and execution are not yet implemented.**

See [ARCHITECTURE.md](ARCHITECTURE.md) for the implementation structure, [ROADMAP.md](ROADMAP.md) for milestone progress, and [benchmarks/README.md](../benchmarks/README.md) for benchmark methodology and results.

## Current decisions

| ID            | Status                                                                                                     | Decision                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                       |
| ------------- | ---------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| D-001         | Decided; frontend implemented (M1/H2/post-H2), name resolution implemented (M2-P1), execution pending (M2) | Eloque's compiler frontend and semantic-analysis stages are implemented in Go. M1/H2 established lexical analysis, parsing, and arena-backed flat AST construction; post-H2 work added structured diagnostic recovery; M2-P1 implemented lexical name resolution and symbol binding. Mandatory static type checking and bytecode compilation remain pending. Execution will use **Monowire**, the selected register-based VM architecture, rather than the previously planned Go stack VM.                                                                                                                                                     |
| D-002         | Decided; CLI provisional                                                                                   | **Eloque** is the language name and `.elq` its official source extension. `eloque` is the planned CLI command, subject to availability and naming-conflict checks.                                                                                                                                                                                                                                                                                                                                                                                                                                                                             |
| D-003         | Decided                                                                                                    | General-purpose language with concise, structured conversational syntax—not unrestricted English. Prioritize readability, predictable agent tooling, and foundations for safe concurrency.                                                                                                                                                                                                                                                                                                                                                                                                                                                     |
| D-004         | Decided; initial lexer implemented (M1)                                                                    | Four-space significant indentation, `:` blocks, no tabs, `#` comments, case-sensitive ASCII identifiers, and double-quoted strings.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                            |
| D-005         | Decided; initial syntax implemented (M1)                                                                   | `let x be value` declares a mutable binding; `change x to value` reassigns. Initialized local bindings may infer types when type checking is available.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        |
| D-006         | Decided                                                                                                    | `let fixed x be value` prevents rebinding; `let frozen x be value` provides deep immutability. Enforcement awaits later stages.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                |
| D-007         | Decided                                                                                                    | `true`, `false`, and `none` are language literals. `none` is valid in optional `T?` contexts, not arbitrary non-optional values.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               |
| D-008         | Decided; checker pending                                                                                   | Strict static typing with `as Type` annotations and local inference; checker errors block normal compilation. No implicit truthiness.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                          |
| D-009         | Decided; functions pending                                                                                 | Function parameters require explicit types; value-returning functions require `giving Type`. Omission denotes no value return.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                 |
| D-010         | Decided; control flow pending                                                                              | `if`, `otherwise if`, and `otherwise`; lexical block scopes; nested lexical shadowing is permitted with a nonfatal warning referencing the shadowed declaration; same-scope duplicates are errors.                                                                                                                                                                                                                                                                                                                                                                                                                                             |
| D-011         | Decided; loops pending                                                                                     | `for each`, `repeat N times`, `repeat while`, `repeat until`, `skip`, and `stop`.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              |
| D-012         | Decided; functions pending                                                                                 | `to name with (parameters) giving Type:` declares functions; `name with (arguments)` invokes them, including zero-argument calls; `give` returns.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              |
| D-013         | Decided; functions pending                                                                                 | Named arguments and defaults use `be`. Verb-first function names are a convention, not a restriction.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                          |
| D-014         | Deferred                                                                                                   | Closures, first-class function details, variadics, and generics require deliberate grammar design; earlier `outer` and anonymous-function syntax are not automatically retained.                                                                                                                                                                                                                                                                                                                                                                                                                                                               |
| D-015         | Decided; arithmetic grouping implemented (M1)                                                              | Parentheses group expressions. Operands and arguments evaluate left-to-right, except Boolean short-circuiting. Runtime evaluation rules remain to be implemented.                                                                                                                                                                                                                                                                                                                                                                                                                                                                              |
| D-016         | Decided; custom types pending                                                                              | `define` introduces structured types with fields and methods; implicit `self`; structural interfaces; no traditional class inheritance.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        |
| D-017         | Decided; exceptions pending                                                                                | Typed unchecked exceptions use `raise`, `try`, `catch`, and `finally`; a `try` requires at least one `catch` or `finally`.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                     |
| D-018         | Decided; modules pending                                                                                   | `import` loads modules and `named` specifies aliases. Modules initialize once; circular imports are initially rejected; underscore-prefixed top-level names are compiler-private.                                                                                                                                                                                                                                                                                                                                                                                                                                                              |
| D-019         | Decided; output syntax implemented (M1)                                                                    | `say` is the output statement. Interpolation syntax is planned as `"Hello, {name}"`, but M1 parses strings as ordinary literals without interpolation evaluation. Input API remains unspecified.                                                                                                                                                                                                                                                                                                                                                                                                                                               |
| D-020         | Decided; collections pending                                                                               | Lists and maps use conversational access and mutation; `remove all` removes all matching list values; `find` returns optional values.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                          |
| D-021         | Decided; comparisons pending                                                                               | Word-based comparisons (`is`, `is not`, `is greater than`, etc.) and structural equality where supported. Chained comparisons are excluded from the initial grammar.                                                                                                                                                                                                                                                                                                                                                                                                                                                                           |
| D-022         | Decided; Monowire implementation pending (M2)                                                              | Adopt **Monowire**, a register-based bytecode virtual machine, as Eloque's production execution architecture. The previously approved stack-first VM plan is superseded. Monowire uses explicit register operands, call frames, and overlapping caller-outgoing/callee-incoming register windows. Its initial implementation belongs to M2 — First Execution. The stack-versus-register architecture decision is closed; remaining runtime representation, instruction encoding, safety, and integration details require implementation validation.                                                                                            |
| D-023         | Provisional                                                                                                | CPython runtime performance parity is an aspiration, not a guarantee. M1 parsing benchmarks are not evidence of execution-speed parity; benchmark runtime performance after the VM exists.                                                                                                                                                                                                                                                                                                                                                                                                                                                     |
| D-024         | Provisional                                                                                                | Backend extensibility, standard library scope, packaging, and post-1.0 compatibility policy need further specification.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        |
| D-025         | Decided; collections pending                                                                               | Zero-based indexing, negative indices from the end, end-exclusive slices, and typed failures for invalid direct access. Unicode indexing units remain provisional.                                                                                                                                                                                                                                                                                                                                                                                                                                                                             |
| D-026         | Superseded in part by D-053–D-056; implementation pending                                                  | `Int` is signed 64-bit and `Float` is IEEE 754 binary64. M2 adds `Int8`, `Int16`, and `Int32`. Integer literals use exact context-directed typing with `Int` as the fallback. Numeric conversions and mixed arithmetic follow D-056. Float display uses the shortest round-trippable decimal.                                                                                                                                                                                                                                                                                                                                                  |
| D-027         | Refined by D-054 and D-055; implementation pending                                                         | `/` denotes real division, `//` floor division, and `%` floor-compatible remainder. Integer division by zero is an error. Integer overflow is checked. Floating-point operations follow the IEEE 754 binary64 and explicitly documented Eloque conventions.                                                                                                                                                                                                                                                                                                                                                                                    |
| D-028         | Decided; collections pending                                                                               | Lists, maps, and structured values have logical value semantics. Copy-on-write or structural sharing may implement them without observable aliasing.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           |
| D-029         | Deferred                                                                                                   | Design explicit references, shared mutable state, and concurrency syntax together.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                             |
| D-030         | Decided; pattern matching pending                                                                          | `match`, `when`, and `otherwise`, without fallthrough. Statement matches may be non-exhaustive; value-producing matches must be exhaustive.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                    |
| D-031         | Decided; execution pending                                                                                 | Top-level statements execute in source order; `main` is not invoked automatically.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                             |
| D-032         | Decided; diagnostics implemented in part (M1)                                                              | Normal execution is sequential; runtime exceptions include source locations; tooling requires structured diagnostics. M1 provides structured frontend diagnostics, not runtime exceptions.                                                                                                                                                                                                                                                                                                                                                                                                                                                     |
| D-033         | Decided; runtime pending                                                                                   | Support configurable sandbox limits; do not impose an arbitrary universal instruction limit on ordinary CLI execution.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                         |
| D-034         | Decided; lexer implemented (M1), recovery implemented (post-H2)                                            | Decimal integer literals prohibit leading zeros except `0`. A digit sequence immediately followed by an ASCII letter or underscore is a lexical error; invalid-suffix diagnostics take precedence over leading-zero diagnostics. The original fail-fast lexical behavior remains available, while the production frontend also supports controlled multi-error recovery.                                                                                                                                                                                                                                                                       |
| D-035         | Implemented (M1)                                                                                           | Go lexer token kinds use a `uint8`-backed enumeration with zero reserved for `TokenInvalid`; `String()` supplies readable names. Incremental token emission and reused indentation-stack/pending-queue storage reduce allocations. This is an implementation choice, not a language semantic rule.                                                                                                                                                                                                                                                                                                                                             |
| D-036         | Decided; lexer implemented (M1)                                                                            | `change`, `to`, `with`, and `giving` are reserved keywords with dedicated token kinds, not ordinary identifiers. One token kind may serve multiple grammatical contexts.                                                                                                                                                                                                                                                                                                                                                                                                                                                                       |
| D-037         | Decided; recovery implemented (post-H2); default integration pending                                       | Recovery-enabled parsing is the default for Eloque's user-facing compilation and development tooling. `ParseRecovering(sourceText, ParseOptions)` provides structured multi-error recovery, configurable error budgets, and explicit parse outcomes. The original fail-fast `Parse(sourceText)` remains available for internal operations and performance comparisons. Recovery preserves direct pull-based token streaming, deterministic diagnostics, safe termination, and the arena-backed AST.                                                                                                                                            |
| D-038         | Implemented (M1/H2); default parser selection updated (M2 decision)                                        | Preserve direct pull-based token streaming without a redundant independent token cursor. Both fail-fast and recovery-enabled parsing use the production arena-backed frontend. `Parse(sourceText)` remains the explicit fail-fast entry point; `ParseRecovering(sourceText, ParseOptions)` is the selected default for user-facing compilation. Buffered parsing remains available for regression and benchmark comparisons.                                                                                                                                                                                                                   |
| D-039         | Implemented (H2)                                                                                           | Adopt an arena-backed flat AST as the canonical production representation. The parser constructs `FlatNode` and `FlatProgram` structures using arena-managed storage. Retain the original M1 interface-based AST and parser as legacy references for correctness and performance comparisons. This is an internal representation decision, not a language-semantic change.                                                                                                                                                                                                                                                                     |
| D-040         | Implemented (H2)                                                                                           | Organize the frontend under `internal/frontend/{ast,lexer,parser}` while retaining shared source-location utilities in `internal/source`. Preserve the canonical `Parse(sourceText)` entry point and direct pull-based token streaming. Future semantic analysis and backend packages should be introduced when their implementations begin rather than creating speculative empty package structures.                                                                                                                                                                                                                                         |
| D-041 - D-050 | Decided/Provisional                                                                                        | [Jump to October M2 Decisions](#m2-architecture-decisions---october-2026)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      |
| D-051         | Provisional; experimental benchmarks completed                                                             | **Zig is the preferred implementation language for Monowire**, with Rust retained as the fallback. The October 23, 2026 experiment compared Go, Rust, Zig switch dispatch, and Zig labeled-switch dispatch across call-loop, arithmetic, branching, and recursion workloads. Zig led all four measured workloads. The results include subprocess overhead and implementation differences, so they do not establish production VM performance. Confirm Zig's suitability through Go integration, memory-ownership, portability, error-handling, and build-system validation before finalizing the production toolchain.                         |
| D-052         | Provisional; optimization review deferred                                                                  | Use conventional switch dispatch as the initial Monowire implementation baseline. Retain Zig labeled-switch dispatch as an optimization candidate. The October 2026 prototype results show small median advantages for labeled dispatch, but the evidence is insufficient to establish a reliable production benefit. Reevaluate after a functioning Monowire VM and representative VM-only benchmarks, considering execution time, code size, compiler output, and maintenance complexity.                                                                                                                                                    |
| D-053         | Decided; implementation pending (M2-P2)                                                                    | **Numeric type inventory and integer literal typing.** M2 supports signed `Int8`, `Int16`, `Int32`, and `Int` (signed 64-bit), plus IEEE 754 binary64 `Float`. Unsigned integer types are deferred to M3. Unsuffixed integer literals begin as exact compile-time constants, use suitable expected integer context, and default to `Int` when no suitable context exists. Expected-type propagation never silently converts already-typed variables.                                                                                                                                                                                           |
| D-054         | Decided; implementation pending (M2-P2)                                                                    | **Arithmetic result types and checked integer operations.** Same-type integer arithmetic preserves its operand type without automatic widening. Overflow is an error, diagnosed statically when provable and checked at runtime otherwise. `/` produces `Float`; `//` is floor division; `%` is floor-compatible remainder. Integer exponentiation requires same-type operands and a non-negative exponent, preserves the base type, defines `0 ** 0` as `1`, and checks overflow.                                                                                                                                                             |
| D-055         | Decided; implementation pending (M2-P2)                                                                    | **IEEE binary64 and floating-point conventions.** `Float` uses IEEE 754 binary64. Basic arithmetic preserves signed zero, infinities, and NaN; floating-point overflow produces signed infinity. Floating `//` and `%` use the specified floor-compatible convention for finite operands and a finite nonzero divisor; exceptional operands produce NaN. General-purpose `pow` behavior is defined through Eloque conformance cases and mainstream math-library conventions, not an overstated claim of complete IEEE 754 specification.                                                                                                       |
| D-056         | Decided; implementation pending (M2-P2)                                                                    | **Numeric conversions and restricted implicit safety.** Destination-type conversion syntax is used. Integer widening is safe when the full source range fits; narrowing is checked. Integer-to-Float conversion rounds to binary64. Mixed-width integer arithmetic and mixed integer/Float arithmetic require explicit conversions. Implicit conversion is restricted to initialization and assignment when value preservation is provable; the same policy applies to future function arguments and returns. Float-to-integer conversion is always explicit, rejects NaN/infinities, truncates toward zero, and checks the destination range. |
| D-057         | Decided for M3 compatibility; implementation pending                                                       | **Exact mixed numeric comparison contract.** Mixed-width signed integer comparisons are exact. Mixed integer/Float comparisons must not lose integer precision through premature conversion to binary64. NaN comparison behavior follows IEEE 754. Comparison and Boolean syntax remain M3 features.                                                                                                                                                                                                                                                                                                                                           |
| D-058         | Decided; implementation pending (M2-P2)                                                                    | **Immutable AST and compilation-scoped semantic result.** The AST remains unchanged by semantic analysis. A separate compact result retains explicit and inferred symbol types and expression types required by bytecode generation, then is released after that stage. Avoid duplicating AST or resolver information without demonstrated need.                                                                                                                                                                                                                                                                                               |
| D-059         | Decided; implementation pending (M2-P2)                                                                    | **Definite assignment and type-error recovery.** Definite assignment is separate from lexical name resolution. Reads of uninitialized bindings are blocking errors. An internal error type suppresses derivative diagnostics while independent expressions continue to be checked. The internal type is not source-declarable. Blocking semantic errors prevent bytecode generation.                                                                                                                                                                                                                                                           |
| D-060         | Decided; implementation pending (M2-P2)                                                                    | **Bounded local constant evaluation.** M2 uses incremental local constant propagation and evaluation, not arbitrary compile-time program execution. It propagates eligible pure known values, invalidates stale facts after mutation, and can establish new facts from known assignments. It does not execute functions, loops, arbitrary control flow, or I/O. Deterministic work, depth, and memory limits are required; their values must be benchmarked.                                                                                                                                                                                   |
| D-061         | Decided; implementation pending (M2-P2)                                                                    | **M2-P2 implementation sequence.** P2.1 numeric frontend prerequisites; P2.2 semantic foundations; P2.3 definite assignment and compatibility; P2.4 numeric semantics; P2.5 constant evaluator; P2.6 compiler integration; P2.7 validation and performance. This sequence must preserve future extensibility and the approved language semantics rather than optimize only for the initial implementation.                                                                                                                                                                                                                                     |

| D-062 | Decided; implementation pending (M2-P2) | **Explicit numeric conversion expression syntax.** Numeric conversions use `convert expression to Type`. The `convert` keyword is reserved and introduces a dedicated numeric conversion expression. The operand is a complete expression terminated by `to`; parentheses group complex or nested operands. The destination must be a supported numeric type. This syntax does not introduce general function calls, which retain the established `name with (arguments)` form. This decision supersedes the conversion syntax portion of D-056 and the corresponding examples in `LANGUAGE_SPEC.md`; all conversion semantics in D-056 remain unchanged. |

## M2-P2 semantic and numeric decisions — November 2026

**Status:** Approved design; implementation pending. These decisions refine and supersede conflicting provisional numeric wording in earlier entries and in the language specification. They do not claim that M2-P2 has been implemented.

### D-053 — Numeric type inventory and integer literal typing

M2 defines `Int8`, `Int16`, `Int32`, and `Int` as signed 8-, 16-, 32-, and 64-bit integer types respectively, plus IEEE 754 binary64 `Float`. Unsigned integer types are deferred to M3. `Int` remains the default integer type.

Unsuffixed integer literals begin as exact compile-time constants. A suitable expected integer type determines their type when the value fits; a typed peer operand supplies context to an unsuffixed literal; an enclosing expected type propagates through otherwise-untyped arithmetic; and literals without a suitable context default to `Int`. Already-typed variables are never silently converted by expected-type propagation. Values must fit exactly. The minimum signed `Int` literal must be accepted without first requiring its positive magnitude to fit in `Int`.

### D-054 — Arithmetic result types and checked integer operations

Same-type integer arithmetic preserves its operand type. No automatic widening or mixed-width promotion occurs. Integer overflow is an error, diagnosed at compile time when provable and checked at runtime otherwise. Runtime checks may be eliminated only when safety is proven.

`/` produces `Float`; `//` is floor division; `%` is floor-compatible remainder. Integer operations never silently wrap, saturate, or widen. Integer exponentiation requires same-type operands and a non-negative exponent, returns the base type, defines `0 ** 0` as `1`, and checks overflow.

The minimum signed integer divided by `-1` with `//` is overflow; `%` returns zero without computing an overflowing quotient. Explicit wrapping or saturating operations and the exact runtime error transport mechanism remain deferred.

### D-055 — IEEE binary64 and floating-point operator conventions

`Float` is IEEE 754 binary64. Basic arithmetic preserves signed zero, infinities, and NaN; floating-point overflow produces signed infinity. Floating division follows IEEE behavior.

Floating `//` and `%` use floor-compatible semantics for finite operands and finite nonzero divisors. NaN, infinite operands, or a zero divisor produce NaN for these operators. This is explicitly distinct from IEEE 754's nearest-quotient `remainder` operation.

Floating exponentiation follows IEEE-compatible binary64 behavior and mainstream math-library conventions. The conformance suite, rather than a claim that IEEE 754 fully specifies general-purpose `pow`, defines required observable edge cases.

### D-056 — Explicit conversions and restricted implicit safety

umeric conversion syntax uses the explicit form `convert expression to Type`, for example `convert value to Int16` and `convert count to Float`. The `convert` keyword introduces a dedicated numeric conversion expression; it does not invoke a user-defined function. Integer widening is safe when the complete source range fits; narrowing is checked and never silently wraps or saturates. Integer-to-Float conversion rounds to binary64. Mixed-width integer arithmetic and mixed integer/Float arithmetic require explicit conversions.

Implicit conversion is allowed only for initialization and assignment when the compiler proves value preservation from the complete source range or a particular known value. The same policy applies to function arguments and returns when those features arrive. It is not a general conversion permission for every expression context.

Float-to-integer conversion is always explicit. It rejects NaN and infinities, truncates finite values toward zero, then checks the truncated result against the destination range. Compile-time-provable invalid conversions are diagnosed statically; otherwise a runtime check is required.

### D-057 — Exact mixed numeric comparison contract for M3

Mixed-width signed integer comparisons are exact. Mixed integer/Float comparisons must not lose integer precision by first converting the integer to binary64. NaN comparison behavior follows IEEE 754. Comparison and Boolean syntax remain M3 features; recording their numeric contract in M2-P2 prevents a later incompatible promotion rule.

### D-058 — Immutable AST and compilation-scoped semantic result

The AST remains immutable. Semantic analysis produces a separate, compact result scoped to one compilation, retaining explicit and inferred symbol types and expression types required by bytecode generation. It remains available through bytecode generation and is then released. Temporary checker data may be discarded earlier. The representation must avoid duplicating information already available from the AST or resolver without demonstrated need.

### D-059 — Definite assignment and type-error recovery

Definite assignment is a distinct semantic responsibility from lexical name resolution. Reads of uninitialized bindings are blocking errors. Type checking uses an internal error type to suppress diagnostics that are merely consequences of an earlier error while continuing to check independent expressions. The internal error type is not part of the source language. Blocking semantic errors prevent bytecode generation.

### D-060 — Bounded local constant evaluation

M2 adopts incremental local constant propagation and evaluation, not arbitrary compile-time execution. It propagates eligible known values through pure expression chains, invalidates stale facts after mutation, and may establish a new fact when a mutation's assigned value is known. It does not execute functions, loops, arbitrary control flow, or I/O. It must preserve runtime numeric semantics.

Deterministic work, depth, and memory budgets are required; their numeric thresholds are to be selected from stress tests and compile-time/memory benchmarks. Budget exhaustion makes a value unknown and retains valid runtime computation; it does not turn a valid program into an error.

### D-061 — M2-P2 implementation sequence

The accepted work packages are:

1. **P2.1 — Numeric frontend prerequisites:** Float literal recognition/parsing and built-in numeric conversion expressions, without enabling general function calls.
2. **P2.2 — Semantic foundations:** Type representation, expression typing, symbol type information, and compilation-scoped semantic results.
3. **P2.3 — Definite assignment and compatibility:** Binding initialization state, assignment/initialization compatibility, and error recovery.
4. **P2.4 — Numeric semantics:** Literal context propagation, arithmetic typing, overflow, division, exponentiation, and conversions.
5. **P2.5 — Constant evaluator:** Bounded local propagation and evaluation with mutation invalidation.
6. **P2.6 — Compiler integration:** Semantic-result consumption and blocking-error eligibility for bytecode generation.
7. **P2.7 — Validation and performance:** Conformance, stress, runtime parity, compile-time, allocation, and memory benchmarks.

The work packages express dependencies and review boundaries. They do not justify implementing a shortcut that compromises future type extensibility, semantic clarity, or runtime compatibility.

## Superseded syntax and semantics

The following are **historical examples only**, not current lexer/parser fixtures.

| Earlier draft                                                 | Current direction                                                             |
| ------------------------------------------------------------- | ----------------------------------------------------------------------------- |
| `x = 2`, `x += 1`                                             | `change x to 2`; compound assignment not retained                             |
| `fix x at 1`                                                  | `let fixed x be 1` or `let frozen x be 1`                                     |
| `nothing`                                                     | `none`                                                                        |
| `name: int`, `name of type int`                               | `name as Int`                                                                 |
| Optional gradual type checker                                 | Mandatory static checker with inference                                       |
| `else if`                                                     | `otherwise if`                                                                |
| `each item in items`                                          | `for each item in items`                                                      |
| `while ready`, `until ready`                                  | `repeat while ready`, `repeat until ready`                                    |
| `call greet with "Alex"`                                      | `greet with ("Alex")`                                                         |
| `call greet`                                                  | `greet with ()`                                                               |
| Default/named argument `as`                                   | `be`                                                                          |
| `f"Hello, {name}"`                                            | `"Hello, {name}"` (interpolation planned)                                     |
| `bring maths`, `bring square as sq from maths`                | `import maths`, `import square named sq from maths`                           |
| `define Child from Parent`, `setup`, `create`, `my`, `parent` | Structured `define` types, implicit `self`, `Type with (...)`; no inheritance |
| `always` exception cleanup                                    | `finally`                                                                     |
| `items[0]`, `items at 0`                                      | `item at 0 in items`                                                          |
| `remove every x from items`                                   | `remove all x from items`                                                     |
| `==`, `!=`, `<`, `>=` as comparison alternatives              | Word-based comparison forms are the documented grammar                        |
| Tuple and set literals as core collection types               | Deferred pending separate design                                              |
| Uninitialized `let x`                                         | Initializers required in current design                                       |
| `outer`, `with x give ...`, legacy variadics                  | Deferred pending deliberate redesign                                          |

### Historical origin

Eloque grew out of the earlier BakeScript culinary-syntax concept. That transition is historical context, not a competing active grammar. Additional old mappings remain available in Git history.

## Decisions requiring refinement

These are future tasks, **not unfinished M1 requirements**:

- **Lexing and parsing:** Extend token coverage, grammar, source spans, and conformance fixtures for future features; specify comparison/Boolean precedence and associativity. Preserve existing reserved-keyword rules and arithmetic parser behavior.
- **Diagnostics and frontend performance:** Frontend recovery and M2-P1 lexical name resolution are implemented. Complete integration of recovery-enabled parsing as the default user-facing compiler path when the compiler driver is introduced. Preserve fail-fast parsing and resolution as a separate correctness and performance baseline. Maintain malformed-source recovery benchmarks independently. The M2 cross-language v2 comparison has been completed; future comparisons must continue to distinguish in-process Eloque processing, CPython frontend APIs, and native compiler subprocess execution.
- **Types and numbers:** M2-P2 has settled the initial signed integer inventory, context-directed integer literal typing, checked arithmetic, IEEE binary64 behavior, conversion rules, and bounded local constant evaluation. Implementation and conformance validation remain pending. Future work includes unsigned types, comparison implementation in M3, optional-value analysis, generic compatibility, and more advanced flow-sensitive range proofs.
- **Collections and text:** Define iteration and mutation details, including Unicode code-point versus grapheme-cluster indexing.
- **Language features:** Specify exception taxonomy and cleanup precedence, module resolution, structural-interface diagnostics, and pattern exhaustiveness.
- **Execution and tooling:** Define CLI behavior, sandbox budgets, and concurrency/ownership semantics as their implementation milestones approach.
- **VM architecture and implementation:** Monowire is the selected register-based bytecode VM, superseding the previously planned Go stack VM. Its overlapping caller-outgoing and callee-incoming register-window architecture is an approved design direction, not an open stack-versus-register decision. Zig is provisionally preferred as the runtime implementation language, with Rust retained as a fallback pending integration, portability, correctness, and maintenance validation. Finalize the production bytecode format, register-window invariants, runtime value representation, memory ownership, and compiler/runtime integration during M2. Use conventional switch dispatch as the initial baseline; reconsider labeled-switch dispatch and other optimizations only when representative production benchmarks justify them.
- **Frontend representation and optimization:** Preserve H2's arena-backed flat AST as the production baseline. Revisit arena capacity and growth policies only when representative profiling demonstrates a meaningful opportunity. Evaluate any proposed optimization against correctness, allocation overhead, retained memory, small-program performance, and implementation complexity. The review is tracked under TB-025 in the technical backlog.

## M2 architecture decisions — October 2026

**Status:** Approved architectural decisions. Frontend recovery and M2-P1 lexical name resolution have been implemented; mandatory static type checking, bytecode compilation, and Monowire execution remain pending.

The decisions below establish the architectural contract for Eloque's diagnostic migration and M2 name-resolution stage. They also preserve the original design rationale and implementation requirements. Some stages have since been implemented, while other requirements remain pending or require verification against the current source and tests. Changes to agreed language semantics must be documented explicitly.

### D-041 — Dedicated name-resolution stage

**Status:** Implemented (M2-P1) for the supported language subset.

Name resolution is a dedicated semantic-analysis pass performed after syntactic validation and before mandatory static type checking.

The resolver must not mutate the AST. It produces a separate resolution result containing symbol definitions and bindings between identifier references and declarations.

Variable declarations are visible from their declaration point onward. Future function and named-type declarations may support forward references through a declaration prepass within the appropriate scope.

The precise interaction between forward declarations, module imports, and visibility will be specified when those language features are introduced.

### D-042 — Lexical scopes and shadowing

**Status:** Implemented (M2-P1) for supported scope behavior; future nested-language constructs remain pending.

Scopes are lexically nested and parent-linked.

Each scope owns a mapping from locally declared names to symbol identities. Name lookup searches the current scope first, then successive enclosing scopes.

- Duplicate declarations within the same scope are errors.
- A declaration in a nested scope may shadow a visible enclosing declaration.
- Shadowing emits a nonfatal warning identifying the earlier declaration.
- The nearest visible lexical declaration takes precedence.
- Independent scopes may reuse identifiers.
- A duplicate declaration must not replace the original binding in the scope table.

This decision supersedes the no-shadowing restriction previously recorded in D-010 and Section 4 of the language specification.

### D-043 — Declaration and initializer resolution

**Status:** Name-resolution behavior implemented (M2-P1); definite-assignment checking pending.

Variable declarations are registered before their initializer expressions are resolved.

Therefore:

```elq
let x be x + 1
```

The initializer reference to `x` resolves to the newly introduced symbol.

Name resolution itself does not reject this as an undefined name. A separate definite-assignment analysis must diagnose the read of an uninitialized value.

The same rule applies when a nested declaration shadows an enclosing variable: references within the initializer resolve to the new declaration.

Name resolution and definite-assignment checking remain separate responsibilities.

### D-044 — Symbol representation and ownership

**Status:** Implemented (M2-P1); performance and representation refinements remain subject to validation.

Symbols are stored in a contiguous, append-only arena represented conceptually by `[]Symbol`.

Symbol identity uses a compact `uint32`-backed `SymbolID` with an explicit invalid sentinel.

Symbol identities:

- Are stable within one compilation unit.
- Do not represent runtime storage slots.
- Are not guaranteed to remain stable across recompilations.
- Are not globally unique across modules.

Scope lookup initially uses local `map[string]SymbolID` tables linked through parent scopes. Inherited environments are not copied.

Identifier names may reference substrings of retained source text, avoiding unnecessary allocations. The resolution result must retain the necessary source/program lifetime.

Global string interning is deferred until representative benchmarks justify it.

### D-045 — Reference binding representation

**Status:** Implemented (M2-P1); internal representation remains subject to benchmark-driven review.

Reference bindings initially use a dense `[]SymbolID` indexed by AST `NodeID`.

Unbound entries use an explicit invalid sentinel.

The storage representation is hidden behind accessors, conceptually:

- `Symbol(SymbolID)`
- `Reference(ast.NodeID)`
- `Declaration(ast.NodeID)`

The implementation must accommodate the existing H2 AST:

- Identifier expressions are represented by `NodeIdentifier`.
- Declaration names are source spans in `NodeLetStatement` binding payloads.
- Reassignment targets are source spans in `NodeChangeStatement` binding payloads.

Consequently, a `NodeChangeStatement` may serve as the lookup key for its assignment-target binding, while a `NodeLetStatement` identifies a declaration.

The compiler should consume resolved symbol identities rather than repeating name lookups.

Dense versus sparse binding storage remains an explicitly tracked benchmark decision.

### D-046 — Shared diagnostic representation

**Status:** Approved diagnostic contract; implementation details require verification against the current compiler stages.

Compiler diagnostics use a common structured representation containing:

- Stable machine-readable diagnostic code.
- Severity: error, warning, or informational.
- Human-readable message.
- Primary source location.
- Optional related locations with explanatory messages.
- Optional notes.

Locations use half-open source byte spans and compilation-local file identities.

Line and column information is derived from source text for presentation rather than duplicated throughout the compiler.

A minimal source registry maps file identities to source information. File identities are local to a compilation snapshot and are not permanent identifiers across recompilations.

Parser-specific expected/actual-token information is retained as structured stage-specific metadata. The shared diagnostic package must not depend on lexer-specific token types.

Diagnostic codes are grouped by stage, with stable named identifiers such as `LEX_UNEXPECTED_CHARACTER`, `PARSE_EXPECTED_TOKEN`, `RESOLVE_UNDEFINED_NAME`, and `TYPE_MISMATCH`.

Diagnostic messages may evolve without changing the meaning of established codes.

### D-047 — Diagnostic collection and ordering

**Status:** Approved compiler-wide policy; complete cross-stage driver integration remains pending.

Each compiler stage owns its diagnostic collection.

There is no global mutable diagnostic collector.

The compilation driver coordinates a shared error budget and merges stage diagnostics deterministically.

A configurable overall diagnostic limit is required. A default of 100 errors is an initial candidate, subject to validation rather than a permanent language-level guarantee.

Diagnostics should be ordered by source location, with stable tie-breaking for coincident locations.

Independent errors should be collected where safe. Diagnostics caused solely by an earlier error should be suppressed where practical.

Warnings are nonblocking by default.

Successful compiler paths should avoid unnecessary diagnostic allocations. Optional diagnostic slices remain nil when unused.

### D-048 — Recovery and compilation eligibility

**Status:** Frontend recovery implemented (post-H2); end-to-end compilation eligibility enforcement remains pending.

**Default frontend policy (approved October 2026):** User-facing compilation, source checking, and editor tooling use recovery-enabled parsing by default. Fail-fast parsing remains available as an explicit alternative. A `ParseComplete` result permits progression to semantic validation but does not itself establish semantic correctness. `ParseRecovered` and `ParseAborted` results are ineligible for normal bytecode generation or execution. Semantic errors also block executable compilation.

Diagnostic collection and error recovery are separate concerns.

Each stage determines how to recover safely while using the common diagnostic representation.

Stage results explicitly distinguish:

- **Complete:** Processing finished without blocking errors.
- **Recovered:** Processing continued after blocking errors; output is not eligible for normal executable compilation.
- **Aborted:** Processing could not safely continue.

A warning-only stage result remains complete.

A recovered AST may support diagnostic reporting or future editor tooling, but must not be treated as a fully valid executable program.

Any blocking compilation error prevents normal bytecode generation and execution.

Fatal structural failures may terminate an affected region or stage when continued recovery would be unsafe.

### D-049 — Frontend migration prerequisite

**Status:** Completed (post-H2); retained as the historical prerequisite for M2-P1.

Before M2 name resolution begins, the existing H2 lexer, streaming token source, and arena-backed parser must be adapted to controlled multi-error reporting.

The migration must preserve:

- Direct pull-based lexical streaming.
- Arena-backed flat AST representation.
- Existing correct syntax behavior.
- Source-location accuracy.
- Safe token and parser progress.
- The separation between recovered and valid programs.

The lexer must distinguish independently recoverable malformed input from failures that compromise structural interpretation.

The parser must synchronize at contextually safe boundaries and avoid infinite recovery loops.

Partially constructed arena nodes may remain allocated internally, but abandoned nodes must not be exposed as valid statements.

Correctness, diagnostic quality, allocations, and representative performance must be evaluated against the existing M1/H2 baseline.

This replaces the earlier decision to defer multi-error recovery indefinitely.

### D-050 — Implementation sequence

**Status:** Decided.

The approved sequence is:

1. Finalize and document shared diagnostic architecture.
2. Migrate frontend lexical and syntax diagnostics to controlled recovery.
3. Validate frontend regression tests, diagnostic behavior, and benchmarks.
4. Implement dedicated M2 name resolution.
5. Implement mandatory static type checking and definite-assignment checks.
6. Implement bytecode compilation.
7. Implement the baseline Monowire register-based VM.
8. Establish execution benchmarks before reconsidering major runtime architecture choices.

**Progress as of November 2026:**

- Steps 1–3: Frontend diagnostic architecture, controlled recovery, and regression/performance validation completed.
- Step 4: Dedicated M2-P1 lexical name resolution implemented and benchmarked.
- Step 5: Mandatory static type checking and definite-assignment analysis pending.
- Step 6: Production Monowire bytecode compilation pending.
- Step 7: Production Monowire VM execution pending.
- Step 8: Executable runtime benchmarking pending. Historical frontend and synthetic VM benchmarks remain separate evidence.

No stage should be considered implemented solely because its design is recorded here.

See `ARCHITECTURE.md` for the resolver algorithm, test matrix, and acceptance criteria, and `TECHNICAL_BACKLOG.md` for deferred reevaluation triggers.

## M2-P1 implementation and benchmark record — October 2026

**Status:** Lexical name resolution implemented; benchmark comparison completed.

M2-P1 established the first dedicated semantic-analysis stage after parsing.

The implementation resolves identifier references and assignment targets, records declaration-symbol bindings, and reports name-resolution errors for the supported grammar.

The M2 cross-language v2 benchmark suite evaluates declarations, references, mixed programs, sparse undefined references, and undefined-reference workloads at multiple sizes.

For the 3,000-unit mixed workload, recorded median latencies include:

| Implementation or operation             | Median latency |
| --------------------------------------- | -------------: |
| Eloque recovery-capable parse + resolve |    2.059516 ms |
| Eloque fail-fast parse + resolve        |    1.993480 ms |
| CPython `symtable`                      |   17.534834 ms |
| CPython `ast.parse`                     |   25.515754 ms |

These measurements represent different operations and implementation boundaries. They do not establish equivalent compiler functionality or runtime performance superiority.

The benchmark results are historical measurements, not language-semantic decisions. They should remain reproducible through the corresponding benchmark scripts and recorded result files.

**Next implementation phase:** Mandatory static type checking and definite-assignment analysis.

## Change process

When changing a language-level decision:

1. Update [LANGUAGE_SPEC.md](LANGUAGE_SPEC.md) first.
2. Amend the corresponding decision here and explicitly identify superseded behavior.
3. Update affected examples, architecture, roadmap, technical backlog, and conformance tests.
4. Distinguish agreed design from implemented, tested behavior; record benchmark evidence and reevaluation triggers for performance-sensitive choices.

**A decision's presence in this register does not mean its feature is executable.**

## Monowire architecture decision — October 23, 2026

**Status:** Decided architecture; production implementation pending.

Eloque will use **Monowire**, a register-based bytecode virtual machine, beginning with M2 — First Execution.

This decision supersedes the earlier plan to implement a Go-based stack VM first and reconsider register-based execution only after a working stack interpreter.

### Rationale

The register-based architecture was selected to support explicit operand locations, efficient register-oriented bytecode execution, and a runtime design with well-defined call frames and overlapping register windows.

An independent implementation-language experiment compared Go, Rust, and two Zig dispatch variants across four synthetic bytecode workloads.

Zig recorded the lowest median subprocess wall-clock time in every workload. This supports its provisional selection as Monowire's implementation language, but the experiment was not a complete or fully controlled production-runtime comparison.

The architecture decision does not depend on Zig remaining the implementation language. Monowire's register-based execution model can be retained if implementation constraints eventually favor Rust or another suitable systems language.

### Implementation boundary

The production pipeline is planned as:

```text
Eloque source (.elq)
    |
    v
Go lexer and parser
    |
    v
Arena-backed flat AST
    |
    v
Go name resolver
    |
    v
Go static type checker
    |
    v
Monowire bytecode compiler
    |
    v
Monowire register-based VM
    |
    v
Program execution
```

The compiler-to-runtime interface, bytecode representation, runtime value layout, allocation and ownership rules, and error-propagation mechanisms require further specification.

Go–Zig integration remains a validation task rather than an implemented capability.

### Milestone impact

Monowire does not change the functional boundary of M2.

M2 remains responsible for name resolution, mandatory static checking, bytecode generation, runtime execution, and end-to-end validation of Eloque's initial executable subset.

M3 continues to introduce control flow, while M4 introduces functions at the language level. Monowire may implement internal call-frame mechanisms earlier without implying that Eloque function syntax is already supported.

### Superseded decision

**Previous D-022:** Implement a Go stack VM first; reconsider a register VM after obtaining a working baseline and comparative measurements.

**Replacement:** Implement Monowire as the initial production VM, with register-based execution selected from the outset.

The previous decision is retained here for historical traceability, not as an active implementation requirement.

### Remaining validation

Before considering the production runtime architecture operational:

- Validate the Go–Zig integration boundary and error propagation.
- Specify the bytecode format, register allocation, and call-frame invariants.
- Establish runtime value representation and memory ownership.
- Implement instruction-level correctness tests and resource-boundary safeguards.
- Validate executable Eloque programs through the mandatory checker.
- Establish representative VM-only and end-to-end performance baselines.

The corresponding unresolved engineering work is tracked in `TECHNICAL_BACKLOG.md`.

**Evidence:** `benchmarks/monowire/README.md`, `benchmarks/monowire/SPEC.md`, and the recorded October 23, 2026 benchmark results.
