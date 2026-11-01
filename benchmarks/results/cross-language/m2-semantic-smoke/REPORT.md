# Eloque M2 cross-language semantic comparison

**Different measurement boundaries: do not interpret this as a compiler speed ranking.**

## In-process measurements (Eloque and CPython)

| Engine | Workload | N | Median (µs/op) | Boundary |
|---|---|---:|---:|---|
| CPython ast.parse | declarations | 10 | 20.327 | in-process parse + AST |
| CPython ast.parse | declarations | 100 | 152.450 | in-process parse + AST |
| CPython ast.parse | declarations | 1,000 | 1816.468 | in-process parse + AST |
| CPython ast.parse | declarations | 3,000 | 6743.134 | in-process parse + AST |
| CPython ast.parse | mixed | 10 | 31.254 | in-process parse + AST |
| CPython ast.parse | mixed | 100 | 300.491 | in-process parse + AST |
| CPython ast.parse | mixed | 1,000 | 4210.699 | in-process parse + AST |
| CPython ast.parse | mixed | 3,000 | 12797.661 | in-process parse + AST |
| CPython ast.parse | references | 10 | 18.923 | in-process parse + AST |
| CPython ast.parse | references | 100 | 119.577 | in-process parse + AST |
| CPython ast.parse | references | 1,000 | 2051.015 | in-process parse + AST |
| CPython ast.parse | references | 3,000 | 4889.068 | in-process parse + AST |
| CPython ast.parse | sparse-undefined | 10 | 17.535 | in-process parse + AST |
| CPython ast.parse | sparse-undefined | 100 | 119.703 | in-process parse + AST |
| CPython ast.parse | sparse-undefined | 1,000 | 1683.815 | in-process parse + AST |
| CPython ast.parse | sparse-undefined | 3,000 | 4337.060 | in-process parse + AST |
| CPython ast.parse | undefined | 10 | 14.822 | in-process parse + AST |
| CPython ast.parse | undefined | 100 | 128.272 | in-process parse + AST |
| CPython ast.parse | undefined | 1,000 | 1867.434 | in-process parse + AST |
| CPython ast.parse | undefined | 3,000 | 5969.444 | in-process parse + AST |
| CPython symtable | declarations | 10 | 21.702 | in-process symbol table (internally reparses source) |
| CPython symtable | declarations | 100 | 117.757 | in-process symbol table (internally reparses source) |
| CPython symtable | declarations | 1,000 | 1573.178 | in-process symbol table (internally reparses source) |
| CPython symtable | declarations | 3,000 | 4727.454 | in-process symbol table (internally reparses source) |
| CPython symtable | mixed | 10 | 24.721 | in-process symbol table (internally reparses source) |
| CPython symtable | mixed | 100 | 220.957 | in-process symbol table (internally reparses source) |
| CPython symtable | mixed | 1,000 | 3237.420 | in-process symbol table (internally reparses source) |
| CPython symtable | mixed | 3,000 | 11327.745 | in-process symbol table (internally reparses source) |
| CPython symtable | references | 10 | 16.489 | in-process symbol table (internally reparses source) |
| CPython symtable | references | 100 | 97.423 | in-process symbol table (internally reparses source) |
| CPython symtable | references | 1,000 | 1434.869 | in-process symbol table (internally reparses source) |
| CPython symtable | references | 3,000 | 3767.750 | in-process symbol table (internally reparses source) |
| CPython symtable | sparse-undefined | 10 | 16.757 | in-process symbol table (internally reparses source) |
| CPython symtable | sparse-undefined | 100 | 105.872 | in-process symbol table (internally reparses source) |
| CPython symtable | sparse-undefined | 1,000 | 1258.913 | in-process symbol table (internally reparses source) |
| CPython symtable | sparse-undefined | 3,000 | 3645.269 | in-process symbol table (internally reparses source) |
| CPython symtable | undefined | 10 | 14.963 | in-process symbol table (internally reparses source) |
| CPython symtable | undefined | 100 | 102.930 | in-process symbol table (internally reparses source) |
| CPython symtable | undefined | 1,000 | 1376.486 | in-process symbol table (internally reparses source) |
| CPython symtable | undefined | 3,000 | 4417.484 | in-process symbol table (internally reparses source) |
| Eloque resolver | declarations | 10 | 0.624 | in-process resolver only (prebuilt AST) |
| Eloque resolver | declarations | 100 | 4.736 | in-process resolver only (prebuilt AST) |
| Eloque resolver | declarations | 1,000 | 46.043 | in-process resolver only (prebuilt AST) |
| Eloque resolver | declarations | 3,000 | 130.247 | in-process resolver only (prebuilt AST) |
| Eloque resolver | mixed | 10 | 1.366 | in-process resolver only (prebuilt AST) |
| Eloque resolver | mixed | 100 | 12.894 | in-process resolver only (prebuilt AST) |
| Eloque resolver | mixed | 1,000 | 108.098 | in-process resolver only (prebuilt AST) |
| Eloque resolver | mixed | 3,000 | 340.579 | in-process resolver only (prebuilt AST) |
| Eloque resolver | references | 10 | 0.325 | in-process resolver only (prebuilt AST) |
| Eloque resolver | references | 100 | 1.578 | in-process resolver only (prebuilt AST) |
| Eloque resolver | references | 1,000 | 14.121 | in-process resolver only (prebuilt AST) |
| Eloque resolver | references | 3,000 | 41.638 | in-process resolver only (prebuilt AST) |
| Eloque resolver | sparse-undefined | 10 | 0.682 | in-process resolver only (prebuilt AST) |
| Eloque resolver | sparse-undefined | 100 | 2.023 | in-process resolver only (prebuilt AST) |
| Eloque resolver | sparse-undefined | 1,000 | 17.435 | in-process resolver only (prebuilt AST) |
| Eloque resolver | sparse-undefined | 3,000 | 48.992 | in-process resolver only (prebuilt AST) |
| Eloque resolver | undefined | 10 | 1.184 | in-process resolver only (prebuilt AST) |
| Eloque resolver | undefined | 100 | 13.139 | in-process resolver only (prebuilt AST) |
| Eloque resolver | undefined | 1,000 | 100.391 | in-process resolver only (prebuilt AST) |
| Eloque resolver | undefined | 3,000 | 438.136 | in-process resolver only (prebuilt AST) |

