---
translation_set_id: install
path: getting-started/install
locale: en
group: getting-started
group_order: 1
order: 2
title: Install Wave
summary: Install the released compiler and verify the toolchain.
---

## Official installer

```shell
curl -fsSL https://wave-lang.dev/install.sh | bash -s -- latest
```

Open a new shell after installation, or apply the PATH instruction printed by the installer.

## Verify

```shell
wavec --version
wavec --help
```

> **Security**
> 
> Review an installation script before piping it to a shell when required by your environment.

## Build from source

Use the build instructions shipped with the exact Wave source revision you intend to compile. Compiler dependencies and supported targets can change between releases.

