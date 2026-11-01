# Eloque M2 cross-language semantic comparison

**Different measurement boundaries: do not interpret this as a compiler speed ranking.**

## In-process measurements (Eloque and CPython)

| Engine | Workload | N | Median (µs/op) | Boundary |
|---|---|---:|---:|---|
| CPython ast.parse | declarations | 10 | 21.381 | in-process parse + AST |
| CPython ast.parse | declarations | 100 | 176.741 | in-process parse + AST |
| CPython ast.parse | declarations | 1,000 | 2198.739 | in-process parse + AST |
| CPython ast.parse | declarations | 3,000 | 6945.701 | in-process parse + AST |
| CPython ast.parse | mixed | 10 | 34.970 | in-process parse + AST |
| CPython ast.parse | mixed | 100 | 309.471 | in-process parse + AST |
| CPython ast.parse | mixed | 1,000 | 4575.726 | in-process parse + AST |
| CPython ast.parse | mixed | 3,000 | 13975.634 | in-process parse + AST |
| CPython ast.parse | references | 10 | 18.931 | in-process parse + AST |
| CPython ast.parse | references | 100 | 137.836 | in-process parse + AST |
| CPython ast.parse | references | 1,000 | 1789.489 | in-process parse + AST |
| CPython ast.parse | references | 3,000 | 5543.547 | in-process parse + AST |
| CPython ast.parse | sparse-undefined | 10 | 19.223 | in-process parse + AST |
| CPython ast.parse | sparse-undefined | 100 | 138.603 | in-process parse + AST |
| CPython ast.parse | sparse-undefined | 1,000 | 1695.918 | in-process parse + AST |
| CPython ast.parse | sparse-undefined | 3,000 | 5273.948 | in-process parse + AST |
| CPython ast.parse | undefined | 10 | 17.141 | in-process parse + AST |
| CPython ast.parse | undefined | 100 | 139.065 | in-process parse + AST |
| CPython ast.parse | undefined | 1,000 | 1903.916 | in-process parse + AST |
| CPython ast.parse | undefined | 3,000 | 6117.620 | in-process parse + AST |
| CPython symtable | declarations | 10 | 18.308 | in-process symbol table (internally reparses source) |
| CPython symtable | declarations | 100 | 127.982 | in-process symbol table (internally reparses source) |
| CPython symtable | declarations | 1,000 | 1630.384 | in-process symbol table (internally reparses source) |
| CPython symtable | declarations | 3,000 | 5141.783 | in-process symbol table (internally reparses source) |
| CPython symtable | mixed | 10 | 28.865 | in-process symbol table (internally reparses source) |
| CPython symtable | mixed | 100 | 235.476 | in-process symbol table (internally reparses source) |
| CPython symtable | mixed | 1,000 | 3123.231 | in-process symbol table (internally reparses source) |
| CPython symtable | mixed | 3,000 | 10345.063 | in-process symbol table (internally reparses source) |
| CPython symtable | references | 10 | 17.798 | in-process symbol table (internally reparses source) |
| CPython symtable | references | 100 | 108.361 | in-process symbol table (internally reparses source) |
| CPython symtable | references | 1,000 | 1327.406 | in-process symbol table (internally reparses source) |
| CPython symtable | references | 3,000 | 4203.049 | in-process symbol table (internally reparses source) |
| CPython symtable | sparse-undefined | 10 | 17.371 | in-process symbol table (internally reparses source) |
| CPython symtable | sparse-undefined | 100 | 109.033 | in-process symbol table (internally reparses source) |
| CPython symtable | sparse-undefined | 1,000 | 1273.225 | in-process symbol table (internally reparses source) |
| CPython symtable | sparse-undefined | 3,000 | 4040.492 | in-process symbol table (internally reparses source) |
| CPython symtable | undefined | 10 | 17.463 | in-process symbol table (internally reparses source) |
| CPython symtable | undefined | 100 | 118.078 | in-process symbol table (internally reparses source) |
| CPython symtable | undefined | 1,000 | 1555.929 | in-process symbol table (internally reparses source) |
| CPython symtable | undefined | 3,000 | 4740.064 | in-process symbol table (internally reparses source) |
| Eloque resolver | declarations | 10 | 0.647 | in-process resolver only (prebuilt AST) |
| Eloque resolver | declarations | 100 | 5.022 | in-process resolver only (prebuilt AST) |
| Eloque resolver | declarations | 1,000 | 47.184 | in-process resolver only (prebuilt AST) |
| Eloque resolver | declarations | 3,000 | 150.994 | in-process resolver only (prebuilt AST) |
| Eloque resolver | mixed | 10 | 1.483 | in-process resolver only (prebuilt AST) |
| Eloque resolver | mixed | 100 | 12.598 | in-process resolver only (prebuilt AST) |
| Eloque resolver | mixed | 1,000 | 119.822 | in-process resolver only (prebuilt AST) |
| Eloque resolver | mixed | 3,000 | 380.836 | in-process resolver only (prebuilt AST) |
| Eloque resolver | references | 10 | 0.369 | in-process resolver only (prebuilt AST) |
| Eloque resolver | references | 100 | 1.853 | in-process resolver only (prebuilt AST) |
| Eloque resolver | references | 1,000 | 16.238 | in-process resolver only (prebuilt AST) |
| Eloque resolver | references | 3,000 | 46.798 | in-process resolver only (prebuilt AST) |
| Eloque resolver | sparse-undefined | 10 | 0.776 | in-process resolver only (prebuilt AST) |
| Eloque resolver | sparse-undefined | 100 | 2.129 | in-process resolver only (prebuilt AST) |
| Eloque resolver | sparse-undefined | 1,000 | 18.663 | in-process resolver only (prebuilt AST) |
| Eloque resolver | sparse-undefined | 3,000 | 53.243 | in-process resolver only (prebuilt AST) |
| Eloque resolver | undefined | 10 | 1.268 | in-process resolver only (prebuilt AST) |
| Eloque resolver | undefined | 100 | 16.084 | in-process resolver only (prebuilt AST) |
| Eloque resolver | undefined | 1,000 | 117.745 | in-process resolver only (prebuilt AST) |
| Eloque resolver | undefined | 3,000 | 464.897 | in-process resolver only (prebuilt AST) |

