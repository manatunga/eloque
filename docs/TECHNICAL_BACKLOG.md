# Eloque Technical Backlog

**Status:** Living engineering decision register. M1, post-M1 H2 frontend optimization, and pre-M2 structured diagnostic recovery are implemented. M2 remains the next implementation milestone. A review commitment is not a commitment to implement a particular technology.

See [LANGUAGE_SPEC.md](LANGUAGE_SPEC.md) for intended language behavior, [DESIGN_DECISIONS.md](DESIGN_DECISIONS.md) for settled decisions, and [ROADMAP.md](ROADMAP.md) for implementation milestones.

## How to use this register

- **Committed:** intended capability, with details or delivery timing still subject to engineering validation.
- **Review required:** specific engineering decision with an explicit review trigger.
- **Exploratory:** possible technology or optimization, not approved for implementation.
- **Resolved:** reviewed, with decision and evidence recorded.

At each milestone, inspect items whose triggers have arrived. Record the decision, evidence and next action. If deferred, assign a new explicit trigger. Language rules already settled in the specification must not be silently reopened as runtime implementation choices.

## VM and performance

### TB-001 — Stack-based versus register-based VM
- **Status:** Review required.
- **Baseline:** Go stack-based bytecode VM, with local slots and call frames.
- **Trigger:** After first working VM and repeatable benchmark harness (approximately M2).
- **Method:** Profile dispatch, stack traffic, bytecode size and representative programs. Compare simpler stack optimizations before prototyping a register VM.
- **Criterion:** Measurable end-to-end gains that justify additional compiler/runtime complexity.
- **Evidence:** Pending.

### TB-002 — Safe native concurrency
- **Status:** Committed long-term goal; design review required.
- **Baseline:** Sequential VM; avoid unnecessary global mutable state.
- **Trigger:** Once functions, value ownership and mutation semantics are stable, before concurrency implementation (M9 or later).
- **Method:** Design task lifetime, cancellation, isolation, error propagation, and explicit shared-state rules. Prototype goroutine-backed execution contexts and test race safety.
- **Criterion:** Safe, understandable semantics with useful concurrency; Go goroutines are an implementation mechanism, not the language contract.
- **Evidence:** Pending.

### TB-003 — Adaptive bytecode specialization
- **Status:** Exploratory.
- **Baseline:** Correct generic bytecode operations.
- **Trigger:** After meaningful workloads and profiling expose hot operations.
- **Method:** Prototype numeric/call/property specialization; measure allocation, dispatch and regression behavior.
- **Criterion:** Repeatable benefits without unacceptable complexity.
- **Evidence:** Pending.

### TB-004 — JIT compilation
- **Status:** Exploratory.
- **Baseline:** Bytecode interpretation.
- **Trigger:** Baseline VM optimized and remaining bottlenecks documented; formal review after initial stabilization.
- **Method:** Assess compilation overhead, portability, warmup, runtime integration and maintenance.
- **Criterion:** Clear benefit beyond improving the interpreter.
- **Evidence:** Pending.

### TB-005 — AOT/native backend
- **Status:** Exploratory.
- **Baseline:** Go-hosted bytecode VM.
- **Trigger:** Stable semantics and a concrete deployment/performance use case, no later than 1.0 architecture planning.
- **Method:** Compare native backend, transpilation and bytecode packaging.
- **Criterion:** Practical advantage and manageable compatibility burden.
- **Evidence:** Pending.

### TB-006 — WebAssembly
- **Status:** Exploratory.
- **Baseline:** Native Go runtime.
- **Trigger:** Concrete browser, portable sandbox or embedded execution requirement; otherwise 1.0 ecosystem planning.
- **Method:** Compare compiling Go runtime to Wasm against a dedicated backend.
- **Criterion:** Clear use case with sustainable support costs.
- **Evidence:** Pending.

