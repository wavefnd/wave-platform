---
translation_set_id: install
path: getting-started/install
locale: en
group: getting-started
group_order: 1
order: 2
title: Install Wave
summary: Install Wave on Linux, macOS, or Windows and run your first program.
---

## Installation policy

The installers install only the latest public versioned release, including versioned prereleases. Drafts and Nightly are excluded. Use manual installation below for older versions and Nightly. If the latest release has no package for your platform, installation stops without selecting an older version.

| Operating system | Architectures |
| --- | --- |
| Linux | amd64, arm64, riscv64, loong64 |
| macOS | amd64, arm64 |
| Windows | amd64, arm64 |
| FreeBSD | amd64 |

## Linux, macOS and FreeBSD

Requires Bash, curl, jq, tar and a SHA-256 tool. Install bash, curl and jq first on FreeBSD. Linux requires compatible glibc and system libraries; macOS requires Apple Command Line Tools; FreeBSD requires a compatible base system.

```shell
curl -fsSL https://wave-lang.dev/install.sh | bash
```

## Windows

Run in PowerShell. The installer detects Windows amd64 or arm64 and selects the MSVC package. The Visual C++ runtime, Windows SDK and MSVC/UCRT libraries are required. A Visual Studio developer shell can provide the environment needed for the compilation check.

```powershell
irm https://wave-lang.dev/install.ps1 -OutFile install.ps1
powershell -ExecutionPolicy Bypass -File .\install.ps1
```

## Installation options

By default, install Wave and the latest Vex when a package exists for your platform. If Vex has no package, install Wave only and report that choice. Vex download or verification failures are errors and are never silently skipped.

| | Bash option | PowerShell option |
| --- | --- | --- |
| Require Vex | `--with-vex` | `-WithVex` |
| Install Wave only | `--without-vex` | `-WithoutVex` |
| Do not modify PATH | `--no-modify-path` | `-NoModifyPath` |

Run the same command to update. Keep the previous installation until the new compiler and bundled std pass execution checks; restore it on failure. Default locations are ~/.wave/bin on Unix and %LOCALAPPDATA%\Wave\bin on Windows. Set WAVE_INSTALL_DIR to choose a dedicated installation directory. Open a new terminal after automatic PATH setup.

## Manually install older versions and Nightly

Download the .tar.gz or .zip for the required version, OS and architecture from [GitHub Releases](https://github.com/wavefnd/Wave/releases). Extract it while keeping the relative paths of wavec, std, llvm and all other bundled files. Add the directory containing wavec to PATH. Follow the notes for older releases whose layout or prerequisites differ. Installer options --version, --vex-version, -Version and -VexVersion are not supported.

## First run

The installer compiles and runs a small program using the bundled std. There is no need to download a separate latest std. Run the example below after installation. If Vex was installed, you can also check it with vex --version.

```shell
wavec --version
```

Create a plain-text file named `main.wave` in a directory of your choice and save the complete program below. Make sure the filename is not `main.wave.txt`.

<!-- wave-example: install-stdlib -->
```wave
import("std::string::len")::{
    len
};

fun main() {
    println("Wave: {} bytes", len("Wave"));
}
```

Open a terminal in the directory containing `main.wave` (or use `cd` to navigate there), then run:

```shell
wavec run main.wave
```

Expected output:

```text
Wave: 4 bytes
```
