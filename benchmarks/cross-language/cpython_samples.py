"""
CPython controlled arithmetic AST parsing; raw independent timeit batches.
"""

import argparse
import ast
import json
import timeit


def source_for(n: int) -> str:
    return "".join(f"value_{i} = (({i} + 3) * (7 - 2)) + ({i} * 4)\n" for i in range(n))


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--repeats", type=int, default=10)
    args = parser.parse_args()
    for n in (10, 100, 1000, 10000):
        source = source_for(n)
        ast.parse(source)
        timer = timeit.Timer(lambda: ast.parse(source))
        iterations, _ = timer.autorange()
        for _ in range(args.repeats):
            elapsed = timer.timeit(number=iterations)
            print(
                json.dumps(
                    {
                        "size": n,
                        "ns": elapsed * 1e9 / iterations,
                        "iterations": iterations,
                    }
                ),
                flush=True,
            )


if __name__ == "__main__":
    main()