### TB-007 — CPython comparative performance
- **Status:** Committed measurement program; parity is an aspiration, not a guarantee.
- **Baseline:** Frontend-only comparisons have been completed during M1-P5 and H2. No Eloque runtime execution comparisons are possible yet.
- **Trigger:** First usable VM (M2), then major runtime milestones.
- **Method:** Pin versions, hardware, warmup and inputs. Compare equivalent, correctness-tested arithmetic, loops, calls, strings, collections and object workloads as implemented.
- **Criterion:** Reproducible findings that guide optimization, not marketing claims based on isolated microbenchmarks.
- **Evidence:** M1-P5 compared the streaming frontend with Go and CPython 3.15.0rc2. H2's official controlled arithmetic comparison recorded median frontend latencies at 10,000 units of 7.16 ms for Eloque H2, 14.12 ms for the Go compiler syntax parser, 23.92 ms for Go go/parser, and 147.86 ms for CPython 3.14.4 ast.parse. These results concern different language grammars and AST representations and do not establish execution-performance parity.
- **Next action:** Preserve the frontend measurements as historical baselines. Begin correctness-equivalent execution benchmarks after the first working Eloque VM.

### TB-008 — Runtime value representation
- **Status:** Review required.
- **Baseline:** Undecided; Go GC can manage Go-allocated objects.
- **Trigger:** Before implementing the first VM value type (M2).
- **Method:** Compare Go interfaces, tagged representations and pointer-based values against checked `Int`, binary64 `Float`, `Bool`, `String`, optional values and future collections.
- **Criterion:** Correct semantics, maintainability, acceptable allocation/dispatch costs.
- **Evidence:** Pending.

### TB-009 — Intermediate representation and backend separation
- **Status:** Review required.
- **Baseline:** AST directly to bytecode; frontend semantics independent of VM.
- **Trigger:** Second backend approved or optimization passes become unwieldy.
- **Method:** Identify real requirements and prototype the smallest useful IR.
- **Criterion:** Lower overall complexity, not speculative extensibility.
- **Evidence:** Pending.

### TB-022 — Lexer representation and allocation efficiency

- **Status:** Resolved for the current lexer implementation.
- **Baseline:** String-backed token kinds, an indentation stack with minimal initial capacity, and a pending-token queue that discarded reusable backing storage after draining, causing repeated allocations.
- **Trigger:** Initial M1 lexer implementation and performance profiling.
- **Decision:** Use `uint8`-backed token kinds with readable names, reuse pending-token queue storage through an indexed queue, and preallocate modest indentation-stack capacity.
- **Evidence:** Mixed-source allocations decreased from 248 B/op and 7 allocs/op to 80 B/op and 2 allocs/op. Token structure size decreased from 48 to 40 bytes on amd64. Scaling benchmarks from 138 bytes to 138 KB showed approximately linear execution time, with constant allocation counts for the tested bounded-depth workload. The M1 completion audit reconfirmed approximately linear scaling across 138-byte to 138-KB benchmark inputs. The saved ten-run baseline reports 96 B/op and 2 allocs/op for mixed input, with approximately 695 µs/op for the ×1000 scaling benchmark. The benchmark baseline is retained for future comparisons.
- **Complexity:** O(N) overall scanning time; O(d + q) auxiliary working space, excluding source storage and caller-retained tokens.
- **Criterion:** Maintain correctness, predictable linear scaling, and low allocation overhead without introducing unnecessary implementation complexity.
- **Next action:** Retain the benchmark suite and reprofile when new lexical features materially change scanner behavior.
- **Next review trigger:** Completion of M1 lexical coverage or a measured regression in representative workloads.

## Frontend and language semantics

### TB-010 — Conversational grammar and call parsing

- **Status:** Initial grammar design resolved; parser implementation pending.
- **Baseline:** Function calls use `name with (...)`, including zero-argument calls. Function definitions use `to name with (...) giving Type:` where a return type is declared. Multiword operators are recognized according to the documented grammar.
- **Decision:** The initial expression grammar uses Pratt parsing. Exponentiation is right-associative and binds more tightly than unary negation on its left, while accepting unary expressions on its right. Multiplicative and additive operators are left-associative. Chained comparisons are excluded from the initial grammar.
- **Evidence:** The M1-P2 Pratt expression parser implements integer and string literals, identifiers, grouping, unary numeric operators, and binary arithmetic operators. Tests cover precedence, left- and right-associativity, exponentiation's interaction with unary operators, source spans, malformed expressions, and expression boundaries. Stress tests with 1,000 nested parentheses and 1,000 exponentiation operators pass. Function-call parsing and conversational comparison operators are not yet implemented.
- **Next action:** Retain the implemented arithmetic grammar as the baseline. Extend expression parsing when comparison, Boolean, call, and member-access syntax enters its scheduled milestone. Implement function calls according to the established `name with (...)` grammar rather than introducing alternative syntax.
- **Next review trigger:** M3 for comparison and Boolean expression parsing; M4 for function-call parsing.
- **Criterion:** Deterministic, unambiguous parsing with precise source locations and diagnostics.
- **Note:** The historical `call ... with ...` and bare built-in invocation alternatives remain superseded.

