# Eloque Architecture

**Status:** Early implementation — Milestone 1 (Frontend foundations) in progress. The lexer, foundational AST structures, parser infrastructure, Pratt arithmetic expression parser, and `let`, `change`, and `say` statement/program parser are implemented and tested. Following M1-P3, the existing lexer and parser underwent canonical frontend consolidation and performance optimization, now merged into `main`. M1-P4 completed broader syntax diagnostics and positive/negative conformance testing, adding 29 source-based fixtures covering valid and malformed programs, accurate source locations, structured diagnostic metadata, fail-fast error propagation, and lexer/parser error boundaries. The conformance suite and existing Go tests, race detection, and static analysis checks passed without requiring production-code changes. The parser continues to consume prepared token streams; production lexer–parser integration, formal end-to-end validation, performance regression review, and milestone closure remain scheduled for M1-P5. Static analysis, bytecode compilation, and VM execution are not yet implemented. The working language name, `.elq` extension, and CLI command remain provisional.

## Goals and boundaries

Eloque is a general-purpose, structured conversational language designed for novice readability, precise static analysis and eventual safe concurrency. The implementation is written in Go, beginning with a **stack-based bytecode virtual machine**. Language semantics must not depend on the bytecode representation.

This document describes the target architecture and the order in which it will be built. The [language specification](LANGUAGE_SPEC.md) governs syntax and observable behavior; [design decisions](DESIGN_DECISIONS.md) records settled and superseded choices.

## Compilation and execution pipeline

```text
Source file (.elq)
       |
       v
Lexer (tokens, indentation, source spans)
       |
       v
Parser (AST and syntax diagnostics)
       |
       v
Name resolution and semantic analysis
       |
       v
Static type checker (mandatory for normal compilation)
       |
       v
Bytecode compiler (constants, instructions, debug metadata)
       |
       v
Stack-based Go VM
       |
       v
Program output or typed runtime error
```

**Static checking is not optional or gradual.** Eloque permits local type inference, but programs with unresolved type errors must not proceed to normal bytecode execution. A future `eloque check` command should run the frontend without executing user code. The initial lexer and parser use fail-fast diagnostics: the first lexical or syntax error terminates the corresponding operation. Multi-error collection and recovery are deferred until a dedicated frontend enhancement.

The implementation may initially recognize only a small subset of the full design. Unsupported syntax should be rejected clearly, never silently accepted with guessed semantics.

## Component responsibilities

| Component | Responsibility |
|---|---|
| CLI | Route `run`, future `check`, and eventual formatting/debugging commands |
| Source manager | Track file names, line/column positions and source spans |
| Lexer | Recognize identifiers, keywords, literals, operators, newlines, indentation and dedentation |
| Parser | Construct AST nodes and report grammatical errors, including malformed conversational phrases |
| Resolver | Bind names, enforce scope rules and detect undeclared names and prohibited shadowing |
| Type checker | Infer eligible local types, validate explicit signatures, optional types, calls, operators and return paths |
| Bytecode compiler | Lower validated AST into executable instructions and constant pools |
| VM | Manage value stack, frames, instructions, calls, control flow and exception propagation |
| Runtime | Represent values, implement built-ins and collection semantics, enforce checked operations and immutability |
| Diagnostics | Render human-readable errors and provide a machine-readable structured form |

A single implementation package may initially handle more than one responsibility. Package boundaries should emerge from real dependencies rather than an elaborate empty directory tree.

## Frontend constraints

### Lexical and syntactic rules

- Four spaces per indentation level; tabs prohibited. Blank/comment-only lines do not change indentation.
- A colon introduces an indented block; newlines ordinarily terminate statements.
- The current lexer supports implicit line continuation inside matched parentheses. Continuation inside brackets and braces is part of the target language design and will be implemented alongside their respective collection syntax.
- Keywords and conversational operator phrases have precise grammar; this is **not** natural-language parsing.
- Prefer tokenizing individual words and recognizing multiword constructs in the parser.
- Preserve source spans on tokens and AST nodes for precise diagnostics.

### Name resolution and typing

