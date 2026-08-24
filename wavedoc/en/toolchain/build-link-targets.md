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

Artifact emit kinds are `ast`, `ir`, `bc`, `asm`, `obj`, and `bin`. `check` validates source without producing an artifact and is used alone.

```shell
wavec print supported-emit-kinds
```

## Input kinds and link-only mode

The compiler distinguishes Wave source, IR, bitcode, assembly, object, and archive inputs. Query the installed compiler with:

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

## Supported target families

A compiler built with the full target set provides these target contracts. Use `wavec print supported-targets` to see the targets included in a particular compiler build.

| Target | Environment | Object format |
| --- | --- | --- |
| `x86_64-unknown-linux-gnu` | Hosted Linux GNU | ELF |
| `x86_64-apple-darwin` | Hosted macOS | Mach-O |
| `x86_64-w64-windows-gnu` | Hosted Windows GNU | COFF |
| `x86_64-pc-windows-gnu` | Hosted Windows GNU alias | COFF |
| `x86_64-unknown-none-elf` | Freestanding | ELF |
| `aarch64-unknown-linux-gnu` | Hosted Linux GNU | ELF |
| `aarch64-apple-darwin` | Hosted macOS | Mach-O |
| `aarch64-unknown-none-elf` | Freestanding | ELF |
| `riscv64-unknown-linux-gnu` | Hosted Linux GNU | ELF |
| `riscv64-unknown-none-elf` | Freestanding | ELF |

## RISC-V 64 contract

The hosted RISC-V target defaults to `generic-rv64`, RV64GC, and the `lp64d` ABI. The freestanding target defaults to `generic-rv64`, RV64IMAC, and `lp64`.

```shell
wavec print target-spec --target riscv64-unknown-linux-gnu --format=json
wavec print target-spec --target riscv64-unknown-none-elf --format=json
```

Supported RISC-V CPUs are `generic`, `generic-rv64`, `rocket-rv64`, and `sifive-u74`. Feature overrides use signed comma-separated names from `m`, `a`, `f`, `d`, `c`, `zicsr`, and `zifencei`:

```shell
wavec build main.wave \
  --target riscv64-unknown-linux-gnu \
  --features=+m,+a,+f,-d,+c,+zicsr \
  --abi=lp64f
```

RISC-V validation rejects inconsistent combinations: `d` requires `f`, `f` requires `zicsr`, and `lp64`, `lp64f`, or `lp64d` must agree with the enabled floating-point features. When no ABI override is supplied, the compiler derives the ABI from those features.

## Freestanding linking

```shell
wavec build kernel.wave \
  --target riscv64-unknown-none-elf \
  --freestanding \
  --entry=_start \
  --linker-script=linker.ld \
  --no-start-files \
  -o kernel.elf
```

`--freestanding` adjusts the build away from default libraries. `--entry` sets the linker entry symbol, `--linker-script` supplies a linker script, and `--no-start-files` omits hosted startup files.

Use `--dry-run` to inspect the planned build and link steps before execution.

## Hosted cross-linking

Code generation for a target does not provide that target's C runtime, startup objects, or libraries. A hosted cross-build needs a compatible sysroot and, when necessary, an explicit linker:

```shell
wavec build main.wave \
  --target riscv64-unknown-linux-gnu \
  --sysroot /path/to/riscv64-sysroot \
  -C linker=/path/to/target-linker \
  -o app-riscv64
```

The sysroot must contain files for the selected ABI. A host library with the same name is not a substitute for a target library.

## Cross-build checklist

- The target triple appears in the compiler's supported-target list.
- The sysroot and linker match the target ABI.
- Linked libraries were built for the target architecture.
- Requested CPU features are valid for the selected CPU.
- In freestanding builds, the entry symbol and memory layout match the linker script.
