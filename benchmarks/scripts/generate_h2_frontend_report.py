#!/usr/bin/env python3
"""Generate a reproducible H2 frontend benchmark report from raw log files.

Run at the Eloque repository root. Standard-library Python only.
"""

from __future__ import annotations
import argparse
import csv
import json
import re
import statistics
from datetime import date
from pathlib import Path

GO_RE = re.compile(
    r"^Benchmark(?P<name>ControlledEloqueFrontend|ControlledEloqueRecoveringFrontend|ControlledGoFrontend|ControlledGoCompilerFrontend)/N(?P<n>10|100|1000|10000)-\d+\s+\d+\s+(?P<ns>[\d.]+) ns/op\s+(?P<mb>[\d.]+) MB/s\s+(?P<bytes>\d+) B/op\s+(?P<allocs>\d+) allocs/op",
    re.M,
)
PY_RE = re.compile(
    r"^N=(?P<n>\d+)\s+bytes=(?P<bytes>\d+)\s+frontend=\s*(?P<ns>[\d,]+) ns/op", re.M
)
LABELS = {
    "ControlledEloqueFrontend": "Eloque fail-fast",
    "ControlledEloqueRecoveringFrontend": "Eloque recovering",
    "ControlledGoCompilerFrontend": "Go compiler-native",
    "ControlledGoFrontend": "Go go/parser",
}
SIZES = (10, 100, 1000, 10000)
ORDER = (*LABELS.values(), "CPython 3.15.0rc2")


def median(xs):
    return statistics.median(xs)


