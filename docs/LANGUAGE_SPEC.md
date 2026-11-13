# Eloque Language Specification

> **Status:** Active language design specification, reviewed November 2026. Milestone 1, post-M1 H2 frontend optimization, post-H2 structured diagnostic recovery, and M2-P1 lexical name resolution are implemented. The Go frontend supports the initial syntax subset through direct lexer token streaming and arena-backed flat AST construction. Both fail-fast and recovery-capable parsing are available. The dedicated resolver performs lexical name resolution for the supported language subset. Mandatory static type checking, production Monowire bytecode compilation, and program execution are not yet implemented. Examples throughout this specification illustrate the target language and are not necessarily supported by the current compiler.

**Language name:** Eloque
**Source file extension:** `.elq` (selected)
**Planned CLI command:** `eloque` (subject to availability checks)
**Implementation architecture:** Go-based compiler frontend and semantic analysis, targeting the Monowire register-based bytecode VM. Zig is provisionally preferred for the Monowire runtime implementation, with Rust retained as a fallback pending integration and portability validation.

The lexer, parser, and arena-backed flat AST construction are implemented. The canonical `Parse(sourceText)` entry point returns `(*ast.FlatProgram, error)`. Recovery-capable parsing is also implemented through `ParseRecovering(sourceText, ParseOptions)`.

M2-P1 lexical name resolution is implemented for the supported language subset. Mandatory static type checking, Monowire bytecode compilation, and production VM execution remain pending within M2 — First Execution.

**Monowire** is the selected register-based execution architecture. The previous stack-based VM plan has been superseded.

**Status labels:**

- **Decided:** An agreed language rule, not necessarily implemented.
- **Provisional:** A design direction is accepted, but details remain unresolved.
- **Deferred:** A feature or decision is reserved for a later milestone.
- **Implemented:** A feature exists in the current codebase and has been validated by tests.

Unless explicitly identified as implemented, features described in this specification represent the target language rather than the current frontend's capabilities.

The first executable Eloque release will support a narrower subset than the complete specification.

## 1. Principles [Decided]

Eloque is a general-purpose programming language with structured conversational syntax, not unrestricted natural-language parsing. It prioritizes novice readability, unambiguous static semantics and machine-readable diagnostics for AI-assisted development, and a foundation for safe native concurrency. Prefer concise, regular grammar over excessive English verbosity. Execution is deterministic subject to explicit external inputs and eventual concurrency semantics.

## Current implementation coverage (M1 + H2 + post-H2 recovery + M2-P1)

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

Post-H2 work introduced recovery-capable parsing through `ParseRecovering(sourceText, ParseOptions)`. This API collects structured diagnostics, reports whether parsing completed, recovered, or aborted, and may return an AST containing successfully retained statements.

The existing `Parse(sourceText)` entry point continues to provide fail-fast parsing.

Recovery behavior, diagnostic budgets, and expression-depth safeguards are implementation facilities. They do not introduce new language syntax or alter the static typing rules defined by this specification. Recovery results must not be treated as semantically validated or executable programs.

M2-P1 added a dedicated lexical name-resolution stage after parsing. For the supported language subset, the resolver establishes declaration-symbol bindings, resolves identifier expressions and assignment targets, and reports name-resolution errors such as undefined identifiers and duplicate declarations.

The implementation does not yet provide mandatory static type checking, definite-assignment analysis, production Monowire bytecode compilation, or execution.

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

| Type          | Meaning                                                       | Examples      |
| ------------- | ------------------------------------------------------------- | ------------- |
| `Int8`        | Signed 8-bit integer; checked overflow                        | `42`, `-7`    |
| `Int16`       | Signed 16-bit integer; checked overflow                       | `1200`        |
| `Int32`       | Signed 32-bit integer; checked overflow                       | `100000`      |
| `Int`         | Signed 64-bit integer; checked overflow; default integer type | `42`, `-7`    |
| `Float`       | IEEE 754 binary64                                             | `3.14`, `0.5` |
| `String`      | Immutable Unicode text, stored as UTF-8                       | `"Hello"`     |
| `Bool`        | Exactly `true` or `false`                                     | `true`        |
| Optional `T?` | Value of type `T` or absence                                  | `String?`     |

Unsigned integer types are deferred to M3. The canonical default integer type remains `Int`.