- `let x be value` declares a mutable binding; `change x to value` updates it.
- `let fixed` prohibits rebinding; `let frozen` requires deeply immutable values.
- Names cannot shadow visible bindings in enclosing scopes.
- Function parameters require `as Type`; functions returning values require `giving Type`.
- Initialized local bindings can infer types. `none` is permitted only where an optional type is expected.
- Type checking is a required compilation stage, including for CLI execution.
- Any flow-sensitive narrowing must be invalidated where mutation or concurrency makes it unsafe.

### Evaluation

- Evaluate operands and function arguments left to right; `and`/`or` short-circuit.
- Preserve lexical scope and deterministic sequential execution by default.
- Top-level statements execute in source order.
- Imports initialize modules once per execution; reject circular imports in the initial design.
- No automatic `main` invocation.

## Current frontend implementation

The lexer, foundational AST structures, parser infrastructure, arithmetic expression parser, and initial statement/program parser are implemented in Go.

M1-P1 established the shared source-location model, AST foundations, token navigation, and structured syntax diagnostics. M1-P2 introduced Pratt-based arithmetic expression parsing, including precedence, associativity, grouping, and source-extent tracking. M1-P3 extends the parser to support `let`, `change`, and `say` statements and complete programs.

The parser currently operates on prepared token streams. M1-P4 has completed broader syntax conformance and diagnostic verification using lexer-generated tokens in test fixtures. Production lexer–parser integration and formal end-to-end frontend validation remain scheduled for M1-P5.

### Lexer functionality

The lexer currently supports:

- Byte-oriented source scanning with 1-based line and column tracking.
- ASCII identifiers and reserved keywords, including `let`, `be`, `say`, `change`, `to`, `with`, and `giving`.
- Decimal integer literals, rejecting prohibited leading zeros and invalid identifier suffixes.
- Single- and double-quoted string literals with validation of the supported escape sequences.
- Arithmetic operators `+`, `-`, `*`, `**`, `/`, `//`, and `%`.
- Parentheses and colons.
- Significant newlines, `INDENT`, and `DEDENT` tokens.
- Four-space indentation alignment and rejection of invalid indentation.
- `#` comments, including inline and comment-only lines.
- Suppression of physical newlines inside parentheses.
- Half-open source spans and positioned lexical errors.
- Fail-fast lexical analysis.

The lexer recognizes fixed reserved words as dedicated token kinds. Their grammatical roles are determined by the parser. For example, `to` participates in both assignment and function-definition syntax.

### Token representation

Each token contains:

- `Kind`: a `TokenKind` identifying its lexical category.
- `Lexeme`: the original source text.
- `Span`: a half-open source range whose start is inclusive and end is exclusive.

`TokenKind` uses a `uint8`-backed enumeration. Its `String()` method provides readable names for diagnostics and debugging. The zero value, `TokenInvalid`, represents an invalid or uninitialized token kind.

The token representation uses shared source-location types from `internal/source`. Compatibility aliases in the lexer preserve the existing `lexer.Position` and `lexer.Span` names.

Token lexemes refer to the original Go source string where applicable, avoiding unnecessary copies of source text.

### Indentation and pending tokens

The lexer maintains an indentation stack initially containing the base indentation level of zero. Indentation increases generate `INDENT` tokens; decreases generate the necessary `DEDENT` tokens. Remaining indentation levels are closed at end of file.

Generated structural tokens are stored in a reusable pending-token queue. An index tracks the next unread token, and the queue resets its length after draining while retaining its backing storage.

The indentation stack starts with capacity for four entries, including the base indentation level. Its capacity grows automatically when deeper nesting requires it.

### AST foundations

The `internal/ast` package defines three principal interfaces:

- `Node`: any AST node that exposes its source span.
- `Expression`: a value-producing syntactic construct.
- `Statement`: a statement-level construct.

Expression and statement interfaces use package-private marker methods to distinguish their implementations. The `Program` root implements `Node` but is neither an expression nor a statement.

The initial expression nodes are `IntegerLiteral`, `StringLiteral`, `Identifier`, `UnaryExpression`, and `BinaryExpression`. The initial statement nodes are `LetStatement`, `ChangeStatement`, and `SayStatement`.

