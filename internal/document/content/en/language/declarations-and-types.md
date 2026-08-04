---
translation_set_id: types
path: language/declarations-and-types
locale: en
group: language
group_order: 2
order: 2
title: Declarations and types
summary: Variables, constants, mutability, primitive types, arrays, and aliases.
---

## Variables and constrained bindings

```wave
var count: i32 = 1;
var total: i64 = 0;
const LIMIT: u64 = 64;
static requests: u64 = 0;
```

var is Wave's normal variable declaration and should be the default in ordinary programs. const declares a constant, while static provides static storage.

> **OS and security code**
> 
> let and let mut are constrained binding forms intended for code that needs to make immutability and state transitions especially explicit, such as operating-system and security-sensitive components. They are available, but they are not Wave's default variable style.

## Primitive types

| Family | Types |
| --- | --- |
| Signed integers | i8 through i1024 |
| Unsigned integers | u8 through u1024 |
| Floating point | f32, f64 |
| Other | bool, char, byte, str, ptr<T>, array<T, N> |

## Type aliases

```wave
type UserId = u64;
var id: UserId = 7;
```

> **Release limitation**
> 
> Although isz and usz are reserved spellings, the 0.2.0-pre-beta parser does not handle them correctly. Use a fixed-width integer type in this release.

## Arrays

```wave
var values: array<i32, 4> = [10, 20, 30, 40];
var first: i32 = values[0];
values[1] = 25;
```

array<T, N> has an element type and a compile-time length. Indexing uses square brackets.

