"""uvpip: ultra-fast transparent pip wrapper powered by uv."""

from __future__ import annotations

import os
import shutil
import subprocess
import sys

__version__ = "1.0.7"


def translate_to_uv(args: list[str]) -> list[str]:
    """Translate pip CLI arguments to uv CLI arguments."""
    if not args:
        return ["pip"]

    cmd = args[0].lower()
    if cmd == "cache":
        return ["cache", *args[1:]]

    if cmd in ("uninstall", "remove"):
        filtered: list[str] = []
        options = True
        for a in args[1:]:
            if a == "--":
                options = False
            if options and a in ("-y", "--yes"):
                continue
            filtered.append(a)
        return ["pip", "uninstall", *filtered]

    if cmd == "upgrade":
        return ["pip", "install", "--upgrade", *args[1:]]

    return ["pip", *args]


def main() -> int:
    """Entry point for the uvpip CLI command."""
    args = sys.argv[1:]

    # Check for native uvpip executable on PATH first
    uvpip_exe = shutil.which("uvpip")
    if uvpip_exe and os.path.abspath(uvpip_exe) != os.path.abspath(sys.argv[0]):
        try:
            return subprocess.call([uvpip_exe, *args])
        except OSError:
            pass

    # Discover uv
    uv_exe = os.getenv("UVPIP_UV") or shutil.which("uv")
    if not uv_exe:
        # Check standard installation locations
        home = os.path.expanduser("~")
        candidates = [
            os.path.join(home, ".local", "bin", "uv"),
            os.path.join(home, ".cargo", "bin", "uv"),
            os.path.join(home, ".local", "bin", "uv.exe"),
            os.path.join(home, ".cargo", "bin", "uv.exe"),
        ]
        for c in candidates:
            if os.path.isfile(c) and os.access(c, os.X_OK):
                uv_exe = c
                break

    if not uv_exe:
        sys.stderr.write(
            "[uvpip] uv not found on PATH or candidate locations; "
            "install uv: https://docs.astral.sh/uv/getting-started/installation/\n"
        )
        return 127

    # Translate arguments
    uv_args = translate_to_uv(args)
    cmd = [uv_exe, *uv_args]

    # Preserve environment, default to UV_SYSTEM_PYTHON=1 if outside venv
    env = os.environ.copy()
    if "VIRTUAL_ENV" not in env and "CONDA_PREFIX" not in env and "UV_SYSTEM_PYTHON" not in env:
        env["UV_SYSTEM_PYTHON"] = "1"

    try:
        return subprocess.call(cmd, env=env)
    except OSError as err:
        sys.stderr.write(f"[uvpip] failed to execute uv: {err}\n")
        return 126


if __name__ == "__main__":
    sys.exit(main())
