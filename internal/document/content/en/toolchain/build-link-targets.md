---
translation_set_id: build-link-targets
path: toolchain/build-link-targets
locale: en
group: toolchain
group_order: 4
order: 2
title: Build, link, and target options
summary: Control emitted artifacts, linking, target selection, and freestanding builds.
---

## Emit artifacts

```shell
wavec build main.wave --emit=check
wavec build main.wave --emit=ast
wavec build main.wave --emit=ir
wavec build main.wave --emit=obj -o main.o
wavec build main.wave --emit=bin -o app
```

Supported emit kinds are check, ast, ir, bc, asm, obj, and bin. check must be used alone. Use wavec print supported-emit-kinds for a machine-readable capability check.

## Linking

```shell
wavec build main.wave --link=m -L ./lib
wavec build module.o --input-type=obj --link-only --emit=bin -o app
```

## Targets and freestanding mode

--target, --cpu, --features, --abi, and --sysroot describe the compilation target. --freestanding, --entry, --linker-script, --no-start-files, and -C no-default-libs support kernel and OS-style link plans. Validate a plan with --dry-run before executing it.

