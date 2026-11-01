#!/usr/bin/env python3
"""Run cross-language compiler subprocess benchmarks.

All source generation and correctness validation occur before timing. Timed
operations spawn a compiler process, so these results include process startup,
compiler initialization, frontend processing, and any selected later phases.
They must not be ranked as equivalent to in-process parser benchmarks.
"""

from __future__ import annotations

import argparse
import csv
import hashlib
import json
import os
from pathlib import Path
import platform
import random
import shutil
import subprocess
import sys
import tempfile
import time
from datetime import datetime, timezone

from workloads import LANGUAGES, SIZES, WORKLOADS, generate


def compiler_command(
    language: str, source: Path, output: Path, mode: str = "default"
) -> list[str]:
    """Choose a compiler mode that avoids linking or running user programs.

    Clang uses syntax-only checking; rustc emits metadata; Zig emits an object;
    Go invokes the standard compiler directly to emit an object. These are
    *different phase boundaries*, intentionally recorded in the results.
    """
    if mode == "syntax" and language == "zig":
        return ["zig", "ast-check", str(source)]
    if mode == "syntax" and language not in ("c", "cpp", "zig"):
        raise ValueError(f"syntax-only mode unavailable for {language}; use default")
    if language == "c":
        return [
            "clang",
            "-std=c17",
            "-fsyntax-only",
            "-Wno-unused-variable",
            str(source),
        ]
    if language == "cpp":
        return [
            "clang++",
            "-std=c++20",
            "-fsyntax-only",
            "-Wno-unused-variable",
            str(source),
        ]
    if language == "rust":
        return [
            "rustc",
            "--edition=2024",
            "--emit=metadata",
            "-C",
            "opt-level=0",
            "-o",
            str(output),
            str(source),
        ]
    if language == "zig":
        return ["zig", "build-obj", "-O", "Debug", f"-femit-bin={output}", str(source)]
    if language == "go":
        return ["go", "tool", "compile", "-o", str(output), str(source)]
    raise ValueError(language)


PHASES = {
    "c": "clang syntax-only (parse + semantic validation)",
    "cpp": "clang++ syntax-only (parse + semantic validation)",
    "rust": "rustc metadata emission (includes semantic analysis)",
    "zig": "zig object emission (includes code generation)",
    "go": "go tool compile object emission (includes code generation)",
}


def run_once(cmd: list[str], cwd: Path) -> tuple[int, str]:
    start = time.perf_counter_ns()
    result = subprocess.run(cmd, cwd=cwd, capture_output=True, text=True, check=False)
    elapsed = time.perf_counter_ns() - start
    if result.returncode:
        raise RuntimeError(f"command failed: {' '.join(cmd)}\n{result.stderr[:4000]}")
    return elapsed, result.stderr


