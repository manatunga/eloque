"""
M2 semantic benchmark collector; run from Eloque repository root.

Collects Eloque parse+resolve measurements and CPython in-process
AST parsing and symbol-table measurements.

These operations have different measurement boundaries and must not be ranked
as equivalent compiler phases.

Optional C, C++, Rust, Zig, and Go compiler subprocess measurements are
reported separately; these are not isolated name-resolution timings.
"""

from __future__ import annotations
import argparse
import ast
import csv
import json
import os
from pathlib import Path
import re
import statistics
import subprocess
import sys
import time
from datetime import datetime, timezone

from m2_native import LANGUAGES as NATIVE_LANGUAGES, collect as native_collect

ROOT = Path("benchmarks")
PIPELINE = re.compile(
    r"^BenchmarkPipeline(?P<mode>Recovering|FailFast)/(?P<workload>[^/\s]+)/N(?P<n>\d+)-\d+\s+\d+\s+(?P<ns>[\d.]+) ns/op(?:\s+[\d.]+ MB/s)?\s+(?P<bytes>\d+) B/op\s+(?P<allocs>\d+) allocs/op$",
    re.M,
)
SIZES = (10, 100, 1000, 3000)
CASES = ("declarations", "references", "mixed", "sparse-undefined", "undefined")


def run(cmd: list[str]) -> str:
    print("+", " ".join(map(str, cmd)), flush=True)
    proc = subprocess.run(
        cmd, text=True, stdout=subprocess.PIPE, stderr=subprocess.STDOUT
    )
    if proc.returncode:
        raise RuntimeError(
            f"Command failed ({proc.returncode}): {cmd}\n{proc.stdout[-5000:]}"
        )
    return proc.stdout


def python_source(n: int, kind: str) -> str:
    lines = ["def workload():"]
    if kind == "declarations":
        lines += [f"    value_{i} = {i}" for i in range(n)]
    elif kind == "references":
        lines += ["    seed = 7"] + ["    seed" for _ in range(n)]
    elif kind == "mixed":
        lines += [f"    value_{i} = {i}\n    value_{i}" for i in range(n)]
    elif kind == "undefined":
        lines += [f"    missing_{i}" for i in range(n)]
    elif kind == "sparse-undefined":
        lines += ["    seed = 7"] + [
            f"    missing_{i}" if i > 0 and i % 100 == 0 else "    seed" for i in range(n)
        ]
    else:
        raise ValueError(kind)
    return "\n".join(lines) + "\n"


def python_bench(repeats: int, duration: float, out: Path) -> list[dict]:
    import symtable

    results = []
    for kind in CASES:
        for n in SIZES:
            source = python_source(n, kind)
            # Parse+symtable on the same source; symtable independently reparses.
            # This is intentionally NOT described as a single compiler pipeline.
            ast.parse(source)
            symtable.symtable(source, "<benchmark>", "exec")
            for engine, action, boundary in (
                (
                    "CPython ast.parse",
                    lambda: ast.parse(source),
                    "in-process parse + AST",
                ),
                (
                    "CPython symtable",
                    lambda: symtable.symtable(source, "<benchmark>", "exec"),
                    "in-process symbol table (internally reparses source)",
                ),
            ):
                # Warm-up and auto-calibrate iterations to avoid clock resolution artifacts.
                for _ in range(3):
                    action()
                iterations = 1
                while iterations < 1_000_000:
                    start = time.perf_counter_ns()
                    for _ in range(iterations):
                        action()
                    elapsed = time.perf_counter_ns() - start
                    if elapsed >= int(duration * 1e9):
                        break
                    iterations *= 2
                for rep in range(repeats):
                    start = time.perf_counter_ns()
                    for _ in range(iterations):
                        action()
                    ns = (time.perf_counter_ns() - start) / iterations
                    results.append(
                        dict(
                            engine=engine,
                            workload=kind,
                            size=n,
                            repetition=rep + 1,
                            ns_per_op=ns,
                            bytes_per_op="",
                            allocs_per_op="",
                            boundary=boundary,
                            source_bytes=len(source.encode()),
                        )
                    )
                print(
                    f'{engine:21} {kind:16} N={n:<5} {statistics.median(float(x["ns_per_op"]) for x in results[-repeats:])/1e3:.2f} us'
                )
    return results


