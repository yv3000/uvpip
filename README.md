<p align="center">
  <a href="https://pypi.org/project/pip/"><img src="https://img.shields.io/badge/pip-3776AB?style=for-the-badge&logo=pypi&logoColor=white" alt="pip" /></a>
  <a href="https://github.com/astral-sh/uv"><img src="https://img.shields.io/badge/powered%20by-uv-de5fe9?style=for-the-badge&logo=rust&logoColor=white" alt="uv" /></a>
  <a href="#install"><img src="https://img.shields.io/badge/platform-Windows%20%7C%20macOS%20%7C%20Linux-0078D6?style=for-the-badge" alt="Cross-platform" /></a>
  <a href="#license"><img src="https://img.shields.io/badge/license-MIT-green?style=for-the-badge" alt="MIT License" /></a>
</p>

<h1 align="center">uvpip</h1>
<p align="center"><b>pip at the speed of uv.</b></p>
<p align="center">Type <code>pip</code>, run <b>uv</b>. Completely transparent. 10-100x faster installs.</p>

---

## What is uvpip?

**uvpip** is a transparent drop-in wrapper that intercepts all `pip` and `pip3` commands and silently runs them through [uv](https://github.com/astral-sh/uv) — the blazing-fast Python package manager written in Rust by Astral (the team behind Ruff).

- You keep typing `pip install requests`, `pip uninstall X`, `pip freeze` — same commands, same flags, same muscle memory.
- Under the hood, uv does the heavy lifting at 10-100x the speed.
- Your original pip is **never touched or modified**. uvpip uses a PATH priority trick — plus a shell function for venv compatibility — to intercept calls first.

> **Think of it as a turbocharger for pip.** You don't change how you drive — the car just goes faster.

---

## Install

### Windows

Open **PowerShell** and run:

```powershell
iex (irm https://raw.githubusercontent.com/yv3000/uvpip/main/installer/install.ps1)
```

### macOS / Linux

```bash
curl -fsSL https://raw.githubusercontent.com/yv3000/uvpip/main/installer/install.sh | sh
```

That's it. **One command. No other steps.**

The installer will:
1. Detect your OS and architecture
2. Auto-install uv if not present
3. Download the right uvpip binary
4. Create `pip`/`pip3` shims and shell functions
5. Add itself to your PATH and shell profile
6. Verify everything works

Restart your terminal, then:

```bash
pip install requests   # now running through uv
uvpip doctor            # verify everything is working
```

---

## Uninstall

### Windows

```powershell
iex (irm https://raw.githubusercontent.com/yv3000/uvpip/main/uninstaller/uninstall.ps1)
```

### macOS / Linux

```bash
curl -fsSL https://raw.githubusercontent.com/yv3000/uvpip/main/uninstaller/uninstall.sh | sh
```

This removes the `~/.uvpip` folder, cleans your PATH, and removes the shell function from your profile — entirely. Original pip is instantly restored. uv stays installed (it's not removed, since you may want to keep using it directly).

---

## Proof it works

```bash
pip --version              # uvpip v1.0.0 (wrapper active, not real pip)
pip install requests       # installed in ~0.5-2s instead of ~8-15s
pip install -r requirements.txt   # same flags, same files, just faster
pip uninstall requests     # clean uninstall, no confirmation prompt needed
uvpip doctor                # shows uv version, shim status, PATH order
```

---

## Speed Comparison

```
Before uvpip                          After uvpip
--------------                        ---------------
pip install requests   (~10s)  -->    pip install requests   (~1s)
numpy + pandas          (~45s) -->    numpy + pandas          (~3s)
Cache hit (reinstall)    (~5s)  -->    Cache hit (reinstall)   (<0.5s)
```

Same commands. Same output. Just faster.

---

## Command Mapping

### Commands routed through uv

| pip command | uv equivalent | Notes |
|---|---|---|
| `pip install X` | `uv pip install X` | Install a package |
| `pip install -r requirements.txt` | `uv pip install -r requirements.txt` | Install from file |
| `pip install -e .` | `uv pip install -e .` | Editable install |
| `pip install --upgrade X` | `uv pip install --upgrade X` | Upgrade a package |
| `pip uninstall X` | `uv pip uninstall X` | Uninstall — `-y`/`--yes` is stripped automatically since uv never prompts |
| `pip list` | `uv pip list` | List installed packages |
| `pip show X` | `uv pip show X` | Show package info |
| `pip freeze` | `uv pip freeze` | Freeze installed versions |
| `pip check` | `uv pip check` | Verify dependency compatibility |
| `pip cache dir` / `pip cache list` | `uv cache dir` / `uv cache list` | Routed to uv's top-level cache command, not `uv pip` |
| `pip3 install X` | `uv pip install X` | `pip3` is an alias for the same binary |

### Commands that fall through to uv as-is

Any pip command not explicitly mapped above is passed straight through as `uv pip <command> <args>` — uv will report its own error if that exact command/flag doesn't exist on its side, rather than uvpip pretending to support something it doesn't.

---

## How It Works

### Architecture

```
~/.uvpip/
└── bin/
    ├── uvpip          # the Go binary — does the actual translation + execution
    ├── pip / pip.cmd   # shim — calls uvpip
    └── pip3 / pip3.cmd # shim — calls uvpip
```

### The Flow

```
You type:  pip install requests
               |
               v
   pip shim / pip shell function (found before real pip)
               |
               v
        uvpip binary (Go)
               |
               v
   translator: pip install requests -> uv pip install requests
               |
               v
   uv pip install requests   <- actually runs, stdio fully inherited
               |
               v
   You see: Resolved 5 packages... Installed 5 packages   (in ~1s instead of ~10s)
```

No output reformatting. No interception of uv's own output — it goes straight to your terminal, including progress bars and interactive prompts.

### PATH Hijack + Shell Function (Safe, and venv-aware)

uvpip works two ways at once:

1. **PATH priority** — `~/.uvpip/bin/` is prepended to both User and System PATH. `pip`/`pip3` shims here are found before the real pip.
2. **Shell function** — a `pip`/`pip3` function is added to your shell profile (`.bashrc`/`.zshrc`/fish config, or PowerShell's `$PROFILE`). Shells resolve functions *before* searching PATH — and since activating a Python venv only modifies PATH (never shell functions), the function keeps working even inside an activated venv, where a plain PATH-based shim alone would lose to the venv's own pip.

The original pip is never modified, deleted, or renamed — both mechanisms sit *in front of* it, never replacing it.

---

## Working inside a virtual environment

uvpip installs a shell function (not just a PATH entry) for `pip` and `pip3`. Shell functions are resolved before PATH on every common shell, and venv activation only modifies PATH — so `pip install X` continues to route through uv even inside an activated venv, in any interactive terminal session.

```bash
python3 -m venv .venv
source .venv/bin/activate
pip install requests     # still routes through uv — fast, even inside the venv
```

The one exception: non-interactive contexts that don't load your shell profile (some CI runners, certain subprocess/exec calls that bypass an interactive shell) fall back to the plain PATH-based shim, which may resolve to the venv's own pip if a venv is active in that specific non-interactive context. Normal terminal usage, including inside venvs, is fully covered.

---

## Fallback Behavior

| Scenario | What happens |
|---|---|
| uv not found at install time | Auto-installed silently via the official uv installer |
| uv not found at runtime | uvpip attempts auto-install, then retries once |
| Unknown/unmapped pip command | Passed through to `uv pip <command>` as-is; uv reports its own error if unsupported |
| Non-interactive shell (no profile loaded) | Falls back to the PATH-based shim only |

---

## Shell Compatibility

| Shell | Status |
|---|---|
| bash | ✅ Fully supported (function + PATH) |
| zsh | ✅ Fully supported (function + PATH) |
| fish | ✅ Fully supported (function + PATH) |
| PowerShell 5.1 / 7+ | ✅ Fully supported (function + PATH) |
| CMD | ✅ Supported (PATH shim only — CMD has no function concept) |
| Git Bash | ✅ Supported (PATH shim) |

---

## Requirements

| Requirement | Details |
|---|---|
| **OS** | Windows 10/11, macOS (Intel + Apple Silicon), Linux (amd64 + arm64) |
| **Python** | Any version with pip or venv support |
| **uv** | Auto-installed if missing |
| **Admin** | Only for the optional System PATH step on Windows |

---

## Platform Support

| Platform | Status |
|---|---|
| Windows (amd64, arm64) | ✅ Supported |
| macOS (Intel, Apple Silicon) | ✅ Supported |
| Linux (amd64, arm64) | ✅ Supported |

---

## FAQ

**Will this break my existing projects?**
No. Your `requirements.txt`, virtual environments, and project structure all work the same way.

**Does this affect `python` or other commands?**
No. Only `pip` and `pip3` are intercepted.

**Can I temporarily bypass uvpip?**
Yes — call the real pip directly with its full path, or use `python -m pip` (which is never intercepted):
```bash
python -m pip install requests
```

**How do I update uvpip?**
Re-run the install command — it detects and skips a clean re-install if already present; uninstall first if you want a fresh install.

**Why does `pip uninstall X` not ask for confirmation anymore?**
uv never prompts for uninstall confirmation in the first place — uvpip strips any `-y`/`--yes` flag automatically before handing off, since uv would otherwise reject it as an unrecognized argument.

---

## Tech Stack

- Go — single static binary per platform, zero runtime dependencies
- PowerShell installer (Windows) / POSIX shell installer (macOS, Linux) — no external dependencies
- `stdio` fully inherited on every uv call — interactive prompts, progress bars, and confirmations all work normally
- No output reformatting — what you see is exactly what uv printed

---

## Changelog

- **v1.0.0** — Initial cross-platform release: full pip → uv command mapping, PATH hijack + shell-function install/uninstall, venv-aware interception, auto-install of uv and detection of pip/Python

---

## License <a name="license"></a>

MIT License

Copyright (c) 2026 YV

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.

---

<p align="center">
  <sub>YV 🖤 ~ I EXPECT NOTHING FROM YOU...</sub>
</p>
