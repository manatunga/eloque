# Eloque M2 cross-language semantic comparison v2

**Seven languages, different measurement boundaries: these are ranked observed timings, not equivalent compiler-phase speed claims.**

## In-process measurements (Eloque and CPython)

| Engine | Workload | N | Median (µs/op) | Boundary |
|---|---|---:|---:|---|
| CPython ast.parse | declarations | 10 | 33.970 | in-process parse + AST |
| CPython ast.parse | declarations | 100 | 271.421 | in-process parse + AST |
| CPython ast.parse | declarations | 1,000 | 3677.015 | in-process parse + AST |
| CPython ast.parse | declarations | 3,000 | 12166.904 | in-process parse + AST |
| CPython ast.parse | mixed | 10 | 53.110 | in-process parse + AST |
| CPython ast.parse | mixed | 100 | 415.679 | in-process parse + AST |
| CPython ast.parse | mixed | 1,000 | 7478.221 | in-process parse + AST |
| CPython ast.parse | mixed | 3,000 | 25515.754 | in-process parse + AST |
| CPython ast.parse | references | 10 | 30.852 | in-process parse + AST |
| CPython ast.parse | references | 100 | 202.610 | in-process parse + AST |
| CPython ast.parse | references | 1,000 | 2860.245 | in-process parse + AST |
| CPython ast.parse | references | 3,000 | 8202.715 | in-process parse + AST |
| CPython ast.parse | sparse-undefined | 10 | 28.597 | in-process parse + AST |
| CPython ast.parse | sparse-undefined | 100 | 202.470 | in-process parse + AST |
| CPython ast.parse | sparse-undefined | 1,000 | 2806.021 | in-process parse + AST |
| CPython ast.parse | sparse-undefined | 3,000 | 8996.443 | in-process parse + AST |
| CPython ast.parse | undefined | 10 | 27.876 | in-process parse + AST |
| CPython ast.parse | undefined | 100 | 206.530 | in-process parse + AST |
| CPython ast.parse | undefined | 1,000 | 2561.350 | in-process parse + AST |
| CPython ast.parse | undefined | 3,000 | 8558.334 | in-process parse + AST |
| CPython symtable | declarations | 10 | 27.880 | in-process symbol table (internally reparses source) |
| CPython symtable | declarations | 100 | 203.928 | in-process symbol table (internally reparses source) |
| CPython symtable | declarations | 1,000 | 2644.215 | in-process symbol table (internally reparses source) |
| CPython symtable | declarations | 3,000 | 8366.105 | in-process symbol table (internally reparses source) |
| CPython symtable | mixed | 10 | 46.799 | in-process symbol table (internally reparses source) |
| CPython symtable | mixed | 100 | 381.509 | in-process symbol table (internally reparses source) |
| CPython symtable | mixed | 1,000 | 5582.418 | in-process symbol table (internally reparses source) |
| CPython symtable | mixed | 3,000 | 17534.834 | in-process symbol table (internally reparses source) |
| CPython symtable | references | 10 | 27.606 | in-process symbol table (internally reparses source) |
| CPython symtable | references | 100 | 183.580 | in-process symbol table (internally reparses source) |
| CPython symtable | references | 1,000 | 2077.128 | in-process symbol table (internally reparses source) |
| CPython symtable | references | 3,000 | 6524.007 | in-process symbol table (internally reparses source) |
| CPython symtable | sparse-undefined | 10 | 27.917 | in-process symbol table (internally reparses source) |
| CPython symtable | sparse-undefined | 100 | 146.674 | in-process symbol table (internally reparses source) |
| CPython symtable | sparse-undefined | 1,000 | 2016.639 | in-process symbol table (internally reparses source) |
| CPython symtable | sparse-undefined | 3,000 | 6938.886 | in-process symbol table (internally reparses source) |
| CPython symtable | undefined | 10 | 28.148 | in-process symbol table (internally reparses source) |
| CPython symtable | undefined | 100 | 157.355 | in-process symbol table (internally reparses source) |
| CPython symtable | undefined | 1,000 | 2492.515 | in-process symbol table (internally reparses source) |
| CPython symtable | undefined | 3,000 | 6548.321 | in-process symbol table (internally reparses source) |
| Eloque FailFast parse+resolve | declarations | 10 | 7.547 | in-process failfast parse + name resolution |
| Eloque FailFast parse+resolve | declarations | 100 | 59.691 | in-process failfast parse + name resolution |
| Eloque FailFast parse+resolve | declarations | 1,000 | 631.338 | in-process failfast parse + name resolution |
| Eloque FailFast parse+resolve | declarations | 3,000 | 1932.346 | in-process failfast parse + name resolution |
| Eloque FailFast parse+resolve | mixed | 10 | 7.644 | in-process failfast parse + name resolution |
| Eloque FailFast parse+resolve | mixed | 100 | 59.457 | in-process failfast parse + name resolution |
| Eloque FailFast parse+resolve | mixed | 1,000 | 616.103 | in-process failfast parse + name resolution |
| Eloque FailFast parse+resolve | mixed | 3,000 | 1993.480 | in-process failfast parse + name resolution |
| Eloque FailFast parse+resolve | references | 10 | 3.443 | in-process failfast parse + name resolution |
| Eloque FailFast parse+resolve | references | 100 | 21.907 | in-process failfast parse + name resolution |
| Eloque FailFast parse+resolve | references | 1,000 | 201.350 | in-process failfast parse + name resolution |
| Eloque FailFast parse+resolve | references | 3,000 | 598.344 | in-process failfast parse + name resolution |
| Eloque FailFast parse+resolve | sparse-undefined | 10 | 3.980 | in-process failfast parse + name resolution |
| Eloque FailFast parse+resolve | sparse-undefined | 100 | 22.165 | in-process failfast parse + name resolution |
| Eloque FailFast parse+resolve | sparse-undefined | 1,000 | 211.876 | in-process failfast parse + name resolution |
| Eloque FailFast parse+resolve | sparse-undefined | 3,000 | 621.641 | in-process failfast parse + name resolution |
| Eloque FailFast parse+resolve | undefined | 10 | 4.905 | in-process failfast parse + name resolution |
| Eloque FailFast parse+resolve | undefined | 100 | 37.972 | in-process failfast parse + name resolution |
| Eloque FailFast parse+resolve | undefined | 1,000 | 364.151 | in-process failfast parse + name resolution |
| Eloque FailFast parse+resolve | undefined | 3,000 | 1171.557 | in-process failfast parse + name resolution |
| Eloque Recovering parse+resolve | declarations | 10 | 7.563 | in-process recovering parse + name resolution |
| Eloque Recovering parse+resolve | declarations | 100 | 61.742 | in-process recovering parse + name resolution |
| Eloque Recovering parse+resolve | declarations | 1,000 | 640.125 | in-process recovering parse + name resolution |
| Eloque Recovering parse+resolve | declarations | 3,000 | 1968.826 | in-process recovering parse + name resolution |
| Eloque Recovering parse+resolve | mixed | 10 | 7.748 | in-process recovering parse + name resolution |
| Eloque Recovering parse+resolve | mixed | 100 | 62.815 | in-process recovering parse + name resolution |
| Eloque Recovering parse+resolve | mixed | 1,000 | 677.803 | in-process recovering parse + name resolution |
| Eloque Recovering parse+resolve | mixed | 3,000 | 2059.516 | in-process recovering parse + name resolution |
| Eloque Recovering parse+resolve | references | 10 | 3.577 | in-process recovering parse + name resolution |
| Eloque Recovering parse+resolve | references | 100 | 22.535 | in-process recovering parse + name resolution |
| Eloque Recovering parse+resolve | references | 1,000 | 225.641 | in-process recovering parse + name resolution |
| Eloque Recovering parse+resolve | references | 3,000 | 679.809 | in-process recovering parse + name resolution |
| Eloque Recovering parse+resolve | sparse-undefined | 10 | 3.894 | in-process recovering parse + name resolution |
| Eloque Recovering parse+resolve | sparse-undefined | 100 | 23.983 | in-process recovering parse + name resolution |
| Eloque Recovering parse+resolve | sparse-undefined | 1,000 | 219.337 | in-process recovering parse + name resolution |
| Eloque Recovering parse+resolve | sparse-undefined | 3,000 | 659.643 | in-process recovering parse + name resolution |
| Eloque Recovering parse+resolve | undefined | 10 | 4.957 | in-process recovering parse + name resolution |
| Eloque Recovering parse+resolve | undefined | 100 | 39.023 | in-process recovering parse + name resolution |
| Eloque Recovering parse+resolve | undefined | 1,000 | 377.635 | in-process recovering parse + name resolution |
| Eloque Recovering parse+resolve | undefined | 3,000 | 1218.360 | in-process recovering parse + name resolution |

