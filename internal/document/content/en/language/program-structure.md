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

A v0.2.0-pre-beta source file can contain top-level items such as:

- `import(...)`
- `const` and `static`
- `type`
- `struct`, `enum`, and `proto`
- `extern(...)` and `export(...)`
- `fun`
- `#[target(...)]` conditions before supported items

Local `var` and `let` declarations belong inside functions or blocks.

## Hosted executable

```wave
import("std::string::len");

const EXIT_OK: i32 = 0;

fun main() -> i32 {
    let message: str = "Wave";
    println("{} {}", message, len(message));
    return EXIT_OK;
}
```

A normal hosted executable uses `main` as its program entry function.

## Declaration organization and imports

Imports are processed by combining the imported file's AST with the program. The compiler tracks already imported units, while standard, external-package, and local paths follow different resolution rules.

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
