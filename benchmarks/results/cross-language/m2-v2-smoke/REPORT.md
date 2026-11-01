# Eloque M2 cross-language semantic comparison v2

**Seven languages, different measurement boundaries: these are ranked observed timings, not equivalent compiler-phase speed claims.**

## In-process measurements (Eloque and CPython)

| Engine | Workload | N | Median (µs/op) | Boundary |
|---|---|---:|---:|---|
| CPython ast.parse | declarations | 10 | 32.085 | in-process parse + AST |
| CPython ast.parse | declarations | 100 | 242.409 | in-process parse + AST |
| CPython ast.parse | declarations | 1,000 | 3380.622 | in-process parse + AST |
| CPython ast.parse | declarations | 3,000 | 10275.560 | in-process parse + AST |
| CPython ast.parse | mixed | 10 | 51.702 | in-process parse + AST |
| CPython ast.parse | mixed | 100 | 752.410 | in-process parse + AST |
| CPython ast.parse | mixed | 1,000 | 6488.504 | in-process parse + AST |
| CPython ast.parse | mixed | 3,000 | 22406.506 | in-process parse + AST |
| CPython ast.parse | references | 10 | 27.968 | in-process parse + AST |
| CPython ast.parse | references | 100 | 202.416 | in-process parse + AST |
| CPython ast.parse | references | 1,000 | 2880.193 | in-process parse + AST |
| CPython ast.parse | references | 3,000 | 10587.694 | in-process parse + AST |
| CPython ast.parse | sparse-undefined | 10 | 28.950 | in-process parse + AST |
| CPython ast.parse | sparse-undefined | 100 | 207.519 | in-process parse + AST |
| CPython ast.parse | sparse-undefined | 1,000 | 3187.740 | in-process parse + AST |
| CPython ast.parse | sparse-undefined | 3,000 | 9182.055 | in-process parse + AST |
| CPython ast.parse | undefined | 10 | 32.702 | in-process parse + AST |
| CPython ast.parse | undefined | 100 | 212.292 | in-process parse + AST |
| CPython ast.parse | undefined | 1,000 | 3418.425 | in-process parse + AST |
| CPython ast.parse | undefined | 3,000 | 9041.721 | in-process parse + AST |
| CPython symtable | declarations | 10 | 31.575 | in-process symbol table (internally reparses source) |
| CPython symtable | declarations | 100 | 189.552 | in-process symbol table (internally reparses source) |
| CPython symtable | declarations | 1,000 | 2972.135 | in-process symbol table (internally reparses source) |
| CPython symtable | declarations | 3,000 | 11644.709 | in-process symbol table (internally reparses source) |
| CPython symtable | mixed | 10 | 44.008 | in-process symbol table (internally reparses source) |
| CPython symtable | mixed | 100 | 415.330 | in-process symbol table (internally reparses source) |
| CPython symtable | mixed | 1,000 | 5567.283 | in-process symbol table (internally reparses source) |
| CPython symtable | mixed | 3,000 | 15095.012 | in-process symbol table (internally reparses source) |
| CPython symtable | references | 10 | 30.074 | in-process symbol table (internally reparses source) |
| CPython symtable | references | 100 | 172.402 | in-process symbol table (internally reparses source) |
| CPython symtable | references | 1,000 | 2042.039 | in-process symbol table (internally reparses source) |
| CPython symtable | references | 3,000 | 6157.840 | in-process symbol table (internally reparses source) |
| CPython symtable | sparse-undefined | 10 | 40.365 | in-process symbol table (internally reparses source) |
| CPython symtable | sparse-undefined | 100 | 161.585 | in-process symbol table (internally reparses source) |
| CPython symtable | sparse-undefined | 1,000 | 2087.709 | in-process symbol table (internally reparses source) |
| CPython symtable | sparse-undefined | 3,000 | 10877.534 | in-process symbol table (internally reparses source) |
| CPython symtable | undefined | 10 | 26.228 | in-process symbol table (internally reparses source) |
| CPython symtable | undefined | 100 | 183.952 | in-process symbol table (internally reparses source) |
| CPython symtable | undefined | 1,000 | 2906.349 | in-process symbol table (internally reparses source) |
| CPython symtable | undefined | 3,000 | 18676.076 | in-process symbol table (internally reparses source) |
| Eloque FailFast parse+resolve | declarations | 10 | 7.049 | in-process failfast parse + name resolution |
| Eloque FailFast parse+resolve | declarations | 100 | 58.747 | in-process failfast parse + name resolution |
| Eloque FailFast parse+resolve | declarations | 1,000 | 661.027 | in-process failfast parse + name resolution |
| Eloque FailFast parse+resolve | declarations | 3,000 | 1963.657 | in-process failfast parse + name resolution |
| Eloque FailFast parse+resolve | mixed | 10 | 7.398 | in-process failfast parse + name resolution |
| Eloque FailFast parse+resolve | mixed | 100 | 58.110 | in-process failfast parse + name resolution |
| Eloque FailFast parse+resolve | mixed | 1,000 | 611.597 | in-process failfast parse + name resolution |
| Eloque FailFast parse+resolve | mixed | 3,000 | 1900.162 | in-process failfast parse + name resolution |
| Eloque FailFast parse+resolve | references | 10 | 3.439 | in-process failfast parse + name resolution |
| Eloque FailFast parse+resolve | references | 100 | 21.501 | in-process failfast parse + name resolution |
| Eloque FailFast parse+resolve | references | 1,000 | 200.675 | in-process failfast parse + name resolution |
| Eloque FailFast parse+resolve | references | 3,000 | 622.312 | in-process failfast parse + name resolution |
| Eloque FailFast parse+resolve | sparse-undefined | 10 | 3.975 | in-process failfast parse + name resolution |
| Eloque FailFast parse+resolve | sparse-undefined | 100 | 21.043 | in-process failfast parse + name resolution |
| Eloque FailFast parse+resolve | sparse-undefined | 1,000 | 219.699 | in-process failfast parse + name resolution |
| Eloque FailFast parse+resolve | sparse-undefined | 3,000 | 635.288 | in-process failfast parse + name resolution |
| Eloque FailFast parse+resolve | undefined | 10 | 5.814 | in-process failfast parse + name resolution |
| Eloque FailFast parse+resolve | undefined | 100 | 45.168 | in-process failfast parse + name resolution |
| Eloque FailFast parse+resolve | undefined | 1,000 | 330.947 | in-process failfast parse + name resolution |
| Eloque FailFast parse+resolve | undefined | 3,000 | 1157.896 | in-process failfast parse + name resolution |
| Eloque Recovering parse+resolve | declarations | 10 | 7.714 | in-process recovering parse + name resolution |
| Eloque Recovering parse+resolve | declarations | 100 | 63.823 | in-process recovering parse + name resolution |
| Eloque Recovering parse+resolve | declarations | 1,000 | 663.810 | in-process recovering parse + name resolution |
| Eloque Recovering parse+resolve | declarations | 3,000 | 2101.845 | in-process recovering parse + name resolution |
| Eloque Recovering parse+resolve | mixed | 10 | 7.563 | in-process recovering parse + name resolution |
| Eloque Recovering parse+resolve | mixed | 100 | 59.794 | in-process recovering parse + name resolution |
| Eloque Recovering parse+resolve | mixed | 1,000 | 609.261 | in-process recovering parse + name resolution |
| Eloque Recovering parse+resolve | mixed | 3,000 | 2039.673 | in-process recovering parse + name resolution |
| Eloque Recovering parse+resolve | references | 10 | 3.608 | in-process recovering parse + name resolution |
| Eloque Recovering parse+resolve | references | 100 | 23.157 | in-process recovering parse + name resolution |
| Eloque Recovering parse+resolve | references | 1,000 | 223.203 | in-process recovering parse + name resolution |
| Eloque Recovering parse+resolve | references | 3,000 | 612.445 | in-process recovering parse + name resolution |
| Eloque Recovering parse+resolve | sparse-undefined | 10 | 3.959 | in-process recovering parse + name resolution |
| Eloque Recovering parse+resolve | sparse-undefined | 100 | 23.121 | in-process recovering parse + name resolution |
| Eloque Recovering parse+resolve | sparse-undefined | 1,000 | 221.074 | in-process recovering parse + name resolution |
| Eloque Recovering parse+resolve | sparse-undefined | 3,000 | 612.138 | in-process recovering parse + name resolution |
| Eloque Recovering parse+resolve | undefined | 10 | 4.665 | in-process recovering parse + name resolution |
| Eloque Recovering parse+resolve | undefined | 100 | 40.430 | in-process recovering parse + name resolution |
| Eloque Recovering parse+resolve | undefined | 1,000 | 363.564 | in-process recovering parse + name resolution |
| Eloque Recovering parse+resolve | undefined | 3,000 | 1162.812 | in-process recovering parse + name resolution |