## Observed timing ranks (by workload and size)

Ranks below are descriptive only: operations and process boundaries differ.

| Workload | N | Family | Rank | Engine | Median (ms/op) |
|---|---:|---|---:|---|---:|
| declarations | 10 | in-process | 1 | Eloque FailFast parse+resolve | 0.007547 |
| declarations | 10 | in-process | 2 | Eloque Recovering parse+resolve | 0.007563 |
| declarations | 10 | in-process | 3 | CPython symtable | 0.027880 |
| declarations | 10 | in-process | 4 | CPython ast.parse | 0.033970 |
| declarations | 10 | compiler subprocess | 1 | c compiler | 19.955925 |
| declarations | 10 | compiler subprocess | 2 | cpp compiler | 21.232931 |
| declarations | 10 | compiler subprocess | 3 | rust compiler | 29.367076 |
| declarations | 10 | compiler subprocess | 4 | go compiler | 55.133881 |
| declarations | 10 | compiler subprocess | 5 | zig compiler | 449.523817 |
| declarations | 100 | in-process | 1 | Eloque FailFast parse+resolve | 0.059691 |
| declarations | 100 | in-process | 2 | Eloque Recovering parse+resolve | 0.061742 |
| declarations | 100 | in-process | 3 | CPython symtable | 0.203928 |
| declarations | 100 | in-process | 4 | CPython ast.parse | 0.271421 |
| declarations | 100 | compiler subprocess | 1 | c compiler | 19.789823 |
| declarations | 100 | compiler subprocess | 2 | cpp compiler | 24.788455 |
| declarations | 100 | compiler subprocess | 3 | rust compiler | 32.013947 |
| declarations | 100 | compiler subprocess | 4 | go compiler | 71.310835 |
| declarations | 100 | compiler subprocess | 5 | zig compiler | 449.179840 |
| declarations | 1000 | in-process | 1 | Eloque FailFast parse+resolve | 0.631338 |
| declarations | 1000 | in-process | 2 | Eloque Recovering parse+resolve | 0.640125 |
| declarations | 1000 | in-process | 3 | CPython symtable | 2.644215 |
| declarations | 1000 | in-process | 4 | CPython ast.parse | 3.677015 |
| declarations | 1000 | compiler subprocess | 1 | c compiler | 26.956041 |
| declarations | 1000 | compiler subprocess | 2 | cpp compiler | 31.577652 |
| declarations | 1000 | compiler subprocess | 3 | rust compiler | 53.485724 |
| declarations | 1000 | compiler subprocess | 4 | go compiler | 97.218019 |
| declarations | 1000 | compiler subprocess | 5 | zig compiler | 453.910836 |
| declarations | 3000 | in-process | 1 | Eloque FailFast parse+resolve | 1.932346 |
| declarations | 3000 | in-process | 2 | Eloque Recovering parse+resolve | 1.968826 |
| declarations | 3000 | in-process | 3 | CPython symtable | 8.366105 |
| declarations | 3000 | in-process | 4 | CPython ast.parse | 12.166904 |
| declarations | 3000 | compiler subprocess | 1 | c compiler | 41.601371 |
| declarations | 3000 | compiler subprocess | 2 | cpp compiler | 50.763134 |
| declarations | 3000 | compiler subprocess | 3 | rust compiler | 101.797717 |
| declarations | 3000 | compiler subprocess | 4 | go compiler | 143.327524 |
| declarations | 3000 | compiler subprocess | 5 | zig compiler | 463.070850 |
| references | 10 | in-process | 1 | Eloque FailFast parse+resolve | 0.003443 |
| references | 10 | in-process | 2 | Eloque Recovering parse+resolve | 0.003577 |
| references | 10 | in-process | 3 | CPython symtable | 0.027606 |
| references | 10 | in-process | 4 | CPython ast.parse | 0.030852 |
| references | 10 | compiler subprocess | 1 | cpp compiler | 22.728587 |
| references | 10 | compiler subprocess | 2 | c compiler | 25.084639 |
| references | 10 | compiler subprocess | 3 | rust compiler | 28.525496 |
| references | 10 | compiler subprocess | 4 | go compiler | 74.328621 |
| references | 10 | compiler subprocess | 5 | zig compiler | 441.712293 |
| references | 100 | in-process | 1 | Eloque FailFast parse+resolve | 0.021907 |
| references | 100 | in-process | 2 | Eloque Recovering parse+resolve | 0.022535 |
| references | 100 | in-process | 3 | CPython symtable | 0.183580 |
| references | 100 | in-process | 4 | CPython ast.parse | 0.202610 |
| references | 100 | compiler subprocess | 1 | cpp compiler | 20.355633 |
| references | 100 | compiler subprocess | 2 | c compiler | 20.397674 |
| references | 100 | compiler subprocess | 3 | rust compiler | 30.564234 |
| references | 100 | compiler subprocess | 4 | go compiler | 61.854734 |
| references | 100 | compiler subprocess | 5 | zig compiler | 456.242052 |
| references | 1000 | in-process | 1 | Eloque FailFast parse+resolve | 0.201350 |
| references | 1000 | in-process | 2 | Eloque Recovering parse+resolve | 0.225641 |
| references | 1000 | in-process | 3 | CPython symtable | 2.077128 |
| references | 1000 | in-process | 4 | CPython ast.parse | 2.860245 |
| references | 1000 | compiler subprocess | 1 | c compiler | 24.902728 |
| references | 1000 | compiler subprocess | 2 | cpp compiler | 29.103597 |
| references | 1000 | compiler subprocess | 3 | rust compiler | 41.705290 |
| references | 1000 | compiler subprocess | 4 | go compiler | 63.182582 |
| references | 1000 | compiler subprocess | 5 | zig compiler | 441.389869 |
| references | 3000 | in-process | 1 | Eloque FailFast parse+resolve | 0.598344 |
| references | 3000 | in-process | 2 | Eloque Recovering parse+resolve | 0.679809 |
| references | 3000 | in-process | 3 | CPython symtable | 6.524007 |
| references | 3000 | in-process | 4 | CPython ast.parse | 8.202715 |
| references | 3000 | compiler subprocess | 1 | c compiler | 30.200903 |
| references | 3000 | compiler subprocess | 2 | cpp compiler | 33.117214 |
| references | 3000 | compiler subprocess | 3 | go compiler | 69.583913 |
| references | 3000 | compiler subprocess | 4 | rust compiler | 83.244038 |
| references | 3000 | compiler subprocess | 5 | zig compiler | 427.677786 |
| mixed | 10 | in-process | 1 | Eloque FailFast parse+resolve | 0.007644 |
| mixed | 10 | in-process | 2 | Eloque Recovering parse+resolve | 0.007748 |
| mixed | 10 | in-process | 3 | CPython symtable | 0.046799 |
| mixed | 10 | in-process | 4 | CPython ast.parse | 0.053110 |
| mixed | 10 | compiler subprocess | 1 | c compiler | 21.287447 |
| mixed | 10 | compiler subprocess | 2 | cpp compiler | 22.020867 |
| mixed | 10 | compiler subprocess | 3 | rust compiler | 28.268609 |
| mixed | 10 | compiler subprocess | 4 | go compiler | 49.789974 |
| mixed | 10 | compiler subprocess | 5 | zig compiler | 454.841339 |
| mixed | 100 | in-process | 1 | Eloque FailFast parse+resolve | 0.059457 |
| mixed | 100 | in-process | 2 | Eloque Recovering parse+resolve | 0.062815 |
| mixed | 100 | in-process | 3 | CPython symtable | 0.381509 |
| mixed | 100 | in-process | 4 | CPython ast.parse | 0.415679 |
| mixed | 100 | compiler subprocess | 1 | c compiler | 22.527190 |
| mixed | 100 | compiler subprocess | 2 | cpp compiler | 23.390231 |
| mixed | 100 | compiler subprocess | 3 | rust compiler | 29.657321 |
| mixed | 100 | compiler subprocess | 4 | go compiler | 43.796817 |
| mixed | 100 | compiler subprocess | 5 | zig compiler | 451.283167 |
| mixed | 1000 | in-process | 1 | Eloque FailFast parse+resolve | 0.616103 |
| mixed | 1000 | in-process | 2 | Eloque Recovering parse+resolve | 0.677803 |
| mixed | 1000 | in-process | 3 | CPython symtable | 5.582418 |
| mixed | 1000 | in-process | 4 | CPython ast.parse | 7.478220 |
| mixed | 1000 | compiler subprocess | 1 | c compiler | 34.361745 |
| mixed | 1000 | compiler subprocess | 2 | cpp compiler | 36.188397 |
| mixed | 1000 | compiler subprocess | 3 | rust compiler | 55.095195 |
| mixed | 1000 | compiler subprocess | 4 | go compiler | 103.095209 |
| mixed | 1000 | compiler subprocess | 5 | zig compiler | 446.696404 |
| mixed | 3000 | in-process | 1 | Eloque FailFast parse+resolve | 1.993480 |
| mixed | 3000 | in-process | 2 | Eloque Recovering parse+resolve | 2.059516 |
| mixed | 3000 | in-process | 3 | CPython symtable | 17.534834 |
| mixed | 3000 | in-process | 4 | CPython ast.parse | 25.515754 |
| mixed | 3000 | compiler subprocess | 1 | c compiler | 52.345913 |
| mixed | 3000 | compiler subprocess | 2 | cpp compiler | 58.663674 |
| mixed | 3000 | compiler subprocess | 3 | rust compiler | 104.461692 |
| mixed | 3000 | compiler subprocess | 4 | go compiler | 144.474990 |
| mixed | 3000 | compiler subprocess | 5 | zig compiler | 441.948222 |
| sparse-undefined | 10 | in-process | 1 | Eloque Recovering parse+resolve | 0.003894 |
| sparse-undefined | 10 | in-process | 2 | Eloque FailFast parse+resolve | 0.003980 |
| sparse-undefined | 10 | in-process | 3 | CPython symtable | 0.027917 |
| sparse-undefined | 10 | in-process | 4 | CPython ast.parse | 0.028597 |
| sparse-undefined | 10 | compiler subprocess | 1 | cpp compiler | 22.375037 |
| sparse-undefined | 10 | compiler subprocess | 2 | c compiler | 22.755748 |
| sparse-undefined | 10 | compiler subprocess | 3 | rust compiler | 28.366164 |
| sparse-undefined | 10 | compiler subprocess | 4 | go compiler | 66.574057 |
| sparse-undefined | 10 | compiler subprocess | 5 | zig compiler | 442.542623 |
| sparse-undefined | 100 | in-process | 1 | Eloque FailFast parse+resolve | 0.022165 |
| sparse-undefined | 100 | in-process | 2 | Eloque Recovering parse+resolve | 0.023983 |
| sparse-undefined | 100 | in-process | 3 | CPython symtable | 0.146674 |
| sparse-undefined | 100 | in-process | 4 | CPython ast.parse | 0.202470 |
| sparse-undefined | 100 | compiler subprocess | 1 | c compiler | 21.868605 |
| sparse-undefined | 100 | compiler subprocess | 2 | cpp compiler | 23.359677 |
| sparse-undefined | 100 | compiler subprocess | 3 | rust compiler | 29.707078 |
| sparse-undefined | 100 | compiler subprocess | 4 | go compiler | 44.644079 |
| sparse-undefined | 100 | compiler subprocess | 5 | zig compiler | 446.718267 |
| sparse-undefined | 1000 | in-process | 1 | Eloque FailFast parse+resolve | 0.211876 |
| sparse-undefined | 1000 | in-process | 2 | Eloque Recovering parse+resolve | 0.219337 |
| sparse-undefined | 1000 | in-process | 3 | CPython symtable | 2.016639 |
| sparse-undefined | 1000 | in-process | 4 | CPython ast.parse | 2.806021 |
| sparse-undefined | 1000 | compiler subprocess | 1 | c compiler | 24.923246 |
| sparse-undefined | 1000 | compiler subprocess | 2 | cpp compiler | 27.693056 |
| sparse-undefined | 1000 | compiler subprocess | 3 | go compiler | 49.911661 |
| sparse-undefined | 1000 | compiler subprocess | 4 | rust compiler | 54.043224 |
| sparse-undefined | 1000 | compiler subprocess | 5 | zig compiler | 451.659055 |
| sparse-undefined | 3000 | in-process | 1 | Eloque FailFast parse+resolve | 0.621641 |
| sparse-undefined | 3000 | in-process | 2 | Eloque Recovering parse+resolve | 0.659643 |
| sparse-undefined | 3000 | in-process | 3 | CPython symtable | 6.938886 |
| sparse-undefined | 3000 | in-process | 4 | CPython ast.parse | 8.996443 |
| sparse-undefined | 3000 | compiler subprocess | 1 | c compiler | 32.264954 |
| sparse-undefined | 3000 | compiler subprocess | 2 | cpp compiler | 33.141924 |
| sparse-undefined | 3000 | compiler subprocess | 3 | go compiler | 54.627991 |
| sparse-undefined | 3000 | compiler subprocess | 4 | rust compiler | 93.013503 |
| sparse-undefined | 3000 | compiler subprocess | 5 | zig compiler | 431.561605 |
| undefined | 10 | in-process | 1 | Eloque FailFast parse+resolve | 0.004905 |
| undefined | 10 | in-process | 2 | Eloque Recovering parse+resolve | 0.004957 |
| undefined | 10 | in-process | 3 | CPython ast.parse | 0.027876 |
| undefined | 10 | in-process | 4 | CPython symtable | 0.028148 |
| undefined | 10 | compiler subprocess | 1 | c compiler | 20.545166 |
| undefined | 10 | compiler subprocess | 2 | cpp compiler | 23.261056 |
| undefined | 10 | compiler subprocess | 3 | go compiler | 44.060069 |
| undefined | 10 | compiler subprocess | 4 | rust compiler | 52.734644 |
| undefined | 10 | compiler subprocess | 5 | zig compiler | 463.650353 |
| undefined | 100 | in-process | 1 | Eloque FailFast parse+resolve | 0.037972 |
| undefined | 100 | in-process | 2 | Eloque Recovering parse+resolve | 0.039023 |
| undefined | 100 | in-process | 3 | CPython symtable | 0.157355 |
| undefined | 100 | in-process | 4 | CPython ast.parse | 0.206530 |
| undefined | 100 | compiler subprocess | 1 | c compiler | 23.260039 |
| undefined | 100 | compiler subprocess | 2 | cpp compiler | 23.992555 |
| undefined | 100 | compiler subprocess | 3 | go compiler | 53.882656 |
| undefined | 100 | compiler subprocess | 4 | rust compiler | 133.540675 |
| undefined | 100 | compiler subprocess | 5 | zig compiler | 439.258803 |
| undefined | 1000 | in-process | 1 | Eloque FailFast parse+resolve | 0.364151 |
| undefined | 1000 | in-process | 2 | Eloque Recovering parse+resolve | 0.377635 |
| undefined | 1000 | in-process | 3 | CPython symtable | 2.492515 |
| undefined | 1000 | in-process | 4 | CPython ast.parse | 2.561350 |
| undefined | 1000 | compiler subprocess | 1 | c compiler | 23.185043 |
| undefined | 1000 | compiler subprocess | 2 | cpp compiler | 28.668390 |
| undefined | 1000 | compiler subprocess | 3 | go compiler | 39.134166 |
| undefined | 1000 | compiler subprocess | 4 | zig compiler | 460.129872 |
| undefined | 1000 | compiler subprocess | 5 | rust compiler | 893.219551 |
| undefined | 3000 | in-process | 1 | Eloque FailFast parse+resolve | 1.171557 |
| undefined | 3000 | in-process | 2 | Eloque Recovering parse+resolve | 1.218360 |
| undefined | 3000 | in-process | 3 | CPython symtable | 6.548321 |
| undefined | 3000 | in-process | 4 | CPython ast.parse | 8.558334 |
| undefined | 3000 | compiler subprocess | 1 | c compiler | 25.875811 |
| undefined | 3000 | compiler subprocess | 2 | cpp compiler | 27.563708 |
| undefined | 3000 | compiler subprocess | 3 | go compiler | 46.364822 |
| undefined | 3000 | compiler subprocess | 4 | zig compiler | 459.057577 |
| undefined | 3000 | compiler subprocess | 5 | rust compiler | 2516.183493 |

