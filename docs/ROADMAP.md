# Eloque Roadmap

**Status:** Active implementation roadmap — M0 and M1 completed; post-M1 H2 frontend optimization, package reorganization, and pre-M2 structured diagnostic recovery implemented. The pre-M2 Monowire implementation-language experiment is complete. Monowire has been selected as Eloque's register-based production VM, replacing the previous stack-VM plan. Zig is provisionally preferred as its implementation language, subject to integration validation. M2 (First execution) is next. Milestones and versions are planning aids, not release commitments.

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
| **M2 — First execution** | Name resolution, **mandatory basic static type checking**, Monowire register-based bytecode compiler and VM, literals, `let`, `change`, arithmetic, `say` | Execute statically checked hello-world and arithmetic programs through Monowire; reject invalid programs before execution; establish initial runtime correctness and performance baselines |
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

- [x] Introduce structured diagnostic collection and recovery-capable frontend parsing.
- [x] Implement safe lexical recovery and parser synchronization with guaranteed progress.
- [x] Preserve valid AST structures where appropriate while distinguishing recovered programs from fully valid programs.
- [x] Verify multi-error behavior through conformance, robustness, and regression tests.
- [x] Benchmark successful parsing and recovery workloads against the retained frontend baseline.
- [x] Update architecture, language specification, and diagnostic documentation to reflect verified behavior.

**Completion criteria:** Eloque reports multiple independent lexical and syntax errors from a single source file, terminates safely on malformed input, preserves valid AST structures where appropriate, and maintains acceptable successful-path performance without abandoning direct token streaming.

### Pre-M2 — Monowire implementation-language experiment: Completed

**Completed:** October 23, 2026.

Before beginning production runtime implementation, an independent experimental benchmark suite was developed to evaluate implementation-language options for Monowire.

The experiment compared:

- Go.
- Rust.
- Zig with conventional switch dispatch.
- Zig with labeled-switch dispatch.

Four synthetic workloads were implemented:

1. Repeated function calls and register-window operations.
2. Integer arithmetic.
3. Conditional branching.
4. Recursive calls and frame management.

The comparison harness built and executed all 16 measured configurations, checked expected outputs, and recorded repeated subprocess wall-clock measurements.

**Recorded median results (milliseconds):**

| Implementation | Call loop | Arithmetic | Branch | Recursion |
|---|---:|---:|---:|---:|
| Zig labeled | 18.802 | 9.268 | 15.919 | 14.033 |
| Zig switch | 19.794 | 9.280 | 15.985 | 14.413 |
| Rust | 33.049 | 11.892 | 21.395 | 23.859 |
| Go | 53.362 | 18.262 | 29.022 | 39.246 |

Zig recorded the lowest median wall-clock time in every measured workload.

However, the prototypes differ in memory-allocation strategies, and the measurements include subprocess startup overhead. These results do not establish production VM performance or prove that register-based execution is universally faster than stack-based execution.

**Architectural outcome:** Monowire's register-based architecture is selected.

**Implementation-language outcome:** Zig is provisionally preferred, subject to Go integration, ownership, portability, build, and correctness validation.

**Dispatch outcome:** Conventional switch dispatch is the initial implementation baseline. Labeled-switch dispatch remains a future optimization candidate.

**Evidence:** `benchmarks/monowire/README.md`, `benchmarks/monowire/SPEC.md`, and the recorded October 23, 2026 benchmark outputs.

The experiment is complete, but production Monowire implementation remains part of M2.

### Transition to M2 — First execution: Not started

The next development phase will introduce Eloque's first statically checked executable subset.

M2 builds upon the implemented H2 frontend and structured diagnostic recovery.

Its primary implementation stages are:

1. Dedicated lexical name resolution and scope validation.
2. Mandatory basic static type checking, including definite-assignment analysis.
3. Monowire register-based bytecode compilation.
4. A sequential Monowire virtual machine.
5. End-to-end execution and diagnostic tests.
6. Initial runtime correctness and performance benchmarks.

**VM architecture:** Monowire is the selected register-based execution architecture. The previous Go stack-VM plan has been superseded.

**Implementation language:** Zig is provisionally preferred for the Monowire runtime, with Rust retained as a fallback. Go remains the implementation language for the existing frontend and planned semantic-analysis stages.

Before production VM implementation, the relevant engineering work must establish:

- The compiler-to-runtime integration approach.
- The production bytecode format and register conventions.
- The minimum runtime value representation.
- Memory ownership and lifetime rules.
- Error propagation across the compiler/runtime boundary.

These investigations may proceed alongside name-resolution and static-checker development. They do not require delaying the resolver until the complete runtime design is finished.

The initial executable subset focuses on literals, arithmetic, `let`, `change`, and `say`.

Expanded control flow, language-level functions, collections, custom types, modules, and native concurrency remain assigned to later milestones.

Monowire may introduce internal call-frame machinery before Eloque supports user-defined functions. Internal VM capabilities must not be confused with implemented language features.

**M2 completion criterion:** Eloque executes representative valid programs through the production Monowire bytecode VM, rejects supported type-invalid programs before execution, handles supported runtime failures correctly, and has reproducible correctness and performance baselines.

**Next milestone:** M2 — First execution.

Eloque can currently parse supported source programs into an AST but cannot yet execute them.

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

