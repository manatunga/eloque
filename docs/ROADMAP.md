# Eloque Roadmap

**Status:** Active implementation roadmap — M0 completed; M1 (Frontend foundations) in progress. Milestones and versions are planning aids, not release commitments. Eloque, `.elq`, and the CLI name remain provisional.

## Delivery principles

- Build the smallest **correct, statically checked, executable** subset before broadening syntax.
- Keep the full language design distinct from features implemented in a particular release.
- Require type checking on the normal compilation path from the first executable milestone. A checker can be incremental in *coverage* without being optional.
- Make each milestone demonstrable with fixtures, tests, and readable diagnostics.
- Avoid committing to dates or performance outcomes before measurement.
- The user writes production code; mentoring and reviews can guide Go, architecture, and testing.

## Internal implementation milestones

These are development checkpoints, **not** promises that each becomes a public release.

| Milestone | Scope | Demonstrable outcome |
|---|---|---|
| **M0 — Foundation** | Go module, repository structure as needed, CLI skeleton, tests, initial conformance fixtures | `go test ./...` succeeds; a minimal CLI starts; specification and decisions are aligned |
| **M1 — Frontend foundations** | Lexer, four-space indentation, source spans, parser, AST, syntax diagnostics | Parse a small valid `.elq` program; reject malformed input with locations |
| **M2 — First execution** | Name resolution, **mandatory basic type checking**, bytecode compiler, stack VM, literals, `let`, `change`, arithmetic, `say` | Run a typed hello-world and arithmetic program; reject basic type errors before execution |
| **M3 — Control flow** | Comparisons, Boolean logic, `if`/`otherwise if`/`otherwise`, `repeat`, `for each` after minimal list/iteration support, `skip`, `stop` | Execute conditional and looping programs with verified scope rules |
| **M4 — Functions** | Typed parameters, `giving`, calls with mandatory `with (...)`, return checking, recursion, default/named arguments in stages | Execute statically checked functions, including valid and invalid call examples |
| **M5 — Values and collections** | `Int` overflow, `Float` rules, optional `T?`, Unicode strings, lists/maps, safe access, slicing, value semantics, `fixed`/`frozen` | Verify value independence, mutation rules, numeric failures and safe lookup |
| **M6 — Structured features** | Custom `define` types, methods, structural interfaces, pattern matching | Construct values, call methods, check interface conformance, evaluate match cases |
| **M7 — Integration** | `import`, module initialization, visibility, typed exceptions, `try`/`catch`/`finally` | Run multi-file programs; report module and exception failures predictably |
| **M8 — Developer experience** | `eloque check`, formatter/REPL exploration, structured diagnostics, packaging and documentation | Usable tooling and a conformance corpus for the implemented subset |
| **M9 — Concurrency and agent safeguards (later)** | Design ownership/references, safe task execution, configurable sandbox limits | Prototype and validate semantics before committing public syntax |

The order may change where dependencies demand it. In particular, implement enough list/iterator support before `for each`, and add type rules **with** each runtime feature rather than deferring a large checker rewrite.

## Current implementation progress

### M0 — Foundation: Completed

- Go module and initial repository structure established.
- Minimal CLI skeleton implemented and tested.
- Initial architecture, language specification, roadmap, and design decisions established.
- Go test infrastructure operational.

### M1 — Frontend foundations: In progress

**Lexer checkpoint: Completed**

- Implemented and tested lexical support for the selected initial M1 grammar.
- Implemented ASCII identifiers, reserved keywords, integer and string literals, arithmetic operators, parentheses, colons, newlines, and indentation tokens.
- Implemented source spans, lexical diagnostics, comments, indentation validation, and parenthesized line continuation.
- Completed lexer correctness, race, profiling, allocation, and scaling reviews.
- Retained reproducible benchmarks for future regression comparisons.

**M1-P1 — AST foundations and parser infrastructure: Completed**

- Extracted shared source-position and span types into `internal/source`.
- Implemented AST interfaces and initial expression, statement, and program nodes.
- Added AST structural tests for interfaces, spans, ordering, and child references.
- Implemented parser construction and token-navigation helpers.
- Implemented structured syntax diagnostics and required-token matching.
- Established fail-fast syntax-error handling; multi-error recovery remains deferred.
- Added dedicated lexer token kinds for `change`, `to`, `with`, and `giving`.

