# Eloque Design Decisions

**Status:** Active design decision register, updated October 2026.
**Source of truth for current language behavior:** [LANGUAGE_SPEC.md](LANGUAGE_SPEC.md).  
**Interpretation:** *Decided* is an agreed design, **not** an implemented feature. *Provisional* needs further specification; *Deferred* is intentionally outside the first executable subset. The historical record below is retained for traceability.

## Current decisions

| ID | Status | Decision |
|---|---|---|
| D-001 | Decided | Implement Eloque in Go, initially using a lexer, parser, AST, mandatory static checker, bytecode compiler, and stack-based VM. |
| D-002 | Provisional | Eloque, `eloque`, and `.elq` are working names; naming/conflict checks remain. |
| D-003 | Decided | General-purpose language with concise structured conversational syntax, not unrestricted English. Prioritize novice readability, predictable agent tooling, and a foundation for safe concurrency. |
| D-004 | Decided | Four-space significant indentation, `:` blocks, no tabs, `#` comments, case-sensitive ASCII identifiers, and double-quoted strings. |
| D-005 | Decided | `let x be value` declares a mutable binding; `change x to value` reassigns. Initialized local bindings may infer types. |
| D-006 | Decided | `let fixed x be value` prevents rebinding; `let frozen x be value` provides deep immutability. |
| D-007 | Decided | `true`, `false`, `none`; `none` belongs in optional `T?` contexts, not arbitrary non-optional values. |
| D-008 | Decided | Strict static typing, with `as Type` annotations and local inference; checker errors block normal compilation. No implicit truthiness. |
| D-009 | Decided | Function parameters require explicit types; value-returning functions require `giving Type`. Omission means no value return. |
| D-010 | Decided | `if`, `otherwise if`, `otherwise`; lexical block scopes; no shadowing of a visible enclosing binding. |
| D-011 | Decided | `for each`, `repeat N times`, `repeat while`, `repeat until`; `skip` and `stop`. |
| D-012 | Decided | `to name with (parameters) giving Type:` declares functions; `name with (arguments)` invokes them, including zero-argument calls. `give` returns. |
| D-013 | Decided | Named arguments and defaults use `be`; function names are conventionally verb-first, not enforced. |
| D-014 | Deferred | Closures, first-class function details, variadics, and generics need a new grammar; prior `outer`/anonymous-function syntax is not automatically retained. |
| D-015 | Decided | Parentheses group expressions; evaluation of operands and arguments is left-to-right, except Boolean short-circuiting. |
| D-016 | Decided | `define` introduces structured custom types with fields and methods; implicit `self`; structural interfaces. No traditional class inheritance. |
| D-017 | Decided | Typed unchecked exceptions: `raise`, `try`, `catch`, `finally`; at least one catch or finally. |
| D-018 | Decided | `import` for modules; `named` for aliases. Modules initialize once, circular imports rejected initially, underscore-prefixed top-level names are compiler-private. |
| D-019 | Decided | `say` output; string interpolation uses `"Hello, {name}"`. Input API remains to be specified. |
| D-020 | Decided | Lists/maps support conversational access and mutation; `remove all` removes all matching list values; `find` access returns optional values. |
| D-021 | Decided | Word-based comparison operators (`is`, `is not`, `is greater than`, etc.); structural equality where supported. Chained comparisons excluded from initial grammar. |
| D-022 | Decided | Stack VM first; revisit register VM only with a functioning benchmarked baseline. |
| D-023 | Provisional | CPython parity is an aspirational benchmark, not a guarantee. |
| D-024 | Provisional | Backend extensibility, standard library, packaging, and post-1.0 compatibility policy require further work. |
| D-025 | Decided | Zero-based indexing, negative indices from the end, end-exclusive slices, and typed failures for invalid direct access. Unicode indexing units remain provisional. |
| D-026 | Decided | `Int` is checked signed 64-bit; `Float` is IEEE 754 binary64; display shortest round-trippable decimal; no automatic rounding; implicit numeric promotion only when provably exact. Conversion edge cases remain provisional. |
| D-027 | Decided | `/` performs real division; `//` floors; `%` uses floor-compatible remainder. Division by zero fails. |
| D-028 | Decided | Lists, maps, and custom structured values have logical value semantics. Copy-on-write/structural sharing may implement this without observable aliasing. |
| D-029 | Deferred | Explicit references, shared mutable state, and concurrency syntax are designed together later. |
| D-030 | Decided | Pattern matching uses `match`, `when`, `otherwise`; no fallthrough. Statement matches may be non-exhaustive; value-producing matches must be exhaustive. |
| D-031 | Decided | Top-level statements execute in source order; `main` is not automatically invoked. |
| D-032 | Decided | Normal execution is sequential; runtime exceptions include source locations; structured diagnostics are required for tooling. |
| D-033 | Decided | Configurable sandbox limits are supported by design; no arbitrary universal instruction limit for normal CLI runs. |
| D-034 | Decided | Decimal integer literals prohibit leading zeros except for `0`. A digit sequence immediately followed by an ASCII letter or underscore is a lexical error; invalid-suffix diagnostics take precedence over leading-zero diagnostics. The initial lexer reports source-positioned errors and stops at the first lexical error; multi-error recovery is deferred. |
| D-035 | Decided (implementation) | The Go lexer represents token kinds using a `uint8`-backed enumeration, reserving zero for `TokenInvalid`. A `String()` method preserves readable token names. The lexer emits tokens incrementally and reuses indentation-stack and pending-token-queue storage to reduce allocations. This is an internal representation decision, not a language-level semantic rule. |
| D-036 | Decided | `change`, `to`, `with`, and `giving` are reserved keywords represented by dedicated lexer token kinds. They cannot be used as ordinary identifiers. A reserved token may have different grammatical roles depending on context; for example, `to` appears in assignment and function-definition syntax. This does not require separate token kinds for each grammatical use. |
| D-037 | Decided (initial implementation) | The initial parser uses fail-fast syntax diagnostics: parsing terminates at the first syntax error and returns a structured diagnostic containing its message, source span, and optional expected/actual token metadata. Malformed input must not be accepted as a successfully parsed program. This approach prioritizes correctness, deterministic diagnostics, and simpler parser control flow during M1. Multi-error recovery is deferred rather than rejected permanently; it may be introduced later when statement-boundary synchronization, nested-expression recovery, and diagnostic quality can be tested systematically. |
| D-038 | Decided (implementation) | The optimized canonical
lexer and parser form the baseline for subsequent frontend
development. Historical benchmark results are preserved,
and future performance changes must be evaluated against
correctness, diagnostic quality, and reproducible measurements.
This decision does not alter language-level semantics. |

