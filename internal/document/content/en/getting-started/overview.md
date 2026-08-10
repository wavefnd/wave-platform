---
translation_set_id: overview
path: getting-started/overview
locale: en
group: getting-started
group_order: 1
order: 1
title: Wave language overview
summary: A practical overview of Wave, its program structure, and the documentation's compatibility scope.
---

## Documentation baseline

Wave is a statically typed systems programming language designed for native code generation and explicit low-level control. This documentation describes the current Wave compiler contract: syntax accepted by the parser, semantic restrictions enforced by the compiler, and options reported by `wavec --help` and `wavec print`.

Pre-beta behavior can still change. If an example behaves differently with your compiler, check `wavec --version` and query the installed compiler rather than relying on a hard-coded capability list.

## First program

```wave
fun main() {
    println("Hello, Wave!");
}
```

Save the source as `main.wave` and run it with:

```shell
wavec run main.wave
```

To build an executable without running it:

```shell
wavec build main.wave -o app
```

Hosted executables normally start at `main`. Freestanding builds can select another entry symbol together with the appropriate linker configuration.

## Basic statement style

Wave uses explicit types for variables and function parameters.

```wave
fun add(left: i32, right: i32) -> i32 {
    let result: i32 = left + right;
    return result;
}

fun main() {
    var count: i32 = 1;
    count += 1;
    println("count = {}", count);
}
```

`var` is a reassignable local variable, `let` is an immutable local binding, and `let mut` is an explicitly mutable `let` binding. In this release, `const` and `static` are top-level declarations.

## Console I/O

`print`, `println`, and `input` are parser-recognized I/O statements rather than ordinary standard-library function calls. Their first argument must be a string literal, and the number of `{}` placeholders must match the number of following arguments.

```wave
fun main() {
    var value: i32 = 0;
    input("{}", value);
    print("value = ");
    println("{}", value);
}
```

## Low-level facilities

Wave provides `ptr<T>`, address-of `&`, explicit `deref`, C ABI boundaries through `extern(c)` and `export(c)`, and inline `asm`. These facilities do not automatically establish ownership, bounds, alignment, or lifetime validity; those properties remain part of the surrounding program and API contract.

## Suggested learning order

1. Install the compiler and check `wavec --version` and `wavec --help`.
2. Learn declarations, types, expressions, and control flow.
3. Learn functions, generics, structs, enums, and `proto`.
4. Move on to pointers, imports, FFI, and the standard library.
5. Use the compiler, Vex, Whale, target, and quick-reference pages as lookup material.
