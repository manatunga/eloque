"""
Monowire experimental VM implementation-language comparison.

Run from anywhere: uv run --no-project python benchmarks/monowire/compare.py
Result names follow benchmarks/README.md.
"""

import argparse
from datetime import datetime
from pathlib import Path
import platform
import random
import statistics
import subprocess
import sys
import time

HERE = Path(__file__).resolve().parent
RESULTS = HERE.parent / "results" / "monowire" / "lang-comparison"
WORKLOADS = ("call_loop", "arithmetic", "branch", "recursion")
BINARIES = {
    "Go": "go/monowire-go",
    "Rust": "rust/monowire-rust",
    "Zig switch": "zig/monowire-switch",
    "Zig labeled": "zig/monowire-labeled",
}
EXPECTED = {
    "call_loop": lambda n: n * (n + 1) // 2,
    "arithmetic": lambda n: n * (n + 1) // 2,
    "branch": lambda n: (n // 2) * ((n + 1) // 2),
    "recursion": lambda n: n * (n + 1) // 2,
}


def command_output(*args):
    try:
        return subprocess.check_output(
            args, text=True, stderr=subprocess.STDOUT
        ).strip()
    except (OSError, subprocess.CalledProcessError) as exc:
        return f"unavailable: {exc}"


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--n", type=int, default=10_000)
    parser.add_argument("--reps", type=int, default=100)
    parser.add_argument("--runs", type=int, default=15)
    parser.add_argument("--workload", choices=(*WORKLOADS, "all"), default="all")
    args = parser.parse_args()
    if not (0 <= args.n <= 1_000_000 and args.reps > 0 and args.runs >= 2):
        parser.error("require 0 <= n <= 1000000, reps > 0, runs >= 2")

    workloads = WORKLOADS if args.workload == "all" else (args.workload,)
    if "recursion" in workloads and args.n > 10_000:
        parser.error("recursion workload supports n <= 10000 (fixed Zig VM capacity)")
    rng = random.Random(42)
    date = datetime.now().astimezone().strftime("%Y%m%d")
    git_revision = command_output(
        "git", "-C", str(HERE), "rev-parse", "--short", "HEAD"
    )
    environment = [
        f"timestamp={datetime.now().astimezone().isoformat()}",
        f"git_revision={git_revision}",
        f"system={platform.platform()}",
        f"machine={platform.machine()}",
        f"python={sys.version.split()[0]}",
        f"go={command_output('go', 'version')}",
        f"rust={command_output('rustc', '--version')}",
        f"zig={command_output('zig', 'version')}",
        f"command=uv run --no-project compare.py --n {args.n} --reps {args.reps} --runs {args.runs} --workload {args.workload}",
        "method=external subprocess wall-clock time; includes startup; warm-up once per binary; shuffled order per round",
        "build=go build; rustc -C opt-level=3; zig -O ReleaseFast -lc",
        "measurement_unit=ms",
        "checkpoint=pre-m2 (working experiment label, not a roadmap milestone)",
    ]

    for workload in workloads:
        expected = EXPECTED[workload](args.n)
        expected_checksum = expected if args.reps % 2 else 0
        commands = {
            name: [str(HERE / workload / binary), str(args.n), str(args.reps)]
            for name, binary in BINARIES.items()
        }
        samples = {name: [] for name in commands}

        def execute(name):
            start = time.perf_counter_ns()
            proc = subprocess.run(
                commands[name], text=True, capture_output=True, check=True
            )
            duration = (time.perf_counter_ns() - start) / 1e6
            output = proc.stdout + proc.stderr
            markers = (
                f"workload={workload}",
                f"n={args.n}",
                f"reps={args.reps}",
                f"expected={expected}",
                f"checksum={expected_checksum}",
            )
            if any(marker not in output for marker in markers):
                raise RuntimeError(f"Invalid output from {name}: {output!r}")
            return duration

        for name in commands:
            execute(name)
        for _ in range(args.runs):
            order = list(commands)
            rng.shuffle(order)
            for name in order:
                samples[name].append(execute(name))

        lines = [
            f"workload={workload}",
            f"n={args.n} reps={args.reps} runs={args.runs}",
            f"expected={expected} checksum={expected_checksum}",
            *environment,
            "",
            "summary:",
        ]
        for name, times in sorted(
            samples.items(), key=lambda pair: statistics.median(pair[1])
        ):
            lines.append(
                f"{name:12s} median={statistics.median(times):8.3f} ms "
                f"mean={statistics.mean(times):8.3f} ms "
                f"stdev={statistics.stdev(times):7.3f} ms "
                f"min={min(times):8.3f} ms max={max(times):8.3f} ms"
            )
        lines.append("")
        lines.append("raw_samples_ms:")
        for name, times in samples.items():
            lines.append(f"{name}: " + ", ".join(f"{t:.6f}" for t in times))

        directory = RESULTS / workload
        directory.mkdir(parents=True, exist_ok=True)
        stem = f"pre-m2_monowire-lang-comparison_{workload.replace('_', '-')}_{date}"
        path = directory / f"{stem}.txt"
        suffix = 2
        while path.exists():
            path = directory / f"{stem}_r{suffix:02d}.txt"
            suffix += 1
        path.write_text("\n".join(lines) + "\n", encoding="utf-8")
        print(
            "\n".join(
                lines[:4]
                + [""]
                + lines[lines.index("summary:") : lines.index("raw_samples_ms:")]
            )
        )
        print(f"Saved: {path}")


if __name__ == "__main__":
    main()
