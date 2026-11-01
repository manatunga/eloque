"""Determinism and controlled workload-shape tests."""

import unittest
import re
from workloads import LANGUAGES, WORKLOADS, generate


class WorkloadsTest(unittest.TestCase):
    def test_deterministic_and_counts(self):
        for language in LANGUAGES:
            for workload in WORKLOADS:
                for size in (10, 100):
                    with self.subTest(language=language, workload=workload, size=size):
                        case = generate(language, workload, size)
                        self.assertEqual(case, generate(language, workload, size))
                        self.assertEqual(case.declarations, size)
                        self.assertEqual(case.references, size)
                        for i in range(size):
                            self.assertEqual(
                                len(re.findall(rf"\bvalue_{i}\b", case.source)), 2
                            )

    def test_invalid(self):
        for args in (
            ("invalid", "arithmetic", 10),
            ("go", "invalid", 10),
            ("go", "arithmetic", 0),
        ):
            with self.assertRaises(ValueError):
                generate(*args)


if __name__ == "__main__":
    unittest.main()
