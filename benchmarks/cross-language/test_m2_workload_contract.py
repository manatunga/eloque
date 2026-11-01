"""
Verify the M2 native benchmark workload contract.
"""

from __future__ import annotations

import re
import unittest

from m2_native import LANGUAGES, source_for

WORKLOADS = (
    "declarations",
    "references",
    "mixed",
    "sparse-undefined",
    "undefined",
)

SIZES = (10, 100, 1000, 3000)


class WorkloadContractTests(unittest.TestCase):
    def test_counts_and_success_contract(self) -> None:
        for language in LANGUAGES:
            for workload in WORKLOADS:
                for n in SIZES:
                    with self.subTest(language=language, workload=workload, n=n):
                        source, declarations, references, valid = source_for(
                            language, workload, n
                        )

                        self.assertTrue(source.strip())
                        self.assertEqual(references, n)

                        expected_declarations = (
                            n
                            if workload in ("declarations", "mixed")
                            else 0 if workload == "undefined" else 1
                        )
                        self.assertEqual(declarations, expected_declarations)

                        expected_valid = workload != "undefined" and not (
                            workload == "sparse-undefined" and n > 100
                        )
                        self.assertEqual(valid, expected_valid)

    def test_sparse_undefined_positions(self) -> None:
        for language in LANGUAGES:
            for n in SIZES:
                with self.subTest(language=language, n=n):
                    source, _, _, _ = source_for(language, "sparse-undefined", n)

                    actual = set(re.findall(r"\bmissing_\d+\b", source))
                    expected = {f"missing_{i}" for i in range(1, n) if i % 100 == 0}

                    self.assertEqual(actual, expected)


if __name__ == "__main__":
    unittest.main()
