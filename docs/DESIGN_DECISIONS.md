# Eloque Design Decisions

**Status:** Active design decision register, updated October 2026.
**Normative language specification:** [LANGUAGE_SPEC.md](LANGUAGE_SPEC.md).

This register records accepted design choices, provisional directions, deferred work, and selected implementation decisions. **Decided** means agreed, not necessarily implemented. **Provisional** requires further specification or validation. **Deferred** is deliberately postponed. **Implemented (M1)** identifies behavior or architecture established during Milestone 1. **Implemented (H2)** identifies the post-M1 frontend optimization and architectural migration. Neither status implies that Eloque programs can execute. **Implemented (post-H2)** identifies frontend functionality added after the H2 optimization, including structured diagnostic recovery. It does not imply that semantic analysis or execution is available.

## Current implementation baseline — M1 and H2

Milestone 1 established Eloque's Go lexer, parser, AST construction, structured fail-fast lexical and syntax diagnostics, and canonical `Parse(sourceText)` entry point.

### Post-H2 structured diagnostics and recovery

The production frontend additionally supports `ParseRecovering(sourceText, ParseOptions)`, which returns a `ParseResult` containing a potentially partial arena-backed AST, collected diagnostics, and a termination status.

The possible statuses are `ParseComplete`, `ParseRecovered`, and `ParseAborted`. Recovery is bounded by a configurable error budget: zero selects the default of 20 errors, positive values specify a limit, and negative values disable the limit.

The recovery architecture preserves successfully completed statements where possible, while avoiding retention of incomplete AST construction. The frontend also guards against excessive expression nesting.

The original `Parse(sourceText)` API remains available for fail-fast parsing. Recovery-capable parsing is intended for diagnostics and tooling; a partial AST is not evidence of semantic validity or executability.

These capabilities are implemented frontend behavior, not changes to the language's normative grammar.

The supported syntax subset includes literals, identifiers, arithmetic expressions, and `let`, `change`, and `say` statements.

The post-M1 H2 optimization introduced an arena-backed flat AST and a new production parser implementation while preserving direct pull-based token streaming.

The canonical `Parse(sourceText)` entry point now returns `(*ast.FlatProgram, error)`.

The original M1 interface-based AST and parser remain available internally as legacy reference implementations for regression and performance comparisons.

The frontend packages are organized under `internal/frontend/{ast,lexer,parser}`, while shared source-location utilities remain in `internal/source`.

H2 was validated through correctness and differential testing, arena-invariant tests, and representative performance benchmarks.

At 1,000 workload units, H2 recorded median improvements over M1 ranging from approximately 1.01× for strings to 1.77× for expressions, with lower allocation counts across the tested workloads.

The controlled cross-language benchmarks concern frontend processing only and do not establish general compiler or runtime performance superiority.

**Static type checking, bytecode generation, and execution are not yet implemented.**

See [ARCHITECTURE.md](ARCHITECTURE.md) for the implementation structure, [ROADMAP.md](ROADMAP.md) for milestone progress, and [benchmarks/README.md](../benchmarks/README.md) for benchmark methodology and results.

## Current decisions

