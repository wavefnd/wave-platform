---
translation_set_id: design-goals
path: getting-started/design-goals
locale: en
group: getting-started
group_order: 1
order: 4
title: Language design
summary: Wave's principles for explicit low-level control, static typing, native interoperability, and clear language boundaries.
---

## Design direction

Wave aims to keep systems-level control visible in source code. That direction appears in several concrete facilities:

- Types are explicit at function and data boundaries, including local variable declarations.
- `ptr<T>`, `&`, and `deref` make addresses and memory accesses explicit.
- C ABI boundaries are declared with `extern(c)` and `export(c)`.
- `wavec` exposes target, CPU-feature, linker, and freestanding-build controls.
- Functions, structs, enums, generics, and `proto` provide higher-level program structure.

## Language versus standard library

Language syntax and the standard library are separate layers. For example, `print`, `println`, and `input` are language statements, while string searching, file systems, networking, and memory helpers are standard-library modules.

Keeping that distinction clear makes it easier to tell whether a feature belongs to the language grammar, library API, or toolchain.

## Local variables

```wave
var fixed: i32 = 10;
var state: i32 = 0;
var counter: i32 = 0;
```

`var` is the syntax for declaring a local variable.

## Explicit contracts

Wave documentation presents each language feature through its syntax and observable behavior. Low-level operations also state the memory, ABI, target, or ownership rules that a program must uphold. This keeps source code understandable without relying on compiler internals.
