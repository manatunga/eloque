#!/usr/bin/env python3
"""Run existing Eloque M2 resolver benchmarks and save reproducible results."""

import argparse
import csv
import json
import platform
import re
import statistics
import subprocess
from datetime import datetime, timezone
from pathlib import Path

PATTERN = re.compile(
    r"^(BenchmarkResolve\S*)-\d+\s+(\d+)\s+([\d.]+) ns/op(?:\s+([\d.]+) B/op)?(?:\s+([\d.]+) allocs/op)?",
    re.M,
)


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument("--repeats", type=int, default=10)
    ap.add_argument("--benchtime", default="1s")
    ap.add_argument(
        "--out", type=Path, default=Path("benchmarks/results/m2/resolver-baseline")
    )
    args = ap.parse_args()
    if args.repeats < 1:
        ap.error("--repeats must be >= 1")
    args.out.mkdir(parents=True, exist_ok=True)
    cmd = [
        "go",
        "test",
        "./internal/semantic/resolver",
        "-run",
        "^$",
        "-bench",
        "^BenchmarkResolve",
        "-benchmem",
        f"-benchtime={args.benchtime}",
        f"-count={args.repeats}",
    ]
    print("+", " ".join(cmd), flush=True)
    result = subprocess.run(
        cmd, text=True, stdout=subprocess.PIPE, stderr=subprocess.STDOUT
    )
    (args.out / "go_benchmark.log").write_text(result.stdout, encoding="utf-8")
    if result.returncode:
        raise SystemExit(
            f'Benchmark failed (exit {result.returncode}); see {args.out / "go_benchmark.log"}\n{result.stdout[-4000:]}'
        )
    rows = []
    for name, iterations, ns, b, allocs in PATTERN.findall(result.stdout):
        rows.append(
            {
                "benchmark": name,
                "iterations": int(iterations),
                "ns_per_op": float(ns),
                "bytes_per_op": float(b) if b else "",
                "allocs_per_op": float(allocs) if allocs else "",
            }
        )
    if not rows:
        raise SystemExit(
            "No BenchmarkResolve results found. See go_benchmark.log; check benchmark names."
        )
    with (args.out / "samples.csv").open("w", newline="") as f:
        w = csv.DictWriter(f, fieldnames=list(rows[0]))
        w.writeheader()
        w.writerows(rows)
    by_name = {}
    for row in rows:
        by_name.setdefault(row["benchmark"], []).append(row)
    lines = [
        "# Eloque M2 resolver baseline",
        "",
        f'Go benchmark: `{ " ".join(cmd) }`',
        "",
        "| Benchmark | Samples | Median ns/op | Median B/op | Median allocs/op |",
        "|---|---:|---:|---:|---:|",
    ]
    for name, group in sorted(by_name.items()):

        def med(k):
            vals = [float(r[k]) for r in group if r[k] != ""]
            return f"{statistics.median(vals):,.2f}" if vals else "—"

        lines.append(
            f'| `{name}` | {len(group)} | {med("ns_per_op")} | {med("bytes_per_op")} | {med("allocs_per_op")} |'
        )
    lines += [
        "",
        "These measurements isolate the M2 resolver; they do not include parsing or full semantic analysis.",
        "",
        "Raw results: `samples.csv` and `go_benchmark.log`.",
    ]
    (args.out / "REPORT.md").write_text("\n".join(lines) + "\n", encoding="utf-8")
    ver = subprocess.run(["go", "version"], capture_output=True, text=True)
    (args.out / "manifest.json").write_text(
        json.dumps(
            {
                "created_utc": datetime.now(timezone.utc).isoformat(),
                "platform": platform.platform(),
                "go_version": ver.stdout.strip(),
                "command": cmd,
                "samples": len(rows),
            },
            indent=2,
        )
        + "\n"
    )
    print(f'Saved {len(rows)} samples to {args.out}; report: {args.out / "REPORT.md"}')


if __name__ == "__main__":
    main()
