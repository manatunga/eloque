# Cross-language compiler subprocess benchmark

**Measurement:** subprocess wall-clock latency, including compiler startup. Lower is better within each compiler and workload.

**Important:** compiler phase boundaries differ; this is not a parser-only ranking or an Eloque speed comparison.

| Language | Workload | Units | Median (ms) | Min (ms) | Max (ms) | Samples | Phase boundary |
|---|---|---:|---:|---:|---:|---:|---|
| c | arithmetic | 10 | 16.451 | 15.586 | 16.801 | 5 | clang syntax-only (parse + semantic validation) |
| c | arithmetic | 100 | 18.380 | 14.801 | 19.373 | 5 | clang syntax-only (parse + semantic validation) |
| c | arithmetic | 1,000 | 25.272 | 22.569 | 29.914 | 5 | clang syntax-only (parse + semantic validation) |
| c | arithmetic | 10,000 | 101.898 | 87.716 | 121.483 | 5 | clang syntax-only (parse + semantic validation) |
| c | declarations | 10 | 16.626 | 16.185 | 22.534 | 5 | clang syntax-only (parse + semantic validation) |
| c | declarations | 100 | 17.111 | 15.749 | 17.780 | 5 | clang syntax-only (parse + semantic validation) |
| c | declarations | 1,000 | 20.541 | 19.370 | 22.939 | 5 | clang syntax-only (parse + semantic validation) |
| c | declarations | 10,000 | 51.429 | 46.343 | 54.172 | 5 | clang syntax-only (parse + semantic validation) |
| cpp | arithmetic | 10 | 17.356 | 16.353 | 22.093 | 5 | clang++ syntax-only (parse + semantic validation) |
| cpp | arithmetic | 100 | 17.277 | 15.262 | 28.909 | 5 | clang++ syntax-only (parse + semantic validation) |
| cpp | arithmetic | 1,000 | 26.555 | 26.476 | 28.200 | 5 | clang++ syntax-only (parse + semantic validation) |
| cpp | arithmetic | 10,000 | 121.389 | 98.188 | 134.032 | 5 | clang++ syntax-only (parse + semantic validation) |
| cpp | declarations | 10 | 16.295 | 14.888 | 20.002 | 5 | clang++ syntax-only (parse + semantic validation) |
| cpp | declarations | 100 | 16.950 | 15.846 | 20.350 | 5 | clang++ syntax-only (parse + semantic validation) |
| cpp | declarations | 1,000 | 21.147 | 20.717 | 26.620 | 5 | clang++ syntax-only (parse + semantic validation) |
| cpp | declarations | 10,000 | 63.464 | 61.406 | 79.867 | 5 | clang++ syntax-only (parse + semantic validation) |
| go | arithmetic | 10 | 29.692 | 26.316 | 32.832 | 5 | go tool compile object emission (includes code generation) |
| go | arithmetic | 100 | 32.796 | 25.354 | 35.661 | 5 | go tool compile object emission (includes code generation) |
| go | arithmetic | 1,000 | 48.183 | 42.944 | 57.523 | 5 | go tool compile object emission (includes code generation) |
| go | arithmetic | 10,000 | 180.802 | 176.748 | 219.664 | 5 | go tool compile object emission (includes code generation) |
| go | declarations | 10 | 30.912 | 24.447 | 40.179 | 5 | go tool compile object emission (includes code generation) |
| go | declarations | 100 | 29.825 | 24.904 | 48.864 | 5 | go tool compile object emission (includes code generation) |
| go | declarations | 1,000 | 38.365 | 35.202 | 46.807 | 5 | go tool compile object emission (includes code generation) |
| go | declarations | 10,000 | 113.487 | 109.508 | 121.546 | 5 | go tool compile object emission (includes code generation) |
| rust | arithmetic | 10 | 20.483 | 19.112 | 23.553 | 5 | rustc metadata emission (includes semantic analysis) |
| rust | arithmetic | 100 | 33.563 | 31.038 | 34.651 | 5 | rustc metadata emission (includes semantic analysis) |
| rust | arithmetic | 1,000 | 167.639 | 162.665 | 174.772 | 5 | rustc metadata emission (includes semantic analysis) |
| rust | arithmetic | 10,000 | 5923.842 | 4872.745 | 6969.771 | 5 | rustc metadata emission (includes semantic analysis) |
| rust | declarations | 10 | 19.197 | 18.113 | 20.916 | 5 | rustc metadata emission (includes semantic analysis) |
| rust | declarations | 100 | 22.635 | 22.203 | 24.194 | 5 | rustc metadata emission (includes semantic analysis) |
| rust | declarations | 1,000 | 66.418 | 60.099 | 68.252 | 5 | rustc metadata emission (includes semantic analysis) |
| rust | declarations | 10,000 | 1347.771 | 1329.438 | 1415.412 | 5 | rustc metadata emission (includes semantic analysis) |
| zig | arithmetic | 10 | 338.714 | 312.360 | 353.638 | 5 | zig object emission (includes code generation) |
| zig | arithmetic | 100 | 317.637 | 309.751 | 379.649 | 5 | zig object emission (includes code generation) |
| zig | arithmetic | 1,000 | 318.024 | 310.292 | 439.668 | 5 | zig object emission (includes code generation) |
| zig | arithmetic | 10,000 | 379.830 | 328.306 | 408.590 | 5 | zig object emission (includes code generation) |
| zig | declarations | 10 | 337.226 | 320.393 | 407.737 | 5 | zig object emission (includes code generation) |
| zig | declarations | 100 | 332.119 | 299.819 | 368.810 | 5 | zig object emission (includes code generation) |
| zig | declarations | 1,000 | 317.096 | 284.269 | 331.171 | 5 | zig object emission (includes code generation) |
| zig | declarations | 10,000 | 342.227 | 300.488 | 403.647 | 5 | zig object emission (includes code generation) |

## Environment

- Host: `Linux-6.18.33.2-microsoft-standard-WSL2-x86_64-with-glibc2.43`
- Warmups: 2; repeats: 5; shuffle seed: 42

## Compiler versions

- c: `Ubuntu clang version 21.1.8 (6ubuntu1)`
- cpp: `Ubuntu clang version 21.1.8 (6ubuntu1)`
- rust: `rustc 1.99.1 (58cf8f9c3 2026-10-13)`
- zig: `0.17.0`
- go: `go version go1.27.2 linux/amd64`

## Limitations

The source programs express similar workloads but are not grammar-equivalent. Clang performs syntax and semantic checks without code generation, Rust emits metadata, and Go/Zig emit objects. Startup and filesystem overhead are included. Do not divide these values by Eloque's in-process parser or resolver timings to claim a compiler speedup.