## Observed timing ranks (by workload and size)

Ranks below are descriptive only: operations and process boundaries differ.

| Workload | N | Family | Rank | Engine | Median (ms/op) |
|---|---:|---|---:|---|---:|
| declarations | 10 | in-process | 1 | Eloque FailFast parse+resolve | 0.007049 |
| declarations | 10 | in-process | 2 | Eloque Recovering parse+resolve | 0.007714 |
| declarations | 10 | in-process | 3 | CPython symtable | 0.031575 |
| declarations | 10 | in-process | 4 | CPython ast.parse | 0.032085 |
| declarations | 100 | in-process | 1 | Eloque FailFast parse+resolve | 0.058747 |
| declarations | 100 | in-process | 2 | Eloque Recovering parse+resolve | 0.063823 |
| declarations | 100 | in-process | 3 | CPython symtable | 0.189552 |
| declarations | 100 | in-process | 4 | CPython ast.parse | 0.242409 |
| declarations | 1000 | in-process | 1 | Eloque FailFast parse+resolve | 0.661027 |
| declarations | 1000 | in-process | 2 | Eloque Recovering parse+resolve | 0.663810 |
| declarations | 1000 | in-process | 3 | CPython symtable | 2.972135 |
| declarations | 1000 | in-process | 4 | CPython ast.parse | 3.380622 |
| declarations | 3000 | in-process | 1 | Eloque FailFast parse+resolve | 1.963657 |
| declarations | 3000 | in-process | 2 | Eloque Recovering parse+resolve | 2.101845 |
| declarations | 3000 | in-process | 3 | CPython ast.parse | 10.275560 |
| declarations | 3000 | in-process | 4 | CPython symtable | 11.644709 |
| references | 10 | in-process | 1 | Eloque FailFast parse+resolve | 0.003439 |
| references | 10 | in-process | 2 | Eloque Recovering parse+resolve | 0.003608 |
| references | 10 | in-process | 3 | CPython ast.parse | 0.027968 |
| references | 10 | in-process | 4 | CPython symtable | 0.030074 |
| references | 100 | in-process | 1 | Eloque FailFast parse+resolve | 0.021501 |
| references | 100 | in-process | 2 | Eloque Recovering parse+resolve | 0.023157 |
| references | 100 | in-process | 3 | CPython symtable | 0.172402 |
| references | 100 | in-process | 4 | CPython ast.parse | 0.202416 |
| references | 1000 | in-process | 1 | Eloque FailFast parse+resolve | 0.200675 |
| references | 1000 | in-process | 2 | Eloque Recovering parse+resolve | 0.223203 |
| references | 1000 | in-process | 3 | CPython symtable | 2.042039 |
| references | 1000 | in-process | 4 | CPython ast.parse | 2.880193 |
| references | 3000 | in-process | 1 | Eloque Recovering parse+resolve | 0.612445 |
| references | 3000 | in-process | 2 | Eloque FailFast parse+resolve | 0.622312 |
| references | 3000 | in-process | 3 | CPython symtable | 6.157840 |
| references | 3000 | in-process | 4 | CPython ast.parse | 10.587694 |
| mixed | 10 | in-process | 1 | Eloque FailFast parse+resolve | 0.007398 |
| mixed | 10 | in-process | 2 | Eloque Recovering parse+resolve | 0.007563 |
| mixed | 10 | in-process | 3 | CPython symtable | 0.044008 |
| mixed | 10 | in-process | 4 | CPython ast.parse | 0.051702 |
| mixed | 100 | in-process | 1 | Eloque FailFast parse+resolve | 0.058110 |
| mixed | 100 | in-process | 2 | Eloque Recovering parse+resolve | 0.059794 |
| mixed | 100 | in-process | 3 | CPython symtable | 0.415330 |
| mixed | 100 | in-process | 4 | CPython ast.parse | 0.752410 |
| mixed | 1000 | in-process | 1 | Eloque Recovering parse+resolve | 0.609261 |
| mixed | 1000 | in-process | 2 | Eloque FailFast parse+resolve | 0.611597 |
| mixed | 1000 | in-process | 3 | CPython symtable | 5.567284 |
| mixed | 1000 | in-process | 4 | CPython ast.parse | 6.488504 |
| mixed | 3000 | in-process | 1 | Eloque FailFast parse+resolve | 1.900162 |
| mixed | 3000 | in-process | 2 | Eloque Recovering parse+resolve | 2.039673 |
| mixed | 3000 | in-process | 3 | CPython symtable | 15.095012 |
| mixed | 3000 | in-process | 4 | CPython ast.parse | 22.406506 |
| sparse-undefined | 10 | in-process | 1 | Eloque Recovering parse+resolve | 0.003959 |
| sparse-undefined | 10 | in-process | 2 | Eloque FailFast parse+resolve | 0.003975 |
| sparse-undefined | 10 | in-process | 3 | CPython ast.parse | 0.028950 |
| sparse-undefined | 10 | in-process | 4 | CPython symtable | 0.040365 |
| sparse-undefined | 100 | in-process | 1 | Eloque FailFast parse+resolve | 0.021043 |
| sparse-undefined | 100 | in-process | 2 | Eloque Recovering parse+resolve | 0.023121 |
| sparse-undefined | 100 | in-process | 3 | CPython symtable | 0.161585 |
| sparse-undefined | 100 | in-process | 4 | CPython ast.parse | 0.207519 |
| sparse-undefined | 1000 | in-process | 1 | Eloque FailFast parse+resolve | 0.219699 |
| sparse-undefined | 1000 | in-process | 2 | Eloque Recovering parse+resolve | 0.221074 |
| sparse-undefined | 1000 | in-process | 3 | CPython symtable | 2.087709 |
| sparse-undefined | 1000 | in-process | 4 | CPython ast.parse | 3.187740 |
| sparse-undefined | 3000 | in-process | 1 | Eloque Recovering parse+resolve | 0.612138 |
| sparse-undefined | 3000 | in-process | 2 | Eloque FailFast parse+resolve | 0.635288 |
| sparse-undefined | 3000 | in-process | 3 | CPython ast.parse | 9.182054 |
| sparse-undefined | 3000 | in-process | 4 | CPython symtable | 10.877534 |
| undefined | 10 | in-process | 1 | Eloque Recovering parse+resolve | 0.004665 |
| undefined | 10 | in-process | 2 | Eloque FailFast parse+resolve | 0.005814 |
| undefined | 10 | in-process | 3 | CPython symtable | 0.026228 |
| undefined | 10 | in-process | 4 | CPython ast.parse | 0.032702 |
| undefined | 100 | in-process | 1 | Eloque Recovering parse+resolve | 0.040430 |
| undefined | 100 | in-process | 2 | Eloque FailFast parse+resolve | 0.045168 |
| undefined | 100 | in-process | 3 | CPython symtable | 0.183952 |
| undefined | 100 | in-process | 4 | CPython ast.parse | 0.212292 |
| undefined | 1000 | in-process | 1 | Eloque FailFast parse+resolve | 0.330947 |
| undefined | 1000 | in-process | 2 | Eloque Recovering parse+resolve | 0.363564 |
| undefined | 1000 | in-process | 3 | CPython symtable | 2.906349 |
| undefined | 1000 | in-process | 4 | CPython ast.parse | 3.418425 |
| undefined | 3000 | in-process | 1 | Eloque FailFast parse+resolve | 1.157896 |
| undefined | 3000 | in-process | 2 | Eloque Recovering parse+resolve | 1.162812 |
| undefined | 3000 | in-process | 3 | CPython ast.parse | 9.041721 |
| undefined | 3000 | in-process | 4 | CPython symtable | 18.676076 |

## Interpretation

- Eloque: parse + name resolution in-process; Recovering is the primary path, FailFast is secondary.
- CPython `ast.parse`: parsing only; does not build a symbol table.
- CPython `symtable`: builds symbol tables but also reparses source internally; **not** isolated resolution.
- CPython permits unresolved names at compile time; undefined-name workloads therefore differ fundamentally from Eloque diagnostics.
- Workloads are structurally similar but not semantically equivalent. Source sizes and language-specific syntax differ.
- Native compiler subprocess measurements, when selected, include process startup and compiler phases beyond name resolution; they are displayed separately and cannot be ranked against Eloque resolver timings.
- Undefined-name cases intentionally fail native compilation; they measure error-path compiler invocation time, not successful resolution.
- Native compiler invocations are ranked separately because they include startup and extra compiler phases.

Raw measurements: `samples.csv`; Eloque raw output: `eloque-pipeline.log`.