AST nodes retain their source locations and syntactic relationships without evaluating expressions, resolving names, or performing type checking. Integer literal nodes preserve their decimal source text, while string literal nodes store decoded string contents produced during expression parsing. Unary and binary expression nodes retain their operators, operands, and source spans.

The AST uses interface-based child references. Its structure remains independent of the planned bytecode representation.

### Parser infrastructure and syntax diagnostics

The `internal/parser` package currently accepts an ordered token slice terminated by `TokenEOF`. Its navigation helpers inspect, match, and consume tokens while preserving the terminal EOF position.

The parser also provides a token-expectation helper and a structured `Diagnostic` containing:

- A human-readable message.
- A source span identifying the error location.
- Optional expected token kinds.
- An optional actual token kind.

The initial parser follows a fail-fast contract: the first syntax error terminates parsing, and a malformed program must not be accepted as a successfully parsed AST. Diagnostic accumulation, synchronization, and multi-error recovery are deferred.

The parser constructs arithmetic expression ASTs using Pratt parsing. It supports integer and string literals, identifiers, parenthesized expressions, unary numeric plus and minus, and the binary arithmetic operators `+`, `-`, `*`, `/`, `//`, `%`, and `**`.

M1-P3 extends this infrastructure with statement-specific parsers, statement dispatch, and complete-program parsing. Each statement delegates its value expression to the existing Pratt parser. The program parser handles newline-separated statements, blank lines, and EOF termination, enforcing statement boundaries and preserving source spans.

Lexer–parser integration and broader frontend conformance remain scheduled for subsequent M1 checkpoints.

### Pratt expression parsing (M1-P2)

The expression parser uses Pratt parsing with numeric binding powers to implement Eloque's initial arithmetic grammar. It separates prefix parsing from binary-operator parsing and uses binding-power thresholds to determine how expressions associate.

The currently implemented precedence levels, from highest to lowest, are:

| Precedence | Operators | Associativity |
|---|---|---|
| Grouping | `(...)` | Explicit grouping |
| Exponentiation | `**` | Right-associative |
| Unary arithmetic | `+`, `-` | Right-nesting prefix operators |
| Multiplicative | `*`, `/`, `//`, `%` | Left-associative |
| Additive | `+`, `-` | Left-associative |

Exponentiation binds more tightly than unary arithmetic on its left but accepts unary expressions on its right. Consequently, `-2 ** 2` parses as `-(2 ** 2)`, `2 ** -3` parses as `2 ** (-3)`, and `2 ** 3 ** 2` parses as `2 ** (3 ** 2)`.

The implementation uses binding powers of 50 for additive operators, 60 for multiplicative operators, 70 for prefix unary operators, and 80 for exponentiation. Left-associative operators parse their right operands with a higher minimum binding power, while exponentiation uses its own binding power to preserve right associativity.

Parentheses control grouping without introducing an additional semantic AST node. The parser maintains an internal expression result containing both the AST expression and its complete syntactic source extent. This allows enclosing unary and binary expressions to include grouping parentheses in their spans while preserving the original locations of child AST nodes.

String literal parsing decodes the escape sequences already validated by the lexer. Integer literals retain their original decimal lexemes; numeric interpretation and type checking are not performed during parsing.

The lexer suppresses physical newlines inside matched parentheses, permitting multiline grouped expressions. The expression parser operates on the resulting token stream and treats emitted newline tokens as expression boundaries.

Malformed expressions produce fail-fast syntax diagnostics with source locations and expected/actual token information where applicable. The parser does not currently implement error recovery or multiple diagnostics per parse.

The implemented expression subset excludes comparisons, Boolean operators, function calls, member access, and collection expressions. These remain part of the target language specification and will be introduced according to their scheduled implementation milestones.

### Expression parser complexity and performance (M1-P2)

The Pratt expression parser consumes tokens without backtracking and has expected O(n) parsing time for an expression containing n tokens. AST memory grows proportionally to the number of expression nodes. Recursive parsing introduces additional stack usage proportional to expression nesting depth.

The initial M1-P2 benchmark suite measures parsing and AST construction from prepared token streams, excluding lexing and token generation. Five benchmark runs were collected on linux/amd64 using an Intel Core Ultra 7 155H processor.