| ID | Status | Decision |
|---|---|---|
| D-001 | Decided; frontend implemented (M1) | Implement Eloque in Go. M1 established the lexer, parser, AST, and diagnostics. A mandatory static checker, bytecode compiler, and stack-based VM remain planned. |
| D-002 | Decided; CLI provisional | **Eloque** is the language name and `.elq` its official source extension. `eloque` is the planned CLI command, subject to availability and naming-conflict checks. |
| D-003 | Decided | General-purpose language with concise, structured conversational syntax—not unrestricted English. Prioritize readability, predictable agent tooling, and foundations for safe concurrency. |
| D-004 | Decided; initial lexer implemented (M1) | Four-space significant indentation, `:` blocks, no tabs, `#` comments, case-sensitive ASCII identifiers, and double-quoted strings. |
| D-005 | Decided; initial syntax implemented (M1) | `let x be value` declares a mutable binding; `change x to value` reassigns. Initialized local bindings may infer types when type checking is available. |
| D-006 | Decided | `let fixed x be value` prevents rebinding; `let frozen x be value` provides deep immutability. Enforcement awaits later stages. |
| D-007 | Decided | `true`, `false`, and `none` are language literals. `none` is valid in optional `T?` contexts, not arbitrary non-optional values. |
| D-008 | Decided; checker pending | Strict static typing with `as Type` annotations and local inference; checker errors block normal compilation. No implicit truthiness. |
| D-009 | Decided; functions pending | Function parameters require explicit types; value-returning functions require `giving Type`. Omission denotes no value return. |
| D-010 | Decided; control flow pending | `if`, `otherwise if`, and `otherwise`; lexical block scopes; nested lexical shadowing is permitted with a nonfatal warning referencing the shadowed declaration; same-scope duplicates are errors. |
| D-011 | Decided; loops pending | `for each`, `repeat N times`, `repeat while`, `repeat until`, `skip`, and `stop`. |
| D-012 | Decided; functions pending | `to name with (parameters) giving Type:` declares functions; `name with (arguments)` invokes them, including zero-argument calls; `give` returns. |
| D-013 | Decided; functions pending | Named arguments and defaults use `be`. Verb-first function names are a convention, not a restriction. |
| D-014 | Deferred | Closures, first-class function details, variadics, and generics require deliberate grammar design; earlier `outer` and anonymous-function syntax are not automatically retained. |
| D-015 | Decided; arithmetic grouping implemented (M1) | Parentheses group expressions. Operands and arguments evaluate left-to-right, except Boolean short-circuiting. Runtime evaluation rules remain to be implemented. |
| D-016 | Decided; custom types pending | `define` introduces structured types with fields and methods; implicit `self`; structural interfaces; no traditional class inheritance. |
| D-017 | Decided; exceptions pending | Typed unchecked exceptions use `raise`, `try`, `catch`, and `finally`; a `try` requires at least one `catch` or `finally`. |
| D-018 | Decided; modules pending | `import` loads modules and `named` specifies aliases. Modules initialize once; circular imports are initially rejected; underscore-prefixed top-level names are compiler-private. |
| D-019 | Decided; output syntax implemented (M1) | `say` is the output statement. Interpolation syntax is planned as `"Hello, {name}"`, but M1 parses strings as ordinary literals without interpolation evaluation. Input API remains unspecified. |
| D-020 | Decided; collections pending | Lists and maps use conversational access and mutation; `remove all` removes all matching list values; `find` returns optional values. |
| D-021 | Decided; comparisons pending | Word-based comparisons (`is`, `is not`, `is greater than`, etc.) and structural equality where supported. Chained comparisons are excluded from the initial grammar. |
| D-022 | Decided; VM pending | Implement a stack VM first. Reconsider a register VM only after a functioning baseline and reproducible comparative benchmarks. |
| D-023 | Provisional | CPython runtime performance parity is an aspiration, not a guarantee. M1 parsing benchmarks are not evidence of execution-speed parity; benchmark runtime performance after the VM exists. |
| D-024 | Provisional | Backend extensibility, standard library scope, packaging, and post-1.0 compatibility policy need further specification. |
| D-025 | Decided; collections pending | Zero-based indexing, negative indices from the end, end-exclusive slices, and typed failures for invalid direct access. Unicode indexing units remain provisional. |
| D-026 | Decided; numeric semantics pending | `Int` is checked signed 64-bit; `Float` is IEEE 754 binary64. Display the shortest round-trippable decimal; no automatic rounding; implicit promotion only when provably exact. Conversion edge cases require refinement. |
| D-027 | Decided; arithmetic syntax implemented (M1) | `/` denotes real division, `//` floor division, and `%` floor-compatible remainder. Division by zero fails. These runtime semantics are not yet executable. |
| D-028 | Decided; collections pending | Lists, maps, and structured values have logical value semantics. Copy-on-write or structural sharing may implement them without observable aliasing. |
| D-029 | Deferred | Design explicit references, shared mutable state, and concurrency syntax together. |
| D-030 | Decided; pattern matching pending | `match`, `when`, and `otherwise`, without fallthrough. Statement matches may be non-exhaustive; value-producing matches must be exhaustive. |
| D-031 | Decided; execution pending | Top-level statements execute in source order; `main` is not invoked automatically. |
| D-032 | Decided; diagnostics implemented in part (M1) | Normal execution is sequential; runtime exceptions include source locations; tooling requires structured diagnostics. M1 provides structured frontend diagnostics, not runtime exceptions. |
| D-033 | Decided; runtime pending | Support configurable sandbox limits; do not impose an arbitrary universal instruction limit on ordinary CLI execution. |
| D-034 | Decided; lexer implemented (M1) | Decimal integer literals prohibit leading zeros except `0`. A digit sequence immediately followed by an ASCII letter or underscore is a lexical error; invalid-suffix diagnostics take precedence over leading-zero diagnostics. The existing lexer fails fast; controlled multi-error recovery is approved for the pre-M2 frontend migration. |
| D-035 | Implemented (M1) | Go lexer token kinds use a `uint8`-backed enumeration with zero reserved for `TokenInvalid`; `String()` supplies readable names. Incremental token emission and reused indentation-stack/pending-queue storage reduce allocations. This is an implementation choice, not a language semantic rule. |
| D-036 | Decided; lexer implemented (M1) | `change`, `to`, `with`, and `giving` are reserved keywords with dedicated token kinds, not ordinary identifiers. One token kind may serve multiple grammatical contexts. |
| D-037 | Active implementation; M1 fail-fast baseline retained | M1 implemented deterministic fail-fast lexical and syntax diagnostics. The post-M1 structured-diagnostics checkpoint is transitioning the frontend to multi-error recovery. The shared diagnostic model and native lexer/parser diagnostic migration are complete; diagnostic accumulation, lexer recovery, parser synchronization, error budgets, and parse outcomes remain in progress. Recovery must preserve direct pull-based token streaming, deterministic diagnostic reporting, safe termination, and successful-path performance. The M1 fail-fast behavior remains the regression baseline until the recovery implementation passes correctness and performance validation. |
| D-038 | Implemented (M1) | The canonical `Parse(sourceText)` frontend consumes lexer tokens incrementally via `tokenSource`, without a redundant independent token cursor. Buffered parsing remains for regression tests and benchmarks. Future changes must preserve correctness, diagnostic quality, and reproducible performance baselines. |
| D-039 | Implemented (H2) | Adopt an arena-backed flat AST as the canonical production representation. The parser constructs `FlatNode` and `FlatProgram` structures using arena-managed storage. Retain the original M1 interface-based AST and parser as legacy references for correctness and performance comparisons. This is an internal representation decision, not a language-semantic change. |
| D-040 | Implemented (H2) | Organize the frontend under `internal/frontend/{ast,lexer,parser}` while retaining shared source-location utilities in `internal/source`. Preserve the canonical `Parse(sourceText)` entry point and direct pull-based token streaming. Future semantic analysis and backend packages should be introduced when their implementations begin rather than creating speculative empty package structures. |

