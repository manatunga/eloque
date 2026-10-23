# Monowire comparative VM benchmark — v0.1 (draft reference)

## Purpose
Compare implementations of the *same* miniature register-window VM in Zig, Rust, Go, and, if feasible, Mojo. This is **not** an Eloque language specification or an official performance result.

## Instruction semantics
All registers are 64-bit bit patterns; arithmetic uses wrapping two's-complement i64. Each function has `regs`, `args`, and an instruction array. Operands are local register indices unless noted.

- `LOAD dst, imm`: write immediate
- `MOVE dst, src`: copy
- `ADD dst, a, b`: wrapping addition
- `SUB dst, a, b`: wrapping subtraction
- `JUMP target`: absolute instruction index within function
- `JUMP_ZERO src, target`: jump if value is zero
- `CALL fn_id, start, dst`: callee frame base = caller base + `start`; callee parameters r0..r(args-1) alias caller's outgoing registers; callee's other registers may extend beyond caller's frame. Return value written to caller's `dst` after return. Caller must not keep live values in clobbered overlapping region.
- `RETURN src`: restore caller and write result to its destination
- `HALT src`: finish with result

`CALL` is permitted only if `start + callee.args <= caller.regs`, and all active windows fit in the register arena. The benchmark reference uses a fixed-capacity arena for workloads with known maximum depth. Production Monowire will need bounded growth and verifier-enforced call-window liveness.

## Workload 1: call_loop
Compute sum(1..N) by repeatedly calling `add(accumulator, counter)` using an overlapping outgoing window. Expected result N(N+1)/2. `N=1000`, `reps=20` are default smoke-test parameters, not meaningful benchmarking parameters. `N=10000` recommended for timing. Check result before timing. Exclude bytecode construction and VM setup.

## Benchmark fairness requirements
Use identical bytecode streams, register counts, overflow semantics, work done, and correctness checks. Report build modes, CPU, OS, compiler versions, warmup, repetitions, raw per-run timings, medians, dispersion, and allocations. Keep dispatch baseline and optimized dispatch as separate variants. Test multi-level calls, recursion, branches, and heap integration separately; do not claim those are covered by `call_loop`.

## Known limitation of current Go reference
This initial Go implementation is a *correctness and semantic reference* for `call_loop`, not a tuned final Go contender. It allocates a frame stack using `append` and bounds-checks accesses. Go version and compiler optimizations can affect performance. No Zig, Rust, or Mojo timing comparisons have been performed yet.