`none` is the absence literal, valid in optional contexts; it is not a general substitute for a non-optional value. `List[T]`, `Map[K, V]`, and user-defined types are statically checked. Literal examples: `[1, 2, 3]`, `{"name": "Alex"}`. Tuple and set types are **deferred**, not established core syntax.

Integer overflow is an error. Integer division by zero is an error. Integer arithmetic must not silently wrap, saturate, or widen.

`Float` uses IEEE 754 binary64 arithmetic. Floating-point overflow produces signed infinity. Floating-point division preserves IEEE behavior for infinities, NaN, and signed zero.

Use the shortest round-trippable decimal when displaying a `Float`. Rounding is explicit.

Implicit numeric conversions are restricted to the value-preserving cases defined in the numeric conversion rules. Mixed-width arithmetic and mixed integer/Float arithmetic require explicit conversions.

No implicit truthiness; conditions require `Bool`.

### 3.1 Numeric semantics [Decided; M2-P2 implementation pending]

#### Integer literal typing

An unsuffixed integer literal initially represents an exact compile-time integer constant. It is not immediately rounded or restricted to the default runtime integer type.

Its type is determined using the following rules:

1. A suitable expected integer type determines the literal's type when its exact value fits that type.
2. In a binary arithmetic expression, a typed integer peer supplies the expected type for an unsuffixed integer literal on the other side.
3. An enclosing expected integer type propagates through otherwise-untyped arithmetic expressions.
4. Explicitly typed operands retain their declared types. Expected-type propagation never silently converts an already-typed variable.
5. If no suitable context determines the type, an unsuffixed integer literal defaults to `Int`.
6. A literal that does not fit its selected type is a compile-time error.

For example:

```elq
let small as Int8 be 100
let next be small + 10
let total as Int8 be (10 + 20) * 2
```

The literals adopt `Int8` in these contexts. Each arithmetic operation retains `Int8` semantics and is checked for overflow.

Mixed-width arithmetic between already-typed variables requires explicit conversion.

The minimum signed `Int` value, `-9223372036854775808`, is valid. Implementations must not reject it merely because its positive magnitude cannot first be represented as a signed `Int`.

#### Integer arithmetic

For operands of the same integer type, `+`, `-`, `*`, `//`, `%`, and `**` retain that integer type. No automatic widening occurs.

Integer overflow is an error. The compiler diagnoses overflow whenever it can prove it. Otherwise, the runtime checks the operation. A runtime check may be eliminated only when the compiler proves that overflow is impossible.

- `/` is real division and produces `Float` from integer operands.
- `//` is floor division. For example, `-5 // 2` is `-3`.
- Integer `%` is floor-compatible remainder. For example, `-5 % 2` is `1`. A nonzero remainder has the divisor's sign.
- Division by integer zero is an error.
- The minimum signed integer divided by `-1` using `//` is overflow.
- The corresponding `%` result is zero and must be computed without evaluating an overflowing quotient.

Integer exponentiation requires operands of the same integer type, returns the base type, and requires a non-negative exponent. A negative integer exponent is an error, not an implicit conversion to `Float`. `0 ** 0` is `1`. Overflow is checked. Implementations should use an efficient checked algorithm rather than a linear multiplication loop for large exponents.

Ordinary integer arithmetic never silently wraps, saturates, or widens. Explicit wrapping and saturating operations are not part of the M2 contract.

#### Floating-point arithmetic

`Float` uses IEEE 754 binary64 arithmetic.

Basic arithmetic preserves IEEE behavior for signed zero, infinities, and NaN. Floating-point overflow produces signed infinity rather than an integer-style overflow error.

Floating-point division follows IEEE behavior, including:

- A finite nonzero value divided by positive or negative zero produces the corresponding signed infinity.
- Zero divided by zero produces NaN.
- Signed zero is preserved where required by IEEE arithmetic.

Floating-point `//` and `%` use floor-compatible semantics for finite operands and a finite nonzero divisor:

- `a // b` is the mathematical floor of `a / b`.
- `a % b` is `a - floor(a / b) * b`.
- A nonzero remainder has the divisor's sign.

If either operand is NaN or infinite, or the divisor is zero, these two floor-compatible operators produce NaN. This is a language convention, not the IEEE 754 `remainder` operation.

Floating-point exponentiation follows IEEE-compatible binary64 behavior and established mainstream math-library conventions. In particular:

- `(-2.0) ** 0.5` produces NaN.
- `0.0 ** 0.0` produces `1.0`.
- `0.0 ** -1.0` produces positive infinity.