## Superseded syntax and semantics

The following are **historical examples only**, not current lexer/parser fixtures.

| Earlier draft | Current direction |
|---|---|
| `x = 2`, `x += 1` | `change x to 2`; compound assignment not retained |
| `fix x at 1` | `let fixed x be 1` or `let frozen x be 1` |
| `nothing` | `none` |
| `name: int`, `name of type int` | `name as Int` |
| Optional gradual type checker | Mandatory static checker with inference |
| `else if` | `otherwise if` |
| `each item in items` | `for each item in items` |
| `while ready`, `until ready` | `repeat while ready`, `repeat until ready` |
| `call greet with "Alex"` | `greet with ("Alex")` |
| `call greet` | `greet with ()` |
| Default/named argument `as` | `be` |
| `f"Hello, {name}"` | `"Hello, {name}"` (interpolation planned) |
| `bring maths`, `bring square as sq from maths` | `import maths`, `import square named sq from maths` |
| `define Child from Parent`, `setup`, `create`, `my`, `parent` | Structured `define` types, implicit `self`, `Type with (...)`; no inheritance |
| `always` exception cleanup | `finally` |
| `items[0]`, `items at 0` | `item at 0 in items` |
| `remove every x from items` | `remove all x from items` |
| `==`, `!=`, `<`, `>=` as comparison alternatives | Word-based comparison forms are the documented grammar |
| Tuple and set literals as core collection types | Deferred pending separate design |
| Uninitialized `let x` | Initializers required in current design |
| `outer`, `with x give ...`, legacy variadics | Deferred pending deliberate redesign |

### Historical origin

Eloque grew out of the earlier BakeScript culinary-syntax concept. That transition is historical context, not a competing active grammar. Additional old mappings remain available in Git history.

## Decisions requiring refinement