### TB-011 — Scope, capture and immutability
- **Status:** Review required.
- **Baseline:** Block scopes, no shadowing of enclosing visible names, `fixed` binding immutability and `frozen` deep immutability.
- **Trigger:** Scope resolution in M2; revisit alias/capture behavior before mutable collections and closures.
- **Method:** Test nested scopes, assignment, recursion, frozen values and aliasing. Define capture mechanics only when closures enter scope.
- **Criterion:** No accidental mutation through alternate references; diagnostics reflect declared rules.
- **Evidence:** Pending.

### TB-012 — Static type inference and checking
- **Status:** Committed; implementation review required.
- **Baseline:** Strict static types, explicit parameter annotations, local inference and `T?`; type errors block normal execution.
- **Trigger:** Before first executable subset in M2.
- **Method:** Specify expression typing, assignment compatibility, return-path validation, optional narrowing and invalidation, and error recovery. Add positive/negative checker fixtures for every implemented feature.
- **Criterion:** Predictable soundness boundaries and actionable diagnostics; never silently downgrade type errors to warnings.
- **Evidence:** Pending.

### TB-013 — Numeric semantics and conversions
- **Status:** Review required.
- **Baseline:** Checked signed 64-bit `Int`; IEEE 754 binary64 `Float`; floor division/remainder; exact-only implicit Int-to-Float promotion.
- **Trigger:** Before arithmetic checker and VM implementation (M2).
- **Method:** Test overflow, division by zero, negative operands, mixed types, representability, NaN/infinity policy and shortest round-trip formatting.
- **Criterion:** Compiler/runtime agreement and reproducible edge-case behavior.
- **Evidence:** Pending.

### TB-014 — Collection value semantics and mutation
- **Status:** Review required.
- **Baseline:** Lists, maps and structured values are logically independent when assigned; negative sequence indexing and optional safe lookup are designed.
- **Trigger:** Before mutable collections in M5.
- **Method:** Compare eager copying, copy-on-write and structural sharing; test nested mutation, `fixed` versus `frozen`, slicing and map-key behavior.
- **Criterion:** Preserve observable value independence without excessive implementation complexity.
- **Evidence:** Pending.

### TB-015 — Unicode and string indexing
- **Status:** Review required.
- **Baseline:** Immutable UTF-8-backed Unicode strings with character-oriented access, not raw byte indexing.
- **Trigger:** Before string indexing/slicing in M5.
- **Method:** Decide Unicode code points versus grapheme clusters; test combining marks, emoji sequences and malformed source input.
- **Criterion:** Well-documented predictable behavior with viable implementation cost.
- **Evidence:** Pending.

### TB-016 — Exceptions and cleanup
- **Status:** Review required.
- **Baseline:** Typed unchecked exceptions, `try`/`catch`/`finally`, and source-aware failures.
- **Trigger:** Before exception implementation (M7).
- **Method:** Define hierarchy, matching, propagation, return/loop exits through `finally`, and precedence when cleanup also fails.
- **Criterion:** Deterministic unwinding and clear diagnostics.
- **Evidence:** Pending.

### TB-017 — Modules and visibility
- **Status:** Review required.
- **Baseline:** `import`, aliases with `named`, public-by-default top-level symbols, underscore-private names, once-only initialization and rejected circular imports.
- **Trigger:** Before module implementation (M7).
- **Method:** Define path resolution, dependency order, initialization failure behavior and caching.
- **Criterion:** Deterministic behavior and compile-time privacy enforcement.
- **Evidence:** Pending.

### TB-018 — Structured types, interfaces and matching
- **Status:** Review required.
- **Baseline:** `define` structured values, implicit `self`, structural interfaces and `match`/`when` without fallthrough; no class inheritance.
- **Trigger:** Before M6.
- **Method:** Specify construction, methods, structural conformance, recursive values, exhaustiveness and value semantics.
- **Criterion:** Coherent static checks and understandable errors.
- **Evidence:** Pending.

