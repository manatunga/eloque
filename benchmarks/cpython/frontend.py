"""
Benchmark CPython tokenization and AST parsing across frontend workloads.
"""

import ast
import io
from statistics import median
import tokenize
from timeit import Timer

SIZES = (10, 100, 1000)
REPEATS = 5


def build_declarations(count: int) -> str:
    return "".join(f"value_{i} = {i}\n" for i in range(count))


def build_expressions(count: int) -> str:
    return "".join(f"value_{i} = (({i} + 3) * (7 - 2)) ** 2\n" for i in range(count))


def build_strings(count: int) -> str:
    return "".join(
        f'print("message {i}: the quick brown fox jumps over the lazy dog")\n'
        for i in range(count)
    )


def build_comments(count: int) -> str:
    return "".join(
        f"# This comment describes declaration number {i} and its purpose.\n"
        f"value_{i} = {i}\n"
        for i in range(count)
    )


def build_mixed(count: int) -> str:
    return "".join(
        f"# Processing value {i}\n"
        f"value_{i} = ({i} + 3) * 4\n"
        f"value_{i} = value_{i} ** 2\n"
        f'print("processed value {i}")\n'
        f"print(value_{i})\n"
        for i in range(count)
    )


def tokenize_source(source: str) -> None:
    list(tokenize.generate_tokens(io.StringIO(source).readline))


def parse_source(source: str) -> None:
    ast.parse(source)


def benchmark(function, source: str) -> float:

    timer = Timer(lambda: function(source))

    iterations, _ = timer.autorange()
    timings = timer.repeat(repeat=REPEATS, number=iterations)

    return median(timings) / iterations

def main() -> None:
    workloads = {
        "Declarations": build_declarations,
        "Expressions": build_expressions,
        "Strings": build_strings,
        "Comments": build_comments,
        "Mixed": build_mixed,
    }

    for name, build in workloads.items():
        for size in SIZES:
            source = build(size)

            token_ns = benchmark(tokenize_source, source) * 1e9
            frontend_ns = benchmark(parse_source, source) * 1e9

            print(
                f"{name:<14} "
                f"N={size:<5} "
                f"tokenize={token_ns:>12,.0f} ns/op "
                f"frontend={frontend_ns:>12,.0f} ns/op"
            )


if __name__ == "__main__":
    main()