| Benchmark | Input size | Median time | Memory per parse | Allocations per parse |
|---|---:|---:|---:|---:|
| Simple arithmetic | 2 operators | ~287 ns | 496 B | 8 |
| Mixed precedence | 3 operators | ~394 ns | 688 B | 11 |
| Additive chain | 100 operators | ~12.42 µs | 19,314 B | 302 |
| Additive chain | 1,000 operators | ~115.63 µs | ~192 KB | 3,002 |
| Exponentiation chain | 100 operators | ~13.11 µs | 19,314 B | 302 |
| Exponentiation chain | 1,000 operators | ~135.02 µs | ~192 KB | 3,002 |
| Nested parentheses | 100 levels | ~14.38 µs | 12,913 B | 202 |
| Nested parentheses | 1,000 levels | ~160.87 µs | ~128 KB | 2,002 |

The tested workloads demonstrate approximately linear execution-time and allocation scaling. Stress tests with 1,000 nested parentheses and 1,000 exponentiation operators also pass.

No production-code optimisation was justified by the initial measurements. Recursive nesting remains a potential robustness concern for unusually deep or adversarial inputs; no universal nesting limit has been established.

Benchmarks are maintained in `internal/parser/expression_bench_test.go`. These results establish an expression-parser baseline rather than full frontend or compiler throughput.

### Statement and program parsing (M1-P3)

M1-P3 introduces the initial statement grammar and complete-program AST construction using the existing parser infrastructure.

The supported statement forms are:

```elq
let value be 10 + 20
change value to value * 2
say value
```

These examples describe syntax accepted by the statement parser when supplied with the corresponding tokens. They are not yet executable Eloque programs.

**Statement parsing**

The parser implements three statement-specific methods:

- `parseLetStatement()` parses `let identifier be expression` and constructs an `ast.LetStatement`.
- `parseChangeStatement()` parses `change identifier to expression` and constructs an `ast.ChangeStatement`.
- `parseSayStatement()` parses `say expression` and constructs an `ast.SayStatement`.

`parseStatement()` dispatches according to the current token kind. Unsupported statement starters produce structured diagnostics identifying the expected statement keywords and actual token kind.

Each statement parser delegates expression parsing to `parseExpression(0)`. Expression parsing remains independent of statement-specific syntax.

Statement source spans begin at their initiating keyword and end at the complete syntactic extent of their value expression. This includes enclosing grouping parentheses where applicable.

**Complete-program parsing**

`parseProgram()` constructs an `ast.Program` containing an ordered slice of statements.

The program parser:

- Processes statements in source order.
- Skips newline tokens between statements, including blank lines.
- Requires a newline or EOF after each successfully parsed statement.
- Stops at EOF.
- Returns a diagnostic rather than a partial successful program when parsing fails.

A nonempty program's source span begins at its first statement and ends at its last statement. For an empty program, the span is zero-width at the EOF position.

The current implementation does not parse indentation-based blocks or additional statement forms. These will be introduced as their respective language features are implemented.

**Diagnostics and failure handling**

M1-P3 preserves the existing fail-fast diagnostic contract.

Malformed statement syntax, missing required keywords or identifiers, invalid statement starters, missing statement separators, and nested expression failures produce structured diagnostics.

Diagnostics retain the relevant source location and expected/actual token metadata where applicable. Expression diagnostics propagate through statement and program parsing without being replaced by generic errors.

The parser returns no successful program AST when a syntax error is encountered. Multi-error recovery and synchronization remain deferred.

**Architecture and complexity**

The parsing flow is:

```text
parseProgram()
    |
    v
parseStatement()
    |
    +-- parseLetStatement()
    |
    +-- parseChangeStatement()
    |
    +-- parseSayStatement()
                |
                v
         parseExpression(0)
```

All three statement-specific parsers delegate value expressions to `parseExpression(0)`; the diagram illustrates the shared dependency rather than exclusive use by `parseSayStatement()`.

Statement dispatch uses token-kind matching, and program parsing advances through the token stream without backtracking. Expected parsing time is O(n) for n tokens, excluding pathological recursive-expression stack considerations.

AST memory and allocations grow proportionally to the size of the parsed program under the tested workloads.

**Performance baseline**

The initial M1-P3 benchmark suite measures complete-program parsing and AST construction from prepared token streams, excluding lexing and token generation.

