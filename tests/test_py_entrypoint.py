"""Tests for uvpip Python entrypoint wrapper."""

import io
import os
import sys
import unittest
from unittest.mock import patch
from uvpip import __version__, translate_to_uv


class TestUVPipPython(unittest.TestCase):
    def test_version(self):
        self.assertEqual(__version__, "1.0.9")

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
        # Test double-dash disables option filtering so literal package names like '-y' are kept
        self.assertEqual(
            translate_to_uv(["uninstall", "-y", "--", "-y"]),
            ["pip", "uninstall", "--", "-y"],
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


class TestMainCLI(unittest.TestCase):
    def setUp(self):
        self.orig_argv = sys.argv
        self.orig_env = os.environ.copy()

    def tearDown(self):
        sys.argv = self.orig_argv
        os.environ.clear()
        os.environ.update(self.orig_env)

    @patch("shutil.which")
    @patch("subprocess.call")
    def test_main_delegates_to_native_uvpip(self, mock_call, mock_which):
        from uvpip import main
        mock_which.return_value = "/usr/local/bin/uvpip"
        mock_call.return_value = 0
        sys.argv = ["uvpip-script.py", "install", "flask"]

        code = main()
        self.assertEqual(code, 0)
        mock_call.assert_called_once_with(["/usr/local/bin/uvpip", "install", "flask"])

    @patch("shutil.which")
    @patch("subprocess.call")
    def test_main_native_uvpip_oserror_falls_back(self, mock_call, mock_which):
        from uvpip import main
        # First call for uvpip returns path but call raises OSError; second call finds uv
        def fake_which(cmd):
            if cmd == "uvpip":
                return "/usr/local/bin/uvpip"
            if cmd == "uv":
                return "/usr/bin/uv"
            return None

        mock_which.side_effect = fake_which
        # First call fails, second succeeds
        mock_call.side_effect = [OSError("cannot run"), 0]
        sys.argv = ["uvpip-script.py", "list"]

        code = main()
        self.assertEqual(code, 0)
        self.assertEqual(mock_call.call_count, 2)

    @patch("shutil.which")
    @patch("subprocess.call")
    def test_main_finds_uv_via_env_override(self, mock_call, mock_which):
        from uvpip import main
        mock_which.return_value = None
        os.environ["UVPIP_UV"] = "/opt/custom/uv"
        os.environ["VIRTUAL_ENV"] = "/env"
        mock_call.return_value = 42
        sys.argv = ["uvpip", "install", "pytest"]

        code = main()
        self.assertEqual(code, 42)
        mock_call.assert_called_once()
        args, kwargs = mock_call.call_args
        self.assertEqual(args[0], ["/opt/custom/uv", "pip", "install", "pytest"])
        self.assertNotIn("UV_SYSTEM_PYTHON", kwargs["env"])

    @patch("shutil.which")
    @patch("subprocess.call")
    def test_main_sets_uv_system_python_when_outside_venv(self, mock_call, mock_which):
        from uvpip import main
        mock_which.side_effect = lambda cmd: "/usr/bin/uv" if cmd == "uv" else None
        os.environ.pop("VIRTUAL_ENV", None)
        os.environ.pop("CONDA_PREFIX", None)
        os.environ.pop("UV_SYSTEM_PYTHON", None)
        mock_call.return_value = 0
        sys.argv = ["uvpip", "list"]

        code = main()
        self.assertEqual(code, 0)
        args, kwargs = mock_call.call_args
        self.assertEqual(kwargs["env"].get("UV_SYSTEM_PYTHON"), "1")

    @patch("shutil.which")
    @patch("os.path.isfile")
    @patch("os.access")
    @patch("subprocess.call")
    def test_main_finds_candidate_path(self, mock_call, mock_access, mock_isfile, mock_which):
        from uvpip import main
        mock_which.return_value = None
        os.environ.pop("UVPIP_UV", None)
        # Pretend the second candidate (~/.cargo/bin/uv) exists and is executable
        home = os.path.expanduser("~")
        cargo_uv = os.path.join(home, ".cargo", "bin", "uv")

        def fake_isfile(path):
            return path == cargo_uv

        def fake_access(path, mode):
            return path == cargo_uv

        mock_isfile.side_effect = fake_isfile
        mock_access.side_effect = fake_access
        mock_call.return_value = 0
        sys.argv = ["uvpip", "freeze"]

        code = main()
        self.assertEqual(code, 0)
        self.assertEqual(mock_call.call_args[0][0][0], cargo_uv)

    @patch("shutil.which")
    @patch("os.path.isfile")
    def test_main_missing_uv_returns_127(self, mock_isfile, mock_which):
        from uvpip import main
        mock_which.return_value = None
        mock_isfile.return_value = False
        os.environ.pop("UVPIP_UV", None)
        sys.argv = ["uvpip", "check"]

        stderr_buf = io.StringIO()
        with patch("sys.stderr", stderr_buf):
            code = main()
        self.assertEqual(code, 127)
        self.assertIn("uv not found on PATH", stderr_buf.getvalue())

    @patch("shutil.which")
    @patch("subprocess.call")
    def test_main_subprocess_oserror_returns_126(self, mock_call, mock_which):
        from uvpip import main
        mock_which.side_effect = lambda cmd: "/usr/bin/uv" if cmd == "uv" else None
        mock_call.side_effect = OSError("permission denied")
        sys.argv = ["uvpip", "install", "pkg"]

        stderr_buf = io.StringIO()
        with patch("sys.stderr", stderr_buf):
            code = main()
        self.assertEqual(code, 126)
        self.assertIn("failed to execute uv", stderr_buf.getvalue())

    @patch("uvpip.main")
    def test_dunder_main_module(self, mock_main):
        import runpy
        mock_main.return_value = 0
        with patch.object(sys, "argv", ["uvpip", "--version"]):
            with self.assertRaises(SystemExit) as ctx:
                runpy.run_module("uvpip.__main__", run_name="__main__")
            self.assertEqual(ctx.exception.code, 0)

    def test_dunder_init_main(self):
        import runpy
        repo_root = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
        init_path = os.path.join(repo_root, "src", "uvpip", "__init__.py")
        with patch("shutil.which", return_value="/mock/bin/uv"), \
             patch("subprocess.call", return_value=0), \
             patch.object(sys, "argv", ["uvpip", "--version"]):
            with self.assertRaises(SystemExit) as ctx:
                runpy.run_path(init_path, run_name="__main__")
            self.assertEqual(ctx.exception.code, 0)


if __name__ == "__main__":
    unittest.main()

