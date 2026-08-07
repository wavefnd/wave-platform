---
translation_set_id: build-link-targets
path: toolchain/build-link-targets
locale: en
group: toolchain
group_order: 4
order: 2
title: Build, link, and target options
summary: Emit artifacts, input kinds, native linking, target/CPU/ABI controls, and freestanding build plans.
---

## Emit artifacts

```shell
wavec build main.wave --emit=check
wavec build main.wave --emit=ast
wavec build main.wave --emit=ir
wavec build main.wave --emit=bc
wavec build main.wave --emit=asm
wavec build main.wave --emit=obj -o main.o
wavec build main.wave --emit=bin -o app
```

The artifact emit kinds in v0.2.0-pre-beta are `ast`, `ir`, `bc`, `asm`, `obj`, and `bin`. `check` is a control mode rather than an artifact kind and is intended to stand alone.

```shell
wavec print supported-emit-kinds
```

## Input kinds and link-only mode

The compiler distinguishes Wave source, IR, bitcode, assembly, object, and archive inputs. Query the current list with:

```shell
wavec print supported-input-types
```

To link already produced objects or archives, combine `--input-type` with `--link-only`:

```shell
wavec build module.o --input-type=obj --link-only --emit=bin -o app
```

## Native linking

```shell
wavec --link=m -L ./lib build main.wave
```

`--link` adds a library and `-L` adds a library search path. Declaring an FFI symbol does not automatically link the library that provides it.

## Target selection

Major LLVM target controls include:

- `--target <triple>`
- `--cpu <name>`
- `--features <csv>`
- `--abi <name>`
- `--sysroot <path>`

Ask the compiler for host defaults and supported targets:

```shell
wavec print host-target
wavec print supported-targets
wavec print target-spec --target <triple>
wavec print cpu-list --target <triple>
wavec print target-features --target <triple>
```

## Freestanding linking

```shell
wavec build kernel.wave \
  --freestanding \
  --entry=_start \
  --linker-script=linker.ld \
  --no-start-files \
  -o kernel.elf
```

`--freestanding` adjusts the build away from default libraries. `--entry` sets the linker entry symbol, `--linker-script` supplies a linker script, and `--no-start-files` omits hosted startup files.

Use `--dry-run` to inspect the planned build and link steps before execution.

## Cross-build checklist

- The target triple appears in the compiler's supported-target list.
- The sysroot and linker match the target ABI.
- Linked libraries were built for the target architecture.
- Requested CPU features are valid for the selected CPU.
- In freestanding builds, the entry symbol and memory layout match the linker script.
