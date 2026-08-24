---
translation_set_id: modules-ffi
path: language/modules-imports-and-ffi
locale: en
group: language
group_order: 2
order: 8
title: Modules, imports, and FFI
summary: Local, standard-library, and package imports; namespaces, visibility, re-exports; and C ABI declarations.
---

## import syntax

```wave
import("std::string::len");
```

`import` takes a string literal and ends with `;`. A normal import creates a namespace for the module's public declarations.

## Standard-library imports

Paths beginning with `std::` resolve against the installed Wave standard library.

```wave
import("std::fs::file");
import("std::io::fd");
```

Use `wavec print std-path` to locate the installed standard library.

## Local module imports

Local modules use a path beginning with `./`. The path is relative to the importing source file, and `.wave` can be omitted.

```wave
import("./math");
```

This form loads `math.wave` and creates the `math` namespace. Local import paths stay inside the module directory and do not use `..` or backslashes.

## External package imports

A bare name identifies an external package root. Additional `::` components identify a module inside that package.

```wave
import("math");
import("math::vector::ops");
```

The package root loads `src/lib.wave` (or `lib.wave`). A package module such as `math::vector::ops` loads `src/vector/ops.wave` (or `vector/ops.wave`). Public declarations are accessed through the import namespace:

```wave
var sum = math::add(1, 2);
var unit = math::vector::ops::normalize(value);
```

Provide external package locations with dependency options:

```shell
wavec --dep-root .vex/deps build main.wave
wavec --dep math=/absolute/path/to/math build main.wave
```

If the same package appears under multiple dependency roots, resolution is ambiguous; pin it with `--dep name=path`.

## Aliases and selective imports

Use `as` to choose a shorter or unambiguous namespace:

```wave
import("./geometry_helpers" as geometry);
var area = geometry::area(width, height);
```

Use a selective import to bring named public declarations into the importing module:

```wave
import("math")::{add, Point};

var sum = add(1, 2);
var origin = Point { x: 0, y: 0 };
```

An import alias and a selective import are separate forms and cannot be combined.

## Public declarations and re-exports

Declarations are private to their module unless they use `pub`:

```wave
pub fun add(left: i32, right: i32) -> i32 {
    return left + right;
}

pub struct Point {
    x: i32;
    y: i32;
}
```

`pub` also applies to `enum`, `type`, `const`, and `static` declarations. A module can re-export selected public names from another module:

```wave
pub import("./arithmetic")::{add, subtract};
```

`pub` controls Wave module visibility. It is separate from `export(c)`, which exposes a native ABI symbol. The `main` entry function remains private.

## Importing C functions

```wave
extern(c) fun puts(text: ptr<i8>) -> i32;
```

Add an explicit external symbol name after the ABI when the Wave name and native symbol differ:

```wave
extern(c, "native_symbol") fun local_name(value: i32) -> i32;
```

## Exporting Wave functions

```wave
export(c) fun wave_add(left: i32, right: i32) -> i32 {
    return left + right;
}
```

`extern` and `export` support both individual functions and block forms. Exported functions cannot be generic. Add `pub` separately when another Wave module must also import an exported function.

## ABI details to verify explicitly

- Integer and pointer widths
- Calling convention and target ABI name
- External symbol names
- Actual string representation
- Pointer lifetime and ownership
- Required libraries and linker search paths

Successful linking does not by itself prove that the function signature and memory contract match.

## Target-condition attributes

Top-level declarations can be selected for a target with `#[target(...)]`:

```wave
#[target(os="linux", arch="x86_64")]
extern(c) fun platform_call(value: i32) -> i32;
```

Supported condition keys are `arch`, `os`, `env`, and `abi`; the attribute applies to the next top-level item.