IEEE 754 does not fully prescribe a general-purpose `pow` function. Eloque's conformance suite defines the required observable cases and tolerances.

#### Numeric conversions

Numeric conversion syntax uses the destination type:

```elq
let narrow be convert small to Int16
let floating be convert count to Float
let widened be convert value to Int
```

Integer-to-integer widening is safe when the entire source type range fits the destination type. Narrowing requires a range check.

If an out-of-range value is known at compile time, compilation fails. Otherwise, the runtime checks it. Integer narrowing never silently wraps, truncates, or saturates.

Integer-to-`Float` conversion rounds according to IEEE 754 binary64 conversion. Mixed integer/Float arithmetic requires explicit conversion, even when a particular integer is exactly representable.

Implicit conversion is permitted for initialization and assignment only when the compiler proves that the conversion preserves the value. Proof may come from the complete source-type range or from a particular known value.

If safety cannot be established, an explicit conversion is required. The same policy applies to function arguments and return values when those features are introduced. This policy does not authorize implicit conversions in unrelated expression contexts.

Float-to-integer conversion is always explicit, even when the compiler knows the Float is an exact integral value in range.

Its conversion order is:

1. Reject NaN and positive or negative infinity.
2. Truncate the finite value toward zero.
3. Check that the truncated result fits the destination integer type.
4. Reject the conversion if it does not fit.

A provably invalid conversion is a compile-time error. Otherwise, the runtime checks it.

Examples:

- `convert 127.9 to Int8` produces `127`.
- `convert -128.9 to Int8` produces `-128`.
- `convert 128.0 to Int8` is invalid.
- `convert -129.0 to Int8` is invalid.

The convert keyword introduces a built-in numeric conversion expression. The source expression is followed by to and a destination numeric type. The operand may contain ordinary arithmetic expressions. Parentheses explicitly group complex or nested conversions. The destination must be one of the supported numeric types. Numeric conversions are dedicated language expressions, not user-defined function calls. General function calls retain the name with (arguments) syntax

#### Mixed numeric comparisons

Comparisons are specified for M3, not part of the M2 executable subset. Their accepted numeric contract is recorded here to prevent later incompatibility.

- Mixed-width signed integer comparisons use exact mathematical comparison.
- Mixed integer/Float comparisons must not first round the integer to binary64 when that would lose information. Comparison is exact with respect to the represented integer and binary64 value.
- NaN comparison behavior follows IEEE 754: equality is false, inequality is true, and ordered comparisons are false, regardless of which operand is NaN.

#### Numeric literal frontend requirements

M2-P2 supports decimal Float literal syntax using the following grammar:

```text
FloatLiteral :=
    DecimalDigits "." DecimalDigits Exponent?
  | DecimalDigits Exponent

Exponent :=
    ("e" | "E") ("+" | "-")? DecimalDigits
```

`DecimalDigits` is a non-empty sequence of ASCII decimal digits.

The integer portion must obey the existing leading-zero restriction: a multi-digit integer portion must not begin with `0`. Both sides of a decimal point must contain at least one digit. An exponent, when present, must contain at least one digit after its optional sign. Both lowercase `e` and uppercase `E` are accepted.

Signs remain separate unary operators and are not part of Float literal tokens.

Valid examples:

```elq
let a be 3.14
let b be 0.5
let c be 1e3
let d be 1.25e-4
let e be 6E+2
let f be 0.0
let g be 1E-10
```

Invalid examples:

```elq
let a be .5
let b be 5.
let c be 01.5
let d be 1e
let e be 1e+
let f be 1e-
```

Float literal suffixes, digit separators, hexadecimal Float forms, and special NaN or infinity literal spellings are not supported.

Malformed Float-looking sequences and unsupported suffixes must produce deterministic lexical diagnostics rather than being silently interpreted as unrelated valid tokens. Existing integer-literal recognition, leading-zero restrictions, invalid-suffix diagnostic precedence, source positioning, and recovery behavior must remain intact.

The grammar above defines lexical recognition. Decimal-to-binary64 conversion details, including rounding and underflow/overflow behavior, remain a separate open design question tracked by TB-035.

Built-in numeric conversion expressions are also parsed in M2-P2. They use the dedicated `convert expression to Type` syntax defined by D-062 and are not general user-defined function calls. General function-call support remains assigned to M4.

````

## 4. Bindings, mutability, scope [Decided]

