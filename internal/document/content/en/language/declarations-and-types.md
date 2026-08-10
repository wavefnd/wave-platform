---
translation_set_id: types
path: language/declarations-and-types
locale: en
group: language
group_order: 2
order: 2
title: Declarations and types
summary: Local bindings, top-level constants and static storage, built-in types, arrays, and type aliases.
---

## Local variables and bindings

```wave
var count: i32 = 1;
let limit: i32 = 10;
let mut index: i32 = 0;
```

| Declaration | Meaning |
| --- | --- |
| `var` | Reassignable local variable |
| `let` | Local binding that cannot be reassigned |
| `let mut` | Reassignable `let` binding |

Reassigning a `let` binding is rejected during compilation.

## Top-level storage declarations

```wave
const PAGE_SIZE: i32 = 4096;
static request_count: i64 = 0;
```

`const` and `static` are top-level declarations. Use `var` or the `let` forms for locals inside function bodies.

## Integer and floating-point types

The documented fixed-width integer spellings are:

- Signed: `i8`, `i16`, `i32`, `i64`, `i128`, `i256`, `i512`, `i1024`
- Unsigned: `u8`, `u16`, `u32`, `u64`, `u128`, `u256`, `u512`, `u1024`
- Floating point: `f32`, `f64`

The lexer recognizes `isz` and `usz`, but the current type-conversion path does not handle them. Do not use them until compiler support is documented.

## Other built-in types

| Type | Purpose |
| --- | --- |
| `bool` | `true` or `false` |
| `char` | Character value |
| `byte` | Byte value |
| `str` | String value |
| `ptr<T>` | Pointer intended to address `T` |
| `array<T, N>` | Fixed-length array with element type `T` and length `N` |

User-defined structs, enums, and type aliases can also appear in type positions.

## Arrays

```wave
var values: array<i32, 4> = [10, 20, 30, 40];
var first: i32 = values[0];
values[1] = 25;
```

When an array is initialized with an array literal, the declared length must match the number of elements. Indexing uses `[]`.

## Type aliases

```wave
type UserId = u64;
var id: UserId = 7;
```

A type alias gives an existing type another name; it does not create additional storage or a new runtime representation by itself.