def parse_go(path, expected):
    text = path.read_text()
    if "PASS" not in text or re.search(r"^FAIL", text, re.M):
        raise ValueError(f"{path}: benchmark did not finish successfully")
    out = []
    for m in GO_RE.finditer(text):
        d = m.groupdict()
        if d["name"] not in expected:
            raise ValueError(f'{path}: unexpected benchmark {d["name"]}')
        out.append(
            dict(
                frontend=LABELS[d["name"]],
                n=int(d["n"]),
                ns=float(d["ns"]),
                mb_s=float(d["mb"]),
                bytes_op=int(d["bytes"]),
                allocs_op=int(d["allocs"]),
            )
        )
    for name in expected:
        for n in SIZES:
            count = sum(r["frontend"] == LABELS[name] and r["n"] == n for r in out)
            if count != 10:
                raise ValueError(
                    f"{path}: {LABELS[name]} N{n}: expected 10 samples, got {count}"
                )
    return out


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument("--eloque", type=Path, required=True)
    ap.add_argument("--go-parser", type=Path, required=True)
    ap.add_argument("--go-compiler", type=Path, required=True)
    ap.add_argument("--cpython", type=Path, required=True)
    ap.add_argument(
        "--out",
        type=Path,
        default=Path("benchmarks/results/experimental/h2-post-recovery-report"),
    )
    args = ap.parse_args()
    raw = (
        parse_go(
            args.eloque,
            {"ControlledEloqueFrontend", "ControlledEloqueRecoveringFrontend"},
        )
        + parse_go(args.go_parser, {"ControlledGoFrontend"})
        + parse_go(args.go_compiler, {"ControlledGoCompilerFrontend"})
    )
    pytext = args.cpython.read_text()
    version = re.search(r"^Python version:\s*(.+)$", pytext, re.M)
    if not version or version.group(1).strip() != "3.15.0rc2":
        raise ValueError(
            "CPython version is missing or differs from the recorded 3.15.0rc2"
        )
    py = {
        int(m["n"]): {"ns": int(m["ns"].replace(",", "")), "bytes": int(m["bytes"])}
        for m in PY_RE.finditer(pytext)
    }
    if set(py) != set(SIZES):
        raise ValueError("CPython must report all four workload sizes")
    args.out.mkdir(parents=True, exist_ok=True)
    with (args.out / "raw_go_samples.csv").open("w", newline="") as f:
        writer = csv.DictWriter(
            f, fieldnames=["frontend", "n", "ns", "mb_s", "bytes_op", "allocs_op"]
        )
        writer.writeheader()
        writer.writerows(raw)
    summary = []
    for name in ORDER:
        for n in SIZES:
            if name.startswith("CPython"):
                row = dict(
                    frontend=name,
                    n=n,
                    median_ns=py[n]["ns"],
                    median_mb_s=None,
                    median_bytes_op=None,
                    median_allocs_op=None,
                    min_ns=None,
                    max_ns=None,
                    samples=0,
                    source_bytes=py[n]["bytes"],
                )
            else:
                samples = [r for r in raw if r["frontend"] == name and r["n"] == n]
                row = dict(
                    frontend=name,
                    n=n,
                    median_ns=median([r["ns"] for r in samples]),
                    median_mb_s=median([r["mb_s"] for r in samples]),
                    median_bytes_op=median([r["bytes_op"] for r in samples]),
                    median_allocs_op=median([r["allocs_op"] for r in samples]),
                    min_ns=min(r["ns"] for r in samples),
                    max_ns=max(r["ns"] for r in samples),
                    samples=len(samples),
                    source_bytes=None,
                )
            row["statements_per_s"] = n * 1e9 / row["median_ns"]
            summary.append(row)
    with (args.out / "summary.csv").open("w", newline="") as f:
        writer = csv.DictWriter(f, fieldnames=list(summary[0]))
        writer.writeheader()
        writer.writerows(summary)
    lookup = {(r["frontend"], r["n"]): r for r in summary}
    lines = [
        "# H2 post-recovery controlled frontend comparison",
        "",
        f"Report generated: {date.today().isoformat()}. Results are from separately executed benchmark runs; see source logs for run dates.",
        "",
        "## Methodology",
        "",
        "- Host: Linux amd64; Intel Core Ultra 7 155H (from Go benchmark output).",
        "- Go version: 1.27.2 (reported by operator). CPython: 3.15.0rc2, Clang 22.1.3.",
        "- Go benchmark measurements: 10 repetitions per frontend per size, medians reported.",
        "- CPython: only four reported medians available; raw repeats and variability unavailable.",
        "- Workloads: 10, 100, 1,000, and 10,000 arithmetic declarations. Languages have different syntax, wrappers, grammars, ASTs, and validation work.",
        "- Go compiler-native benchmark: cmd/compile/internal/syntax via test overlay; Go standard library: go/parser; CPython: ast.parse.",
        "- MB/s reflects bytes of each language-specific source input, not normalized work.",
        "- B/op measures total allocated bytes, **not** peak memory; allocs/op counts allocations, not objects retained.",
        "",
        "## Median frontend latency",
        "",
        "| Frontend | N=10 (µs) | N=100 (µs) | N=1,000 (ms) | N=10,000 (ms) |",
        "|---|---:|---:|---:|---:|",
    ]
    for name in ORDER:
        v = [lookup[(name, n)]["median_ns"] for n in SIZES]
        lines.append(
            f"| {name} | {v[0]/1e3:.2f} | {v[1]/1e3:.2f} | {v[2]/1e6:.3f} | {v[3]/1e6:.3f} |"
        )
    lines += [
        "",
        "## N=10,000 throughput and allocation",
        "",
        "| Frontend | MB/s | Statements/s | Allocated MiB/op | Allocs/op |",
        "|---|---:|---:|---:|---:|",
    ]
    for name in ORDER:
        r = lookup[(name, 10000)]
        fmt = lambda v, format_: "Not reported" if v is None else format_.format(v)
        lines.append(
            f'| {name} | {fmt(r["median_mb_s"], "{:.2f}")} | {r["statements_per_s"]:,.0f} | {fmt(None if r["median_bytes_op"] is None else r["median_bytes_op"]/1048576, "{:.2f}")} | {fmt(r["median_allocs_op"], "{:,.1f}")} |'
        )
    lines += [
        "",
        "## Eloque recovering overhead",
        "",
        "| N | Recovery latency overhead | Allocation count difference |",
        "|---|---:|---:|",
    ]
    for n in SIZES:
        a, b = lookup[("Eloque fail-fast", n)], lookup[("Eloque recovering", n)]
        lines.append(
            f'| {n:,} | {(b["median_ns"]/a["median_ns"]-1)*100:+.2f}% | {b["median_allocs_op"]-a["median_allocs_op"]:+.1f} |'
        )
    lines += ["", "## Relative Eloque fail-fast speedup at N=10,000", ""]
    base = lookup[("Eloque fail-fast", 10000)]["median_ns"]
    for name in ORDER[2:]:
        lines.append(
            f'- Against {name}: **{lookup[(name, 10000)]["median_ns"]/base:.2f}×**.'
        )
    lines += [
        "",
        "## Reproducibility and caveats",
        "",
        "The source logs, overlay harness, generator script, and exact invocation commands should be retained with this report. Benchmark source generation and parser input validation should be checked against the implementation to confirm identical measured operations. The compiler-native overlay benchmark measures parsing and AST construction, not the entire Go compiler pipeline. The Eloque benchmarks measure frontend parsing, not static type checking, bytecode generation, or runtime execution.",
        "",
        "Cross-language comparisons are illustrative, **not apples-to-apples compiler-performance rankings**. Comparisons of MB/s, allocations, and latency are especially sensitive to source length and AST structure. Separate benchmark runs may be affected by thermal and scheduling drift; no statistical significance claim is made. Python sample-level variance cannot be computed from its current summary output.",
        "",
        "Historical October 20 H2 official results must remain unchanged. These post-recovery measurements form a new experimental baseline.",
        "",
    ]
    (args.out / "REPORT.md").write_text("\n".join(lines))
    (args.out / "manifest.json").write_text(
        json.dumps(
            {
                "source_files": {
                    k: str(v)
                    for k, v in [
                        ("eloque", args.eloque),
                        ("go_parser", args.go_parser),
                        ("go_compiler", args.go_compiler),
                        ("cpython", args.cpython),
                    ]
                },
                "go_samples": len(raw),
                "cpython_summary_rows": len(py),
                "go_version": "go1.27.2",
                "python_version": version.group(1).strip(),
            },
            indent=2,
        )
        + "\n"
    )
    print(
        f"Generated {args.out}/REPORT.md, summary.csv, raw_go_samples.csv, manifest.json"
    )


if __name__ == "__main__":
    main()
