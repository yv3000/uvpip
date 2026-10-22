"""Tests for uvpip Python entrypoint wrapper."""

import os
import unittest
from uvpip import __version__, translate_to_uv


class TestUVPipPython(unittest.TestCase):
    def test_version(self):
        self.assertEqual(__version__, "1.0.8")

    def test_lockfile_present_and_consistent(self):
        repo_root = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
        lock_path = os.path.join(repo_root, "uv.lock")
        self.assertTrue(os.path.isfile(lock_path), "uv.lock must exist for reproducible dependencies")
        with open(lock_path, "r", encoding="utf-8") as f:
            content = f.read()
        self.assertIn(f'version = "{__version__}"', content)

    def test_translate_empty(self):
        self.assertEqual(translate_to_uv([]), ["pip"])

    def test_translate_install(self):
        self.assertEqual(
            translate_to_uv(["install", "requests", "--upgrade"]),
            ["pip", "install", "requests", "--upgrade"],
        )

    def test_translate_uninstall_strips_yes(self):
        self.assertEqual(
            translate_to_uv(["uninstall", "-y", "requests"]),
            ["pip", "uninstall", "requests"],
        )
        self.assertEqual(
            translate_to_uv(["uninstall", "--yes", "requests"]),
            ["pip", "uninstall", "requests"],
        )

    def test_translate_upgrade_alias(self):
        self.assertEqual(
            translate_to_uv(["upgrade", "requests"]),
            ["pip", "install", "--upgrade", "requests"],
        )

    def test_translate_cache(self):
        self.assertEqual(
            translate_to_uv(["cache", "dir"]),
            ["cache", "dir"],
        )


if __name__ == "__main__":
    unittest.main()