## Native compiler subprocess measurements (separate boundary)

Includes process startup and compiler work beyond name resolution.

| Engine | Workload | N | Median (ms/invocation) | Boundary |
|---|---|---:|---:|---|
| c compiler | declarations | 10 | 14.606 | compiler subprocess: Clang parse + semantic validation (syntax-only) |
| c compiler | declarations | 100 | 14.295 | compiler subprocess: Clang parse + semantic validation (syntax-only) |
| c compiler | declarations | 1,000 | 19.126 | compiler subprocess: Clang parse + semantic validation (syntax-only) |
| c compiler | declarations | 3,000 | 27.144 | compiler subprocess: Clang parse + semantic validation (syntax-only) |
| c compiler | mixed | 10 | 14.302 | compiler subprocess: Clang parse + semantic validation (syntax-only) |
| c compiler | mixed | 100 | 15.344 | compiler subprocess: Clang parse + semantic validation (syntax-only) |
| c compiler | mixed | 1,000 | 19.916 | compiler subprocess: Clang parse + semantic validation (syntax-only) |
| c compiler | mixed | 3,000 | 31.832 | compiler subprocess: Clang parse + semantic validation (syntax-only) |
| c compiler | references | 10 | 14.772 | compiler subprocess: Clang parse + semantic validation (syntax-only) |
| c compiler | references | 100 | 14.371 | compiler subprocess: Clang parse + semantic validation (syntax-only) |
| c compiler | references | 1,000 | 16.048 | compiler subprocess: Clang parse + semantic validation (syntax-only) |
| c compiler | references | 3,000 | 20.141 | compiler subprocess: Clang parse + semantic validation (syntax-only) |
| c compiler | sparse-undefined | 10 | 14.558 | compiler subprocess: Clang parse + semantic validation (syntax-only) |
| c compiler | sparse-undefined | 100 | 15.275 | compiler subprocess: Clang parse + semantic validation (syntax-only) |
| c compiler | sparse-undefined | 1,000 | 16.205 | compiler subprocess: Clang parse + semantic validation (syntax-only) |
| c compiler | sparse-undefined | 3,000 | 20.755 | compiler subprocess: Clang parse + semantic validation (syntax-only) |
| c compiler | undefined | 10 | 15.962 | compiler subprocess: Clang parse + semantic validation (syntax-only) |
| c compiler | undefined | 100 | 15.066 | compiler subprocess: Clang parse + semantic validation (syntax-only) |
| c compiler | undefined | 1,000 | 20.429 | compiler subprocess: Clang parse + semantic validation (syntax-only) |
| c compiler | undefined | 3,000 | 16.439 | compiler subprocess: Clang parse + semantic validation (syntax-only) |
| cpp compiler | declarations | 10 | 16.346 | compiler subprocess: Clang++ parse + semantic validation (syntax-only) |
| cpp compiler | declarations | 100 | 16.672 | compiler subprocess: Clang++ parse + semantic validation (syntax-only) |
| cpp compiler | declarations | 1,000 | 21.535 | compiler subprocess: Clang++ parse + semantic validation (syntax-only) |
| cpp compiler | declarations | 3,000 | 31.241 | compiler subprocess: Clang++ parse + semantic validation (syntax-only) |
| cpp compiler | mixed | 10 | 14.387 | compiler subprocess: Clang++ parse + semantic validation (syntax-only) |
| cpp compiler | mixed | 100 | 15.913 | compiler subprocess: Clang++ parse + semantic validation (syntax-only) |
| cpp compiler | mixed | 1,000 | 22.400 | compiler subprocess: Clang++ parse + semantic validation (syntax-only) |
| cpp compiler | mixed | 3,000 | 38.992 | compiler subprocess: Clang++ parse + semantic validation (syntax-only) |
| cpp compiler | references | 10 | 14.490 | compiler subprocess: Clang++ parse + semantic validation (syntax-only) |
| cpp compiler | references | 100 | 15.171 | compiler subprocess: Clang++ parse + semantic validation (syntax-only) |
| cpp compiler | references | 1,000 | 16.188 | compiler subprocess: Clang++ parse + semantic validation (syntax-only) |
| cpp compiler | references | 3,000 | 23.975 | compiler subprocess: Clang++ parse + semantic validation (syntax-only) |
| cpp compiler | sparse-undefined | 10 | 15.051 | compiler subprocess: Clang++ parse + semantic validation (syntax-only) |
| cpp compiler | sparse-undefined | 100 | 14.805 | compiler subprocess: Clang++ parse + semantic validation (syntax-only) |
| cpp compiler | sparse-undefined | 1,000 | 16.679 | compiler subprocess: Clang++ parse + semantic validation (syntax-only) |
| cpp compiler | sparse-undefined | 3,000 | 23.294 | compiler subprocess: Clang++ parse + semantic validation (syntax-only) |
| cpp compiler | undefined | 10 | 14.497 | compiler subprocess: Clang++ parse + semantic validation (syntax-only) |
| cpp compiler | undefined | 100 | 15.175 | compiler subprocess: Clang++ parse + semantic validation (syntax-only) |
| cpp compiler | undefined | 1,000 | 15.926 | compiler subprocess: Clang++ parse + semantic validation (syntax-only) |
| cpp compiler | undefined | 3,000 | 18.707 | compiler subprocess: Clang++ parse + semantic validation (syntax-only) |
| go compiler | declarations | 10 | 48.070 | compiler subprocess: Go object compilation (not resolver-only) |
| go compiler | declarations | 100 | 51.274 | compiler subprocess: Go object compilation (not resolver-only) |
| go compiler | declarations | 1,000 | 66.854 | compiler subprocess: Go object compilation (not resolver-only) |
| go compiler | declarations | 3,000 | 106.966 | compiler subprocess: Go object compilation (not resolver-only) |
| go compiler | mixed | 10 | 62.405 | compiler subprocess: Go object compilation (not resolver-only) |
| go compiler | mixed | 100 | 60.841 | compiler subprocess: Go object compilation (not resolver-only) |
| go compiler | mixed | 1,000 | 72.554 | compiler subprocess: Go object compilation (not resolver-only) |
| go compiler | mixed | 3,000 | 94.652 | compiler subprocess: Go object compilation (not resolver-only) |
| go compiler | references | 10 | 83.278 | compiler subprocess: Go object compilation (not resolver-only) |
| go compiler | references | 100 | 51.038 | compiler subprocess: Go object compilation (not resolver-only) |
| go compiler | references | 1,000 | 54.203 | compiler subprocess: Go object compilation (not resolver-only) |
| go compiler | references | 3,000 | 70.919 | compiler subprocess: Go object compilation (not resolver-only) |
| go compiler | sparse-undefined | 10 | 48.615 | compiler subprocess: Go object compilation (not resolver-only) |
| go compiler | sparse-undefined | 100 | 52.731 | compiler subprocess: Go object compilation (not resolver-only) |
| go compiler | sparse-undefined | 1,000 | 63.652 | compiler subprocess: Go object compilation (not resolver-only) |
| go compiler | sparse-undefined | 3,000 | 49.198 | compiler subprocess: Go object compilation (not resolver-only) |
| go compiler | undefined | 10 | 49.269 | compiler subprocess: Go object compilation (not resolver-only) |
| go compiler | undefined | 100 | 52.406 | compiler subprocess: Go object compilation (not resolver-only) |
| go compiler | undefined | 1,000 | 49.204 | compiler subprocess: Go object compilation (not resolver-only) |
| go compiler | undefined | 3,000 | 50.064 | compiler subprocess: Go object compilation (not resolver-only) |
| rust compiler | declarations | 10 | 15.998 | compiler subprocess: rustc metadata compilation (not resolver-only) |
| rust compiler | declarations | 100 | 16.374 | compiler subprocess: rustc metadata compilation (not resolver-only) |
| rust compiler | declarations | 1,000 | 29.602 | compiler subprocess: rustc metadata compilation (not resolver-only) |
| rust compiler | declarations | 3,000 | 57.588 | compiler subprocess: rustc metadata compilation (not resolver-only) |
| rust compiler | mixed | 10 | 14.762 | compiler subprocess: rustc metadata compilation (not resolver-only) |
| rust compiler | mixed | 100 | 17.202 | compiler subprocess: rustc metadata compilation (not resolver-only) |
| rust compiler | mixed | 1,000 | 28.794 | compiler subprocess: rustc metadata compilation (not resolver-only) |
| rust compiler | mixed | 3,000 | 56.253 | compiler subprocess: rustc metadata compilation (not resolver-only) |
| rust compiler | references | 10 | 15.096 | compiler subprocess: rustc metadata compilation (not resolver-only) |
| rust compiler | references | 100 | 16.661 | compiler subprocess: rustc metadata compilation (not resolver-only) |
| rust compiler | references | 1,000 | 23.293 | compiler subprocess: rustc metadata compilation (not resolver-only) |
| rust compiler | references | 3,000 | 47.310 | compiler subprocess: rustc metadata compilation (not resolver-only) |
| rust compiler | sparse-undefined | 10 | 23.755 | compiler subprocess: rustc metadata compilation (not resolver-only) |
| rust compiler | sparse-undefined | 100 | 24.379 | compiler subprocess: rustc metadata compilation (not resolver-only) |
| rust compiler | sparse-undefined | 1,000 | 31.950 | compiler subprocess: rustc metadata compilation (not resolver-only) |
| rust compiler | sparse-undefined | 3,000 | 47.530 | compiler subprocess: rustc metadata compilation (not resolver-only) |
| rust compiler | undefined | 10 | 29.466 | compiler subprocess: rustc metadata compilation (not resolver-only) |
| rust compiler | undefined | 100 | 78.360 | compiler subprocess: rustc metadata compilation (not resolver-only) |
| rust compiler | undefined | 1,000 | 564.926 | compiler subprocess: rustc metadata compilation (not resolver-only) |
| rust compiler | undefined | 3,000 | 1624.094 | compiler subprocess: rustc metadata compilation (not resolver-only) |
| zig compiler | declarations | 10 | 2.724 | compiler subprocess: Zig AST check (not resolver-only) |
| zig compiler | declarations | 100 | 3.338 | compiler subprocess: Zig AST check (not resolver-only) |
| zig compiler | declarations | 1,000 | 5.301 | compiler subprocess: Zig AST check (not resolver-only) |
| zig compiler | declarations | 3,000 | 18.633 | compiler subprocess: Zig AST check (not resolver-only) |
| zig compiler | mixed | 10 | 2.726 | compiler subprocess: Zig AST check (not resolver-only) |
| zig compiler | mixed | 100 | 2.969 | compiler subprocess: Zig AST check (not resolver-only) |
| zig compiler | mixed | 1,000 | 5.400 | compiler subprocess: Zig AST check (not resolver-only) |
| zig compiler | mixed | 3,000 | 13.689 | compiler subprocess: Zig AST check (not resolver-only) |
| zig compiler | references | 10 | 2.535 | compiler subprocess: Zig AST check (not resolver-only) |
| zig compiler | references | 100 | 2.623 | compiler subprocess: Zig AST check (not resolver-only) |
| zig compiler | references | 1,000 | 3.231 | compiler subprocess: Zig AST check (not resolver-only) |
| zig compiler | references | 3,000 | 3.817 | compiler subprocess: Zig AST check (not resolver-only) |
| zig compiler | sparse-undefined | 10 | 2.656 | compiler subprocess: Zig AST check (not resolver-only) |
| zig compiler | sparse-undefined | 100 | 3.356 | compiler subprocess: Zig AST check (not resolver-only) |
| zig compiler | sparse-undefined | 1,000 | 3.023 | compiler subprocess: Zig AST check (not resolver-only) |
| zig compiler | sparse-undefined | 3,000 | 3.359 | compiler subprocess: Zig AST check (not resolver-only) |
| zig compiler | undefined | 10 | 2.788 | compiler subprocess: Zig AST check (not resolver-only) |
| zig compiler | undefined | 100 | 2.868 | compiler subprocess: Zig AST check (not resolver-only) |
| zig compiler | undefined | 1,000 | 2.942 | compiler subprocess: Zig AST check (not resolver-only) |
| zig compiler | undefined | 3,000 | 3.281 | compiler subprocess: Zig AST check (not resolver-only) |

## Interpretation

- Eloque: name resolver only, with a prebuilt AST; no parsing.
- CPython `ast.parse`: parsing only; does not build a symbol table.
- CPython `symtable`: builds symbol tables but also reparses source internally; **not** isolated resolution.
- CPython permits unresolved names at compile time; undefined-name workloads therefore differ fundamentally from Eloque diagnostics.
- Workloads are structurally similar but not semantically equivalent. Source sizes and language-specific syntax differ.
- Native compiler subprocess measurements, when selected, include process startup and compiler phases beyond name resolution; they are displayed separately and cannot be ranked against Eloque resolver timings.
- Undefined-name cases intentionally fail native compilation; they measure error-path compiler invocation time, not successful resolution.
- **Not yet measured:** actual Eloque parse → resolve pipeline. It requires a repository-specific production API benchmark, not a guessed adapter.

Raw measurements: `samples.csv`; Eloque raw output: `eloque-resolver.log`.