### TB-023 — Lexer–parser token delivery strategy

- **Status:** Resolved for the M1 and H2 production frontends.
- **Baseline:** Eagerly buffered lexer tokens supplied to the parser.
- **Trigger:** M1-P5-A lexer–parser integration experiments.
- **Alternatives evaluated:** Complete token buffering, direct pull streaming,
and bounded two-token lookahead.
- **Decision:** Use direct pull streaming for the canonical frontend.
The parser requests tokens incrementally through a token-source
abstraction. Buffered delivery remains available for compatibility
tests and comparative benchmarks.
- **Rationale:** The implemented grammar does not require two-token
lookahead. Direct streaming avoids full token-buffer allocation and
avoids maintaining an unnecessary lookahead window.
- **Evidence:** M1-P5-A token-delivery and complete-frontend
experiments, followed by M1-P5-B production integration.
In the 20-case integrated buffered-versus-streaming comparison,
direct streaming was faster in 19 cases and allocated fewer
bytes in all 20. H2 retained direct pull-based token streaming while replacing the production AST representation with an arena-backed flat AST.
Production integration and diagnostic regression tests passed.
- **Criterion:** Preserve AST equivalence, lexical-error propagation,
EOF behavior, and performance across representative workloads.
- **Next action:** Retain buffered-versus-streaming regression
benchmarks and repeat them when significant grammar changes
are introduced. Reconsider bounded lookahead only if a concrete
grammar requirement or measured performance regression justifies it.
- **Next review trigger:** A future grammar feature demonstrably
requires additional token lookahead, or profiling reveals a
meaningful regression attributable to token delivery.

### TB-024 — Arena-backed flat AST architecture

- **Status:** Resolved for the H2 frontend.
- **Baseline:** M1 used an interface-based AST with individually represented expression and statement nodes.
- **Trigger:** Post-M1 H2 frontend optimization and allocation profiling.
- **Alternatives evaluated:** Retaining the M1 representation versus introducing an arena-backed flat AST.
- **Decision:** Use the arena-backed flat AST as the canonical production representation. Preserve the M1 AST and parser as legacy reference implementations.
- **Rationale:** Reduce allocation overhead and improve frontend performance while preserving syntax, source locations, diagnostics, and AST structure.
- **Evidence:** H2 differential and arena-invariant tests passed. At 1,000 workload units, the ten-run median comparison showed improvements of 1.14× for declarations, 1.77× for expressions, 1.01× for strings, 1.18× for comments, and 1.41× for mixed programs. Allocation counts were reduced across the tested workloads.
- **Criterion:** Preserve correctness and the supported grammar while improving representative frontend performance without unacceptable memory overhead.
- **Next action:** Retain the M1/H2 differential tests and official performance baselines. Extend flat AST coverage alongside future grammar features.
- **Next review trigger:** A substantial AST feature expansion, a measured regression, or evidence that the flat representation materially complicates semantic analysis.

### TB-025 — Arena capacity and growth optimization

- **Status:** Review required; optimization deferred.
- **Baseline:** H2's existing arena allocation and growth strategy.
- **Trigger:** After representative frontend profiling identifies arena growth or allocation as a significant remaining cost, or after a substantial change in AST workload characteristics.
- **Rationale for deferral:** H2 already provides measurable improvements. Additional preallocation or growth-policy changes could increase memory retention or penalize small programs without delivering meaningful end-to-end gains.
- **Evidence:** H2 experimental arena-capacity, growth, and node-capacity measurements are retained in `benchmarks/results/experimental/`. These experiments have not established a sufficient reason to change the production policy.
- **Method:** Benchmark small, medium, large, expression-heavy, and mixed programs. Compare allocation counts, allocated bytes, retained capacity, and latency. Use repeated runs and representative input distributions.
- **Acceptance criterion:** Demonstrate repeatable end-to-end improvements without unacceptable memory overhead, regressions on small programs, or additional correctness risks. Reject optimizations whose benefits do not justify their complexity.
- **Next action:** Preserve the experimental results and retain the current arena policy.
- **Next review trigger:** M2 integration reveals a measurable arena-related bottleneck, or frontend profiling after a significant grammar expansion identifies a new opportunity.

