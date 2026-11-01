#!/usr/bin/env python3
"""Collect Eloque, CPython and Go compiler-parser in-process frontend benchmarks.

Run at the repository root. Each engine uses its own source grammar; the
controlled arithmetic expression and number of declarations are matched.
"""

from __future__ import annotations

import argparse
import csv
import json
import os
from pathlib import Path
import re
import shutil
import statistics
import subprocess
import sys
import tempfile
from datetime import datetime, timezone

SIZES = (10, 100, 1000, 10000)
GO_BENCH = re.compile(
    r"^Benchmark(?P<name>ControlledEloqueFrontend|ControlledEloqueRecoveringFrontend|ControlledGoCompilerFrontend)/N(?P<n>\d+)-\d+\s+\d+\s+(?P<ns>[\d.]+) ns/op",
    re.M,
)

OVERLAY_TEST = r"""package syntax
import (
    "bytes"
    "fmt"
    "strings"
    "testing"
)
var eloqueBenchSink *File
func eloqueControlledSource(count int) []byte {
    var b strings.Builder
    b.WriteString("package benchmark\nfunc workload() {\n")
    for i := 0; i < count; i++ {
        fmt.Fprintf(&b, "value_%d := ((%d + 3) * (7 - 2)) + (%d * 4)\n", i, i, i)
    }
    b.WriteString("}\n")
    return []byte(b.String())
}
func BenchmarkControlledGoCompilerFrontend(b *testing.B) {
    for _, n := range []int{10, 100, 1000, 10000} {
        b.Run(fmt.Sprintf("N%d", n), func(b *testing.B) {
            source := eloqueControlledSource(n)
            var err error
            eloqueBenchSink, err = Parse(nil, bytes.NewReader(source), nil, nil, 0)
            if err != nil { b.Fatal(err) }
            b.SetBytes(int64(len(source)))
            b.ReportAllocs()
            b.ResetTimer()
            for i := 0; i < b.N; i++ {
                var err error
                eloqueBenchSink, err = Parse(nil, bytes.NewReader(source), nil, nil, 0)
                if err != nil { b.Fatal(err) }
            }
        })
    }
}
"""


def checked(command: list[str], *, env: dict[str, str] | None = None) -> str:
    print("+", " ".join(command), flush=True)
    p = subprocess.run(
        command, text=True, stdout=subprocess.PIPE, stderr=subprocess.STDOUT, env=env
    )
    if p.returncode:
        raise RuntimeError(
            f"Command failed ({p.returncode}): {command}\n{p.stdout[-6000:]}"
        )
    return p.stdout


def go_samples(output: str, expected: set[str], repeats: int) -> list[dict]:
    samples = []
    for m in GO_BENCH.finditer(output):
        if m["name"] not in expected or int(m["n"]) not in SIZES:
            continue
        samples.append(
            dict(
                engine={
                    "ControlledEloqueFrontend": "Eloque H2 fail-fast",
                    "ControlledEloqueRecoveringFrontend": "Eloque H2 recovering",
                    "ControlledGoCompilerFrontend": "Go compiler parser",
                }[m["name"]],
                size=int(m["n"]),
                ns=float(m["ns"]),
                boundary="in-process parser + AST",
                workload="controlled arithmetic",
            )
        )
    for engine in (
        {
            "ControlledEloqueFrontend": "Eloque H2 fail-fast",
            "ControlledEloqueRecoveringFrontend": "Eloque H2 recovering",
            "ControlledGoCompilerFrontend": "Go compiler parser",
        }[name]
        for name in expected
    ):
        for n in SIZES:
            actual = sum(s["engine"] == engine and s["size"] == n for s in samples)
            if actual != repeats:
                raise ValueError(
                    f"{engine} N={n}: expected {repeats} samples, got {actual}. Raw output:\n{output[-3000:]}"
                )
    return samples


