---
translation_set_id: quick-reference
path: reference/syntax-quick-reference
locale: en
group: reference
group_order: 3
order: 3
title: Syntax quick reference
summary: Frequently used declarations, control flow, types, pointers, and FFI syntax on one page.
---

## Declarations

```wave
var value: i32 = 1;
var next: i32 = value + 1;
const LIMIT: i32 = 64;
static total: i64 = 0;
type Identifier = u64;
```

`var` is the local form and requires an explicit type. `const` and `static` are top-level declarations.

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

match (status) {
    Ready => { println("ready"); }
    0 => { println("zero"); }
    _ => { println("other"); }
}
```

`if`, `while`, `for`, and `match` use parenthesized headers.

## Arrays and pointers

```wave
var values: array<i32, 4> = [1, 2, 3, 4];
var p: ptr<i32> = &values[0];
var first: i32 = deref p;
```

## Console I/O

```wave
print("value = ");
println("{}", value);
input("{}", value);
```

The first argument is a string literal. Each exact `{}` placeholder requires one following expression; `input` destinations must be assignable.

## Imports and FFI

```wave
import("std::string::len")::{len};
import("./helpers" as helpers);
import("math")::{add, Point};
extern(c) fun native_call(value: i32) -> i32;

export(c) fun wave_call(value: i32) -> i32 {
    return value + 1;
}
```

Use `pub` on declarations imported by other Wave modules, and `pub import("path")::{name};` to re-export selected public names.

## Target-conditioned items

```wave
#[target(os="linux", arch="riscv64")]
extern(c) fun platform_call(value: i32) -> i32;
```

Supported condition keys are `arch`, `os`, `env`, and `abi`. The attribute controls the next top-level item.

## Inline assembly

```wave
var result: i64 = 0;
asm {
    "mv a0, a1"
    in("a1") 7
    out("a0") result
    clobber("memory")
}
```

Instruction text and register names are target-specific. Declare every input, output, and hidden clobber required by the block.

## Compiler queries

```shell
wavec build main.wave --emit=check
wavec print supported-targets
wavec print supported-emit-kinds
```
