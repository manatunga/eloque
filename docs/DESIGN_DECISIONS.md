# Eloque Design Decisions

**Status:** Active design decision register, updated October 2026.
**Normative language specification:** [LANGUAGE_SPEC.md](LANGUAGE_SPEC.md).

This register records accepted design choices, provisional directions, deferred work, and selected implementation decisions. **Decided** means agreed, not necessarily implemented. **Provisional** requires further specification or validation. **Deferred** is deliberately postponed. **Implemented (M1)** identifies behavior or architecture established in the completed Milestone 1 frontend; it does not imply that Eloque programs can execute.

## Milestone 1 implementation baseline

M1 delivered a Go lexer, parser, AST construction, structured fail-fast lexical and syntax diagnostics, and the canonical `Parse(sourceText)` entry point. The supported initial subset includes literals, identifiers, arithmetic expressions, and `let`, `change`, and `say` statements. The canonical parser pulls tokens incrementally from the lexer; buffered parsing is retained for regression testing and benchmarks. **Static type checking, bytecode generation, and execution are not yet implemented.**

M1-P5 comparisons found streaming faster than buffered parsing in **19 of 20** integrated cases and lower in allocated bytes in **20 of 20** cases. Controlled parsing comparisons against Go and CPython are frontend-only measurements; they do not establish runtime performance parity. See [benchmarks/README.md](../benchmarks/README.md) for methodology and results.

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
| D-010 | Decided; control flow pending | `if`, `otherwise if`, and `otherwise`; lexical block scopes; no shadowing of a visible enclosing binding. |
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
| D-034 | Decided; lexer implemented (M1) | Decimal integer literals prohibit leading zeros except `0`. A digit sequence immediately followed by an ASCII letter or underscore is a lexical error; invalid-suffix diagnostics take precedence over leading-zero diagnostics. Lexical errors carry source positions and fail fast; multi-error recovery is deferred. |
| D-035 | Implemented (M1) | Go lexer token kinds use a `uint8`-backed enumeration with zero reserved for `TokenInvalid`; `String()` supplies readable names. Incremental token emission and reused indentation-stack/pending-queue storage reduce allocations. This is an implementation choice, not a language semantic rule. |
| D-036 | Decided; lexer implemented (M1) | `change`, `to`, `with`, and `giving` are reserved keywords with dedicated token kinds, not ordinary identifiers. One token kind may serve multiple grammatical contexts. |
| D-037 | Implemented (M1) | Syntax parsing fails at the first error and returns a structured diagnostic with message, source span, and optional expected/actual token metadata. Malformed input must not be accepted as a valid program. Multi-error recovery may be reconsidered after synchronization and diagnostic-quality criteria are defined. |
| D-038 | Implemented (M1) | The canonical `Parse(sourceText)` frontend consumes lexer tokens incrementally via `tokenSource`, without a redundant independent token cursor. Buffered parsing remains for regression tests and benchmarks. Future changes must preserve correctness, diagnostic quality, and reproducible performance baselines. |

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
- **Diagnostics:** Define machine-readable diagnostic schema and criteria for any future multi-error recovery; retain M1's fail-fast baseline until a replacement is validated.
- **Types and numbers:** Resolve mixed `Int`/`Float` arithmetic, conversion and overflow edge cases, exceptional floating-point values, type narrowing, generic compatibility, and safe mutation analysis.
- **Collections and text:** Define iteration and mutation details, including Unicode code-point versus grapheme-cluster indexing.
- **Language features:** Specify exception taxonomy and cleanup precedence, module resolution, structural-interface diagnostics, and pattern exhaustiveness.
- **Execution and tooling:** Define CLI behavior, sandbox budgets, and concurrency/ownership semantics as their implementation milestones approach.
- **VM architecture:** Benchmark a functioning stack VM before reconsidering a register VM. Keep alternative backends and runtime optimization proposals as tracked future decisions rather than assuming they are adopted.

## Change process

When changing a language-level decision:

1. Update [LANGUAGE_SPEC.md](LANGUAGE_SPEC.md) first.
2. Amend the corresponding decision here and explicitly identify superseded behavior.
3. Update affected examples, architecture, roadmap, technical backlog, and conformance tests.
4. Distinguish agreed design from implemented, tested behavior; record benchmark evidence and reevaluation triggers for performance-sensitive choices.

**A decision's presence in this register does not mean its feature is executable.**
