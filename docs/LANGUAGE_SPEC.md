# Eloque Language Specification

**Status:** Active language design specification, revised October 2026. Milestone 1 and the post-M1 H2 frontend optimization are complete. The Go lexer and parser support the implemented initial syntax subset, with direct pull-based token streaming, arena-backed flat AST construction, and structured fail-fast diagnostics. Name resolution, static type checking, bytecode compilation, and program execution are not yet implemented. Examples throughout this specification illustrate the target language and are not necessarily supported by the current frontend.
**Language name:** Eloque
**Source file extension:** `.elq` (selected)
**Planned CLI command:** `eloque` (subject to availability checks)
**Implementation architecture:** Go-based compiler and runtime.

The lexer, parser, and arena-backed flat AST construction are implemented. The canonical `Parse(sourceText)` entry point returns `(*ast.FlatProgram, error)`. Name resolution, static type checking, bytecode compilation, and the stack-based virtual machine are planned.

**Status labels:**

- **Decided:** An agreed language rule, not necessarily implemented.
- **Provisional:** A design direction is accepted, but details remain unresolved.
- **Deferred:** A feature or decision is reserved for a later milestone.
- **Implemented:** A feature exists in the current codebase and has been validated by tests.

Unless explicitly identified as implemented, features described in this specification represent the target language rather than the current frontend's capabilities.

The first executable Eloque release will support a narrower subset than the complete specification.

## 1. Principles [Decided]

Eloque is a general-purpose programming language with structured conversational syntax, not unrestricted natural-language parsing. It prioritizes novice readability, unambiguous static semantics and machine-readable diagnostics for AI-assisted development, and a foundation for safe native concurrency. Prefer concise, regular grammar over excessive English verbosity. Execution is deterministic subject to explicit external inputs and eventual concurrency semantics.

## Current implementation coverage (M1 + H2)

The completed M1 frontend supports lexical analysis, parsing,
and AST construction for the following syntax subset:

- Decimal integer, string, and boolean literals.
- Identifiers and arithmetic expressions, including grouping,
  unary operators, precedence, and associativity.
- `let` declarations with initializers.
- `change` assignment statements.
- `say` output statements.
- Complete source programs containing supported statements.

The frontend also provides source-position tracking, lexical
error propagation, and structured fail-fast syntax diagnostics.

H2 changed the frontend's internal AST representation and parser implementation without changing the supported M1 grammar. The production frontend uses an arena-backed flat AST and direct lexer token streaming. The original M1 parser remains available internally for differential testing and historical performance comparisons.

The implementation does not yet provide name resolution,
static type checking, bytecode compilation, or execution.

Other syntax and semantics described below remain language
design specifications unless explicitly marked as implemented.

## 2. Lexical rules [Decided]

- Significant indentation uses spaces, with widths aligned to multiples of four; tabs are invalid. An increase to any greater four-space-aligned width introduces one indentation level (`INDENT`), regardless of the size of the increase. A decrease must return to a previously active indentation width, emitting one `DEDENT` for each exited level. A `:` introduces an indented block. Blank and comment-only lines do not affect indentation.
- One statement per logical line. Newlines and indentation are insignificant inside matched parentheses `(...)`, allowing multiline expressions and inline comments. Implicit continuation inside `[...]` and `{...}` is planned for their respective syntax implementations.
- `#` starts a line comment outside a string.
- Identifiers use ASCII letters, digits, and `_`, cannot start with a digit, and are case-sensitive. Reserved keywords cannot be identifiers. Snake_case is the recommended naming convention.
- Strings may use single (`'`) or double (`"`) quotes. The matching quote terminates the literal. The lexer recognizes the escapes `\\`, `\"`, `\'`, `\n`, `\t`, and `\r`, preserving their original source spelling in token lexemes. String interpolation using `{expression}` is part of the planned language syntax but is not supported by the completed M1 frontend. Its parsing rules, expression restrictions, and Unicode behavior require further specification. Precise interpolation and Unicode grapheme rules remain to be specified.
- Multiword operators are parsed from individual keyword tokens, e.g. `is greater than`.
- Decimal integer literals consist of ASCII digits. Leading zeros are forbidden except for the literal `0` itself. A digit sequence immediately followed by an ASCII letter or underscore is a malformed numeric literal and is rejected during lexical analysis (e.g. `123abc`, `42_name`). When both an invalid identifier suffix and leading zeros occur (e.g. `007abc`), the invalid-suffix diagnostic takes precedence.