## Superseded syntax and semantics

These examples are **historical only**. Do not use them as current lexer/parser fixtures.

| Earlier Eloque draft | Current direction |
|---|---|
| `x = 2`, `x += 1` | `change x to 2`; compound assignment syntax not retained |
| `fix x at 1` | `let fixed x be 1` or `let frozen x be 1` |
| `nothing` | `none` |
| `name: int`, `name of type int` | `name as Int` |
| Optional gradual type checker | Mandatory static type checker with inference |
| `else if` | `otherwise if` |
| `each item in items` | `for each item in items` |
| `while ready`, `until ready` | `repeat while ready`, `repeat until ready` |
| `call greet with "Alex"` | `greet with ("Alex")` |
| `call greet` | `greet with ()` |
| Default/named argument `as` | `be` |
| `f"Hello, {name}"` | `"Hello, {name}"` |
| `bring maths`, `bring square as sq from maths` | `import maths`, `import square named sq from maths` |
| `define Child from Parent`, `setup`, `create`, `my`, `parent` | Structured `define` types, implicit `self`, `Type with (...)`; no inheritance |
| `always` exception cleanup | `finally` |
| `items[0]`, `items at 0` | `item at 0 in items` |
| `remove every x from items` | `remove all x from items` |
| `==`, `!=`, `<`, `>=` as comparison alternatives | Word-based comparison forms are the documented grammar |
| Tuple and set literals as core collection types | Deferred pending a separate design |
| Uninitialized `let x` | Initializers required in current design |
| `outer`, `with x give ...`, legacy variadics | Deferred; require deliberate redesign |

### Historical origin

The project began with BakeScript-inspired culinary syntax and then moved to Eloque's conversational style. That earlier transition remains historical context, not a second set of active language rules. The repository's previous version of this file records additional BakeScript-to-Eloque mappings; retain them in Git history rather than mixing them into the active grammar.

## Decisions requiring implementation-time refinement

- Extend lexical coverage as later language features require it; preserve the documented reserved-keyword rules.
- Implement the established expression precedence and associativity rules, including right-associative exponentiation and its interaction with unary operators.
- Complete parser grammar, source-span construction, and syntax conformance fixtures.
- Revisit multi-error syntax recovery only when its implementation is explicitly scheduled.
- `Int`/`Float` mixed arithmetic, conversions, exceptional floating-point values, and overflow behavior for every operator.
- Type narrowing, generic type compatibility, and safe mutation analysis.
- Unicode indexing (code points versus grapheme clusters), collection iteration and mutation rules.
- Exception taxonomy, cleanup precedence, and module resolution.
- Structural interface diagnostics, pattern exhaustiveness, and concurrency/ownership.
- CLI command behavior, machine-readable diagnostic schema, and sandbox budget mechanics.

## Change process

When a decision changes:

1. Update `LANGUAGE_SPEC.md` first.
2. Amend the relevant row here, marking superseded syntax explicitly.
3. Update README examples, architecture, roadmap, backlog, and conformance tests where affected.
4. Distinguish a design agreement from a tested implementation.

No entry in this register implies the feature already runs.