```elq
let age be 20
change age to 21
let fixed max_attempts as Int be 3
let frozen names as List[String] be ["Alice", "Bob"]
let nickname as String? be none
````

The M1 frontend parses `let` declarations and `change ... to ...` assignments. M2-P1 subsequently added lexical binding resolution for the supported grammar. Binding types, mutability restrictions, and definite-assignment requirements are not yet enforced by the static checker.

The target language defines `let` as a mutable binding, `fixed` as preventing rebinding without necessarily freezing the referenced value, and `frozen` as enforcing deep immutability. The `fixed` and `frozen` modifiers are not implemented in M1.

The M1 frontend requires initializers for `let` declarations. Support for declarations without initializers remains a separate language-design decision and is not implemented. Local binding types may be inferred statically from their initializers.

Block scopes are lexical. **Shadowing is permitted in nested lexical scopes**, but produces a nonfatal compiler warning identifying the shadowed declaration. Redeclaring a name within the same scope is an error. Identifier lookup selects the nearest visible lexical declaration.

Variables must be declared before use. Future function and named-type declarations may be forward-referenced, subject to the rules established when those constructs are implemented.

A declaration introduces its binding before its initializer is resolved. Consequently, in `let x be x + 1`, the initializer's `x` refers to the newly declared binding, not an enclosing binding. Definite-assignment analysis rejects this uninitialized read.

Lexical name resolution is implemented for the supported grammar through M2-P1. Definite-assignment analysis is a separate M2-P2 semantic responsibility. The existing M1 declaration syntax requires an initializer; declarations without initializers are not introduced by this decision.

Initializer self-reference is a name-resolution question: the identifier binds to the newly declared symbol. Rejecting an uninitialized read is a separate definite-assignment responsibility and remains pending.

Lists, maps and structured values use **value semantics**: assignment creates an independent _logical_ value. The runtime may optimize with copy-on-write or structural sharing without observable aliasing. Explicit references/shared mutable state are deferred to the concurrency design.

## 5. Expressions [Decided unless noted]

Arithmetic operators are `+`, `-`, `*`, `/`, `//`, `%`, and `**`. Their numeric typing, overflow, conversion, and exceptional-value semantics are defined in [Section 3.1 — Numeric semantics](#31-numeric-semantics-decided-m2-p2-implementation-pending).

The M1 frontend parses these arithmetic operators and constructs expression ASTs according to their precedence and associativity rules. Exponentiation is right-associative and binds more tightly than unary negation on its left.

**Implementation status:** Arithmetic parsing is implemented. Numeric type checking, conversion validation, constant evaluation, and runtime arithmetic remain pending M2-P2 and subsequent M2 execution work.

Comparisons and Boolean operations are target-language features assigned to M3. Their numeric comparison contract is specified in Section 3.1 to preserve compatibility when those operators are implemented.

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

**Implementation status:** Structured frontend diagnostics, recovery-capable parsing, and M2-P1 lexical name resolution are implemented. Mandatory static type checking, production Monowire bytecode compilation, execution, and runtime resource controls remain unimplemented. Monowire is the selected execution architecture, but no production runtime functionality is implied by that decision.

**Lexical diagnostics [Implemented]:** The original M1 fail-fast lexer behavior stops at the first detected lexical error. The production frontend also supports controlled lexical recovery through its recovery-capable parsing path. Lexical diagnostics retain source-location information and remain compiler errors rather than runtime exceptions.

**Runtime implementation status:** Eloque does not yet have an executable runtime. Sequential execution, typed runtime failures, assignment evaluation semantics, and configurable execution limits are requirements for future implementation.

Native concurrency, ownership rules, and shared-state synchronization remain deferred design work.

The canonical `Parse(sourceText)` frontend propagates lexical errors through the parser while preserving their source locations. Lexical errors are compiler diagnostics, not runtime exceptions.

**Compiler diagnostics [Partially implemented]:** Eloque uses structured diagnostics for lexical and syntactic errors. The production frontend supports both fail-fast parsing through `Parse(sourceText)` and recovery-capable parsing through `ParseRecovering(sourceText, ParseOptions)`.

Recovery-capable parsing can collect multiple independent diagnostics, preserve successfully recovered statements, and report whether parsing completed, recovered, or aborted. Recovery is bounded by configurable diagnostic limits and expression-depth safeguards.

Recovered or incomplete syntax trees are not eligible for normal bytecode compilation or execution.