## 3. Types and literals [Decided unless noted]

| Type | Meaning | Examples |
|---|---|---|
| `Int` | Signed 64-bit integer; checked overflow | `42`, `-7` |
| `Float` | IEEE 754 binary64 | `3.14`, `0.5` |
| `String` | Immutable Unicode text, stored as UTF-8 | `"Hello"` |
| `Bool` | Exactly `true` or `false` | `true` |
| Optional `T?` | Value of type `T` or absence | `String?` |

`none` is the absence literal, valid in optional contexts; it is not a general substitute for a non-optional value. `List[T]`, `Map[K, V]`, and user-defined types are statically checked. Literal examples: `[1, 2, 3]`, `{"name": "Alex"}`. Tuple and set types are **deferred**, not established core syntax.

- Integer overflow raises `IntegerOverflow`; integer division by zero raises an error.
- Float arithmetic is approximate; use the shortest round-trippable decimal when displaying a Float. Rounding is explicit.
- Implicit `Int` to `Float` promotion is permitted only when exactness can be established. Precise rules for value-dependent conversions and Float exceptional values remain provisional.
- No implicit truthiness; conditions require `Bool`.

## 4. Bindings, mutability, scope [Decided]

```elq
let age be 20
change age to 21
let fixed max_attempts as Int be 3
let frozen names as List[String] be ["Alice", "Bob"]
let nickname as String? be none
```

The M1 frontend parses `let` declarations and `change ... to ...`
assignments, but does not yet enforce binding types or mutability.

The target language defines `let` as a mutable binding, `fixed` as preventing rebinding without necessarily freezing the referenced value, and `frozen` as enforcing deep immutability. The `fixed` and `frozen` modifiers are not implemented in M1.

The M1 frontend requires initializers for `let` declarations. Support for declarations without initializers remains a separate language-design decision and is not implemented. Local binding types may be inferred statically from their initializers. Block scopes are lexical. **Shadowing is prohibited** when an enclosing declaration of the same name is visible; unrelated scopes may reuse names.

Lists, maps and structured values use **value semantics**: assignment creates an independent *logical* value. The runtime may optimize with copy-on-write or structural sharing without observable aliasing. Explicit references/shared mutable state are deferred to the concurrency design.

## 5. Expressions [Decided unless noted]

Arithmetic: `+`, `-`, `*`, `/` (real division), `//` (floor division), `%` (floor-compatible remainder). For example, `-5 // 2` is `-3` and `-7 % 3` is `2`. Division by zero raises an exception. Exact rules for mixed numeric operands require conformance tests.

**Implementation status:** The M1 frontend parses the supported arithmetic operators and constructs expression ASTs according to their precedence and associativity rules.

Arithmetic evaluation, numeric type compatibility, overflow checking, division-by-zero handling, and runtime exceptions are not yet implemented. These remain semantic requirements for the future type checker and runtime.

Comparisons: `is`, `is not`, `is greater than`, `is less than`, `is at least`, `is at most`. Structural equality is defined for compatible lists, maps, and custom types where member equality is supported. Chained comparisons are **not supported in v0.1**.

Boolean operations: `not`, `and`, `or`, with short-circuit evaluation. `not age is at least 18` means `not (age is at least 18)`.

**Implementation status:** Comparison operators (`is`, `is not`, and the conversational ordering operators) and Boolean operators (`not`, `and`, `or`) are part of the target language design.