Five runs per workload were collected on linux/amd64 using an Intel Core Ultra 7 155H processor.

| Statements | Mean parsing time | Approximate memory per parse | Allocations per parse |
|---|---:|---:|---:|
| 3 | 864.5 ns | 1,712 B | 29 |
| 300 | 86.5 µs | 162,994 B | 2,510 |
| 3,000 | 1.007 ms | 1,694,141 B | 25,015 |

The results demonstrate approximately linear scaling in execution time, memory consumption, and allocations across the tested workloads. No production-code optimisation was justified at this checkpoint.

Benchmarks are maintained in `internal/parser/program_bench_test.go`, with raw results preserved in `benchmark_m1_p3.txt`. The measurements establish a program-parser baseline, not end-to-end frontend or compiler throughput.

### Canonical frontend optimization (post-M1-P3)

Following M1-P3, the existing lexer and parser were consolidated and optimized before beginning M1-P4. This work establishes the canonical implementation baseline for subsequent frontend development.

The optimization was undertaken without intentionally changing the language's accepted syntax or observable parsing semantics. The existing correctness and robustness requirements remain applicable.

**Recorded optimized measurements**

| Workload | Result |
|---|---:|
| Lexer throughput | Approximately 261.6 MB/s |
| Lexer allocation count | 2 allocs/op |
| Parser, 1,000-operation chains | Approximately 76.0–99.4 µs |
| Parser chain allocation count | 2,002 allocs/op |
| Parser, 3,000-statement program | Approximately 664 µs |
| Parser program allocation count | 13,028 allocs/op |
| Expression-heavy complete frontend throughput | Approximately 34.5 MB/s |

These measurements characterize specific benchmark workloads, not universal frontend throughput.

The optimized 3,000-statement parser workload is approximately 34% faster than the original M1-P3 baseline of 1.007 ms, with approximately 48% fewer allocations than the original 25,015 allocs/op.

Complete frontend benchmarks measure lexing, token collection, parsing, and AST construction. They must be distinguished from prepared-token parser benchmarks, which exclude lexical processing.

Reproducible benchmark commands, workload definitions, recorded results, and the preliminary CPython frontend comparison are maintained in [`benchmarks/README.md`](../benchmarks/README.md).

The optimized frontend served as the baseline for M1-P4 conformance testing and remains the canonical implementation for M1-P5 production integration, end-to-end validation, and performance regression review.

### Complexity and measured lexer performance

For source length `N`, the lexer operates in O(N) time overall. Its auxiliary working space is O(d + q), where `d` is indentation depth and `q` is the maximum number of pending tokens, excluding source storage and returned tokens retained by callers.

The October 2026 performance audit measured:

| Metric | Initial implementation | Optimized implementation |
|---|---:|---:|
| Mixed-source execution time | ~699 ns/op | ~644 ns/op |
| Mixed-source allocation volume | 248 B/op | 80 B/op |
| Mixed-source allocation count | 7 allocs/op | 2 allocs/op |
| Token structure size | 48 bytes | 40 bytes |

Benchmarks were performed on linux/amd64 with an Intel Core Ultra 7 155H processor. Runtime measurements are approximate and subject to benchmark variability.

A separate scaling benchmark, using repeated source samples from 138 bytes to 138 KB, showed approximately linear execution time and constant allocation counts for the tested bounded-indentation workload.

These measurements characterize the lexer baseline before the latest keyword additions, not the performance of the future parser, compiler, or virtual machine. Benchmarks are maintained in `internal/lexer/lexer_bench_test.go`.

## Runtime value model

The observable semantics of lists, maps and structured custom values are **value semantics**: assigning one to another produces logically independent values. The runtime may use copies, copy-on-write or structural sharing if mutation remains observationally independent.

- `Int`: signed 64-bit with checked overflow.
- `Float`: IEEE 754 binary64; no implicit lossy conversion.
- `String`: immutable Unicode text.
- `Bool`: strict booleans, no truthiness conversion.
- Optional values: `T?` can hold `none`.
- Lists/maps: typed operations, negative indexing for supported sequence types, end-exclusive slices, safe `find` access.

