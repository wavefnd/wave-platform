---
translation_set_id: compiler
path: getting-started/compiler
locale: en
group: getting-started
group_order: 1
order: 3
title: Compiler command reference
summary: A practical guide to wavec build, run, emit modes, diagnostics, targets, dependencies, and linking.
---

## Core commands

```shell
wavec build main.wave
wavec run main.wave
wavec build main.wave --emit=check
```

- `build` compiles the input and produces an executable by default.
- `run` builds the program and then executes it.
- `--emit=check` performs Wave front-end checking without producing a normal executable.

Use the installed release's `wavec --help` as the authoritative option list.

## Controlling outputs

```shell
wavec build main.wave -o app
wavec build main.wave --emit=ast
wavec build main.wave --emit=ir
wavec build main.wave --emit=obj -o main.o
```

The artifact emit kinds in v0.2.0-pre-beta are `ast`, `ir`, `bc`, `asm`, `obj`, and `bin`; `check` is a separate control mode.

```shell
wavec print supported-emit-kinds
```

## Optimization and debug output

```shell
wavec -O2 build main.wave
wavec --debug-wave=tokens,ast build main.wave
```

Optimization flags include `-O0`, `-O1`, `-O2`, `-O3`, `-Os`, `-Oz`, and `-Ofast`. Use `--debug-wave` to inspect selected compiler stages.

## Dependencies and linking

```shell
wavec --dep-root .vex/dep build main.wave
wavec --dep math=/opt/wave-deps/math build main.wave
wavec --link=m -L ./lib build main.wave
```

- `--dep-root <dir>` adds a root used to resolve external `package::module` imports.
- `--dep <name>=<path>` pins a package name to a specific directory.
- `--link <lib>` and `-L <path>` add native linker inputs and search paths.

## Machine-readable compiler information

Tooling can query the compiler's actual capabilities through `wavec print`:

```shell
wavec print supported-targets
wavec print supported-input-types
wavec print supported-emit-kinds
wavec print std-path
wavec print target-spec --format=json
```

Use `--error-format=json` when diagnostics need to be consumed by automation.

## Whale option

`--whale` is reserved in v0.2.0-pre-beta but is not an implemented backend. Using it returns a usage error stating that the backend is not implemented.