## Native compiler subprocess measurements (separate boundary)

Includes process startup and compiler work beyond name resolution.

| Engine | Workload | N | Median (ms/invocation) | Boundary |
|---|---|---:|---:|---|
| c compiler | declarations | 10 | 19.956 | compiler subprocess: Clang parse + semantic validation (syntax-only) |
| c compiler | declarations | 100 | 19.790 | compiler subprocess: Clang parse + semantic validation (syntax-only) |
| c compiler | declarations | 1,000 | 26.956 | compiler subprocess: Clang parse + semantic validation (syntax-only) |
| c compiler | declarations | 3,000 | 41.601 | compiler subprocess: Clang parse + semantic validation (syntax-only) |
| c compiler | mixed | 10 | 21.287 | compiler subprocess: Clang parse + semantic validation (syntax-only) |
| c compiler | mixed | 100 | 22.527 | compiler subprocess: Clang parse + semantic validation (syntax-only) |
| c compiler | mixed | 1,000 | 34.362 | compiler subprocess: Clang parse + semantic validation (syntax-only) |
| c compiler | mixed | 3,000 | 52.346 | compiler subprocess: Clang parse + semantic validation (syntax-only) |
| c compiler | references | 10 | 25.085 | compiler subprocess: Clang parse + semantic validation (syntax-only) |
| c compiler | references | 100 | 20.398 | compiler subprocess: Clang parse + semantic validation (syntax-only) |
| c compiler | references | 1,000 | 24.903 | compiler subprocess: Clang parse + semantic validation (syntax-only) |
| c compiler | references | 3,000 | 30.201 | compiler subprocess: Clang parse + semantic validation (syntax-only) |
| c compiler | sparse-undefined | 10 | 22.756 | compiler subprocess: Clang parse + semantic validation (syntax-only) |
| c compiler | sparse-undefined | 100 | 21.869 | compiler subprocess: Clang parse + semantic validation (syntax-only) |
| c compiler | sparse-undefined | 1,000 | 24.923 | compiler subprocess: Clang parse + semantic validation (syntax-only) |
| c compiler | sparse-undefined | 3,000 | 32.265 | compiler subprocess: Clang parse + semantic validation (syntax-only) |
| c compiler | undefined | 10 | 20.545 | compiler subprocess: Clang parse + semantic validation (syntax-only) |
| c compiler | undefined | 100 | 23.260 | compiler subprocess: Clang parse + semantic validation (syntax-only) |
| c compiler | undefined | 1,000 | 23.185 | compiler subprocess: Clang parse + semantic validation (syntax-only) |
| c compiler | undefined | 3,000 | 25.876 | compiler subprocess: Clang parse + semantic validation (syntax-only) |
| cpp compiler | declarations | 10 | 21.233 | compiler subprocess: Clang++ parse + semantic validation (syntax-only) |
| cpp compiler | declarations | 100 | 24.788 | compiler subprocess: Clang++ parse + semantic validation (syntax-only) |
| cpp compiler | declarations | 1,000 | 31.578 | compiler subprocess: Clang++ parse + semantic validation (syntax-only) |
| cpp compiler | declarations | 3,000 | 50.763 | compiler subprocess: Clang++ parse + semantic validation (syntax-only) |
| cpp compiler | mixed | 10 | 22.021 | compiler subprocess: Clang++ parse + semantic validation (syntax-only) |
| cpp compiler | mixed | 100 | 23.390 | compiler subprocess: Clang++ parse + semantic validation (syntax-only) |
| cpp compiler | mixed | 1,000 | 36.188 | compiler subprocess: Clang++ parse + semantic validation (syntax-only) |
| cpp compiler | mixed | 3,000 | 58.664 | compiler subprocess: Clang++ parse + semantic validation (syntax-only) |
| cpp compiler | references | 10 | 22.729 | compiler subprocess: Clang++ parse + semantic validation (syntax-only) |
| cpp compiler | references | 100 | 20.356 | compiler subprocess: Clang++ parse + semantic validation (syntax-only) |
| cpp compiler | references | 1,000 | 29.104 | compiler subprocess: Clang++ parse + semantic validation (syntax-only) |
| cpp compiler | references | 3,000 | 33.117 | compiler subprocess: Clang++ parse + semantic validation (syntax-only) |
| cpp compiler | sparse-undefined | 10 | 22.375 | compiler subprocess: Clang++ parse + semantic validation (syntax-only) |
| cpp compiler | sparse-undefined | 100 | 23.360 | compiler subprocess: Clang++ parse + semantic validation (syntax-only) |
| cpp compiler | sparse-undefined | 1,000 | 27.693 | compiler subprocess: Clang++ parse + semantic validation (syntax-only) |
| cpp compiler | sparse-undefined | 3,000 | 33.142 | compiler subprocess: Clang++ parse + semantic validation (syntax-only) |
| cpp compiler | undefined | 10 | 23.261 | compiler subprocess: Clang++ parse + semantic validation (syntax-only) |
| cpp compiler | undefined | 100 | 23.993 | compiler subprocess: Clang++ parse + semantic validation (syntax-only) |
| cpp compiler | undefined | 1,000 | 28.668 | compiler subprocess: Clang++ parse + semantic validation (syntax-only) |
| cpp compiler | undefined | 3,000 | 27.564 | compiler subprocess: Clang++ parse + semantic validation (syntax-only) |
| go compiler | declarations | 10 | 55.134 | compiler subprocess: Go object compilation (not resolver-only) |
| go compiler | declarations | 100 | 71.311 | compiler subprocess: Go object compilation (not resolver-only) |
| go compiler | declarations | 1,000 | 97.218 | compiler subprocess: Go object compilation (not resolver-only) |
| go compiler | declarations | 3,000 | 143.328 | compiler subprocess: Go object compilation (not resolver-only) |
| go compiler | mixed | 10 | 49.790 | compiler subprocess: Go object compilation (not resolver-only) |
| go compiler | mixed | 100 | 43.797 | compiler subprocess: Go object compilation (not resolver-only) |
| go compiler | mixed | 1,000 | 103.095 | compiler subprocess: Go object compilation (not resolver-only) |
| go compiler | mixed | 3,000 | 144.475 | compiler subprocess: Go object compilation (not resolver-only) |
| go compiler | references | 10 | 74.329 | compiler subprocess: Go object compilation (not resolver-only) |
| go compiler | references | 100 | 61.855 | compiler subprocess: Go object compilation (not resolver-only) |
| go compiler | references | 1,000 | 63.183 | compiler subprocess: Go object compilation (not resolver-only) |
| go compiler | references | 3,000 | 69.584 | compiler subprocess: Go object compilation (not resolver-only) |
| go compiler | sparse-undefined | 10 | 66.574 | compiler subprocess: Go object compilation (not resolver-only) |
| go compiler | sparse-undefined | 100 | 44.644 | compiler subprocess: Go object compilation (not resolver-only) |
| go compiler | sparse-undefined | 1,000 | 49.912 | compiler subprocess: Go object compilation (not resolver-only) |
| go compiler | sparse-undefined | 3,000 | 54.628 | compiler subprocess: Go object compilation (not resolver-only) |
| go compiler | undefined | 10 | 44.060 | compiler subprocess: Go object compilation (not resolver-only) |
| go compiler | undefined | 100 | 53.883 | compiler subprocess: Go object compilation (not resolver-only) |
| go compiler | undefined | 1,000 | 39.134 | compiler subprocess: Go object compilation (not resolver-only) |
| go compiler | undefined | 3,000 | 46.365 | compiler subprocess: Go object compilation (not resolver-only) |
| rust compiler | declarations | 10 | 29.367 | compiler subprocess: rustc metadata compilation (not resolver-only) |
| rust compiler | declarations | 100 | 32.014 | compiler subprocess: rustc metadata compilation (not resolver-only) |
| rust compiler | declarations | 1,000 | 53.486 | compiler subprocess: rustc metadata compilation (not resolver-only) |
| rust compiler | declarations | 3,000 | 101.798 | compiler subprocess: rustc metadata compilation (not resolver-only) |
| rust compiler | mixed | 10 | 28.269 | compiler subprocess: rustc metadata compilation (not resolver-only) |
| rust compiler | mixed | 100 | 29.657 | compiler subprocess: rustc metadata compilation (not resolver-only) |
| rust compiler | mixed | 1,000 | 55.095 | compiler subprocess: rustc metadata compilation (not resolver-only) |
| rust compiler | mixed | 3,000 | 104.462 | compiler subprocess: rustc metadata compilation (not resolver-only) |
| rust compiler | references | 10 | 28.525 | compiler subprocess: rustc metadata compilation (not resolver-only) |
| rust compiler | references | 100 | 30.564 | compiler subprocess: rustc metadata compilation (not resolver-only) |
| rust compiler | references | 1,000 | 41.705 | compiler subprocess: rustc metadata compilation (not resolver-only) |
| rust compiler | references | 3,000 | 83.244 | compiler subprocess: rustc metadata compilation (not resolver-only) |
| rust compiler | sparse-undefined | 10 | 28.366 | compiler subprocess: rustc metadata compilation (not resolver-only) |
| rust compiler | sparse-undefined | 100 | 29.707 | compiler subprocess: rustc metadata compilation (not resolver-only) |
| rust compiler | sparse-undefined | 1,000 | 54.043 | compiler subprocess: rustc metadata compilation (not resolver-only) |
| rust compiler | sparse-undefined | 3,000 | 93.014 | compiler subprocess: rustc metadata compilation (not resolver-only) |
| rust compiler | undefined | 10 | 52.735 | compiler subprocess: rustc metadata compilation (not resolver-only) |
| rust compiler | undefined | 100 | 133.541 | compiler subprocess: rustc metadata compilation (not resolver-only) |
| rust compiler | undefined | 1,000 | 893.220 | compiler subprocess: rustc metadata compilation (not resolver-only) |
| rust compiler | undefined | 3,000 | 2516.183 | compiler subprocess: rustc metadata compilation (not resolver-only) |
| zig compiler | declarations | 10 | 449.524 | compiler subprocess: Zig AST check (not resolver-only) |
| zig compiler | declarations | 100 | 449.180 | compiler subprocess: Zig AST check (not resolver-only) |
| zig compiler | declarations | 1,000 | 453.911 | compiler subprocess: Zig AST check (not resolver-only) |
| zig compiler | declarations | 3,000 | 463.071 | compiler subprocess: Zig AST check (not resolver-only) |
| zig compiler | mixed | 10 | 454.841 | compiler subprocess: Zig AST check (not resolver-only) |
| zig compiler | mixed | 100 | 451.283 | compiler subprocess: Zig AST check (not resolver-only) |
| zig compiler | mixed | 1,000 | 446.696 | compiler subprocess: Zig AST check (not resolver-only) |
| zig compiler | mixed | 3,000 | 441.948 | compiler subprocess: Zig AST check (not resolver-only) |
| zig compiler | references | 10 | 441.712 | compiler subprocess: Zig AST check (not resolver-only) |
| zig compiler | references | 100 | 456.242 | compiler subprocess: Zig AST check (not resolver-only) |
| zig compiler | references | 1,000 | 441.390 | compiler subprocess: Zig AST check (not resolver-only) |
| zig compiler | references | 3,000 | 427.678 | compiler subprocess: Zig AST check (not resolver-only) |
| zig compiler | sparse-undefined | 10 | 442.543 | compiler subprocess: Zig AST check (not resolver-only) |
| zig compiler | sparse-undefined | 100 | 446.718 | compiler subprocess: Zig AST check (not resolver-only) |
| zig compiler | sparse-undefined | 1,000 | 451.659 | compiler subprocess: Zig AST check (not resolver-only) |
| zig compiler | sparse-undefined | 3,000 | 431.562 | compiler subprocess: Zig AST check (not resolver-only) |
| zig compiler | undefined | 10 | 463.650 | compiler subprocess: Zig AST check (not resolver-only) |
| zig compiler | undefined | 100 | 439.259 | compiler subprocess: Zig AST check (not resolver-only) |
| zig compiler | undefined | 1,000 | 460.130 | compiler subprocess: Zig AST check (not resolver-only) |
| zig compiler | undefined | 3,000 | 459.058 | compiler subprocess: Zig AST check (not resolver-only) |

## Interpretation

- Eloque: parse + name resolution in-process; Recovering is the primary path, FailFast is secondary.
- CPython `ast.parse`: parsing only; does not build a symbol table.
- CPython `symtable`: builds symbol tables but also reparses source internally; **not** isolated resolution.
- CPython permits unresolved names at compile time; undefined-name workloads therefore differ fundamentally from Eloque diagnostics.
- Workloads are structurally similar but not semantically equivalent. Source sizes and language-specific syntax differ.
- Native compiler subprocess measurements, when selected, include process startup and compiler phases beyond name resolution; they are displayed separately and cannot be ranked against Eloque resolver timings.
- Native workloads with unresolved names intentionally fail compilation and measure error-path invocation time; sparse-undefined cases at smaller sizes may compile successfully.
- Native compiler invocations are ranked separately because they include startup and extra compiler phases.

Raw measurements: `samples.csv`; Eloque raw output: `eloque-pipeline.log`.
