---
translation_set_id: program-structure
path: language/program-structure
locale: en
group: language
group_order: 2
order: 10
title: Program structure and imports
summary: Organize source files, entry points, declarations, and dependency imports.
---

## Source file

A Wave source file can contain imports, top-level const and static declarations, type declarations, structures, enums, protocol blocks, extern or export declarations, and functions.

```wave
import("std::string::len");

const EXIT_OK: i32 = 0;

fun main() -> i32 {
    var message: str = "Wave";
    println("{} {}", message, len(message));
    return EXIT_OK;
}
```

## Import resolution

import receives a qualified string path. Standard-library paths start with std::. Dependency paths are resolved from the roots supplied to wavec or the package tool.

## Entry point

Hosted executables normally define main. Freestanding builds can select another entry symbol with the compiler's --entry option and matching linker configuration.