M2-P1 extended diagnostic handling to lexical name resolution. Mandatory static type checking will introduce additional semantic diagnostics during the remaining M2 work. Compiler stages will report stable machine-readable codes, severity, source locations, and relevant contextual information while suppressing misleading cascading errors.

Warnings are nonfatal by default. Blocking compilation errors prevent executable code generation and execution.

**Implementation boundary:** Lexical and syntactic recovery, together with M2-P1 resolver diagnostics, are implemented. Type-checker diagnostics, compilation-wide diagnostic coordination, and runtime error reporting remain future implementation work.

## 14. Tooling and backend [Decided architecture; implementation in progress]

Planned CLI commands include `eloque run`, `eloque check`, `eloque fmt`, and eventually a REPL.

### Compiler frontend

The Go implementation currently provides:

- Lexical analysis with direct token streaming.
- Pratt expression parsing.
- Arena-backed flat AST construction.
- Source-aware structured diagnostics.
- Fail-fast and recovery-capable parsing.
- M2-P1 lexical name resolution and symbol binding for the supported grammar.
- Resolver diagnostics for supported name-resolution errors.

The canonical frontend entry point is `Parse(sourceText)`. Recovery-capable parsing is available through `ParseRecovering(sourceText, ParseOptions)`.

M2-P1 lexical name resolution is implemented. Mandatory static type checking, definite-assignment analysis, and production Monowire bytecode compilation remain pending within M2.

### Execution backend

Eloque's selected execution backend is **Monowire**, a register-based bytecode virtual machine.

Monowire uses explicit register operands and is designed around call frames with overlapping caller-outgoing and callee-incoming register windows.

The previous Go-hosted stack-based VM plan has been superseded.

Zig is provisionally preferred as the implementation language for Monowire, with Rust retained as a fallback pending integration, correctness, portability, and maintenance validation.

The production Monowire VM is not yet implemented.

### First executable milestone

M2 — First Execution is in progress. Its implementation stages are:

- Lexical name resolution — completed (M2-P1).
- Mandatory static type checking, including definite-assignment analysis — pending.
- Monowire register-based bytecode compilation — pending.
- Sequential Monowire execution — pending.
- End-to-end correctness and runtime diagnostics for the supported language subset — pending.

The initial executable subset covers supported literals, arithmetic, `let`, `change`, and `say` statements.

Control flow, language-level functions, collections, custom types, modules, and native concurrency remain assigned to subsequent milestones.

### Language semantics and backend independence

Monowire is an implementation architecture, not a change to Eloque's language semantics.

Observable program behavior must remain consistent with this specification regardless of the runtime implementation language, bytecode representation, register allocation, or instruction-dispatch strategy.

The production bytecode format, runtime value layout, garbage collection strategy, and Go-to-Monowire integration mechanism remain engineering decisions to be finalized during their relevant implementation stages.

These details are documented in `ARCHITECTURE.md` and tracked in `TECHNICAL_BACKLOG.md` rather than defined as source-language syntax.

The first executable release will support a narrower subset than the complete language specification.

## 15. Remaining specification work

Milestone 1 established the foundational lexer, parser, AST, and source diagnostics for Eloque's initial syntax subset.

The following areas require further specification or validation during subsequent milestones:

- **Lexical rules:** Formal token and escape-sequence reference tables, alongside rules for future lexical features.
- **Expressions and grammar:** Precedence and associativity for additional operators, and grammar extensions for control flow,functions, collections, and other planned constructs.
- **Numeric semantics:** Type conversions, mixed numeric operations, overflow behavior, and floating-point edge cases.
- **Diagnostics:** Compiler-wide diagnostic coordination, cross-stage error classification, compilation-wide diagnostic budgets, and public machine-readable diagnostic serialization. Internal structured diagnostics and frontend recovery are already implemented.
- **Collections and text:** Iteration behavior, mutation rules, indexing, slicing, and Unicode character and grapheme semantics.
- **Type system:** Type compatibility, inference, narrowing, generics, and related static-analysis rules.
- **Exceptions:** Exception hierarchy, propagation, handler matching, and cleanup guarantees.
- **Modules and interfaces:** Module resolution, initialization, visibility, structural compatibility, and interface validation.
- **Concurrency and ownership:** Task execution, shared-state semantics, synchronization, and memory-safety guarantees.

These items are tracked as future specification and implementation work. They do not prevent Milestone 1 from being considered complete.

Each feature must have its relevant semantics defined and validated before being considered stable.