These are future tasks, **not unfinished M1 requirements**:

- **Lexing and parsing:** Extend token coverage, grammar, source spans, and conformance fixtures for future features; specify comparison/Boolean precedence and associativity. Preserve existing reserved-keyword rules and arithmetic parser behavior.
- **Diagnostics:** Implement the approved shared diagnostic contract and controlled multi-error recovery before M2 resolution. Preserve M1/H2 behavior as a historical implementation baseline until the migration is validated.
- **Types and numbers:** Resolve mixed `Int`/`Float` arithmetic, conversion and overflow edge cases, exceptional floating-point values, type narrowing, generic compatibility, and safe mutation analysis.
- **Collections and text:** Define iteration and mutation details, including Unicode code-point versus grapheme-cluster indexing.
- **Language features:** Specify exception taxonomy and cleanup precedence, module resolution, structural-interface diagnostics, and pattern exhaustiveness.
- **Execution and tooling:** Define CLI behavior, sandbox budgets, and concurrency/ownership semantics as their implementation milestones approach.
- **VM architecture:** Benchmark a functioning stack VM before reconsidering a register VM. Keep alternative backends and runtime optimization proposals as tracked future decisions rather than assuming they are adopted.
- **Frontend representation and optimization:** Preserve H2's arena-backed flat AST as the production baseline. Revisit arena capacity and growth policies only when representative profiling demonstrates a meaningful opportunity. Evaluate any proposed optimization against correctness, allocation overhead, retained memory, small-program performance, and implementation complexity. The review is tracked under TB-025 in the technical backlog.

## M2 architecture decisions — October 2026

**Status:** Approved design, not yet implemented.

The decisions below establish the architectural contract for Eloque's diagnostic migration and M2 name-resolution stage. They supersede earlier design choices where explicitly identified. Implementation details may be refined through testing, but changes to agreed semantics must be documented.

### D-041 — Dedicated name-resolution stage

**Status:** Decided; implementation pending.

Name resolution is a dedicated semantic-analysis pass performed after syntactic validation and before mandatory static type checking.

The resolver must not mutate the AST. It produces a separate resolution result containing symbol definitions and bindings between identifier references and declarations.

Variable declarations are visible from their declaration point onward. Future function and named-type declarations may support forward references through a declaration prepass within the appropriate scope.

The precise interaction between forward declarations, module imports, and visibility will be specified when those language features are introduced.

### D-042 — Lexical scopes and shadowing

**Status:** Decided; implementation pending.

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

**Status:** Decided; implementation pending.

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

**Status:** Decided implementation direction; performance subject to validation.

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

**Status:** Decided initial implementation; representation remains private.

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

**Status:** Decided; implementation pending.

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

**Status:** Decided; implementation pending.

Each compiler stage owns its diagnostic collection.

There is no global mutable diagnostic collector.

The compilation driver coordinates a shared error budget and merges stage diagnostics deterministically.

A configurable overall diagnostic limit is required. A default of 100 errors is an initial candidate, subject to validation rather than a permanent language-level guarantee.

Diagnostics should be ordered by source location, with stable tie-breaking for coincident locations.

Independent errors should be collected where safe. Diagnostics caused solely by an earlier error should be suppressed where practical.

Warnings are nonblocking by default.

Successful compiler paths should avoid unnecessary diagnostic allocations. Optional diagnostic slices remain nil when unused.

### D-048 — Recovery and compilation eligibility

**Status:** Decided; implementation pending.

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

**Status:** Approved implementation prerequisite.

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
7. Implement the baseline stack-based VM.
8. Establish execution benchmarks before reconsidering major runtime architecture choices.

No stage should be considered implemented solely because its design is recorded here.

See `ARCHITECTURE.md` for the resolver algorithm, test matrix, and acceptance criteria, and `TECHNICAL_BACKLOG.md` for deferred reevaluation triggers.

## Change process

When changing a language-level decision:

1. Update [LANGUAGE_SPEC.md](LANGUAGE_SPEC.md) first.
2. Amend the corresponding decision here and explicitly identify superseded behavior.
3. Update affected examples, architecture, roadmap, technical backlog, and conformance tests.
4. Distinguish agreed design from implemented, tested behavior; record benchmark evidence and reevaluation triggers for performance-sensitive choices.

**A decision's presence in this register does not mean its feature is executable.**