**M1-P2 — Pratt expression parsing: Completed**

- Implemented Pratt parsing for integer literals, string literals, identifiers, and parenthesized expressions.
- Implemented unary numeric plus and minus and binary arithmetic operators `+`, `-`, `*`, `/`, `//`, `%`, and `**`.
- Verified operator precedence, left- and right-associativity, and exponentiation's interaction with unary arithmetic.
- Preserved semantic AST spans separately from complete syntactic grouping extents.
- Implemented string-literal escape decoding using lexically validated token contents.
- Added expression-boundary, malformed-input diagnostic, precedence, associativity, and conformance tests.
- Validated parsing of 1,000 nested parentheses and 1,000 chained exponentiation operators.
- Established initial expression-parsing benchmarks for arithmetic, precedence, operator chains, and nested grouping.
- Observed approximately linear execution-time and allocation scaling; no immediate production-code optimisation required.

**M1-P3 — Statements and program parsing: Completed**

- Implemented `let`, `change`, and `say` statement parsing, including delegation to the existing Pratt expression parser.
- Implemented statement dispatch based on the initiating keyword.
- Implemented complete-program parsing with ordered statements, newline separation, blank-line handling, and EOF termination.
- Preserved statement and program source spans, including complete syntactic expression extents.
- Added positive and negative tests for statement syntax, program structure, expression integration, and source locations.
- Verified structured, fail-fast diagnostics for malformed statements, invalid statement starters, missing separators, and nested expression errors.
- Completed implementation and architecture review without identifying blocking issues.
- Completed performance review using 3-, 300-, and 3,000-statement workloads, observing approximately linear scaling and retaining benchmark results for future comparisons.
- No production-code optimisation required at the current checkpoint.

**Post-M1-P3 — Canonical frontend optimization: Completed**

- Consolidated and optimized the existing lexer and parser implementations before M1-P4.
- Established updated lexer, parser, and complete frontend performance baselines.
- Preserved historical benchmark measurements for regression comparisons.
- Added reproducible benchmark documentation and a preliminary comparison with CPython's frontend.
- Merged the canonical frontend optimization into `main`.

**Current checkpoint:** M1-P4 — Syntax diagnostics and conformance. M1-P5 remains the subsequent integration and milestone-closure checkpoint.

**Remaining M1 checkpoints**

- **M1-P4 — Syntax diagnostics and conformance:** Verify malformed syntax, source locations, and positive/negative grammar fixtures.
- **M1-P5 — Frontend integration and milestone closure:** Complete lexer–parser integration, end-to-end tests, benchmarks, documentation, and release audit.

M1 remains in progress. The canonical lexer and parser support the implemented arithmetic, literal, binding, assignment, and output syntax, with complete-program AST construction and structured fail-fast diagnostics. Integrated frontend benchmarks exercise lexing through AST construction, but formal end-to-end conformance and integration validation remain scheduled for M1-P4 and M1-P5. Eloque cannot yet execute source programs.

## Public release planning

A release version describes what actually works, not everything present in the language specification.

| Proposed version | Minimum release theme |
|---|---|
| `0.1.0` | Frontend foundations: tokens, AST, parsing, syntax diagnostics, minimal CLI |
| `0.2.0` | First executable **statically checked** subset with a bytecode VM |
| `0.3.x` | Control flow and expanded expression typing |
| `0.4.x` | Functions, typed calls and return validation |
| `0.5.x` | Collections, optional values and value-semantics foundations |
| `0.6.x` | Custom types, methods and interfaces |
| `0.7.x` | Modules and exception handling |
| `0.8–0.9.x` | Tooling, conformance, stability and further feature breadth |
| `1.0.0` | Stable, documented, well-tested language contract **when ready** |

These versions are illustrative. Internal milestones can be combined, split, or shipped under different numbers. Native concurrency need not be a 1.0 prerequisite; it should not be rushed into the initial VM.

## Go readiness and learning strategy