def version(command: list[str]) -> str:
    try:
        proc = subprocess.run(
            command, capture_output=True, text=True, timeout=15, check=False
        )
        return (proc.stdout or proc.stderr).splitlines()[0].strip()
    except (OSError, subprocess.TimeoutExpired, IndexError):
        return "unavailable"


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument(
        "--mode",
        choices=("default", "syntax"),
        default="default",
        help="syntax is supported only by C, C++ and Zig",
    )
    parser.add_argument(
        "--languages", nargs="+", choices=LANGUAGES, default=list(LANGUAGES)
    )
    parser.add_argument(
        "--workloads", nargs="+", choices=WORKLOADS, default=list(WORKLOADS)
    )
    parser.add_argument("--sizes", nargs="+", type=int, default=list(SIZES))
    parser.add_argument("--repeats", type=int, default=10)
    parser.add_argument("--warmups", type=int, default=2)
    parser.add_argument("--seed", type=int, default=42)
    parser.add_argument(
        "--out", type=Path, default=Path("benchmarks/results/cross-language/compiler-subprocess")
    )
    args = parser.parse_args()
    if args.repeats < 1 or args.warmups < 0 or any(n < 1 for n in args.sizes):
        parser.error("repeats and sizes must be positive; warmups cannot be negative")
    missing = [
        lang
        for lang in args.languages
        if shutil.which(
            {"c": "clang", "cpp": "clang++", "rust": "rustc", "zig": "zig", "go": "go"}[
                lang
            ]
        )
        is None
    ]
    if missing:
        parser.error(f"missing compilers: {', '.join(missing)}")

    if args.mode == "syntax" and any(
        lang not in ("c", "cpp", "zig") for lang in args.languages
    ):
        parser.error(
            "--mode syntax supports only c, cpp, zig; select --languages explicitly"
        )
    args.out.mkdir(parents=True, exist_ok=True)
    rows: list[dict[str, object]] = []
    cases = [
        (lang, workload, size)
        for lang in args.languages
        for workload in args.workloads
        for size in args.sizes
    ]
    rng = random.Random(args.seed)
    with tempfile.TemporaryDirectory(prefix="eloque-crosslang-") as temp:
        root = Path(temp)
        prepared: list[tuple[str, str, int, int, list[str], str]] = []
        input_manifest = []
        for index, (lang, workload, size) in enumerate(cases):
            case = generate(lang, workload, size)
            src = root / f"case_{index}{case.extension}"
            src.write_text(case.source, encoding="utf-8")
            output = root / f"case_{index}.out"
            cmd = compiler_command(lang, src, output, args.mode)
            sha256 = hashlib.sha256(case.source.encode("utf-8")).hexdigest()
            input_manifest.append(
                {
                    "language": lang,
                    "workload": workload,
                    "size": size,
                    "sha256": sha256,
                    "bytes": len(case.source.encode("utf-8")),
                    "declarations": case.declarations,
                    "references": case.references,
                    "command": cmd,
                }
            )
            # Validate before any timed work. Zig/Go/Rust may emit an object.
            run_once(cmd, root)
            prepared.append(
                (lang, workload, size, len(case.source.encode()), cmd, sha256)
            )
        for _ in range(args.warmups):
            rng.shuffle(prepared)
            for _, _, _, _, cmd, _ in prepared:
                run_once(cmd, root)
        for repetition in range(args.repeats):
            rng.shuffle(prepared)
            for lang, workload, size, source_bytes, cmd, sha256 in prepared:
                elapsed, _ = run_once(cmd, root)
                rows.append(
                    {
                        "language": lang,
                        "workload": workload,
                        "size": size,
                        "repetition": repetition + 1,
                        "ns": elapsed,
                        "source_bytes": source_bytes,
                        "boundary": (
                            "zig AST check (parse only)"
                            if args.mode == "syntax" and lang == "zig"
                            else PHASES[lang]
                        ),
                        "source_sha256": sha256,
                    }
                )
                print(
                    f"{lang:5} {workload:12} N={size:<6} run={repetition+1:<3} {elapsed/1e6:.3f} ms",
                    flush=True,
                )
    with (args.out / "samples.csv").open("w", newline="") as file:
        writer = csv.DictWriter(
            file,
            fieldnames=[
                "language",
                "workload",
                "size",
                "repetition",
                "ns",
                "source_bytes",
                "boundary",
                "source_sha256",
            ],
        )
        writer.writeheader()
        writer.writerows(rows)
    metadata = {
        "created_utc": datetime.now(timezone.utc).isoformat(),
        "host": platform.platform(),
        "python": sys.version.split()[0],
        "versions": {
            "c": version(["clang", "--version"]),
            "cpp": version(["clang++", "--version"]),
            "rust": version(["rustc", "--version"]),
            "zig": version(["zig", "version"]),
            "go": version(["go", "version"]),
        },
        "inputs": input_manifest,
        "configuration": {
            "mode": args.mode,
            "languages": args.languages,
            "workloads": args.workloads,
            "sizes": args.sizes,
            "repeats": args.repeats,
            "warmups": args.warmups,
            "seed": args.seed,
        },
        "phase_boundaries": PHASES,
        "limitations": "Subprocess wall time, including startup; non-equivalent compiler phases; not comparable directly to Eloque in-process parser benchmarks.",
    }
    (args.out / "manifest.json").write_text(json.dumps(metadata, indent=2) + "\n")
    print(f"Saved {len(rows)} samples to {args.out}")


if __name__ == "__main__":
    main()
