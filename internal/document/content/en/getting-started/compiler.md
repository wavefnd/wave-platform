---
translation_set_id: compiler
path: getting-started/compiler
locale: en
group: getting-started
group_order: 1
order: 3
title: Compiler command reference
summary: The current wavec commands, build pipeline, outputs, targets, diagnostics, dependencies, and tool queries.
---

## Command model

`wavec` is the compiler CLI. It can compile individual inputs directly, expose compiler capabilities to tools, and manage the installed standard-library source.

```text
wavec [global-options] <command> [command-options]
```

| Command | Purpose |
| --- | --- |
| `wavec build <input...>` | Run the check, code-generation, link, or execution pipeline selected by flags. |
| `wavec check <file>` | Alias for `build <file> --emit=check`. |
| `wavec run <file> [-- <args...>]` | Alias for `build <file> --run`; arguments after `--` go to the program. |
| `wavec print <item>` | Query target and toolchain capabilities. |
| `wavec install std` | Install the standard library. |
| `wavec update std` | Update the installed standard library. |
| `wavec --version` | Print compiler and LLVM backend versions. |

Use the installed release's `wavec --help` as the authoritative option list.

## Build, check, and run

```shell
wavec build main.wave
wavec check main.wave
wavec run main.wave -- first-argument second-argument
```

`build` emits an executable by default. `check` stops after front-end validation. `run` requires a binary output and cannot be combined with a shared-library build.

Use `--dry-run` to validate the request and inspect the planned stages without compiling, linking, or running anything:

```shell
wavec build main.wave --target riscv64-unknown-linux-gnu --dry-run
wavec build main.wave --dry-run --error-format=json
```

The JSON form is the stable integration surface used by build tools such as Vex.

## Emit and input kinds

```shell
wavec build main.wave --emit=ast
wavec build main.wave --emit=ir
wavec build main.wave --emit=bc
wavec build main.wave --emit=asm
wavec build main.wave --emit=obj -o main.o
wavec build main.wave --emit=bin -o app
```

Artifact emit kinds are `ast`, `ir`, `bc`, `asm`, `obj`, and `bin`. `check` is a control mode and must be used alone. Multiple artifact kinds can be comma-separated where the pipeline permits it.

The accepted input kinds are `wave`, `ir`, `bc`, `asm`, `obj`, and `archive`. `--input-type=<kind>` forces one kind for all inputs. Use `--link-only` with object or archive inputs and a binary emit:

```shell
wavec build module.o --input-type=obj --link-only --emit=bin -o app
```

## Output locations

| Option | Effect |
| --- | --- |
| `-o <file>` | Set the primary output path. |
| `--out-dir <dir>` | Place emitted artifacts in a selected directory. |
| `--target-dir <dir>` | Select the intermediate and default artifact root. |

## Optimization and compiler inspection

```shell
wavec -O2 build main.wave
wavec --debug-wave=tokens,ast build main.wave
```

Optimization levels are `-O0`, `-O1`, `-O2`, `-O3`, `-Os`, `-Oz`, and `-Ofast`. `--debug-wave` accepts `tokens`, `ast`, `ir`, `mc`, `hex`, or `all`; comma-separated stages may be combined.

## Native linking

```shell
wavec --link=m -L ./lib build main.wave
wavec build main.wave --shared -o libexample.so
wavec build main.wave --static -o app
wavec build main.wave --pie -o app
```

`--link=<lib>` adds a native library and `-L <path>` adds a search path. Link modes are `--shared`, `--static`, `--pie`, and `--no-pie` subject to their compatibility rules.

Backend and linker controls include:

- `--target`, `--cpu`, `--features`, `--abi`, and `--sysroot`
- `-C linker=<path>` and `-C link-arg=<arg>`
- `-C link-sysroot=<path>` and `-C relocation-model=<model>`
- `-C no-default-libs`

For kernels and other freestanding outputs, use `--freestanding` with the appropriate `--entry`, `--linker-script`, and `--no-start-files` settings.

## External package resolution

```shell
wavec --dep-root .vex/deps build main.wave
wavec --dep math=/opt/wave-deps/math build main.wave
```

`--dep-root <dir>` adds a root used to resolve external `package::module` imports. `--dep <name>=<path>` pins a package name to one directory. These are compiler integration points; project manifests, dependency fetching, and lockfiles belong to Vex.

## Capability queries

Do not hard-code compiler capabilities in tooling. Query the installed compiler:

```shell
wavec print host-target
wavec print target-spec --format=json
wavec print supported-targets
wavec print supported-input-types
wavec print supported-emit-kinds
wavec print supported-print-items
wavec print cpu-list --target riscv64-unknown-linux-gnu
wavec print target-features --target riscv64-unknown-linux-gnu
wavec print default-linker
wavec print sysroot
wavec print std-path
wavec print dep-search-paths
```

Other discoverable items include `host`, `default-target`, and `target-list`. Items that support structured output accept `--format=json`.

## Compiler and toolchain boundaries

`wavec` owns compilation details. Vex owns package manifests, dependency graphs, lockfiles, and repeatable package builds. Whale is a separate low-level toolchain project; it is not selected with a `wavec --whale` flag in the current CLI.