def main() -> None:
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument("--repeats", type=int, default=10)
    p.add_argument("--benchtime", default="1s")
    p.add_argument(
        "--out",
        type=Path,
        default=Path("benchmarks/results/cross-language/frontend-v3"),
    )
    p.add_argument("--skip-go-compiler", action="store_true")
    p.add_argument("--skip-eloque", action="store_true")
    p.add_argument("--skip-cpython", action="store_true")
    args = p.parse_args()
    if args.repeats < 1:
        p.error("--repeats must be positive")
    if not Path("go.mod").exists():
        p.error("run this script from the Eloque repository root")
    args.out.mkdir(parents=True, exist_ok=True)
    rows = []
    versions = {"go": checked(["go", "version"]).strip()}
    if not args.skip_eloque:
        output = checked(
            [
                "go",
                "test",
                "./internal/frontend/parser",
                "-run",
                "^$",
                "-bench",
                "^BenchmarkControlledEloque(Frontend|RecoveringFrontend)$",
                "-benchmem",
                f"-benchtime={args.benchtime}",
                f"-count={args.repeats}",
            ]
        )
        (args.out / "eloque.log").write_text(output)
        rows += go_samples(
            output,
            {"ControlledEloqueFrontend", "ControlledEloqueRecoveringFrontend"},
            args.repeats,
        )
    if not args.skip_go_compiler:
        goroot = Path(checked(["go", "env", "GOROOT"]).strip())
        target = (
            goroot / "src/cmd/compile/internal/syntax/eloque_frontend_benchmark_test.go"
        )
        with tempfile.TemporaryDirectory(prefix="eloque-go-overlay-") as td:
            root = Path(td)
            replacement = root / "eloque_frontend_benchmark_test.go"
            replacement.write_text(OVERLAY_TEST)
            overlay = root / "overlay.json"
            overlay.write_text(json.dumps({"Replace": {str(target): str(replacement)}}))
            output = checked(
                [
                    "go",
                    "test",
                    "-vet=off",
                    f"-overlay={overlay}",
                    "cmd/compile/internal/syntax",
                    "-run",
                    "^$",
                    "-bench",
                    "^BenchmarkControlledGoCompilerFrontend$",
                    "-benchmem",
                    f"-benchtime={args.benchtime}",
                    f"-count={args.repeats}",
                ]
            )
        (args.out / "go-compiler.log").write_text(output)
        rows += go_samples(output, {"ControlledGoCompilerFrontend"}, args.repeats)
    if not args.skip_cpython:
        python_cmd = (
            ["uv", "run", "--no-project", "--python", "3.15", "python"]
            if shutil.which("uv")
            else [sys.executable]
        )
        versions["cpython"] = checked(python_cmd + ["--version"]).strip()
        script = Path(__file__).with_name("cpython_samples.py")
        output = checked(python_cmd + [str(script), "--repeats", str(args.repeats)])
        (args.out / "cpython.log").write_text(output)
        for line in output.splitlines():
            if line.startswith("{"):
                d = json.loads(line)
                rows.append(
                    dict(
                        engine="CPython ast.parse",
                        size=d["size"],
                        ns=d["ns"],
                        boundary="in-process parser + AST",
                        workload="controlled arithmetic",
                    )
                )
        for n in SIZES:
            if (
                sum(r["engine"] == "CPython ast.parse" and r["size"] == n for r in rows)
                != args.repeats
            ):
                raise ValueError(f"CPython N={n}: missing measurements")
    if not rows:
        p.error("no engines selected")
    with (args.out / "samples.csv").open("w", newline="") as f:
        writer = csv.DictWriter(
            f, fieldnames=["engine", "workload", "size", "ns", "boundary"]
        )
        writer.writeheader()
        writer.writerows(rows)
    (args.out / "manifest.json").write_text(
        json.dumps(
            {
                "created_utc": datetime.now(timezone.utc).isoformat(),
                "versions": versions,
                "repeats": args.repeats,
                "benchtime": args.benchtime,
                "note": "Language-specific wrappers and AST structures; in-process parsing; Go and Eloque via Go benchmark framework, CPython via timeit.",
            },
            indent=2,
        )
        + "\n"
    )
    groups = {}
    for row in rows:
        groups.setdefault((row["engine"], row["size"]), []).append(row["ns"])
    report = [
        "# Eloque / Go compiler / CPython controlled frontend benchmark",
        "",
        f"Each engine: {args.repeats} repetitions per size. Lower is better.",
        "",
        "| Engine | N=10 (µs) | N=100 (µs) | N=1,000 (ms) | N=10,000 (ms) |",
        "|---|---:|---:|---:|---:|",
    ]
    for engine in [
        "Eloque H2 fail-fast",
        "Eloque H2 recovering",
        "Go compiler parser",
        "CPython ast.parse",
    ]:
        if (engine, 10) not in groups:
            continue
        values = [statistics.median(groups[(engine, n)]) for n in SIZES]
        report.append(
            f"| {engine} | {values[0]/1e3:.2f} | {values[1]/1e3:.2f} | {values[2]/1e6:.3f} | {values[3]/1e6:.3f} |"
        )
    report += [
        "",
        "## Scope",
        "",
        "All measurements exclude subprocess startup from the timed operation. The Go compiler benchmark uses cmd/compile/internal/syntax via a generated test overlay. Eloque benchmarks reuse the existing production parser benchmark functions. CPython uses ast.parse. Grammars, wrappers, ASTs, and diagnostic responsibilities differ. This is not a full compiler or runtime performance ranking.",
        "",
        "Raw samples: samples.csv; Go logs and Python log are retained; environment: manifest.json.",
        "",
    ]
    (args.out / "REPORT.md").write_text("\n".join(report))
    print(f"Wrote {args.out}/REPORT.md and samples.csv")


if __name__ == "__main__":
    main()