The M1 frontend does not yet parse these operators. Their precedence, type constraints, and short-circuit behavior are specified for future implementation and conformance testing.

**Expression precedence, highest to lowest:**

1. Grouping, function calls, and member access.
2. Exponentiation (`**`), right-associative.
3. Unary numeric plus and minus (`+`, `-`).
4. Multiplication, real division, floor division, and remainder (`*`, `/`, `//`, `%`), left-associative.
5. Addition and subtraction (`+`, `-`), left-associative.
6. Ordering comparisons (`is greater than`, `is less than`, `is at least`, `is at most`).
7. Equality comparisons (`is`, `is not`).
8. Boolean negation (`not`).
9. Boolean conjunction (`and`).
10. Boolean disjunction (`or`).

Exponentiation binds more tightly than unary plus or minus on its left, but accepts unary operators on its right. Thus, `-2 ** 2` means `-(2 ** 2)`, `2 ** -3` means `2 ** (-3)`, and `2 ** 3 ** 2` means `2 ** (3 ** 2)`.

Unary plus preserves a numeric operand's value. Unary minus negates it. Both require numeric operands.

Chained comparisons are invalid in v0.1, including chains that mix ordering and equality operators. Parentheses may be used to compare the result of another comparison.

Boolean `not` associates to the right, so `not not ready` means `not (not ready)`. Boolean `and` and `or` short-circuit, evaluating operands from left to right. Other binary arithmetic operands and function-call arguments evaluate from left to right.

Function calls are part of the decided language design but are not supported by the completed M1 frontend.

The planned syntax uses `with (...)`, including zero-argument calls. Positional arguments must precede named arguments, duplicate named arguments are invalid, and argument expressions may contain nested calls.

When implemented, the parser will enforce argument ordering and reject duplicate named arguments. The static checker will validate parameter compatibility.

Multiword comparison operators are recognized by the parser from individual lexical tokens. Incomplete operator phrases produce syntax errors.

## 6. Conditions and loops [Decided]

```elq
if age is at least 18:
    say "Adult"
otherwise if age is at least 13:
    say "Teen"
otherwise:
    say "Child"

for each name in names:
    say name

repeat 3 times:
    say "Again"

repeat while attempts is less than 3:
    change attempts to attempts + 1

repeat until ready:
    stop
```

In this example, `attempts` and `ready` are assumed to have been declared earlier. `repeat while` and `repeat until` test their condition before each iteration. `stop` exits the innermost loop; `skip` continues to its next iteration. Collection mutation during iteration and repeat-count edge cases remain provisional.

**Implementation status:** Conditional statements (`if`, `otherwise if`, `otherwise`) and iteration constructs (`for each`, `repeat`, `repeat while`, and `repeat until`) are part of the target language design but are not implemented in the M1 frontend.

Their parsing, control-flow validation, and runtime behavior will be introduced and tested in subsequent milestones.

## 7. Functions [Decided]

```elq
to calculate_total with (price as Float, quantity as Float) giving Float:
    give price * quantity

to greet with (name as String, excited as Bool be false):
    say "Hello, {name}!"

let total be calculate_total with (20.0, 3.0)
greet with (name be "Alex")
greet with ("Alex", excited be true)
```

Function definitions, parameter declarations, return statements, and function calls are part of the decided language design but are not implemented in the M1 frontend.

The planned syntax uses `to` to define functions and `give` to return values. Parameters require explicit type annotations. Functions returning a value require `giving Type`; omitting it defines a function with no value return, where `give value` is invalid.

Calls use `with (...)`, including zero-argument calls. Named arguments use `be`, as do parameter defaults. Arguments are evaluated from left to right.

Function names conventionally start with verbs, but this is a style guideline rather than a grammar requirement.

First-class functions, closures, variadics and generics require further specification; older `call`, `outer`, and anonymous `with x give ...` forms are **not** grandfathered into the new grammar.