## Native compiler subprocess measurements (separate boundary)

Includes process startup and compiler work beyond name resolution.

| Engine | Workload | N | Median (ms/invocation) | Boundary |
|---|---|---:|---:|---|
| c compiler | declarations | 10 | 11.304 | compiler subprocess: Clang parse + semantic validation (syntax-only) |
| c compiler | declarations | 100 | 13.494 | compiler subprocess: Clang parse + semantic validation (syntax-only) |
| c compiler | declarations | 1,000 | 16.815 | compiler subprocess: Clang parse + semantic validation (syntax-only) |
| c compiler | declarations | 3,000 | 22.163 | compiler subprocess: Clang parse + semantic validation (syntax-only) |
| c compiler | mixed | 10 | 17.030 | compiler subprocess: Clang parse + semantic validation (syntax-only) |
| c compiler | mixed | 100 | 12.718 | compiler subprocess: Clang parse + semantic validation (syntax-only) |
| c compiler | mixed | 1,000 | 16.300 | compiler subprocess: Clang parse + semantic validation (syntax-only) |
| c compiler | mixed | 3,000 | 24.972 | compiler subprocess: Clang parse + semantic validation (syntax-only) |
| c compiler | references | 10 | 13.551 | compiler subprocess: Clang parse + semantic validation (syntax-only) |
| c compiler | references | 100 | 13.041 | compiler subprocess: Clang parse + semantic validation (syntax-only) |
| c compiler | references | 1,000 | 12.879 | compiler subprocess: Clang parse + semantic validation (syntax-only) |
| c compiler | references | 3,000 | 19.882 | compiler subprocess: Clang parse + semantic validation (syntax-only) |
| c compiler | sparse-undefined | 10 | 11.212 | compiler subprocess: Clang parse + semantic validation (syntax-only) |
| c compiler | sparse-undefined | 100 | 12.326 | compiler subprocess: Clang parse + semantic validation (syntax-only) |
| c compiler | sparse-undefined | 1,000 | 14.315 | compiler subprocess: Clang parse + semantic validation (syntax-only) |
| c compiler | sparse-undefined | 3,000 | 16.692 | compiler subprocess: Clang parse + semantic validation (syntax-only) |
| c compiler | undefined | 10 | 11.875 | compiler subprocess: Clang parse + semantic validation (syntax-only) |
| c compiler | undefined | 100 | 12.910 | compiler subprocess: Clang parse + semantic validation (syntax-only) |
| c compiler | undefined | 1,000 | 12.263 | compiler subprocess: Clang parse + semantic validation (syntax-only) |
| c compiler | undefined | 3,000 | 13.992 | compiler subprocess: Clang parse + semantic validation (syntax-only) |
| cpp compiler | declarations | 10 | 12.939 | compiler subprocess: Clang++ parse + semantic validation (syntax-only) |
| cpp compiler | declarations | 100 | 13.151 | compiler subprocess: Clang++ parse + semantic validation (syntax-only) |
| cpp compiler | declarations | 1,000 | 20.626 | compiler subprocess: Clang++ parse + semantic validation (syntax-only) |
| cpp compiler | declarations | 3,000 | 28.429 | compiler subprocess: Clang++ parse + semantic validation (syntax-only) |
| cpp compiler | mixed | 10 | 11.618 | compiler subprocess: Clang++ parse + semantic validation (syntax-only) |
| cpp compiler | mixed | 100 | 13.507 | compiler subprocess: Clang++ parse + semantic validation (syntax-only) |
| cpp compiler | mixed | 1,000 | 19.029 | compiler subprocess: Clang++ parse + semantic validation (syntax-only) |
| cpp compiler | mixed | 3,000 | 30.678 | compiler subprocess: Clang++ parse + semantic validation (syntax-only) |
| cpp compiler | references | 10 | 11.966 | compiler subprocess: Clang++ parse + semantic validation (syntax-only) |
| cpp compiler | references | 100 | 12.940 | compiler subprocess: Clang++ parse + semantic validation (syntax-only) |
| cpp compiler | references | 1,000 | 15.154 | compiler subprocess: Clang++ parse + semantic validation (syntax-only) |
| cpp compiler | references | 3,000 | 17.460 | compiler subprocess: Clang++ parse + semantic validation (syntax-only) |
| cpp compiler | sparse-undefined | 10 | 12.614 | compiler subprocess: Clang++ parse + semantic validation (syntax-only) |
| cpp compiler | sparse-undefined | 100 | 12.816 | compiler subprocess: Clang++ parse + semantic validation (syntax-only) |
| cpp compiler | sparse-undefined | 1,000 | 14.288 | compiler subprocess: Clang++ parse + semantic validation (syntax-only) |
| cpp compiler | sparse-undefined | 3,000 | 18.269 | compiler subprocess: Clang++ parse + semantic validation (syntax-only) |
| cpp compiler | undefined | 10 | 12.921 | compiler subprocess: Clang++ parse + semantic validation (syntax-only) |
| cpp compiler | undefined | 100 | 14.051 | compiler subprocess: Clang++ parse + semantic validation (syntax-only) |
| cpp compiler | undefined | 1,000 | 25.076 | compiler subprocess: Clang++ parse + semantic validation (syntax-only) |
| cpp compiler | undefined | 3,000 | 21.867 | compiler subprocess: Clang++ parse + semantic validation (syntax-only) |
| go compiler | declarations | 10 | 82.057 | compiler subprocess: Go object compilation (not resolver-only) |
| go compiler | declarations | 100 | 89.241 | compiler subprocess: Go object compilation (not resolver-only) |
| go compiler | declarations | 1,000 | 90.473 | compiler subprocess: Go object compilation (not resolver-only) |
| go compiler | declarations | 3,000 | 105.162 | compiler subprocess: Go object compilation (not resolver-only) |
| go compiler | mixed | 10 | 45.949 | compiler subprocess: Go object compilation (not resolver-only) |
| go compiler | mixed | 100 | 50.362 | compiler subprocess: Go object compilation (not resolver-only) |
| go compiler | mixed | 1,000 | 65.795 | compiler subprocess: Go object compilation (not resolver-only) |
| go compiler | mixed | 3,000 | 103.265 | compiler subprocess: Go object compilation (not resolver-only) |
| go compiler | references | 10 | 67.021 | compiler subprocess: Go object compilation (not resolver-only) |
| go compiler | references | 100 | 93.165 | compiler subprocess: Go object compilation (not resolver-only) |
| go compiler | references | 1,000 | 56.187 | compiler subprocess: Go object compilation (not resolver-only) |
| go compiler | references | 3,000 | 60.136 | compiler subprocess: Go object compilation (not resolver-only) |
| go compiler | sparse-undefined | 10 | 47.947 | compiler subprocess: Go object compilation (not resolver-only) |
| go compiler | sparse-undefined | 100 | 59.177 | compiler subprocess: Go object compilation (not resolver-only) |
| go compiler | sparse-undefined | 1,000 | 47.627 | compiler subprocess: Go object compilation (not resolver-only) |
| go compiler | sparse-undefined | 3,000 | 57.641 | compiler subprocess: Go object compilation (not resolver-only) |
| go compiler | undefined | 10 | 62.215 | compiler subprocess: Go object compilation (not resolver-only) |
| go compiler | undefined | 100 | 50.441 | compiler subprocess: Go object compilation (not resolver-only) |
| go compiler | undefined | 1,000 | 48.737 | compiler subprocess: Go object compilation (not resolver-only) |
| go compiler | undefined | 3,000 | 78.197 | compiler subprocess: Go object compilation (not resolver-only) |
| rust compiler | declarations | 10 | 14.825 | compiler subprocess: rustc metadata compilation (not resolver-only) |
| rust compiler | declarations | 100 | 16.124 | compiler subprocess: rustc metadata compilation (not resolver-only) |
| rust compiler | declarations | 1,000 | 27.907 | compiler subprocess: rustc metadata compilation (not resolver-only) |
| rust compiler | declarations | 3,000 | 61.173 | compiler subprocess: rustc metadata compilation (not resolver-only) |
| rust compiler | mixed | 10 | 15.351 | compiler subprocess: rustc metadata compilation (not resolver-only) |
| rust compiler | mixed | 100 | 15.577 | compiler subprocess: rustc metadata compilation (not resolver-only) |
| rust compiler | mixed | 1,000 | 30.370 | compiler subprocess: rustc metadata compilation (not resolver-only) |
| rust compiler | mixed | 3,000 | 56.864 | compiler subprocess: rustc metadata compilation (not resolver-only) |
| rust compiler | references | 10 | 14.503 | compiler subprocess: rustc metadata compilation (not resolver-only) |
| rust compiler | references | 100 | 21.275 | compiler subprocess: rustc metadata compilation (not resolver-only) |
| rust compiler | references | 1,000 | 23.141 | compiler subprocess: rustc metadata compilation (not resolver-only) |
| rust compiler | references | 3,000 | 50.315 | compiler subprocess: rustc metadata compilation (not resolver-only) |
| rust compiler | sparse-undefined | 10 | 21.232 | compiler subprocess: rustc metadata compilation (not resolver-only) |
| rust compiler | sparse-undefined | 100 | 24.154 | compiler subprocess: rustc metadata compilation (not resolver-only) |
| rust compiler | sparse-undefined | 1,000 | 28.511 | compiler subprocess: rustc metadata compilation (not resolver-only) |
| rust compiler | sparse-undefined | 3,000 | 43.934 | compiler subprocess: rustc metadata compilation (not resolver-only) |
| rust compiler | undefined | 10 | 29.843 | compiler subprocess: rustc metadata compilation (not resolver-only) |
| rust compiler | undefined | 100 | 68.838 | compiler subprocess: rustc metadata compilation (not resolver-only) |
| rust compiler | undefined | 1,000 | 522.969 | compiler subprocess: rustc metadata compilation (not resolver-only) |
| rust compiler | undefined | 3,000 | 1615.658 | compiler subprocess: rustc metadata compilation (not resolver-only) |
| zig compiler | declarations | 10 | 2.409 | compiler subprocess: Zig AST check (not resolver-only) |
| zig compiler | declarations | 100 | 2.764 | compiler subprocess: Zig AST check (not resolver-only) |
| zig compiler | declarations | 1,000 | 4.785 | compiler subprocess: Zig AST check (not resolver-only) |
| zig compiler | declarations | 3,000 | 20.552 | compiler subprocess: Zig AST check (not resolver-only) |
| zig compiler | mixed | 10 | 3.056 | compiler subprocess: Zig AST check (not resolver-only) |
| zig compiler | mixed | 100 | 2.939 | compiler subprocess: Zig AST check (not resolver-only) |
| zig compiler | mixed | 1,000 | 4.492 | compiler subprocess: Zig AST check (not resolver-only) |
| zig compiler | mixed | 3,000 | 13.560 | compiler subprocess: Zig AST check (not resolver-only) |
| zig compiler | references | 10 | 3.105 | compiler subprocess: Zig AST check (not resolver-only) |
| zig compiler | references | 100 | 3.745 | compiler subprocess: Zig AST check (not resolver-only) |
| zig compiler | references | 1,000 | 3.091 | compiler subprocess: Zig AST check (not resolver-only) |
| zig compiler | references | 3,000 | 3.512 | compiler subprocess: Zig AST check (not resolver-only) |
| zig compiler | sparse-undefined | 10 | 2.739 | compiler subprocess: Zig AST check (not resolver-only) |
| zig compiler | sparse-undefined | 100 | 2.496 | compiler subprocess: Zig AST check (not resolver-only) |
| zig compiler | sparse-undefined | 1,000 | 2.523 | compiler subprocess: Zig AST check (not resolver-only) |
| zig compiler | sparse-undefined | 3,000 | 3.099 | compiler subprocess: Zig AST check (not resolver-only) |
| zig compiler | undefined | 10 | 3.277 | compiler subprocess: Zig AST check (not resolver-only) |
| zig compiler | undefined | 100 | 2.923 | compiler subprocess: Zig AST check (not resolver-only) |
| zig compiler | undefined | 1,000 | 3.149 | compiler subprocess: Zig AST check (not resolver-only) |
| zig compiler | undefined | 3,000 | 2.793 | compiler subprocess: Zig AST check (not resolver-only) |

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
