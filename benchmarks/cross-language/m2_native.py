"""
Native compiler adapters for M2 identifier-binding workloads.

Compiler subprocess timings are NOT isolated resolver timings. Source generation,
validation, and tool discovery happen outside the timed region.
"""

from __future__ import annotations

import hashlib
from pathlib import Path
import shutil
import subprocess
import tempfile
import time

LANGUAGES = ("c", "cpp", "rust", "zig", "go")
EXTENSIONS = {"c": ".c", "cpp": ".cpp", "rust": ".rs", "zig": ".zig", "go": ".go"}
EXECUTABLES = {
    "c": "clang",
    "cpp": "clang++",
    "rust": "rustc",
    "zig": "zig",
    "go": "go",
}
BOUNDARIES = {
    "c": "compiler subprocess: Clang parse + semantic validation (syntax-only)",
    "cpp": "compiler subprocess: Clang++ parse + semantic validation (syntax-only)",
    "rust": "compiler subprocess: rustc metadata compilation (not resolver-only)",
    "zig": "compiler subprocess: Zig AST check (not resolver-only)",
    "go": "compiler subprocess: Go object compilation (not resolver-only)",
}


def source_for(language: str, workload: str, n: int) -> tuple[str, int, int, bool]:
    """Return source, declaration count, reference count, expected success.

    'undefined' cases intentionally fail name lookup. Every language gets the
    same logical number of declarations/references; syntax differs by language.
    """
    if language not in LANGUAGES or n < 1:
        raise ValueError((language, workload, n))
    if workload not in (
        "declarations",
        "references",
        "mixed",
        "sparse-undefined",
        "undefined",
    ):
        raise ValueError(workload)

    names: list[tuple[str, int]] = []
    if workload == "declarations":
        names = [(f"value_{i}", i) for i in range(n)]
        uses = [f"value_{i}" for i in range(n)]
    elif workload == "references":
        names = [("seed", 7)]
        uses = ["seed"] * n
    elif workload == "mixed":
        names = [(f"value_{i}", i) for i in range(n)]
        uses = [f"value_{i}" for i in range(n)]
    elif workload == "undefined":
        uses = [f"missing_{i}" for i in range(n)]
    else:
        names = [("seed", 7)]
        uses = [f"missing_{i}" if i > 0 and i % 100 == 0 else "seed" for i in range(n)]

    statements: list[str] = []

    def declaration(name: str, value: int) -> str:
        if language in ("c", "cpp"):
            return f"long long {name} = {value};"
        if language == "rust":
            return f"let {name}: i64 = {value};"
        if language == "zig":
            return f"const {name}: i64 = {value};"
        return f"{name} := {value}"

    def reference(name: str) -> str:
        if language in ("c", "cpp"):
            return f"(void)&{name};"
        if language == "rust":
            return f"let _ = &{name};"
        if language == "zig":
            return f"_ = &{name};"
        return f"_ = &{name}"

    if workload == "mixed":
        for (name, value), use in zip(names, uses):
            statements.extend((declaration(name, value), reference(use)))
    else:
        statements.extend(declaration(name, value) for name, value in names)
        statements.extend(reference(use) for use in uses)

    body = "".join("    " + line + "\n" for line in statements)

    if language == "c":
        source = "int main(void) {\n" + body + "    return 0;\n}\n"
    elif language == "cpp":
        source = "int main() {\n" + body + "    return 0;\n}\n"
    elif language == "rust":
        source = "fn main() {\n" + body + "}\n"
    elif language == "zig":
        source = "pub fn main() void {\n" + body + "}\n"
    else:
        source = "package main\nfunc main() {\n" + body + "}\n"

    return (
        source,
        len(names),
        len(uses),
        workload != "undefined" and not (workload == "sparse-undefined" and n > 100),
    )


def command_for(language: str, source: Path, output: Path) -> list[str]:
    if language == "c":
        return [
            "clang",
            "-std=c17",
            "-fsyntax-only",
            "-Wno-unused-variable",
            "-ferror-limit=1",
            str(source),
        ]
    if language == "cpp":
        return [
            "clang++",
            "-std=c++20",
            "-fsyntax-only",
            "-Wno-unused-variable",
            "-ferror-limit=1",
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
        return [
            "zig",
            "build-obj",
            str(source),
            "-fno-emit-bin",
        ]
    if language == "go":
        return ["go", "tool", "compile", "-o", str(output), str(source)]
    raise ValueError(language)


def collect(
    languages: list[str],
    cases: tuple[str, ...],
    sizes: tuple[int, ...],
    repeats: int,
    warmups: int,
    out: Path,
) -> tuple[list[dict], list[dict]]:
    """Return timing rows and manifest inputs; never time source generation."""
    missing = [lang for lang in languages if shutil.which(EXECUTABLES[lang]) is None]
    if missing:
        raise RuntimeError("Missing compilers: " + ", ".join(missing))

    rows: list[dict] = []
    inputs: list[dict] = []

    with tempfile.TemporaryDirectory(prefix="eloque-m2-native-") as temp:
        root = Path(temp)

        for language in languages:
            for workload in cases:
                for n in sizes:
                    source, declarations, references, valid = source_for(
                        language, workload, n
                    )
                    stem = f"{language}_{workload}_{n}"
                    src = root / (stem + EXTENSIONS[language])
                    dst = root / (stem + ".out")
                    src.write_text(source, encoding="utf-8")
                    cmd = command_for(language, src, dst)

                    inputs.append(
                        {
                            "language": language,
                            "workload": workload,
                            "size": n,
                            "source_bytes": len(source.encode()),
                            "sha256": hashlib.sha256(source.encode()).hexdigest(),
                            "declarations": declarations,
                            "references": references,
                            "expected_success": valid,
                            "command": [
                                part.replace(str(root), "<temporary-dir>")
                                for part in cmd
                            ],
                        }
                    )

                    check = subprocess.run(cmd, capture_output=True, text=True)
                    if (check.returncode == 0) != valid:
                        raise RuntimeError(
                            f"Unexpected compiler result for {stem}: "
                            f"exit={check.returncode}\n"
                            + (check.stderr or check.stdout)[-3000:]
                        )

                    for repetition in range(-warmups, repeats):
                        start = time.perf_counter_ns()
                        proc = subprocess.run(
                            cmd,
                            stdout=subprocess.DEVNULL,
                            stderr=subprocess.DEVNULL,
                        )
                        elapsed = time.perf_counter_ns() - start

                        if (proc.returncode == 0) != valid:
                            raise RuntimeError(
                                f"Compiler result changed during timing: {stem}"
                            )

                        if repetition >= 0:
                            rows.append(
                                {
                                    "engine": f"{language} compiler",
                                    "workload": workload,
                                    "size": n,
                                    "repetition": repetition + 1,
                                    "ns_per_op": elapsed,
                                    "bytes_per_op": "",
                                    "allocs_per_op": "",
                                    "boundary": BOUNDARIES[language],
                                    "source_bytes": len(source.encode()),
                                }
                            )

                    print(
                        f"{language:5} {workload:16} N={n:<5} completed",
                        flush=True,
                    )

    return rows, inputs
