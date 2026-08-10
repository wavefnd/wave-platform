---
translation_set_id: vex-package-manager
path: toolchain/vex-package-manager
locale: en
group: toolchain
group_order: 4
order: 3
title: Vex package manager
summary: Manifest-based Wave projects, Git and path dependencies, lockfiles, offline builds, and the wavec boundary.
---

## Role

Vex is the package manager and build tool for Wave. It sits above `wavec`: Vex owns project structure and dependency resolution, while `wavec` owns compiler flags and the compilation pipeline.

Vex commands are manifest-based. Raw `wavec` flags are intentionally not accepted by `vex build`, `vex check`, or `vex run`.

## Create a package

```shell
vex init
vex init --lib
```

An application uses `src/main.wave`; a library uses `src/lib.wave`. The package root contains:

```text
my_project/
├── src/
│   └── main.wave
├── vex.ws
├── vex.lock
└── .vex/
    └── deps/
```

`vex.ws` is the manifest. Vex does not use a `.wson` manifest extension.

```wson
{
    name = "my_project",
    version = 0.1.0,
    lib = false,
    description = "my_project Project",
    author = "unknown",
    license = "Unknown",
    dependencies = []
}
```

## Build commands

```shell
vex build [--target <triple>] [--release] [--dry-run] [--locked] [--offline]
vex check [--target <triple>] [--release] [--dry-run] [--locked] [--offline]
vex run   [--target <triple>] [--release] [--dry-run] [--locked] [--offline] [-- <args...>]
```

Vex keeps this surface small. Use `wavec` directly when you need compiler-specific emit, linker, CPU, ABI, or debug controls. Set `VEX_WAVEC=/path/to/wavec` when Vex must use a specific compiler binary.

Progress stages such as `Resolving`, `Fetching`, `Compiling`, `Checking`, `Running`, and `Finished` are written to stderr. Program output remains on stdout.

## Git-first dependencies

Vex currently needs no central registry. A dependency uses exactly one of a local `path` or a Git URL.

```wson
{
    name = "app",
    version = 0.1.0,
    dependencies = [
        { name = "local_math", path = "../local_math" },
        { name = "remote_math", git = "https://github.com/example/math.git", tag = "v0.1.0" }
    ]
}
```

A Git dependency may select at most one of `branch`, `tag`, or `rev`. Every dependency root must contain its own `vex.ws`. Vex resolves dependency manifests recursively, rejects conflicting package identities, and stores managed Git checkouts under `.vex/deps/<name>`.

## Lockfile contract

The schema-v2 `vex.lock` records the complete transitive graph and exact Git commits. Commit it with the manifest. Given the same manifest and valid lockfile, Vex selects the same dependency graph rather than following a branch or tag again.

Commands that need dependencies resolve them automatically, or you can prepare the graph explicitly:

```shell
vex fetch
vex update
vex update math shared_core
```

`vex update` refreshes all Git packages, or only the named packages and their affected transitive graph. Unrelated locked packages retain their selected commits.

## Locked and offline workflows

`--locked` forbids creating or changing `vex.lock`. It fails when the file is missing, uses an unsupported schema, or no longer matches the manifest graph. It may still fetch a commit already pinned by the lockfile.

`--offline` prohibits Git network operations. Required checkouts and commits must already exist locally.

```shell
vex fetch --locked
vex build --locked --offline
```

This pair is the strict CI workflow: fetch the exact locked commits while network access is allowed, then compile without network access or lockfile changes. A dry run never fetches dependencies or rewrites the lockfile.

## Compiler setup and inspection

```shell
vex info
vex setup wavec
vex setup wavec --version <version>
vex --version
```

Vex validates `wavec`'s dry-run JSON schema before the real build. A compiler that does not implement the required schema is rejected with a compatibility error rather than being invoked with an unknown plan.
