---
translation_set_id: install
path: getting-started/install
locale: en
group: getting-started
group_order: 1
order: 2
title: Installing Wave
summary: Install Wave with the official script or release archives, verify the toolchain, or build the compiler from source.
---

## Official install scripts

On Linux and macOS, the official shell installer can install the latest release:

```shell
curl -fsSL https://wave-lang.dev/install.sh | bash -s -- latest
```

On Windows x86-64, run the PowerShell installer:

```powershell
irm https://wave-lang.dev/install.ps1 -OutFile install.ps1
powershell -ExecutionPolicy Bypass -File .\install.ps1 -Latest
```

If an installer cannot update the current shell environment, apply the PATH instructions it prints or start a new shell. The Windows installer uses `%LOCALAPPDATA%\Wave\bin` by default and adds it to the user PATH.

## Verify the installation

```shell
wavec --version
wavec --help
```

Record this output when reporting a documentation mismatch. The pages describe the current compiler contract; older installed binaries may expose a smaller command or syntax surface.

## Release archives

Release asset names include the selected version. Check the current release page for the exact version string and available platforms.

| Platform | Archive pattern |
| --- | --- |
| Linux x86-64 GNU | `wave-<version>-x86_64-linux-gnu.tar.gz` |
| Windows x86-64 GNU | `wave-<version>-x86_64-pc-windows-gnu.zip` |
| macOS Apple Silicon | `wave-<version>-aarch64-apple-darwin.tar.gz` |

A `SHA256SUMS` asset is also published for verifying manually downloaded archives.

## Build from source

Building the compiler repository requires a Rust toolchain.

```shell
git clone https://github.com/wavefnd/Wave.git
cd Wave
cargo build
```

This builds the default branch. To reproduce a published release, check out its tag before running Cargo:

```shell
git checkout <release-tag>
```

For an optimized compiler binary:

```shell
cargo build --release
```

The resulting executable is normally `target/debug/wavec` or `target/release/wavec`.

## Post-install checks

- Confirm the expected version with `wavec --version`.
- Inspect recognized targets with `wavec print supported-targets`.
- Locate the standard library with `wavec print std-path`.
- If the shell cannot find `wavec`, check the installation directory and PATH first.

> **Reviewing remote scripts**
>
> Download and review `install.sh` or `install.ps1` before running it when required by your environment's security policy.