## 8. Collections and text [Decided unless noted]

```elq
let names be ["Alice", "Bob", "Charlie"]
say item at 0 in names
say item at -1 in names
let missing_name as String? be find item at 99 in names
say slice of names from 0 to 2
say size of names
add "Dana" to names
remove "Bob" from names
remove all "Alice" from names

let data be {"name": "Alex"}
say value for "name" in data
let missing_value as String? be find value for "missing" in data
add key "city" with value "Colombo" to data
change value for "city" in data to "Kandy"
remove key "city" from data
```

Indexing is zero-based; negative indices count backward from the end (`-1` is last). Slices include the start and exclude the end. Direct out-of-range indexing raises `IndexOutOfRange`; missing map keys raise `KeyNotFound`. `find` variants return an optional result (`none` when absent). List removal removes the first matching element; `remove all` removes all matches. Map equality is independent of entry order. String concatenation uses `+`; `size of` and slicing apply to strings as well. Unicode character versus grapheme indexing details remain provisional.

**Implementation status:** Collection literals, indexing, slicing, map operations, and collection mutation are part of the target language design but are not supported by the completed M1 frontend.

The M1 lexer recognizes the supported string literal syntax, but string indexing, slicing, interpolation, and collection operations require future parser and runtime implementation.

Collection value semantics, bounds checking, and optional lookup behavior remain requirements for subsequent milestones.

## 9. Custom types and interfaces [Decided design; implementation deferred]

```elq
define Person:
    name as String
    age as Int

    to greet with ():
        say "Hello, I'm {self.name}!"

let alice be Person with (name be "Alice", age be 25)
alice.greet with ()
change alice.age to 26

define interface Speaker:
    to speak with () giving String
```

`self` is implicit inside methods. Interfaces are satisfied **structurally and implicitly** by compatible method signatures. Traditional class inheritance, `setup`, `create`, `my`, and `parent` from the older draft are superseded. Field initialization, visibility diagnostics, and mutation rules for frozen instances require additional tests.

**Implementation status:** Custom type definitions, methods, field access, constructors, and interfaces are not implemented in the M1 frontend.

Structural interface compatibility, implicit `self`, member visibility, and instance mutation rules describe the intended language semantics. Their parser, static-checker, and runtime implementations are deferred to future milestones.

## 10. Pattern matching [Decided design; implementation deferred]

```elq
match status:
    when 200:
        say "OK"
    when 404:
        say "Missing"
    otherwise:
        say "Other"
```

Only the first matching branch executes; there is no fallthrough. `when value named name:` binds the matched value. Statement matches may omit `otherwise`; value-producing matches must be exhaustive. The expression form and exhaustiveness algorithm remain to be specified.

**Implementation status:** Pattern matching is not implemented in the M1 frontend.

The `match`, `when`, and `otherwise` constructs describe the intended language syntax and semantics. Pattern parsing, exhaustiveness checking, value-producing matches, and runtime branch selection are deferred to future milestones.

## 11. Exceptions [Decided]

```elq
try:
    raise DivisionError with ("Cannot divide by zero")
catch DivisionError named exc:
    say "Invalid calculation"
catch error named exc:
    say "Unexpected error"
finally:
    say "Cleanup"
```

At least one `catch` or `finally` is required. Uncaught exceptions propagate after `finally`; `finally` also runs on normal exits and `give`. Exceptions are **unchecked**: functions need not declare their possible exceptions. Static tooling may warn about potentially unhandled failures. Detailed exception hierarchy and precedence when `finally` itself fails remain provisional.

**Implementation status:** Exception handling is not implemented in the M1 frontend.

The `try`, `catch`, `finally`, and `raise` constructs describe the intended language syntax and semantics. Exception propagation, handler matching, cleanup guarantees, and runtime error types require future parser, static-checker, and runtime implementation.

Lexical and syntax errors produced by the existing frontend are compiler diagnostics, not Eloque runtime exceptions.

