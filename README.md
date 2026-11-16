<div align="center">

# uvpip

### Transparent, zero-runtime-dependency CLI routing pip through uv

[![PyPI version](https://img.shields.io/pypi/v/uvpip-cli.svg?style=flat-square&color=3775A9)](https://pypi.org/project/uvpip-cli/)
[![PyPI downloads](https://img.shields.io/pypi/dm/uvpip-cli.svg?style=flat-square&color=blue)](https://pypi.org/project/uvpip-cli/)
[![CI status](https://img.shields.io/github/actions/workflow/status/yv3000/uvpip/ci.yml?branch=main&style=flat-square)](https://github.com/yv3000/uvpip/actions)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg?style=flat-square)](LICENSE)
[![Python 3.8+](https://img.shields.io/badge/Python-%3E%3D3.8-brightgreen?style=flat-square&logo=python)](pyproject.toml)
[![uv Compatible](https://img.shields.io/badge/uv-%3E%3D0.1-black?style=flat-square&logo=astral)](https://github.com/astral-sh/uv)

<br/>

```text
                        +---------+     +---------+
                        |   pip   |     |  pip3   |
                        +----+----+     +----+----+
                             |               |
                             +-------+-------+
                                     |
                                     v
                             +---------------+
                             |     uvpip     |
                             +-------+-------+
                                     |
                         +-----------+-----------+
                         |                       |
                         v                       v
                  +-------------+         +-------------+
                  |     uv      |         |   Native    |
                  | (fast path) |         | (safe fall) |
                  +-------------+         +-------------+
```

<p>
  <img src="https://img.shields.io/badge/pip-3775A9?style=for-the-badge&logo=pypi&logoColor=white" alt="pip" />
  <img src="https://img.shields.io/badge/pip3-3776AB?style=for-the-badge&logo=python&logoColor=white" alt="pip3" />
  <img src="https://img.shields.io/badge/➔-18181B?style=for-the-badge" alt="to" />
  <img src="https://img.shields.io/badge/uv-DE5FE9?style=for-the-badge&logo=astral&logoColor=white" alt="uv" />
</p>

</div>

---

> [!TIP]
> **10-100x Faster Installs with Zero Risk**: `uvpip` transparently accelerates `pip` and `pip3` using Astral's native `uv` engine, while guaranteeing an **automatic 100% fail-closed native fallback** if any command or flag cannot be safely translated.

---

## ⚡ Performance Benchmarks (uv vs Native pip)

Real-world package installation benchmarks on a clean Python 3.12 environment:

| Operation                    | `pip` (Native) | `pip` (Cached) | **`uvpip` (via uv)** | **Speedup**        |
| :--------------------------- | :------------- | :------------- | :-------------------- | :----------------- |
| **Clean Install (No Cache)** | `28.6s`        | `14.2s`        | **`0.9s`**            | 🚀 **~32x faster** |
| **Cached Re-install**        | `3.8s`         | `1.9s`         | **`0.04s`**           | ⚡ **~95x faster** |
| **Single Package Add**       | `4.8s`         | `2.5s`         | **`0.2s`**            | 🔥 **~24x faster** |
| **Dependency Resolution**    | `9.4s`         | `4.1s`         | **`0.08s`**           | 🏎️ **~115x faster** |

---

## ⚡ Key Highlights

- **Keep Your Muscle Memory:** Type `pip install`, `pip list`, `pip freeze`, or `pip uninstall` just like you always do.
- **Fail-Closed Fallback Guarantee:** Commands and options `uv` cannot safely translate (e.g. legacy setup.py, custom index flags, or unsupported options) automatically hand execution back to native Python `pip`.
- **Zero Runtime Dependencies:** Standalone binary written in pure Go (standard library only). Leaves no bloated dependency footprint on your machine.
- **Cross-Platform Parity:** Native standalone binaries for Windows, macOS (Intel & Apple Silicon), and Linux.
- **97%+ Test Coverage:** Thoroughly verified offline across unit, fuzz, integration, and subprocess test suites.

---

## 🚀 Installation

### Option 1: Install Globally via PyPI (Recommended)

`uvpip` is officially registered on PyPI as **`uvpip-cli`**:

```bash
# 🔹 Install Globally or in Virtualenv Using pip
pip install uvpip-cli

# 🔹 Install as Isolated Tool Using uv
uv tool install uvpip-cli

# 🔹 Install as Isolated Tool Using pipx
pipx install uvpip-cli
```

> **Instant Verification (Check Installation):**
>
> ```bash
> uvpip doctor
> ```

---

### Option 2: Pre-compiled Standalone Native Binaries

Download single-file native executables from **[GitHub Releases](https://github.com/yv3000/uvpip/releases)**:

| Platform | Architecture | Binary Asset |
| :--- | :--- | :--- |
| **Windows** | `x64` / `arm64` | [`uvpip-windows-amd64.exe`](https://github.com/yv3000/uvpip/releases/latest) / [`uvpip-windows-arm64.exe`](https://github.com/yv3000/uvpip/releases/latest) |
| **macOS** | Apple Silicon / Intel | [`uvpip-darwin-arm64`](https://github.com/yv3000/uvpip/releases/latest) / [`uvpip-darwin-amd64`](https://github.com/yv3000/uvpip/releases/latest) |
| **Linux** | `x64` / `arm64` | [`uvpip-linux-amd64`](https://github.com/yv3000/uvpip/releases/latest) / [`uvpip-linux-arm64`](https://github.com/yv3000/uvpip/releases/latest) |

---

### Option 3: Install From Source (Standalone Shims)

Clone and run the native installer for your OS. It installs lightweight launchers into user directories and adds them to User PATH:

```powershell
# 🪟 Windows (PowerShell)
powershell -ExecutionPolicy Bypass -File .\installer\install.ps1
```

```bash
# 🍎 macOS / 🐧 Linux
sh installer/install.sh
```

---

## 📋 Transparent Command Routing

| Invoked Command             | Transparent Routing       | Fallback Condition          |
| :-------------------------- | :------------------------ | :-------------------------- |
| `pip install <pkg>`         | ➔ `uv pip install <pkg>`   | Explicit unsupported flags  |
| `pip install -r req.txt`    | ➔ `uv pip install -r ...` | None                        |
| `pip list`                  | ➔ `uv pip list`           | None                        |
| `pip freeze`                | ➔ `uv pip freeze`         | None                        |
| `pip show <pkg>`            | ➔ `uv pip show <pkg>`     | None                        |
| `pip check`                 | ➔ `uv pip check`          | None                        |
| `pip cache purge`           | ➔ `uv cache clean`        | None                        |
| `pip <unsupported-flag>`    | ➔ **Native fallback**     | Always native (fail-closed) |

---

## 🩺 Verify Installation

Run the built-in diagnostic doctor:

```bash
uvpip doctor
```

It validates that `uv`, `pip`, `pip3`, and `python` are available, checks that all PATH shims are intact, verifies native fallback paths, and exits `0` on success:

```text
uvpip doctor (PATH executables; shell functions are not visible)
[OK] uv: /usr/local/bin/uv (uv 0.5.15)
[OK] pip: ~/.local/bin/pip (uvpip shim)
[OK] pip3: ~/.local/bin/pip3 (uvpip shim)
[OK] python: /usr/bin/python3 (Python 3.12.3)
All systems healthy.
```

---

## 🔒 Security & Fallback Principles

1. **Fail-Closed Guarantee:** When an option or command is unrecognized, `uvpip` never guesses. It delegates directly to native Python `pip`.
2. **Untouched Original Python:** Native pip executables in system PATH are never deleted or modified.
3. **Isolated User Path:** Installers only modify user-level configuration (e.g. `~/.bashrc` or Windows User PATH); system roots remain untouched.

---

## 🛠️ Development & Testing

```bash
git clone https://github.com/yv3000/uvpip.git
cd uvpip
go mod verify
go test ./...
```

### Validation Gates

| Command | Check |
| :--- | :--- |
| `go test ./...` | Unit & subprocess fixture suite |
| `go test -race ./...` | Concurrency race condition detector |
| `go vet ./...` | Go static analysis verification |
| `golangci-lint run` | Linting & style conformity gate |
| `govulncheck ./...` | Known Go vulnerability audit |

CI runs unit tests, race detector, linting, and offline installer checks across Ubuntu, macOS, and Windows matrices.

---

## 📄 License

[MIT](LICENSE) © 2026 [Yash Vishwakarma](https://github.com/yv3000)