### TB-026 — Post-recovery frontend performance validation

- **Status:** Review required; further optimization deferred.
- **Baseline:** The October 22–23 experimental post-recovery measurements, alongside the unchanged October 20 official H2 baseline.
- **Trigger:** Representative mixed-program workloads, significant parser changes, or a measured frontend performance regression.
- **Rationale:** The current controlled arithmetic benchmarks show modest recovery overhead on large valid inputs, but they do not establish performance across realistic source files or malformed-input distributions.
- **Method:** Compare fail-fast and recovery-capable parsing using equivalent source inputs. Include valid and malformed programs, small and large files, repeated measurements, allocated bytes, allocation counts, and latency variability.
- **Acceptance criteria:** Preserve correctness and recovery termination guarantees. Introduce optimizations only when repeatable improvements justify their complexity and memory trade-offs.
- **Evidence:** `benchmarks/results/experimental/h2-post-recovery-report/REPORT.md`.
- **Next action:** Retain the current implementation. Do not block M2 on further frontend optimization.

## Tooling, agents and execution safety

### TB-019 — Structured diagnostics
- **Status:** Committed; schema review required.
- **Baseline:** Human-readable diagnostics plus machine-readable output for agents and editor tools.
- **Trigger:** Begin source spans in M1; finalize initial schema before M8 tooling.
- **Method:** Define stable error codes, severity, file/line/column/spans, related locations and optional suggested fixes.
- **Criterion:** Diagnostics are actionable, deterministic and easy to parse without requiring prose interpretation.
- **Evidence:** Pending.
- **M1 progress:** Tokens and AST nodes use shared half-open source spans. M1-P2 established expression source-location tracking, including the distinction between semantic AST node locations and complete syntactic grouping extents. M1-P3 extended source-span preservation and structured fail-fast diagnostics to `let`, `change`, and `say` statements and complete-program parsing. M1-P4 completed broader positive and negative frontend syntax conformance coverage with 29 additional fixtures covering valid programs, malformed syntax, EOF diagnostics, multiline source locations, nested expression errors, first-error behavior, and lexer/parser error boundaries. Tests verify diagnostic messages, source spans, expected/actual token metadata, and rejection of malformed programs without returning successful partial ASTs. Full Go tests, race detection, and `go vet` passed.
- **Next action:** Preserve the completed M1 diagnostic contract through H2 and subsequent frontend changes. Extend diagnostics to name resolution and static type checking during M2. Multi-error recovery remains deferred under D-037, and the broader machine-readable diagnostic schema remains scheduled for review before M8 tooling.
- **Post-M1 status:** Active implementation on `feat/structured-diagnostics`.
- **Completed foundation:** Introduced the shared `internal/diagnostic` model with stable codes, severity, file identity, half-open byte spans, related locations, notes, and token-expectation metadata. Migrated the lexer, H2 parser, frontend, and associated tests to native structured diagnostics. Retired the legacy M1 pointer-based parser.
- **Current objective:** Replace fail-fast frontend behavior with bounded multi-error recovery while preserving the canonical arena-backed AST and direct pull-based token streaming.
- **Remaining work:** Diagnostic accumulation and ownership; explicit lexical recovery; parser synchronization at safe statement and nesting boundaries; error budgets; parse completion, recovery, and abortion outcomes; partial AST retention; and diagnostic cascade suppression.
- **Acceptance criteria:** Multiple independent errors are reported deterministically; recovery always terminates; lexical diagnostics are not obscured by secondary parser errors; valid statements survive recovery where safe; malformed programs are never treated as valid for execution; and successful-path latency and allocations are validated against the preserved frontend baseline.
- **Next action:** Finalize recovery contracts and integrate diagnostic accumulation with the existing lexer, token source, and H2 parser. Add recovery conformance tests and reproducible performance benchmarks.
- **Next review trigger:** Completion of multi-error recovery integration and performance validation, before proceeding to M2.

### TB-020 — Sandboxed execution budgets
- **Status:** Committed long-term capability; design review required.
- **Baseline:** Normal CLI has no arbitrary global instruction limit. Sandboxed runs should support configurable instruction, time and memory budgets.
- **Trigger:** Before agent-oriented execution or untrusted-code hosting (M9 or earlier if required).
- **Method:** Define metering points, cancellation, resource isolation, host guarantees and failure categories.
- **Criterion:** Explicit, testable limits with no misleading security guarantees.
- **Evidence:** Pending.