def eloque_bench(repeats: int, benchtime: str, out: Path) -> list[dict]:
    log = run([
        "go", "test", "./internal/semantic/resolver", "-run", "^$",
        "-bench", "^BenchmarkPipeline(Recovering|FailFast)$",
        "-benchmem", f"-benchtime={benchtime}", f"-count={repeats}",
    ])
    (out / "eloque-pipeline.log").write_text(log)
    rows = []
    for m in PIPELINE.finditer(log):
        kind, n, mode = m["workload"], int(m["n"]), m["mode"]
        if kind not in CASES or n not in SIZES:
            continue
        rows.append(dict(
            engine=f"Eloque {mode} parse+resolve", workload=kind, size=n,
            repetition=0, ns_per_op=float(m["ns"]),
            bytes_per_op=int(m["bytes"]), allocs_per_op=int(m["allocs"]),
            boundary=f"in-process {mode.lower()} parse + name resolution",
            source_bytes="",
        ))
    for mode in ("Recovering", "FailFast"):
        for kind in CASES:
            for n in SIZES:
                subset = [r for r in rows if r["engine"] == f"Eloque {mode} parse+resolve"
                          and r["workload"] == kind and r["size"] == n]
                if len(subset) != repeats:
                    raise RuntimeError(f"Missing Eloque samples: {mode}/{kind}/N{n}: "
                                       f"{len(subset)}/{repeats}. See {out}/eloque-pipeline.log")
                for i, row in enumerate(subset, 1):
                    row["repetition"] = i
    return rows


