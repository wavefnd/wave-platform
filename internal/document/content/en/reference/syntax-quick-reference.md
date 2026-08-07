---
translation_set_id: quick-reference
path: reference/syntax-quick-reference
locale: en
group: reference
group_order: 3
order: 3
title: Syntax quick reference
summary: Frequently used v0.2.0-pre-beta declarations, control flow, types, pointers, and FFI syntax on one page.
---

## Declarations

```wave
var value: i32 = 1;
let fixed: i32 = 2;
let mut mutable: i32 = 3;
const LIMIT: i32 = 64;
static total: i64 = 0;
type Identifier = u64;
```

`var`/`let` are local forms; `const`/`static` are top-level declarations.

## Functions

```wave
fun max(left: i32, right: i32) -> i32 {
    if (left > right) {
        return left;
    }
    return right;
}
```

## Generics

```wave
fun identity<T>(value: T) -> T {
    return value;
}

var value: i32 = identity<i32>(10);
```

Generic calls require explicit type arguments.

## Structs and enums

```wave
struct Pair {
    left: i32;
    right: i32;
}

enum Result -> i32 {
    Ok = 0,
    Error,
}
```

## Conditionals and loops

```wave
if (ready) {
    println("ready");
}

while (count < 10) {
    count += 1;
}

for (var i: i32 = 0; i < 10; i += 1) {
    println("{}", i);
}
```

`if`, `while`, `for`, and `match` use parenthesized headers in this release.

## Arrays and pointers

```wave
var values: array<i32, 4> = [1, 2, 3, 4];
var p: ptr<i32> = &values[0];
var first: i32 = deref p;
```

## Imports and FFI

```wave
import("std::string::len");
extern(c) fun native_call(value: i32) -> i32;

export(c) fun wave_call(value: i32) -> i32 {
    return value + 1;
}
```

## Compiler queries

```shell
wavec build main.wave --emit=check
wavec print supported-targets
wavec print supported-emit-kinds
```