The implementation uses Go. Prior Python experience transfers conceptually, but Go-specific behavior deserves explicit study.

Learn concepts **just before using them**:
- Module/package layout and `go mod`, `go test`.
- Structs, methods, slices, maps, and interfaces.
- Pointers and value-versus-reference behavior.
- Error returns and error wrapping.
- Table-driven tests.
- Rune/UTF-8 handling when implementing strings.
- Component-level profiling and benchmarks as soon as meaningful implementations exist; end-to-end interpreter benchmarks after the first working VM.

Do not require a lengthy Go course before M0. Introduce small Go exercises alongside real compiler tasks, and favor guided implementation over copying a completed compiler.

## Definition of done for early milestones

**M0**
- Existing design documents are consistent.
- Go module path matches the actual repository.
- Minimal CLI can be built and tested.
- `go test ./...` succeeds.
- At least one `.elq` example is explicitly marked as illustrative until supported.

**M1**
- Indentation and source spans have tests.
- The lexer handles keywords as precise tokens rather than free-form English.
- Parser produces stable AST shapes for the selected subset.
- Invalid syntax reports useful locations.
- The grammar covered by the implementation has conformance examples.

**M2**
- Resolver rejects unknown names and prohibited shadowing.
- Checker rejects type-invalid programs **before** VM execution.
- Bytecode and VM run representative arithmetic, binding and `say` examples.
- Checked failures are tested for supported operations.
- End-to-end tests distinguish syntax, type and runtime errors.

## Review gates and deferred work

Keep the [technical backlog](TECHNICAL_BACKLOG.md) active:

- **Before the first VM:** choose a practical value representation; test it against `Int`, `Float`, `Bool`, `String`, and eventual collections.
- **After the first working VM:** establish a reproducible performance baseline and review the stack-versus-register question based on evidence, not preference.
- **Before mutable collections:** validate deep immutability and logical value independence.
- **Before modules/exceptions:** settle initialization order, typed error matching, and cleanup semantics.
- **Before concurrency:** settle explicit references, ownership, cancellation, and safe shared state.
- **Before agent sandboxing:** define measurable instruction/time/memory controls and realistic guarantees.
- **Before 1.0:** audit documented semantics, test coverage, diagnostics, packaging, and compatibility.

JIT, AOT, WebAssembly, aggressive specialization, and CPython parity remain research/performance ambitions, not prerequisites for the initial language.

## Performance measurement

Once there is a working VM:

1. Establish correctness tests for each benchmark.
2. Pin the Go, Eloque, and CPython versions and machine environment.
3. Compare equivalent algorithms for arithmetic, loops, calls, strings, collections, and objects as features become available.
4. Report wall-clock time, variance, memory usage, and warmup methodology.
5. Profile before optimizing; rerun correctness and performance regression suites.

**CPython parity is a long-term aspiration, not a forecast.** A stack-based VM remains the initial baseline until evidence supports a different architecture.

## Versioning and compatibility

Eloque intends to use Semantic Versioning as a guideline:
- Before `1.0`, breaking changes are possible and should be documented.
- After `1.0`, breaking public language behavior warrants a major version.
- Internal bytecode format is not part of the public compatibility contract unless standardized later.
- Adding an optional backend without changing public behavior need not trigger a major release.

## Immediate next steps

- [x] Establish the Go module and repository structure.
- [x] Build and test the minimal CLI skeleton.
- [x] Implement and test the initial lexer token categories.
- [x] Implement indentation handling and lexical diagnostics for the supported subset.
- [x] Establish lexer profiling and scaling benchmarks.
- [x] Complete lexer functionality for the selected initial M1 grammar.
- [x] Implement initial AST nodes and parser navigation infrastructure (M1-P1).
- [x] Establish the initial structured, fail-fast parser diagnostic contract.
- [x] Implement expression parsing with the agreed precedence and associativity rules (M1-P2).
- [x] Implement statement and complete-program parsing (M1-P3).
- [ ] Complete syntax diagnostics and frontend conformance fixtures (M1-P4).
- [ ] Integrate and validate the complete frontend (M1-P5).
- [ ] Expand grammar and static-checking rules alongside subsequent executable features.
