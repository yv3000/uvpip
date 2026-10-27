"""Dedicated test suite for uvpip CLI translation and argument passthrough."""

import io
import os
import sys
import unittest
from unittest.mock import patch
from uvpip import main, translate_to_uv


class TestCLITranslation(unittest.TestCase):
    """Test console script argument passthrough and error propagation."""

    def setUp(self):
        self.orig_argv = sys.argv
        self.orig_env = os.environ.copy()

    def tearDown(self):
        sys.argv = self.orig_argv
        os.environ.clear()
        os.environ.update(self.orig_env)

    @patch("shutil.which")
    @patch("subprocess.call")
    def test_console_script_passthrough_install(self, mock_call, mock_which):
        mock_which.return_value = "/usr/local/bin/uvpip"
        mock_call.return_value = 0
        sys.argv = ["uvpip-cli", "install", "flask", "--upgrade"]

        code = main()
        self.assertEqual(code, 0)
        mock_call.assert_called_once_with(["/usr/local/bin/uvpip", "install", "flask", "--upgrade"])

    @patch("shutil.which")
    @patch("subprocess.call")
    def test_console_script_passthrough_uninstall(self, mock_call, mock_which):
        mock_which.return_value = "/usr/local/bin/uvpip"
        mock_call.return_value = 0
        sys.argv = ["uvpip-cli", "uninstall", "-y", "flask"]

        code = main()
        self.assertEqual(code, 0)
        mock_call.assert_called_once_with(["/usr/local/bin/uvpip", "uninstall", "-y", "flask"])

    @patch("shutil.which")
    @patch("subprocess.call")
    def test_console_script_passthrough_freeze(self, mock_call, mock_which):
        mock_which.return_value = "/usr/local/bin/uvpip"
        mock_call.return_value = 0
        sys.argv = ["uvpip-cli", "freeze"]

        code = main()
        self.assertEqual(code, 0)
        mock_call.assert_called_once_with(["/usr/local/bin/uvpip", "freeze"])

    @patch("shutil.which")
    @patch("subprocess.call")
    def test_non_zero_exit_propagation(self, mock_call, mock_which):
        mock_which.return_value = "/usr/local/bin/uvpip"
        mock_call.return_value = 23
        sys.argv = ["uvpip-cli", "install", "nonexistent-pkg"]

        code = main()
        self.assertEqual(code, 23)

    @patch("shutil.which")
    @patch("subprocess.call")
    def test_stderr_passthrough_on_oserror(self, mock_call, mock_which):
        mock_which.side_effect = lambda cmd: "/usr/bin/uv" if cmd == "uv" else None
        mock_call.side_effect = OSError("exec format error")
        sys.argv = ["uvpip-cli", "install", "pytest"]

        stderr_buf = io.StringIO()
        with patch("sys.stderr", stderr_buf):
            code = main()
        self.assertEqual(code, 126)
        self.assertIn("[uvpip] failed to execute uv: exec format error", stderr_buf.getvalue())

    @patch("shutil.which")
    @patch("subprocess.call")
    def test_fallback_uv_translation_rules(self, mock_call, mock_which):
        mock_which.side_effect = lambda cmd: "/opt/uv/bin/uv" if cmd == "uv" else None
        mock_call.return_value = 0

        # install command
        sys.argv = ["uvpip-cli", "install", "scipy"]
        main()
        self.assertEqual(mock_call.call_args[0][0], ["/opt/uv/bin/uv", "pip", "install", "scipy"])

        # uninstall with -y stripped
        sys.argv = ["uvpip-cli", "uninstall", "-y", "scipy"]
        main()
        self.assertEqual(mock_call.call_args[0][0], ["/opt/uv/bin/uv", "pip", "uninstall", "scipy"])

        # freeze command
        sys.argv = ["uvpip-cli", "freeze"]
        main()
        self.assertEqual(mock_call.call_args[0][0], ["/opt/uv/bin/uv", "pip", "freeze"])

    def test_input_validation_patterns_exported(self):
        """Assert input_validation_patterns list is exported and non-empty."""
        from uvpip import input_validation_patterns
        self.assertIsInstance(input_validation_patterns, list)
        self.assertGreaterEqual(len(input_validation_patterns), 2)
        for pattern in input_validation_patterns:
            self.assertTrue(pattern.startswith("^") and pattern.endswith("$"))


if __name__ == "__main__":
    unittest.main()
