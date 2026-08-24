---
translation_set_id: program-structure
path: language/program-structure
locale: en
group: language
group_order: 2
order: 10
title: Program structure
summary: Top-level declarations, the main entry point, import organization, and freestanding entry symbols.
---

## Top-level source items

A Wave source file can contain top-level items such as:

- `import(...)`
- `const` and `static`
- `type`
- `struct`, `enum`, and `proto`
- `extern(...)` and `export(...)`
- `fun`
- `#[target(...)]` conditions before supported items

Local `var` declarations belong inside functions or blocks.

Prefix an importable declaration with `pub` to expose it to other Wave modules. `pub` can be used with functions, structs, enums, type aliases, constants, statics, selected re-exports, and individual ABI exports. Declarations without `pub` stay private to their source module.

## Hosted executable

```wave
import("std::string::len")::{len};

const EXIT_OK: i32 = 0;

fun main() -> i32 {
    var message: str = "Wave";
    println("{} {}", message, len(message));
    return EXIT_OK;
}
```

A normal hosted executable uses `main` as its program entry function.

## Declaration organization and imports

Imports make another Wave source unit's public declarations available through a namespace or a selective import. Standard-library, external-package, and local imports use distinct path forms described in the Modules, imports, and FFI guide.

Keep source files focused and import the modules each file directly depends on; this makes dependency flow easier to read.

## Freestanding programs

Kernels, boot code, or targets without a hosted runtime can use freestanding build controls.

```shell
wavec build kernel.wave \
  --freestanding \
  --entry=_start \
  --linker-script=linker.ld \
  --no-start-files
```

`--freestanding` participates in a build plan without default libraries, while `--entry` configures the linker entry symbol. A bootable result still requires a target-appropriate linker script, object format, and platform startup design.
