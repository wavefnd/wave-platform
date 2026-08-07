---
translation_set_id: ecosystem
path: toolchain/ecosystem
locale: en
group: toolchain
group_order: 4
order: 1
title: Wave toolchain
summary: Separate the roles of wavec, the standard library, Vex, the reserved Whale path, and native interoperability.
---

## wavec

`wavec` is the central compiler CLI in v0.2.0-pre-beta. It checks, compiles, links, and runs Wave sources and exposes a `print` interface for querying target and capability information.

```shell
wavec --version
wavec build main.wave
wavec run main.wave
wavec print supported-targets
```

## Standard library

`std` ships as Wave source in the compiler repository and provides runtime functionality for strings, memory, files, networking, processes, and more.

```shell
wavec print std-path
```

Use the printed directory to inspect the exact `.wave` APIs used by the installed compiler.

## Vex

Vex is a separate package-management project for Wave dependencies. `wavec` provides stable integration points such as `--dep-root` and `--dep name=path` for resolving external package imports.

This allows a package manager to prepare a dependency tree while the compiler receives the final resolution paths.

## Whale

`wavec` still recognizes `--whale`, but in v0.2.0-pre-beta it is reserved and does not select an implemented backend. Do not use it for normal builds.

## Native interoperability

Wave can connect to C/C++ and platform facilities through `extern(c)`, `export(c)`, native linker options, and inline assembly. Capabilities not yet covered by the standard library can be exposed behind small native boundaries.

## Capability discovery for tools

Query the compiler instead of hardcoding target lists:

```shell
wavec print supported-targets
wavec print supported-input-types
wavec print supported-emit-kinds
wavec print supported-print-items
```

Tooling that needs structured data can use `wavec print ... --format=json`.
