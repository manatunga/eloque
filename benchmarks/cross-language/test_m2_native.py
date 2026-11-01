"""
Tests for M2 native compiler benchmark adapters.

Run from the Eloque repository root:

    python3 -m unittest discover \
        -s benchmarks/cross-language \
        -p 'test_m2_native.py' -v

Compiler integration tests are skipped when the required tool is unavailable.
"""

from __future__ import annotations

from pathlib import Path
import shutil
import subprocess
import tempfile
import unittest

from m2_native import (
    BOUNDARIES,
    EXECUTABLES,
    EXTENSIONS,
    LANGUAGES,
    collect,
    command_for,
    source_for,
)

CASES = (
    "declarations",
    "references",
    "mixed",
    "sparse-undefined",
    "undefined",
)


class SourceGenerationTests(unittest.TestCase):
    def test_all_languages_and_workloads(self) -> None:
        for language in LANGUAGES:
            for workload in CASES:
                with self.subTest(language=language, workload=workload):
                    source, declarations, references, valid = source_for(
                        language, workload, 10
                    )

                    self.assertIsInstance(source, str)
                    self.assertTrue(source.strip())
                    self.assertEqual(references, 10)

                    if workload in ("declarations", "mixed"):
                        self.assertEqual(declarations, 10)
                    elif workload in ("references", "sparse-undefined"):
                        self.assertEqual(declarations, 1)
                    else:
                        self.assertEqual(declarations, 0)

                    self.assertEqual(
                        valid,
                        workload != "undefined"
                        and not (workload == "sparse-undefined" and 10 > 100),
                    )

    def test_references_reuse_one_binding(self) -> None:
        for language in LANGUAGES:
            with self.subTest(language=language):
                source, declarations, references, valid = source_for(
                    language, "references", 100
                )

                self.assertEqual(declarations, 1)
                self.assertEqual(references, 100)
                self.assertTrue(valid)
                self.assertIn("seed", source)

    def test_mixed_interleaves_declarations_and_references(self) -> None:
        for language in LANGUAGES:
            with self.subTest(language=language):
                source, declarations, references, valid = source_for(
                    language, "mixed", 10
                )

                self.assertEqual(declarations, 10)
                self.assertEqual(references, 10)
                self.assertTrue(valid)

                self.assertLess(
                    source.index("value_0"),
                    source.index("value_1"),
                )

    def test_sparse_undefined_frequency(self) -> None:
        for language in LANGUAGES:
            with self.subTest(language=language):
                source, declarations, references, valid = source_for(
                    language, "sparse-undefined", 300
                )

                self.assertEqual(declarations, 1)
                self.assertEqual(references, 300)
                self.assertFalse(valid)

                self.assertNotRegex(source, r"\bmissing_0\b")
                self.assertIn("missing_100", source)
                self.assertIn("missing_200", source)
                self.assertNotRegex(source, r"\bmissing_1\b")

    def test_deterministic_generation(self) -> None:
        for language in LANGUAGES:
            for workload in CASES:
                with self.subTest(language=language, workload=workload):
                    first = source_for(language, workload, 100)
                    second = source_for(language, workload, 100)
                    self.assertEqual(first, second)

    def test_invalid_arguments(self) -> None:
        with self.assertRaises(ValueError):
            source_for("unknown", "mixed", 10)

        with self.assertRaises(ValueError):
            source_for("c", "unknown", 10)

        with self.assertRaises(ValueError):
            source_for("c", "mixed", 0)

    def test_all_languages_have_boundaries(self) -> None:
        for language in LANGUAGES:
            self.assertIn(language, BOUNDARIES)
            self.assertTrue(BOUNDARIES[language])


class CompilerCommandTests(unittest.TestCase):
    def test_commands_use_expected_compilers(self) -> None:
        source = Path("/tmp/example.c")
        output = Path("/tmp/example.out")

        for language in LANGUAGES:
            with self.subTest(language=language):
                command = command_for(language, source, output)

                self.assertIsInstance(command, list)
                self.assertTrue(command)
                self.assertEqual(command[0], EXECUTABLES[language])

    def test_clang_uses_syntax_only(self) -> None:
        for language in ("c", "cpp"):
            with self.subTest(language=language):
                command = command_for(
                    language,
                    Path("example.c"),
                    Path("example.out"),
                )
                self.assertIn("-fsyntax-only", command)

    def test_rust_uses_metadata(self) -> None:
        command = command_for(
            "rust",
            Path("example.rs"),
            Path("example.rmeta"),
        )
        self.assertIn("--emit=metadata", command)

    def test_zig_uses_semantic_compilation(self) -> None:
        command = command_for(
            "zig",
            Path("example.zig"),
            Path("example.o"),
        )
        self.assertEqual(command[:2], ["zig", "build-obj"])
        self.assertIn("-fno-emit-bin", command)

    def test_go_uses_compiler(self) -> None:
        command = command_for(
            "go",
            Path("example.go"),
            Path("example.o"),
        )
        self.assertEqual(command[:3], ["go", "tool", "compile"])


class NativeCompilerIntegrationTests(unittest.TestCase):
    def test_generated_programs(self) -> None:
        """Check successful and intentionally invalid programs.

        This test checks compiler acceptance, not performance.
        """
        with tempfile.TemporaryDirectory(prefix="eloque-m2-test-") as temp:
            root = Path(temp)

            for language in LANGUAGES:
                if shutil.which(EXECUTABLES[language]) is None:
                    continue

                for workload in CASES:
                    with self.subTest(language=language, workload=workload):
                        source, _, _, expected_success = source_for(
                            language, workload, 10
                        )

                        source_path = root / (
                            f"{language}_{workload}" + EXTENSIONS[language]
                        )
                        output_path = root / (f"{language}_{workload}.out")

                        source_path.write_text(
                            source,
                            encoding="utf-8",
                        )

                        command = command_for(
                            language,
                            source_path,
                            output_path,
                        )

                        result = subprocess.run(
                            command,
                            capture_output=True,
                            text=True,
                            timeout=60,
                        )

                        self.assertEqual(
                            result.returncode == 0,
                            expected_success,
                            msg=(
                                f"{language}/{workload}: "
                                f"exit={result.returncode}\n"
                                f"{result.stderr[-2000:]}"
                            ),
                        )

    def test_collector_smoke(self) -> None:
        """Run one small collection using an available native compiler."""
        available = [
            language
            for language in LANGUAGES
            if shutil.which(EXECUTABLES[language]) is not None
        ]

        if not available:
            self.skipTest("No supported native compiler installed")

        with tempfile.TemporaryDirectory(prefix="eloque-m2-collect-") as temp:
            rows, inputs = collect(
                languages=[available[0]],
                cases=("declarations", "undefined"),
                sizes=(10,),
                repeats=1,
                warmups=0,
                out=Path(temp),
            )

        self.assertEqual(len(rows), 2)
        self.assertEqual(len(inputs), 2)

        for row in rows:
            self.assertGreater(row["ns_per_op"], 0)
            self.assertIn("boundary", row)

        for entry in inputs:
            self.assertEqual(len(entry["sha256"]), 64)
            self.assertIn("expected_success", entry)
            self.assertIn("command", entry)


if __name__ == "__main__":
    unittest.main()