def main() -> None:
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument("--repeats", type=int, default=10)
    p.add_argument("--benchtime", default="1s")
    p.add_argument(
        "--python-duration",
        type=float,
        default=0.1,
        help="minimum calibration duration per CPython case, seconds",
    )
    p.add_argument(
        "--out",
        type=Path,
        default=Path("benchmarks/results/cross-language/m2-semantic-v2"),
    )
    p.add_argument("--skip-eloque", action="store_true")
    p.add_argument("--skip-cpython", action="store_true")
    p.add_argument("--native-languages", nargs="+", choices=NATIVE_LANGUAGES,
                   default=list(NATIVE_LANGUAGES), help="native compiler subprocess comparisons")
    p.add_argument("--native-warmups", type=int, default=1)
    p.add_argument("--skip-native", action="store_true")
    args = p.parse_args()
    if args.native_warmups < 0:
        p.error("native-warmups cannot be negative")
    if args.repeats < 1 or args.python_duration <= 0:
        p.error("repeats and python-duration must be positive")
    if not Path("go.mod").is_file():
        p.error("run from the Eloque repository root")
    args.out.mkdir(parents=True, exist_ok=True)
    rows = []
    if not args.skip_eloque:
        rows += eloque_bench(args.repeats, args.benchtime, args.out)
    if not args.skip_cpython:
        rows += python_bench(args.repeats, args.python_duration, args.out)
    native_inputs = []
    if args.native_languages and not args.skip_native:
        native_rows, native_inputs = native_collect(
            args.native_languages, CASES, SIZES, args.repeats,
            args.native_warmups, args.out,
        )
        rows += native_rows
    if not rows:
        p.error("all engines skipped")
    columns = [
        "engine",
        "workload",
        "size",
        "repetition",
        "ns_per_op",
        "bytes_per_op",
        "allocs_per_op",
        "boundary",
        "source_bytes",
    ]
    with (args.out / "samples.csv").open("w", newline="") as f:
        writer = csv.DictWriter(f, fieldnames=columns)
        writer.writeheader()
        writer.writerows(rows)
    groups = {}
    for r in rows:
        groups.setdefault((r["engine"], r["workload"], r["size"]), []).append(
            float(r["ns_per_op"])
        )
    report = [
        "# Eloque M2 cross-language semantic comparison v2",
        "",
        "**Seven languages, different measurement boundaries: these are ranked observed timings, not equivalent compiler-phase speed claims.**",
        "",
        "## In-process measurements (Eloque and CPython)",
        "",
        "| Engine | Workload | N | Median (µs/op) | Boundary |",
        "|---|---|---:|---:|---|",
    ]
    for (engine, kind, n), values in sorted(groups.items()):
        if engine.endswith(" compiler"):
            continue
        boundary = next(
            r["boundary"]
            for r in rows
            if (r["engine"], r["workload"], r["size"]) == (engine, kind, n)
        )
        report.append(
            f"| {engine} | {kind} | {n:,} | {statistics.median(values)/1e3:.3f} | {boundary} |"
        )
    # Rank observed timings within each measurement family, not as equal work.
    report.extend(["", "## Observed timing ranks (by workload and size)", "",
                   "Ranks below are descriptive only: operations and process boundaries differ.", "",
                   "| Workload | N | Family | Rank | Engine | Median (ms/op) |",
                   "|---|---:|---|---:|---|---:|"])
    for kind in CASES:
        for n in SIZES:
            for family in ("in-process", "compiler subprocess"):
                entries = []
                for (engine, k, size), values in groups.items():
                    if k != kind or size != n:
                        continue
                    is_native = engine.endswith(" compiler")
                    if (family == "compiler subprocess") != is_native:
                        continue
                    entries.append((statistics.median(values), engine))
                for rank, (median, engine) in enumerate(sorted(entries), 1):
                    report.append(f"| {kind} | {n} | {family} | {rank} | {engine} | {median/1e6:.6f} |")
    if native_inputs:
        report.extend(["", "## Native compiler subprocess measurements (separate boundary)",
                       "", "Includes process startup and compiler work beyond name resolution.",
                       "", "| Engine | Workload | N | Median (ms/invocation) | Boundary |",
                       "|---|---|---:|---:|---|"])
        for (engine, kind, n), values in sorted(groups.items()):
            if not engine.endswith(" compiler"):
                continue
            boundary = next(r["boundary"] for r in rows
                            if (r["engine"], r["workload"], r["size"]) == (engine, kind, n))
            report.append(f"| {engine} | {kind} | {n:,} | {statistics.median(values)/1e6:.3f} | {boundary} |")
    report += [
        "",
        "## Interpretation",
        "",
        "- Eloque: parse + name resolution in-process; Recovering is the primary path, FailFast is secondary.",
        "- CPython `ast.parse`: parsing only; does not build a symbol table.",
        "- CPython `symtable`: builds symbol tables but also reparses source internally; **not** isolated resolution.",
        "- CPython permits unresolved names at compile time; undefined-name workloads therefore differ fundamentally from Eloque diagnostics.",
        "- Workloads are structurally similar but not semantically equivalent. Source sizes and language-specific syntax differ.",
        "- Native compiler subprocess measurements, when selected, include process startup and compiler phases beyond name resolution; they are displayed separately and cannot be ranked against Eloque resolver timings.",
        "- Undefined-name cases intentionally fail native compilation; they measure error-path compiler invocation time, not successful resolution.",
        "- Native compiler invocations are ranked separately because they include startup and extra compiler phases.",
        "",
        "Raw measurements: `samples.csv`; Eloque raw output: `eloque-pipeline.log`.",
        "",
    ]
    (args.out / "REPORT.md").write_text("\n".join(report))
    (args.out / "manifest.json").write_text(
        json.dumps(
            {
                "created_utc": datetime.now(timezone.utc).isoformat(),
                "python": sys.version,
                "repeats": args.repeats,
                "benchtime": args.benchtime,
                "python_duration": args.python_duration,
                "scope": "M2 Eloque parse+resolve, CPython AST/symbol table, native compiler subprocesses; separate boundaries",
                "native_languages": args.native_languages if not args.skip_native else [],
                "native_warmups": args.native_warmups,
                "native_inputs": native_inputs,
            },
            indent=2,
        )
        + "\n"
    )
    print(f"Wrote {args.out}/REPORT.md and samples.csv")


if __name__ == "__main__":
    main()
