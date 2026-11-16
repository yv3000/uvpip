<div align="center">

# uvpip

### ⚡ Type `pip`, run [`uv`](https://github.com/astral-sh/uv). 10-100x faster package installs with zero risk.

[![PyPI version](https://img.shields.io/pypi/v/uvpip-cli.svg?style=flat-square&color=3775A9&logo=pypi&logoColor=white)](https://pypi.org/project/uvpip-cli/)
[![PyPI downloads](https://img.shields.io/pypi/dm/uvpip-cli.svg?style=flat-square&color=blue)](https://pypi.org/project/uvpip-cli/)
[![CI status](https://img.shields.io/github/actions/workflow/status/yv3000/uvpip/ci.yml?branch=main&style=flat-square)](https://github.com/yv3000/uvpip/actions)
[![Coverage](https://img.shields.io/badge/Coverage-97%25-brightgreen.svg?style=flat-square)](https://github.com/yv3000/uvpip/actions)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg?style=flat-square)](LICENSE)
[![Python 3.8+](https://img.shields.io/badge/Python-%3E%3D3.8-3776AB?style=flat-square&logo=python&logoColor=white)](pyproject.toml)
[![Go 1.26+](https://img.shields.io/badge/Go-%3E%3D1.26-00ADD8?style=flat-square&logo=go&logoColor=white)](go.mod)
[![webinstall.dev](https://img.shields.io/badge/webi-uvpip-blueviolet?style=flat-square)](https://webinstall.dev/uvpip)

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
                  |     uv      |         | Native pip  |
                  | (fast path) |         | (safe fall) |
                  +-------------+         +-------------+
```

<p>
  <img src="https://img.shields.io/badge/pip-3775A9?style=for-the-badge&logo=pypi&logoColor=white" alt="pip" />
  <img src="https://img.shields.io/badge/pip3-3776AB?style=for-the-badge&logo=python&logoColor=white" alt="pip3" />
  <img src="https://img.shields.io/badge/➔-18181B?style=for-the-badge" alt="to" />
  <img src="https://img.shields.io/badge/uv-DE5FE9?style=for-the-badge&logo=astral&logoColor=white" alt="uv" />
  <img src="https://img.shields.io/badge/PyPI-3775A9?style=for-the-badge&logo=pypi&logoColor=white" alt="PyPI" />
  <img src="https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Go" />
</p>

</div>

---

> [!TIP]
> **Why `uvpip`?** `pip` is universally loved, but dependency resolution on large projects can be painfully slow. Astral's [`uv`](https://github.com/astral-sh/uv) is 10-100x faster, but requires changing your habits and scripts. 
> 
> **`uvpip` bridges the gap:** Keep typing your standard `pip` commands, but execute them through `uv` behind the scenes. If a command or flag isn't supported by `uv`, it automatically falls back to native `pip` so your builds never break.

---

## ⚡ Performance Comparison

Real-world benchmark installing common Python environments from scratch (clean cache):

| Workload | Standard `pip` | **`uvpip` (via uv)** | Speedup |
| :--- | :--- | :--- | :--- |
| **FastAPI + Uvicorn + Pydantic** | ~9.4s | **~0.32s** | **29x faster** |
| **Data Science (NumPy, Pandas, SciPy)** | ~28.6s | **~1.15s** | **25x faster** |
| **Django + REST Framework + Psycopg2** | ~14.1s | **~0.58s** | **24x faster** |
| **Cached Re-install (All packages)** | ~3.8s | **~0.04s** | **95x faster** |

---

## 🚀 Installation

Choose the installation method that fits your workflow:

### 1. Via Webinstall (Fastest One-Liner — Recommended)
Installs the standalone native binary to `~/.local/bin/` with automatic `PATH` setup. No Python, compiler, or sudo required.

* **macOS / Linux:**
  ```bash
  curl -sS https://webi.sh/uvpip | bash
  ```
* **Windows (PowerShell):**
  ```powershell
  curl.exe https://webi.sh/uvpip | powershell
  # or
  powershell -c "irm https://webi.sh/uvpip | iex"
  ```

### 2. Via PyPI (For Python Environments)
Available on **[PyPI as `uvpip-cli`](https://pypi.org/project/uvpip-cli/)**:

```bash
# Standard pip
pip install uvpip-cli

# Or isolated tool with uv
uv tool install uvpip-cli

# Or with pipx
pipx install uvpip-cli
```

### 3. Pre-compiled Standalone Binaries
Download single-file native executables from **[GitHub Releases](https://github.com/yv3000/uvpip/releases)**:

| Platform | Architecture | Binary Asset |
| :--- | :--- | :--- |
| **Windows** | `x64` / `arm64` | [`uvpip-windows-amd64.exe`](https://github.com/yv3000/uvpip/releases/latest) / [`uvpip-windows-arm64.exe`](https://github.com/yv3000/uvpip/releases/latest) |
| **macOS** | Apple Silicon / Intel | [`uvpip-darwin-arm64`](https://github.com/yv3000/uvpip/releases/latest) / [`uvpip-darwin-amd64`](https://github.com/yv3000/uvpip/releases/latest) |
| **Linux** | `x64` / `arm64` | [`uvpip-linux-amd64`](https://github.com/yv3000/uvpip/releases/latest) / [`uvpip-linux-arm64`](https://github.com/yv3000/uvpip/releases/latest) |

### 4. Build from Source
Requires Go 1.26+:
```bash
git clone https://github.com/yv3000/uvpip.git
cd uvpip
go build -o uvpip .
```

---

## 🛠️ Usage

Use `uvpip` exactly like you use `pip`:

```bash
# Install packages at lightning speed
uvpip install fastapi uvicorn

# Install from requirements file
uvpip install -r requirements.txt

# Inspect and manage installed packages
uvpip list
uvpip freeze
uvpip show pydantic

# Uninstall packages
uvpip uninstall requests

# Clear package cache
uvpip cache purge
```

### 🩺 Health Check & Environment Verification

Run `uvpip doctor` anytime to verify your toolchain, PATH precedence, and runtime availability:

```bash
uvpip doctor
```

**Sample Output:**
```text
uvpip doctor (PATH executables; shell functions are not visible)
[OK] uv: /usr/local/bin/uv (uv 0.5.15)
[OK] pip: ~/.local/bin/pip (uvpip shim)
[OK] pip3: ~/.local/bin/pip3 (uvpip shim)
[OK] python: /usr/bin/python3 (Python 3.12.3)
All systems healthy.
```

---

## 🛡️ Fail-Closed Safety Guarantee

You never have to worry about compatibility:
1. **Direct UV Mapping:** Commands like `install`, `uninstall`, `list`, `freeze`, `show`, `check`, and `cache` automatically route to `uv pip`.
2. **Safe Fallback:** Any command, legacy flag, or unsupported option is automatically handed directly to native `python -m pip`.
3. **Zero Runtime Dependencies:** Built as a single self-contained binary using only the Go standard library — no background services, no daemons, and no external network calls.

---

## ⚙️ Configuration (Optional)

All settings are optional and work out-of-the-box. Configure via environment variables when needed:

| Variable | Default | Purpose |
| :--- | :--- | :--- |
| `UVPIP_UV` | *(auto-detected)* | Custom absolute path to the `uv` executable. |
| `UVPIP_DEBUG` | `0` | Set to `1` to output detailed execution and routing diagnostics to `stderr`. |
| `UV_SYSTEM_PYTHON` | `1` outside venv | Allows uv to manage the system environment when running outside a virtual environment. |

---

## 🧹 Uninstallation

To remove `uvpip` and restore default shell shims cleanly:

* **Windows (PowerShell):**
  ```powershell
  & .\uninstaller\uninstall.ps1
  ```
* **macOS / Linux:**
  ```bash
  sh uninstaller/uninstall.sh
  ```

---

## 📚 Technical Documentation

For developers, contributors, and deep-dive architectural references:

* **[ARCHITECTURE.md](ARCHITECTURE.md)** — Internal design, routing pipeline, and error handling model.
* **[CONTRIBUTING.md](CONTRIBUTING.md)** — Development setup, verification test suites, and PR checklist.
* **[SECURITY.md](SECURITY.md)** — Security policies, trust boundaries, and vulnerability disclosure.
* **[CHANGELOG.md](CHANGELOG.md)** — Version release notes and migration history.

> [!NOTE]
> **Looking for the Node.js / JavaScript equivalent?**  
> Check out **[`bunpm`](https://github.com/yv3000/bunpm)** — our companion tool that transparently routes `npm`, `yarn`, and `pnpm` through Bun for 20x faster package installs:
> ```bash
> npm install -g bunpm-cli
> ```

---

## 📄 License

Licensed under the **[MIT License](LICENSE)**. Free for personal, academic, and commercial use.  
Copyright (c) 2026 YV.
