---
translation_set_id: modules-ffi
path: language/modules-imports-and-ffi
locale: en
group: language
group_order: 2
order: 8
title: Modules, imports, and FFI
summary: Organize Wave code and declare native C boundaries.
---

## Modules and imports

```wave
import("std::io::fd");
import("std::math::int");
```

import loads a module by its qualified string path. The exact path follows the installed standard library or dependency tree.

## C interoperability

```wave
extern(c) fun puts(text: ptr<i8>) -> i32;

export(c) fun wave_add(left: i32, right: i32) -> i32 {
    return left + right;
}
```

extern(c) declares a function implemented by a C-compatible library. export(c) exposes a Wave function with a C-compatible calling boundary.

> **ABI contract**
> 
> Match integer widths, calling convention, symbol name, pointer lifetime, string representation, and ownership with the native declaration. A successful link does not prove the ABI is correct.

