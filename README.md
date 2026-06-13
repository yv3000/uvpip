# uvpip

**Transparent `pip` → `uv` wrapper. You type `pip`. `uv` runs. You win.**

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](#license)
[![Platform](https://img.shields.io/badge/platform-Windows%20%7C%20macOS%20%7C%20Linux-blue)](#install)
[![Powered by uv](https://img.shields.io/badge/powered%20by-uv-orange)](https://github.com/astral-sh/uv)

---

## What is this?

`uvpip` intercepts your `pip` and `pip3` commands and silently runs them through [uv](https://github.com/astral-sh/uv) — a blazing-fast Python package manager written in Rust.

| | pip | uvpip (via uv) |
|---|---|---|
| `pip install requests` | ~8–15s | ~0.5–2s |
| `pip install numpy pandas` | ~30–60s | ~2–5s |
| Cache hit (reinstall) | ~5s | <0.5s |

Same commands. Same output. 10–100x faster.

---

## Install

### Windows

```powershell
irm https://raw.githubusercontent.com/yv3000/uvpip/main/installer/install.ps1 | iex
```

### macOS / Linux

```bash
curl -fsSL https://raw.githubusercontent.com/yv3000/uvpip/main/installer/install.sh | sh
```

Restart your terminal after install, then:

```bash
pip install requests   # now running through uv
uvpip doctor           # verify everything is working
```

---

## How it works

```
You type:  pip install requests
               ↓
        pip.cmd / pip shim
               ↓
        uvpip binary (Go)
               ↓
        translator: pip install → uv pip install
               ↓
        uv pip install requests  ← actually runs
               ↓
        output → your terminal (stdio: inherit)
```

No output reformatting. No interception. `uv`'s actual output goes straight to your terminal.

---

## Supported commands

Everything under `uv pip` works:

```bash
pip install X
pip install -r requirements.txt
pip install -e .              # editable installs
pip install X==1.2.3          # pinned versions
pip uninstall X
pip list
pip show X
pip freeze
pip freeze > requirements.txt
pip check
pip download X
pip install --upgrade X
pip3 install X                # pip3 also works
```

---

## Uninstall

### Windows
```powershell
irm https://raw.githubusercontent.com/yv3000/uvpip/main/uninstaller/uninstall.ps1 | iex
```

### macOS / Linux
```bash
curl -fsSL https://raw.githubusercontent.com/yv3000/uvpip/main/uninstaller/uninstall.sh | sh
```

`uv` itself is **not** removed. Only the uvpip shims and PATH entries are cleaned up.

---

## Diagnostics

```bash
uvpip doctor
```

Shows: uvpip binary path, uv version, pip shim status, PATH order, Python version.

---

## Known limitations

- `python -m pip install X` — not intercepted (direct Python call bypasses the shim)
- Tools that hardcode `/usr/local/bin/pip` — bypass the shim
- `pip publish` / `pip index` — some registry commands may behave differently under uv

---

## License

MIT — by **YV** 🖤

> I EXPECT NOTHING FROM YOU...

---

## Dev / Cross-Compilation

```bash
# Windows amd64
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o dist/uvpip-windows-amd64.exe .

# macOS Intel
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o dist/uvpip-darwin-amd64 .

# macOS Apple Silicon
GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o dist/uvpip-darwin-arm64 .

# Linux amd64
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o dist/uvpip-linux-amd64 .

# Linux arm64
GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o dist/uvpip-linux-arm64 .
```