- **Before production Monowire implementation:** Validate the compiler/runtime integration boundary, register-window contract, bytecode format, runtime value representation, and memory-ownership rules.
- **After the first working Monowire VM:** Establish reproducible correctness and execution-performance baselines. Reevaluate dispatch optimizations and other implementation details only when representative profiling justifies them.
- **VM architecture:** Resolved — Monowire's register-based execution model is selected. No stack-based VM prerequisite remains.
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

**CPython parity is a long-term aspiration, not a forecast.** Monowire is the selected register-based execution architecture. The pre-M2 implementation-language benchmarks support provisional selection of Zig but do not establish complete Eloque execution performance. Comparisons with CPython must use equivalent, correctness-tested executable workloads after the production Monowire VM is operational.

---

## Pre-M2 prerequisite: structured diagnostics and frontend recovery

**Status:** Completed — implemented, validated, and merged before M2.

The H2 frontend now supports structured diagnostic collection and recovery-capable parsing through `ParseRecovering(sourceText, ParseOptions)`, while retaining the original fail-fast `Parse(sourceText)` API.

The implementation includes lexical recovery, parser synchronization, statement-level AST preservation, configurable error budgets, expression-depth protection, and dedicated regression, robustness, stress, and performance tests.

Recovery-capable parsing returns a termination status and may retain a partial AST. It does not establish semantic validity or execution eligibility. Name resolution, mandatory static type checking, bytecode compilation, and VM execution remain M2 work.

The work includes:

- [x] Introduce the shared diagnostic representation and minimal compilation-local source registry.
- [x] Define stage-specific result and recovery contracts.
- [x] Adapt the streaming lexer and token source to safe lexical recovery.
- [x] Adapt the arena-backed parser to statement-level recovery and guaranteed progress.
- [x] Enforce compilation eligibility so recovered or malformed programs cannot execute.
- [x] Add diagnostic and recovery regression tests.
- [x] Benchmark the migration against the existing M1/H2 frontend baseline.
- [x] Complete the recovery migration's acceptance criteria and clear the prerequisite for beginning M2 name resolution.

**Outcome:** The pre-M2 structured diagnostic and frontend recovery prerequisite is complete. M2 name resolution is the next implementation stage.

### M2 implementation order

After the completed frontend recovery work, proceed through the following implementation stages:

1. **Name resolution**
   - Implement the dedicated lexical resolver.
   - Establish symbol identities, scope relationships, and identifier bindings.
   - Validate the approved resolver behavioral matrix.

2. **Mandatory static type checking**
   - Implement typing rules for the supported executable subset.
   - Include definite-assignment analysis.
   - Prevent execution of programs with blocking semantic diagnostics.

3. **Monowire execution contracts**
   - Validate the Go compiler-to-Monowire integration boundary.
   - Define the production bytecode instruction subset.
   - Specify register operands, register-window invariants, and runtime value representation.
   - Establish memory ownership and runtime error-propagation rules.

4. **Monowire bytecode compiler**
   - Translate statically validated programs into register-based bytecode.
   - Preserve Eloque's required evaluation order and runtime semantics.
   - Produce valid instructions, register references, and source metadata.

5. **Monowire virtual machine**
   - Implement the initial sequential register-based interpreter.
   - Support constants, arithmetic, bindings, reassignment, and output.
   - Validate register access, instruction semantics, and supported runtime failures.

6. **First executable Eloque programs**
   - Integrate the frontend, resolver, checker, compiler, and VM.
   - Execute representative valid programs.
   - Reject invalid programs before execution.

7. **Runtime correctness and performance benchmarks**
   - Establish instruction-level and end-to-end correctness coverage.
   - Record reproducible VM-only and process-level performance baselines.
   - Preserve environment and toolchain metadata.
   - Compare with CPython only using equivalent, correctness-tested executable workloads.

The stages describe logical dependencies, not a requirement to complete all runtime design work before beginning name resolution.

Compiler/runtime integration investigation may proceed in parallel with semantic-analysis development.

The M2 design is documented in `DESIGN_DECISIONS.md` and `ARCHITECTURE.md`.

Outstanding engineering reviews and their acceptance criteria are tracked in `TECHNICAL_BACKLOG.md`.

**Implementation status:** Name resolution, static checking, production Monowire bytecode compilation, and production VM execution are not yet implemented.

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
- [x] Synchronize documentation with the completed H2 implementation.
- [x] Implement structured diagnostics and recovery-capable frontend parsing.
- [x] Establish post-recovery experimental frontend benchmark evidence.
- [x] Complete final recovery-branch verification and merge.
- [x] Complete the pre-M2 Monowire implementation-language experiment.
- [x] Select Monowire as Eloque's register-based VM architecture.
- [ ] Begin M2 dedicated lexical name resolution.
- [ ] Implement mandatory basic static type checking and definite-assignment analysis.
- [ ] Validate the Go–Monowire integration boundary and provisional Zig implementation choice.
- [ ] Specify the production Monowire bytecode format, register-window invariants, and runtime value representation.
- [ ] Implement the Monowire register-based bytecode compiler.
- [ ] Implement the initial sequential Monowire VM.
- [ ] Validate the first executable Eloque subset end to end.
- [ ] Establish production runtime correctness and performance baselines.
- [ ] Expand grammar and static-checking rules alongside subsequent executable features.
