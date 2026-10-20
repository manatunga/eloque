"""
Controlled arithmetic frontend benchmark for CPython.
"""

import ast
import platform
from statistics import median
from timeit import Timer

SIZES = (10, 100, 1000, 10000)
REPEATS = 10


def build_source(count: int) -> str:
    return "".join(
        f"value_{i} = (({i} + 3) * (7 - 2)) + ({i} * 4)\n" for i in range(count)
    )


def main() -> None:
    print(f"Python implementation: {platform.python_implementation()}")
    print(f"Python version: {platform.python_version()}")
    print(f"Python compiler: {platform.python_compiler()}")
    print()

    for size in SIZES:
        source = build_source(size)

        # Validate before timing.
        ast.parse(source)

        timer = Timer(lambda: ast.parse(source))
        iterations, _ = timer.autorange()
        timings = timer.repeat(
            repeat=REPEATS,
            number=iterations,
        )

        ns_per_op = median(timings) / iterations * 1e9

        print(
            f"N={size:<5} "
            f"bytes={len(source.encode('utf-8')):<8} "
            f"frontend={ns_per_op:>12,.0f} ns/op"
        )


if __name__ == "__main__":
    main()
