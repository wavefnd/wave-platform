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

## Local variables

```wave
var count: i32 = 1;
var limit: i32 = 10;
var index = count + 1;
```

| Declaration | Meaning |
| --- | --- |
| `var` | Declares a local variable whose value can be reassigned |

`var` is the syntax for declaring local variables.

A type can be written after the variable name, or inferred from an initializer:

```wave
var capacity: i64 = 4096;
var doubled = capacity * 2;
```

Use `var name: Type;` when declaring storage without an initializer. Inferred declarations require an initializer whose type can be determined from the expression.

## Top-level storage declarations

```wave
const PAGE_SIZE: i32 = 4096;
static request_count: i64 = 0;
```

`const` and `static` are top-level declarations. Use `var` for locals inside function bodies.

## Integer and floating-point types

Wave provides these integer and floating-point types:

- Signed: `i8`, `i16`, `i32`, `i64`, `i128`
- Unsigned: `u8`, `u16`, `u32`, `u64`, `u128`
- Pointer-sized: `isz`, `usz`
- Floating point: `f32`, `f64`

`isz` and `usz` represent signed and unsigned integers sized for the target's address space.

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

A type alias is a readable alternative name for another type. In this example, `UserId` can be used anywhere a `u64` is expected.
