---
translation_set_id: overview
path: getting-started/overview
locale: en
group: getting-started
group_order: 1
order: 1
title: Wave language overview
summary: A practical introduction to Wave syntax, program structure, console I/O, and low-level facilities.
---

## About Wave

Wave is a statically typed systems programming language designed for native code generation and explicit low-level control. Types, memory access, native interfaces, and target settings remain visible in source code and build commands.

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

Wave keeps function signatures explicit. Local variables declare their type after the variable name.

```wave
fun add(left: i32, right: i32) -> i32 {
    var result: i32 = left + right;
    return result;
}

fun main() {
    var count: i32 = 1;
    var next: i32 = count + 1;
    count += 1;
    println("count = {}, next = {}", count, next);
}
```

`var` is the syntax for declaring a local variable. Local declarations use `var name: Type = value;` to state the variable type explicitly. `const` and `static` are top-level declarations.

## Console I/O

`print`, `println`, and `input` are Wave console I/O statements. Their first argument is a string literal, and each `{}` placeholder corresponds to one following value.

```wave
fun main() {
    var value: i32 = 0;
    input("{}", value);
    print("value = ");
    println("{}", value);
}
```

## Low-level facilities

Wave provides `ptr<T>`, address-of `&`, explicit `deref`, C ABI boundaries through `extern(c)` and `export(c)`, and inline `asm`. Programs using these facilities define the ownership, bounds, alignment, and lifetime rules required by each memory or native API.

## Suggested learning order

1. Install the compiler and check `wavec --version` and `wavec --help`.
2. Learn declarations, types, expressions, and control flow.
3. Learn functions, generics, structs, enums, and `proto`.
4. Move on to pointers, imports, FFI, and the standard library.
5. Use the compiler, Vex, Whale, target, and quick-reference pages as lookup material.
