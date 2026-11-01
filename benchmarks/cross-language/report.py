#!/usr/bin/env python3
"""Generate a phase-labelled summary of cross-language compiler measurements."""

from __future__ import annotations

import argparse
import csv
import json
from pathlib import Path
from statistics import median


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument(
        "--input", type=Path, default=Path("benchmarks/results/cross-language/compiler-subprocess")
    )
    args = parser.parse_args()
    with (args.input / "samples.csv").open(newline="") as file:
        rows = list(csv.DictReader(file))
    manifest = json.loads((args.input / "manifest.json").read_text())
    groups: dict[tuple[str, str, int], list[dict[str, str]]] = {}
    for row in rows:
        groups.setdefault(
            (row["language"], row["workload"], int(row["size"])), []
        ).append(row)
    lines = [
        "# Cross-language compiler subprocess benchmark",
        "",
        "**Measurement:** subprocess wall-clock latency, including compiler startup. Lower is better within each compiler and workload.",
        "",
        "**Important:** compiler phase boundaries differ; this is not a parser-only ranking or an Eloque speed comparison.",
        "",
        "| Language | Workload | Units | Median (ms) | Min (ms) | Max (ms) | Samples | Phase boundary |",
        "|---|---|---:|---:|---:|---:|---:|---|",
    ]
    summary = []
    for (language, workload, size), samples in sorted(groups.items()):
        times = [int(r["ns"]) / 1e6 for r in samples]
        boundary = samples[0]["boundary"]
        lines.append(
            f"| {language} | {workload} | {size:,} | {median(times):.3f} | {min(times):.3f} | {max(times):.3f} | {len(times)} | {boundary} |"
        )
        summary.append(
            {
                "language": language,
                "workload": workload,
                "size": size,
                "median_ms": median(times),
                "min_ms": min(times),
                "max_ms": max(times),
                "samples": len(times),
                "boundary": boundary,
            }
        )
    lines += [
        "",
        "## Environment",
        "",
        f"- Host: `{manifest['host']}`",
        f"- Warmups: {manifest['configuration']['warmups']}; repeats: {manifest['configuration']['repeats']}; shuffle seed: {manifest['configuration']['seed']}",
        "",
        "## Compiler versions",
        "",
    ]
    for language, version in manifest["versions"].items():
        lines.append(f"- {language}: `{version}`")
    lines += [
        "",
        "## Limitations",
        "",
        "The source programs express similar workloads but are not grammar-equivalent. Clang performs syntax and semantic checks without code generation, Rust emits metadata, and Go/Zig emit objects. Startup and filesystem overhead are included. Do not divide these values by Eloque's in-process parser or resolver timings to claim a compiler speedup.",
        "",
    ]
    (args.input / "REPORT.md").write_text("\n".join(lines))
    with (args.input / "summary.csv").open("w", newline="") as file:
        writer = csv.DictWriter(file, fieldnames=list(summary[0]))
        writer.writeheader()
        writer.writerows(summary)
    print(f"Wrote {args.input / 'REPORT.md'} and summary.csv")


if __name__ == "__main__":
    main()
