---
translation_set_id: modules-ffi
path: language/modules-imports-and-ffi
locale: en
group: language
group_order: 2
order: 8
title: Modules, imports, and FFI
summary: Resolution of local, standard, and package imports plus C ABI extern/export declarations.
---

## import syntax

```wave
import("std::string::len");
```

`import` takes one string literal and ends as a complete `);` statement.

## Standard-library imports

Paths beginning with `std::` resolve against the installed Wave standard library.

```wave
import("std::fs::file");
import("std::io::fd");
```

Use `wavec print std-path` to inspect the standard-library location used by the current compiler.

## Local file imports

A path without `::` resolves relative to the importing source file's base directory. If `.wave` is omitted, the compiler appends it while searching.

```wave
import("math");
```

This form resolves `math.wave` in the corresponding base directory.

## External package imports

A non-`std::` path containing `::` treats the first component as a package name.

```wave
import("math::vector::ops");
```

Provide external package locations with dependency options:

```shell
wavec --dep-root .vex/deps build main.wave
wavec --dep math=/absolute/path/to/math build main.wave
```

If the same package appears under multiple dependency roots, resolution is ambiguous; pin it with `--dep name=path`.

## Importing C functions

```wave
extern(c) fun puts(text: ptr<i8>) -> i32;
```

The parser also accepts an explicit external symbol name after the ABI:

```wave
extern(c, "native_symbol") fun local_name(value: i32) -> i32;
```

## Exporting Wave functions

```wave
export(c) fun wave_add(left: i32, right: i32) -> i32 {
    return left + right;
}
```

`extern` and `export` support both individual functions and block forms. Exported functions cannot be generic.

## ABI details to verify explicitly

- Integer and pointer widths
- Calling convention and target ABI name
- External symbol names
- Actual string representation
- Pointer lifetime and ownership
- Required libraries and linker search paths

Successful linking does not by itself prove that the function signature and memory contract match.

## Target-condition attributes

The import preprocessor can handle top-level target conditions such as:

```wave
#[target(os="linux", arch="x86_64")]
extern(c) fun platform_call(value: i32) -> i32;
```

Supported condition keys are `arch`, `os`, `env`, and `abi`; the attribute applies to the next top-level item.