### TB-021 — References and ownership
- **Status:** Review required.
- **Baseline:** Value semantics by default; explicit shared mutable references are deferred.
- **Trigger:** Before public concurrency/shared-state syntax.
- **Method:** Evaluate ownership, borrowing or other explicit-sharing models against ease of learning and safety.
- **Criterion:** Preserve predictable mutation while enabling useful concurrent programs.
- **Evidence:** Pending.

## M1-P2 expression parser performance review

- **Date / milestone:** October 2026 — M1-P2.
- **Decision:** Retain the current Pratt parser implementation without production-code optimisation.
- **Evidence:** Five-run benchmarks measured arithmetic expressions, mixed precedence, additive chains, exponentiation chains, and nested parentheses using prepared token streams. At 1,000 operators, median parsing times were approximately 115.63 µs for additive chains and 135.02 µs for exponentiation chains. At 1,000 grouping levels, median parsing time was approximately 160.87 µs. Execution time and allocations scaled approximately linearly across the tested sizes. Stress tests at depth 1,000 passed.
- **Reasoning:** No pathological scaling or demonstrated performance bottleneck justifies additional implementation complexity at this checkpoint.
- **Next action:** Preserve the benchmark suite and compare results after substantial parser changes. Investigate recursive parsing depth only if robustness requirements or measurements establish a need.
- **Next review trigger:** Completion of M1-P5 frontend integration, a measured performance regression, or a demonstrated excessive-nesting failure.

## M1-P3 statement and program parser performance review

- **Date / milestone:** October 2026 — M1-P3.
- **Decision:** Retain the current statement and program parser implementation without production-code optimisation.
- **Evidence:** Five-run benchmarks measured complete-program parsing from prepared token streams containing 3, 300, and 3,000 statements, cycling through `let`, `change`, and `say`. Mean execution times were approximately 864.5 ns, 86.5 µs, and 1.007 ms respectively. Memory usage was approximately 1,712 B, 162,994 B, and 1,694,141 B per parse, with 29, 2,510, and 25,015 allocations per parse. Execution time, memory consumption, and allocation counts scaled approximately linearly across the tested workloads. Statement/program correctness and diagnostic tests passed.
- **Reasoning:** The implementation advances through tokens without backtracking, delegates expressions to the existing Pratt parser, and exhibits no demonstrated pathological scaling or bottleneck that would justify additional complexity.
- **Next action:** Retain `internal/parser/program_bench_test.go` and the saved `benchmark_m1_p3.txt` results as a performance baseline. Reassess after frontend integration or substantial grammar changes.
- **Next review trigger:** M1-P5 frontend integration, a measured performance regression, or a demonstrated robustness issue.

## Post-M1-P3 canonical frontend optimization review

- **Date / milestone:** October 2026 — post-M1-P3, before M1-P4.
- **Status:** Resolved for the current lexer and parser subset.
- **Decision:** Adopt the optimized canonical lexer and parser as the implementation baseline for subsequent frontend work.
- **Rationale:** Consolidate frontend implementation paths, reduce avoidable parsing and allocation overhead, and establish reproducible performance evidence before expanding syntax conformance.
- **Evidence:** Recorded optimized lexer throughput of approximately 261.6 MB/s with 2 allocations per operation. Parser benchmarks measured approximately 76.0–99.4 µs for 1,000-operation chains and 664 µs with 13,028 allocations for a 3,000-statement program. The expression-heavy complete frontend benchmark measured approximately 34.5 MB/s.
- **Comparison:** The original M1-P3 3,000-statement benchmark measured approximately 1.007 ms and 25,015 allocations per parse. The optimized measurement represents approximately 34% lower execution time and 48% fewer allocations under the recorded workload.
- **Scope:** These measurements cover the currently implemented lexical and syntactic subset. They do not measure static checking, bytecode compilation, or runtime execution.
- **Next action:** Preserve historical benchmark evidence, use the canonical frontend for M1-P4 conformance testing, and complete the M1-P5 integration and regression review.
- **Next review trigger:** M1-P5, substantial grammar expansion, a reproducible performance regression, or an identified correctness or robustness issue.
- **Acceptance criteria:** Existing language behavior and diagnostics remain correct; new syntax receives positive and negative conformance tests; performance remains measurable and regressions are investigated before introducing additional optimization complexity.