## 12. Modules and entry point [Decided]

```elq
import math_utils
import square from math_utils
import square, cube from math_utils
import math_utils named math
import square named squared from math_utils
```

A `.elq` file can execute top-level statements in source order; no automatic `main` invocation is required. An imported module initializes once per program execution, with dependency initialization in defined order. Circular imports are rejected in the initial implementation. Top-level declarations are public by default, except names beginning `_`, which are **compiler-enforced module-private**. Precise module lookup paths and packaging are deferred.

**Implementation status:** Module imports, dependency resolution, module initialization, and module-level visibility enforcement are not implemented in the M1 frontend.

The current parser can construct an AST for supported top-level statements, but cannot execute them or load imported modules.

The `.elq` source extension and the documented import forms represent the intended module system. Import ordering, circular dependency detection, and module-private visibility require future compiler and runtime implementation.

## 13. Execution, safety and diagnostics [Decided direction]

Normal execution is sequential by default. Assignment evaluates its right-hand side before changing the target. Runtime failures are typed and include source locations. Human-readable diagnostics are the default; structured machine-readable diagnostics (e.g. JSON) are a design requirement for agent tooling. The runtime should support configurable instruction/time/memory limits for sandboxed execution; normal CLI execution has no arbitrary universal instruction cap. Native safe concurrency is a language goal, but syntax, ownership and shared-state rules are **deferred**.

**Lexical diagnostics [Implemented in M1]:** The production lexer uses fail-fast error handling, stopping at the first detected lexical error. Diagnostics identify the source line and column and describe the invalid input.

**Runtime implementation status:** Eloque does not yet have an executable runtime. Sequential execution, typed runtime failures, assignment evaluation semantics, and configurable execution limits are requirements for future implementation.

Native concurrency, ownership rules, and shared-state synchronization remain deferred design work.

The canonical `Parse(sourceText)` frontend propagates lexical errors through the parser while preserving their source locations. Lexical errors are compiler diagnostics, not runtime exceptions.

Error recovery and reporting multiple lexical errors in a single pass remain potential future enhancements.

## 14. Tooling and backend [Provisional roadmap]

Planned commands include `eloque run`, `eloque check`, `eloque fmt` and eventually a REPL. The Go implementation currently provides the lexer, parser, and AST construction stages through the canonical `Parse(sourceText)` frontend. Mandatory static type checking, bytecode compilation, and stack-based VM execution are planned for subsequent milestones. The completed M1 frontend supports only a subset of the specified syntax. The first executable release will extend that subset with the semantic analysis, compilation, and runtime functionality required to execute Eloque programs.

## 15. Remaining specification work

Milestone 1 established the foundational lexer, parser, AST, and source diagnostics for Eloque's initial syntax subset.

The following areas require further specification or validation during subsequent milestones:

- **Lexical rules:** Formal token and escape-sequence reference tables, alongside rules for future lexical features.
- **Expressions and grammar:** Precedence and associativity for additional operators, and grammar extensions for control flow,functions, collections, and other planned constructs.
- **Numeric semantics:** Type conversions, mixed numeric operations, overflow behavior, and floating-point edge cases.
- **Diagnostics:** Structured diagnostic schemas, error classifications, and potential error-recovery mechanisms.
- **Collections and text:** Iteration behavior, mutation rules, indexing, slicing, and Unicode character and grapheme semantics.
- **Type system:** Type compatibility, inference, narrowing, generics, and related static-analysis rules.
- **Exceptions:** Exception hierarchy, propagation, handler matching, and cleanup guarantees.
- **Modules and interfaces:** Module resolution, initialization, visibility, structural compatibility, and interface validation.
- **Concurrency and ownership:** Task execution, shared-state semantics, synchronization, and memory-safety guarantees.

These items are tracked as future specification and implementation work. They do not prevent Milestone 1 from being considered complete.

Each feature must have its relevant semantics defined and validated before being considered stable.