The VM's actual value representation (Go interfaces, tagged union, pointers, etc.) remains **open** until implementation. Go's garbage collector can manage Go-allocated objects, but it does not automatically implement Eloque's value semantics, deep freezing or safe shared-state policy.

## VM direction

Start with a minimal stack machine that can execute validated constants, arithmetic, `say` and bindings. Add branches, loops, calls, collections and exceptions as the frontend supports them.

Likely building blocks:

- Opcode/instruction representation and constant pool.
- Operand stack and local variable slots.
- Call frames and return addresses.
- Branch and loop instructions.
- Source-location/debug metadata.
- Typed runtime failures and exception unwinding.

Avoid premature commitments to bytecode encoding, object layout, garbage-collection strategies beyond Go-managed allocation, or a multi-backend intermediate representation.

A small tree evaluator can be used as a temporary test aid if useful, but **the primary planned runtime is bytecode plus VM**.

## Errors and diagnostics

Diagnostics are a first-class interface for both human programmers and AI agents.

- Compile-time errors should include an error category/code, message, file, line, column and relevant source span.
- Runtime failures should carry typed exceptions and source locations when available.
- Human-readable CLI diagnostics are the default; structured machine-readable output should be supported by the tooling design.
- Invalid direct indexing, missing map keys, integer overflow and division by zero must not silently return misleading values.
- Exceptions are unchecked; static analysis may warn about potentially unhandled errors without requiring exception declarations.

Exact diagnostic codes, JSON schema and exception hierarchy are deferred until their respective implementation milestones.

## Modules

Each `.elq` file is intended to be a module. `import` supports whole-module and selected-name imports, with aliases introduced by `named`. Top-level names beginning with `_` are private and access must be rejected by the compiler.

Module initialization runs at most once per program execution. Initialization order follows a defined dependency order; circular imports are rejected initially. Path resolution, caching representation and failure recovery need further specification.

## Concurrency and sandboxing

The initial runtime is **sequential**. Eventual native concurrency is a language goal, not a feature of the first VM.

Explicit references, ownership rules, shared mutable state, cancellation, task lifetime and exception propagation should be designed together. Go goroutines may be a runtime mechanism, but Eloque's user-facing safety guarantees must be specified independently.

Normal CLI execution has no arbitrary universal instruction budget. Sandboxed/agent execution should eventually support configurable instruction, time and memory limits. Hard guarantees depend on the implementation and hosting environment; not every infinite loop can be proven statically.

## Performance and future backends

Measure correctness first, then benchmark representative workloads. A stack-based VM is the baseline. Compare a register-based alternative only after the working VM and a repeatable benchmark harness exist.

Potential later work includes opcode specialization, reduced allocations, inline caches, optional JIT, AOT/native compilation and WebAssembly. None is a committed v0.1 feature. CPython parity is an aspiration, **not** an established result.

Keep language semantics separate from VM internals so future backends can preserve observable behavior without rewriting the language contract. Do not build a general-purpose IR until a real second-backend or optimization need justifies it.

## Compatibility

- Before 1.0, syntax and semantics may change, with documentation and conformance tests updated together.
- At 1.0, published source-level semantics become the compatibility commitment.
- Internal bytecode is not a stable public interchange format unless standardized later.
- Backend changes that preserve observable behavior need not require a major version.

## Open architecture decisions and review gates

| Decision | Review trigger |
|---|---|
| Runtime value representation and efficient value semantics | Before first VM value implementation |
| Exact instruction encoding and source-map design | First bytecode compiler milestone |
| Float edge cases, conversion rules and formatting | Numeric runtime and checker implementation |
| Collection freezing and safe mutation | Before mutable collection implementation |
| Exception taxonomy and unwinding precedence | Exception milestone |
| Module resolution and caching | Import milestone |
| Closure representation and advanced functions | When closures enter scope |
| Explicit references, ownership and concurrency | Dedicated concurrency design milestone |
| Structured diagnostic schema and sandbox limits | CLI tooling / agent execution milestone |
| Backend-neutral IR | Only when a second backend or real optimization need arises |

Review triggers are tracked in [TECHNICAL_BACKLOG.md](TECHNICAL_BACKLOG.md). Avoid interpreting a deferred implementation detail as an unsettled language principle.
