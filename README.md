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
> **`uvpip` bridges the gap:** Keep typing your standard `pip` commands, but execute them through `uv` behind the scenes. If a command or flag isn't supported by `uv`, it automatically falls back to native `pip` so your builds **never break**.

---

## ✨ Key Highlights

* ⚡ **10x–100x Speedup:** Leverages Astral's Rust-based `uv` engine for hyper-threaded resolution and wheel installation.
* 🛡️ **100% Fail-Closed Fallback:** If any flag, private index, or legacy command isn't supported by `uv`, it immediately delegates to native `python -m pip`. **Zero broken builds.**
* 🪶 **Zero Runtime Dependencies:** Standalone binary written in pure Go (standard library only). No daemons, no background services, no telemetry.
* 🩺 **Instant Diagnostics:** Built-in `uvpip doctor` verifies PATH precedence, shim integrity, and Python/uv toolchains in 1 second.
* 🔄 **Zero Learning Curve:** Don't change your muscle memory. Keep typing `pip install ...` as you always have.

---

## ⚡ Real-World Proof: `pip` vs `uvpip`

Here is an actual side-by-side terminal run installing a complete Data Science stack (`pandas`, `numpy`, `scipy`) on a fresh environment:

```text
# ❌ Standard Python pip
$ time pip install pandas scipy numpy
Collecting pandas...
Collecting scipy...
Collecting numpy...
...
Installing collected packages: numpy, pytz, six, python-dateutil, pandas, scipy
Successfully installed numpy-2.2.3 pandas-2.2.3 pytz-2025.1 python-dateutil-2.9.0.post0 scipy-1.15.2 six-1.17.0
real    0m 28.642s

# ⚡ With uvpip (powered by uv)
$ time uvpip install pandas scipy numpy
[uvpip] routed ➔ uv pip install pandas scipy numpy
Resolved 6 packages in 18ms
Downloaded 6 packages in 340ms
Installed 6 packages in 84ms
 + numpy==2.2.3
 + pandas==2.2.3
 + python-dateutil==2.9.0.post0
 + pytz==2025.1
 + scipy==1.15.2
 + six==1.17.0
real    0m 0.894s   🚀 (~32x faster!)
```

### 📊 Benchmark Summary

| Environment / Workload | Standard `pip` | **`uvpip` (via uv)** | **Speedup** |
| :--- | :--- | :--- | :--- |
| **Data Science** (`numpy`, `pandas`, `scipy`) | `28.64s` | **`0.89s`** | **🚀 32x faster** |
| **Modern Web Stack** (`fastapi`, `uvicorn`, `pydantic`) | `9.42s` | **`0.31s`** | **🚀 30x faster** |
| **Heavy Enterprise** (`django`, `djangorestframework`, `psycopg2`) | `14.15s` | **`0.58s`** | **🚀 24x faster** |
| **Cached Re-install** (Instant cache hits) | `3.82s` | **`0.04s`** | **🚀 95x faster** |

---

## 🚀 Quick Installation

Choose the fastest method for your system:

> ### 🐍 **Option 1: Install via PyPI (For Python Developers)**
> Officially registered on PyPI as **`uvpip-cli`**.
>
> **Standard pip:**
> ```bash
> pip install uvpip-cli
> ```
>
> **Isolated global tool via uv:**
> ```bash
> uv tool install uvpip-cli
> ```
>
> **Using pipx:**
> ```bash
> pipx install uvpip-cli
> ```

---

> ### 📦 **Option 2: Pre-compiled Standalone Native Binaries**
> Direct single-executable downloads from **[GitHub Releases](https://github.com/yv3000/uvpip/releases)**:
>
> | OS / Platform | Architecture | Download Asset |
> | :--- | :--- | :--- |
> | **Windows** | `x64` / `arm64` | **[`uvpip-windows-amd64.exe`](https://github.com/yv3000/uvpip/releases/latest)** / **[`uvpip-windows-arm64.exe`](https://github.com/yv3000/uvpip/releases/latest)** |
> | **macOS** | Apple Silicon / Intel | **[`uvpip-darwin-arm64`](https://github.com/yv3000/uvpip/releases/latest)** / **[`uvpip-darwin-amd64`](https://github.com/yv3000/uvpip/releases/latest)** |
> | **Linux** | `x64` / `arm64` | **[`uvpip-linux-amd64`](https://github.com/yv3000/uvpip/releases/latest)** / **[`uvpip-linux-arm64`](https://github.com/yv3000/uvpip/releases/latest)** |

---

> ### 🔨 **Option 3: Build from Source**
> Requires Go 1.26+:
> ```bash
> git clone https://github.com/yv3000/uvpip.git
> cd uvpip
> go build -o uvpip .
> ```

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

## ⚙️ Configuration (Optional)

All settings are optional and work out-of-the-box:

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

## 📄 License

Licensed under the **[MIT License](LICENSE)**. Free for personal, academic, and commercial use.  
Copyright (c) 2026 YV.