See [`benchmarks/README.md`](../benchmarks/README.md) for reproduction instructions, recorded benchmark results, and methodological limitations.

## H2 frontend optimization review

- **Decision:** Adopt the H2 streaming parser and arena-backed flat AST as the canonical production frontend.
- **Evidence:** Differential correctness testing, arena-invariant testing, representative M1/H2 workload comparisons, and controlled cross-language frontend measurements.
- **Outcome:** The H2 implementation improved representative frontend performance and reduced allocations. The frontend was reorganized under `internal/frontend/`, with shared source utilities retained in `internal/source`.
- **Deferred optimization:** Further arena-capacity tuning remains tracked under TB-025.
- **Next action:** Preserve official and experimental benchmark results, extend the frontend as required by future language milestones, and avoid optimization without profiling evidence.
- **Next review trigger:** M2 integration reveals a meaningful frontend bottleneck, a new grammar feature changes AST allocation patterns, or representative benchmarks demonstrate a regression.

The H2 measurements are frontend-only. They do not establish Eloque runtime performance or justify changing the planned stack-based VM baseline.

---

## M2 architecture follow-through — October 2026

**Status:** Approved decisions and deferred work. No implementation implied.

### Immediate prerequisite: diagnostic migration

**Priority:** Required before M2 resolver implementation.

**Rationale:** First-error termination limits developer feedback and future agent tooling. Recovery must be introduced without compromising compiler correctness or H2's frontend architecture.

**Required work:**

- Shared structured diagnostics with stable codes, severity, locations, related locations, and notes.
- Compilation-local file identity and source ownership.
- Stage-owned diagnostic slices and a compilation-wide budget.
- Explicit complete, recovered, and aborted outcomes.
- Safe lexer and parser recovery.
- Deterministic ordering and cascade suppression.
- Preservation of parser token mismatch metadata.
- Strict prevention of executing recovered or invalid programs.

**Acceptance criteria:** Multiple independent errors reported; no infinite loops; accurate spans; safe structural recovery; deterministic diagnostics; valid-program regression tests pass; representative performance recorded.

### M2 name resolver

**Priority:** Immediately after diagnostic migration.

**Rationale:** Identifier binding must be established before type checking and bytecode generation.

**Required work:**

- Contiguous symbol arena with compilation-local `SymbolID`.
- Parent-linked lexical scopes.
- Local name maps.
- Declaration-before-use rules.
- Duplicate declaration errors.
- Nested shadowing warnings with related locations.
- Initializer self-reference binding.
- Private reference-binding storage indexed initially by AST `NodeID`.
- Correct binding conventions for identifier expressions, declarations, and reassignment targets.
- No AST mutation.

**Acceptance criteria:** Resolver behavioral matrix R01–R28 in `ARCHITECTURE.md`, with scope-specific tests introduced as the relevant grammar becomes available.

### Deferred: diagnostic budget tuning

**Current baseline:** Recovery-capable parsing defaults to a maximum of 20 error-severity diagnostics. Positive limits are configurable; negative values disable the limit.

**Trigger:** Representative malformed-source workloads, editor integration, or evidence that the current default produces insufficient diagnostics or excessive recovery work.

**Rationale:** The default of 20 provides a bounded initial recovery policy. Increasing it without representative evidence could add unnecessary processing or produce overwhelming diagnostic output.

**Acceptance criteria:** Predictable termination, bounded resource consumption under the default configuration, actionable diagnostics, and no meaningful regression in normal frontend performance.

### Deferred: identifier interning

**Trigger:** Repeated-identifier workloads and eventual multi-file compilation.

**Rationale:** Source-backed strings avoid early allocations. Interning may help repeated names but introduces tables and ownership complexity.

**Acceptance criteria:** Demonstrated net benefit in allocations, retained memory, and lookup performance.

### Deferred: expression traversal strategy

**Trigger:** Deeply nested expression stress tests.

**Rationale:** Recursive traversal is straightforward but may require safeguards for extreme nesting.

