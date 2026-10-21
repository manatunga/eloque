# Eloque Roadmap

**Status:** Active implementation roadmap — M0 and M1 completed; post-M1 H2 frontend optimization and package reorganization completed. M2 (First execution) is next. Milestones and versions are planning aids, not release commitments. Eloque, `.elq`, and the CLI name remain provisional.

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

### M1 — Frontend foundations: Completed

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

**M1-P4 — Syntax diagnostics and conformance: Completed**

- Added 29 source-based conformance fixtures: 9 valid programs, 8 malformed programs, 8 diagnostic edge cases, and 4 lexer/parser boundary cases.
- Verified supported statement forms, arithmetic expressions, statement boundaries, source spans, EOF diagnostics, multiline error locations, nested expression diagnostic propagation, and deterministic fail-fast behavior.
- Confirmed that malformed programs do not return successful partial ASTs.
- Passed `go test ./...`, `go test -race ./...`, and `go vet ./...`.
- Preserved the canonical frontend implementation without production-code changes.

**M1-P5-A — Lexer-to-parser token delivery architecture: Completed**

- Evaluated eager token buffering, direct pull-based streaming, and bounded two-token lookahead.
- Completed three performance experiments covering token delivery, integrated lexer/parser/AST processing, and lookahead overhead.
- Selected direct streaming as the preferred production token-delivery architecture.
- Verified AST equivalence across representative frontend workloads.
- Passed targeted diagnostic regression tests, including source-position and lexical/syntax error interaction cases.
- Recorded that the integrated streaming frontend was faster in 19 of 20 workload-size comparisons and allocated fewer bytes in all 20.
- Deferred bounded lookahead until a concrete grammar requirement justifies it.
- Retained eager buffering as a comparison and regression-testing implementation.
- Production integration was subsequently completed in M1-P5-B,
followed by cross-language benchmark validation in M1-P5-C.

**M1-P5-B — Production frontend integration: Completed**

- Established `Parse(sourceText)` as the canonical frontend entry point.
- Integrated direct pull-based token streaming through `tokenSource`.
- Removed the parser's redundant independent token cursor.
- Retained buffered parsing for regression testing and benchmarks.
- Verified AST construction, lexical-error propagation, source
  positions, EOF handling, and syntax diagnostics.
- Passed integration tests, race detection, and static analysis.

**M1-P5-C — Cross-language frontend benchmark validation: Completed**

- Benchmarked Eloque's production streaming frontend against Go's
  `go/parser.ParseFile` and CPython 3.15.0rc2 `ast.parse()`.
- Evaluated five language-native workloads across four input sizes:
  10, 100, 1,000, and 10,000.
- Added controlled arithmetic benchmarks to reduce differences in
  expression complexity across the three languages.
- Observed the lowest frontend latency for Eloque in all measured
  workload-size comparisons.
- Recorded limitations involving grammar coverage, AST structure,
  source sizes, and benchmark methodology.
- Preserved benchmark results for future performance comparisons.

**Milestone 1 — Lexer, Parser, and Frontend: Completed**

M1-P1 through M1-P5 have been completed, establishing Eloque's
foundational source-processing frontend.

The canonical `Parse(sourceText)` entry point integrates incremental
lexical analysis, direct pull-based token streaming, parsing, and
complete-program AST construction.

The frontend supports the implemented arithmetic, literal, binding,
assignment, and output syntax, with source-aware diagnostics and
fail-fast error handling.

M1-P5 validated the production integration through end-to-end tests,
race detection, static analysis, buffered-versus-streaming performance
comparisons, and cross-language frontend benchmarks against Go and
CPython 3.15.

### Post-M1 — H2 streaming frontend and arena-backed AST: Completed

After completing M1, Eloque underwent a second frontend optimization phase, designated **H2**.

The objective was to reduce parser and AST allocation overhead without changing the accepted M1 grammar or observable diagnostic behavior.

**Implementation completed**

- Introduced an arena-backed flat AST representation using `FlatNode` and `FlatProgram`.
- Updated the production parser to construct the flat AST directly.
- Preserved direct pull-based lexer token streaming.
- Retained the original M1 parser and interface-based AST as legacy reference implementations.
- Added arena invariant tests, AST tests, differential parser tests, and regression coverage.
- Reorganized the frontend packages into `internal/frontend/ast`, `internal/frontend/lexer`, and `internal/frontend/parser`.
- Kept shared source-location utilities in `internal/source`.
- Preserved `Parse(sourceText)` as the canonical production entry point.

**Performance validation completed**

H2 was evaluated against the retained M1 implementation using five representative workload categories: declarations, expressions, strings, comments, and mixed programs.

The official comparison used workload sizes of 10, 100, and 1,000 units, with ten benchmark runs.

At 1,000 units:

| Workload | M1 median | H2 median | Speedup |
|---|---:|---:|---:|
| Declarations | 238.1 µs | 209.0 µs | 1.14× |
| Expressions | 1,181.5 µs | 667.4 µs | 1.77× |
| Strings | 400.1 µs | 394.2 µs | 1.01× |
| Comments | 407.8 µs | 346.4 µs | 1.18× |
| Mixed | 1,683.8 µs | 1,191.1 µs | 1.41× |

H2 reduced allocation counts across these workloads. String-heavy parsing showed little latency improvement, while expression-heavy parsing showed the largest improvement.

A separate controlled arithmetic benchmark compared H2 with the Go compiler syntax parser, Go's standard-library parser, and CPython 3.14.4.

At 10,000 units, the recorded median frontend latencies were:

