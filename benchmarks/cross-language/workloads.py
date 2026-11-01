"""Deterministic, controlled compiler inputs with matched declaration/use counts.

Each unit contains one declaration and one address-taking reference. Address
references are intentionally chosen to satisfy Zig's unused-binding rules.
Source construction is never timed.
"""

from __future__ import annotations
from dataclasses import dataclass

SIZES = (10, 100, 1_000, 10_000)
LANGUAGES = ("c", "cpp", "rust", "zig", "go")
WORKLOADS = ("declarations", "arithmetic")


@dataclass(frozen=True)
class SourceCase:
    language: str
    workload: str
    size: int
    source: str
    extension: str
    declarations: int
    references: int


def generate(language: str, workload: str, size: int) -> SourceCase:
    if language not in LANGUAGES or workload not in WORKLOADS or size < 1:
        raise ValueError(f"invalid benchmark case: {language}/{workload}/{size}")
    declarations = []
    uses = []
    for i in range(size):
        expr = (
            str(i)
            if workload == "declarations"
            else f"(({i} + 3) * (7 - 2)) + ({i} * 4)"
        )
        if language in ("c", "cpp"):
            declarations.append(f"    long long value_{i} = {expr};")
            uses.append(f"    (void)&value_{i};")
        elif language == "rust":
            declarations.append(f"    let value_{i}: i64 = {expr};")
            uses.append(f"    let _ = &value_{i};")
        elif language == "zig":
            declarations.append(f"    const value_{i}: i64 = {expr};")
            uses.append(f"    _ = &value_{i};")
        else:
            declarations.append(f"    value_{i} := {expr}")
            uses.append(f"    _ = &value_{i}")
    body = "\n".join(declarations + uses) + "\n"
    if language == "c":
        source, ext = "int main(void) {\n" + body + "    return 0;\n}\n", ".c"
    elif language == "cpp":
        source, ext = "int main() {\n" + body + "    return 0;\n}\n", ".cpp"
    elif language == "rust":
        source, ext = "fn main() {\n" + body + "}\n", ".rs"
    elif language == "zig":
        source, ext = "pub fn main() void {\n" + body + "}\n", ".zig"
    else:
        source, ext = "package main\nfunc main() {\n" + body + "}\n", ".go"
    return SourceCase(language, workload, size, source, ext, size, size)