**Acceptance criteria:** Correctness and stack safety under defined stress workloads, with acceptable traversal overhead.

### Deferred: public diagnostic serialization

**Trigger:** CLI, editor, or AI-agent tooling requiring a stable machine-readable protocol.

**Rationale:** The internal diagnostic model should be stable before committing to a public JSON schema.

**Acceptance criteria:** Deterministic output, stable codes, accurate locations, extensible metadata, and documented compatibility expectations.

### Deferred: diagnostic budget tuning

**Trigger:** Realistic malformed-file and tooling workloads.

**Rationale:** An initial candidate of 100 errors may not suit every compilation context.

**Acceptance criteria:** Bounded resource use, actionable diagnostic output, and predictable truncation behavior.

### Deferred: multi-file source management

**Trigger:** Module imports and multi-file compilation, expected around M7.

**Rationale:** M2 needs compilation-local file identity, not a complete workspace architecture.

**Acceptance criteria:** Correct cross-file locations, source ownership, module identity, and acceptable memory use.

### Deferred: runtime architecture optimization

**Trigger:** First executable stack VM and reproducible runtime benchmarks.

**Rationale:** Register VMs, JIT, AOT, WebAssembly, and alternative backends cannot be justified solely from frontend performance measurements.

**Acceptance criteria:** Demonstrated runtime or deployment advantages against representative workloads, accounting for implementation and maintenance costs.

### Documentation consistency requirement

When implementation begins, review this backlog alongside:

- `LANGUAGE_SPEC.md`
- `DESIGN_DECISIONS.md`
- `ARCHITECTURE.md`
- `ROADMAP.md`

Update implementation status and superseded descriptions without erasing historical M1/H2 context.

All deferred items must retain their rationale, reevaluation trigger, and acceptance criteria until resolved.

## Milestone review checklist

- [x] **M1 lexer performance:** TB-022 token representation, allocation efficiency, and scaling review.
- [x] **M1:** Begin TB-019 source spans through lexer tokens.
- [x] **M1-P1:** Review TB-010 conversational grammar, exponentiation, and Pratt precedence before expression parsing.
- [x] **M1-P1:** Establish shared AST source spans and the initial structured parser diagnostic contract under TB-019.
- [x] **M1-P2:** Verify arithmetic expression precedence, associativity, source spans, syntax diagnostics, robustness, and initial performance scaling.
- [x] **M1-P3:** Verify statement and complete-program source spans, structured syntax diagnostics, nested expression diagnostic propagation, and program-parsing performance.
- [x] **Post-M1-P3:** Complete canonical lexer/parser
  optimization, establish updated performance baselines,
  and preserve reproducible benchmark evidence.
- [x] **M1-P4:** Complete broader syntax diagnostics and positive/negative frontend conformance fixtures.
- [x] **M1-P5:** Complete lexer–parser integration, end-to-end frontend validation, and performance regression review.
- [x] **Pre-M2:** Implement structured diagnostics and recovery-capable parsing.
- [x] **Pre-M2:** Record post-recovery experimental frontend benchmarks.
- [ ] **Future frontend review:** Revisit TB-026 when representative workloads or regressions justify it.
- [ ] **M2:** Extend TB-019 diagnostics to name resolution and static type checking.
- [ ] **Before M2:** TB-008 values, TB-012 checker, TB-013 numeric behavior, TB-011 scope.
- [ ] **After working M2:** TB-001 stack/register review; TB-007 benchmark baseline.
- [ ] **Before M5:** TB-014 collections/immutability; TB-015 Unicode.
- [ ] **Before M6:** TB-018 custom types/interfaces/matching.
- [ ] **Before M7:** TB-016 exceptions; TB-017 modules.
- [ ] **M8:** TB-019 diagnostics/tooling schema.
- [ ] **Before M9 or sandbox deployment:** TB-002 concurrency, TB-020 budgets, TB-021 ownership.
- [ ] **When measured evidence or concrete use cases arise:** TB-003/004/005/006/009.

## Review record template

```text
Date / milestone:
Item ID:
Decision: implement / retain baseline / defer / reject
Evidence (tests, benchmarks, prototypes, links):
Reasoning:
Owner and next action:
Next review trigger (if deferred):
```

Do not mark an item resolved merely because a target release has passed. Review it or explicitly reschedule with a rationale.
