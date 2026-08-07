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

## Official install script

On Unix-like systems, the official installer can install the latest release:

```shell
curl -fsSL https://wave-lang.dev/install.sh | bash -s -- latest
```

If the installer cannot update the current shell environment, apply the PATH instructions it prints or start a new shell.

## Verify the installation

```shell
wavec --version
wavec --help
```

To follow this documentation exactly, verify that the installed version is `v0.2.0-pre-beta`.

## v0.2.0-pre-beta binary archives

The release publishes the following prebuilt archives:

| Platform | Release archive |
| --- | --- |
| Linux x86-64 GNU | `wave-v0.2.0-pre-beta-x86_64-linux-gnu.tar.gz` |
| Windows x86-64 GNU | `wave-v0.2.0-pre-beta-x86_64-pc-windows-gnu.zip` |
| macOS Apple Silicon | `wave-v0.2.0-pre-beta-aarch64-apple-darwin.tar.gz` |

A `SHA256SUMS` asset is also published for verifying manually downloaded archives.

## Build from source

Building the compiler repository requires a Rust toolchain.

```shell
git clone https://github.com/wavefnd/Wave.git
cd Wave
git checkout v0.2.0-pre-beta
cargo build
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
> If your environment does not permit piping a remote script directly into a shell, download and review the installer before running it.
