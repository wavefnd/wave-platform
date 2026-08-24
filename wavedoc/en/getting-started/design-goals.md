---
translation_set_id: design-goals
path: getting-started/design-goals
locale: en
group: getting-started
group_order: 1
order: 4
title: Language design and current scope
summary: Wave's explicit low-level control, static typing, native interoperability, and documented implementation boundaries.
---

## Design direction

Wave aims to keep systems-level control visible in source code. That direction appears in several concrete facilities:

- Types are written explicitly for variables, parameters, return values, and composite data.
- `ptr<T>`, `&`, and `deref` make addresses and memory accesses explicit.
- C ABI boundaries are declared with `extern(c)` and `export(c)`.
- `wavec` exposes target, CPU-feature, linker, and freestanding-build controls.
- Functions, structs, enums, generics, and `proto` provide higher-level program structure.

## Language versus standard library

Compiler-recognized syntax and functionality implemented in `std` are separate layers. For example, `print`, `println`, and `input` are parser-recognized statements, while string searching, file systems, networking, and memory helpers are implemented as standard-library modules.

Keeping that distinction clear makes it easier to tell whether a feature belongs to the language grammar, library API, or toolchain.

## Local variables

```wave
var fixed: i32 = 10;
var state: i32 = 0;
var counter: i32 = 0;
```

All local variables use `var`. The former `let` and `let mut` declarations are
removed syntax, so examples and new source must not use them.

## Documentation policy

The current compiler contains spellings reserved by the lexer alongside implementation paths that are not complete. This documentation therefore does not treat the existence of a token as proof that a language feature is supported. It focuses on syntax that reaches the parser and code generator, with known limitations called out explicitly.