| Implementation | Median latency |
|---|---:|
| Eloque H2 | 7.16 ms |
| Go compiler syntax parser | 14.12 ms |
| Go `go/parser` | 23.92 ms |
| CPython 3.14.4 `ast.parse` | 147.86 ms |

These measurements concern synthetic frontend workloads, not program execution. Differences in language grammar, source wrappers, AST representations, and measurement harnesses limit cross-language comparability.

Official results are archived in `benchmarks/results/official/`.

**Validation**

The H2 implementation includes correctness, differential, arena-invariant, and benchmark coverage. The frontend package migration subsequently passed `go test ./...` and `go vet ./...`.

**Deferred work**

Additional arena-capacity tuning remains deferred pending representative profiling evidence. The existing flat AST type names are retained for now. Further optimization must be justified by measurable benefits without unacceptable memory or correctness regressions.

**Outcome:** H2 is the canonical production frontend. The legacy M1 implementation remains available for regression and comparative testing.

### Post-M1 — Structured diagnostics and multi-error recovery: In progress

This checkpoint transitions Eloque's frontend from the M1 fail-fast diagnostic model to bounded multi-error recovery before beginning M2 execution development.

**Shared diagnostic foundation: Completed**

- Established shared diagnostic codes, severity, file identities, half-open byte spans, related locations, notes, and token-expectation metadata.
- Migrated the lexer and H2 parser to native structured diagnostics.
- Updated frontend, lexer, and parser diagnostic tests.
- Retired the legacy M1 pointer-based parser.
- Verified the migration with `go test ./...` and `go vet ./...`.

**Multi-error recovery: In progress**

- [ ] Finalize diagnostic accumulation, ownership, and recovery-session contracts.
- [ ] Implement explicit lexical recovery with guaranteed forward progress.
- [ ] Implement parser synchronization and suppression of cascading diagnostics.
- [ ] Introduce configurable error budgets and explicit parse outcomes.
- [ ] Preserve valid AST structures where recovery is safe.
- [ ] Verify multi-error behavior through conformance, robustness, and regression tests.
- [ ] Benchmark successful parsing and recovery workloads against the retained frontend baseline.
- [ ] Update architecture, language specification, and diagnostic documentation to reflect verified behavior.

**Completion criteria:** Eloque reports multiple independent lexical and syntax errors from a single source file, terminates safely on malformed input, preserves valid AST structures where appropriate, and maintains acceptable successful-path performance without abandoning direct token streaming.

### Transition to M2 — First execution: Not started

The next development phase will introduce the smallest statically checked executable subset.

M2 will build upon the H2 frontend and add:

1. Name resolution and scope validation.
2. Mandatory basic static type checking.
3. Bytecode compilation.
4. A sequential stack-based virtual machine.
5. End-to-end execution and diagnostic tests.

Before implementing VM values, the runtime value representation must be reviewed.

The initial executable subset will focus on literals, arithmetic, `let`, `change`, and `say`. Expanded control flow, functions, collections, and concurrency remain assigned to later milestones.

**M2 completion criterion:** Eloque executes representative valid programs through the bytecode VM and rejects supported type-invalid programs before execution.

**Next milestone:** M2 — Refer to the milestone definition below.

Eloque can parse supported source programs into an AST but cannot yet
execute them.

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

---

## Pre-M2 prerequisite: structured diagnostics and frontend recovery

**Status:** Approved; not yet implemented.

Before implementing the M2 name resolver, Eloque will migrate its existing H2 frontend from first-error termination to controlled, structured multi-error reporting.

The work includes:

1. Introduce the shared diagnostic representation and minimal compilation-local source registry.
2. Define stage-specific result and recovery contracts.
3. Adapt the streaming lexer and token source to safe lexical recovery.
4. Adapt the arena-backed parser to statement-level recovery and guaranteed progress.
5. Enforce compilation eligibility so recovered or malformed programs cannot execute.
6. Add diagnostic and recovery regression tests.
7. Benchmark the migration against the existing M1/H2 frontend baseline.
8. Begin M2 name resolution only after the migration passes its acceptance criteria.

### M2 implementation order

After frontend recovery is validated:

1. Dedicated lexical name resolver.
2. Mandatory static type checker, including definite-assignment analysis.
3. Bytecode compiler.
4. Baseline stack-based virtual machine.
5. First executable Eloque programs.
6. Runtime correctness and performance benchmarks.

The M2 design is documented in `DESIGN_DECISIONS.md` and `ARCHITECTURE.md`.

The detailed resolver acceptance matrix and deferred optimization triggers are preserved in `ARCHITECTURE.md` and `TECHNICAL_BACKLOG.md`.

This roadmap addition does not mark the diagnostic migration or any M2 execution stage as implemented.

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
- [x] Complete syntax diagnostics and frontend conformance fixtures (M1-P4).
- [x] Integrate and validate the complete streaming frontend (M1-P5).
- [x] Complete H2 arena-backed flat AST and streaming parser optimization.
- [x] Validate H2 against the legacy M1 frontend using differential tests and controlled benchmarks.
- [x] Reorganize the frontend under `internal/frontend/`.
- [ ] Synchronize documentation with the completed H2 implementation.
- [ ] Review the M2 runtime value representation and compiler-stage interfaces.
- [ ] Implement name resolution and mandatory basic static type checking.
- [ ] Implement bytecode compilation and the initial stack-based VM.
- [ ] Validate the first executable Eloque subset end to end.
- [ ] Expand grammar and static-checking rules alongside subsequent executable features.
