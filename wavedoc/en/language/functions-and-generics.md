---
translation_set_id: functions
path: language/functions-and-generics
locale: en
group: language
group_order: 2
order: 5
title: Functions and generics
summary: Function declarations, default parameters, returns, explicit generic instantiation, and current release limitations.
---

## Function declarations

```wave
fun add(left: i32, right: i32) -> i32 {
    return left + right;
}

fun log(message: str) {
    println("{}", message);
}
```

Parameters use `name: type`. Functions that return a value use `-> type`; omitting the return type defines a value-less function.

## return

```wave
fun choose(flag: bool) -> i32 {
    if (flag) {
        return 1;
    }
    return 0;
}
```

A value-less function can use `return;`.

## Generic functions

```wave
fun identity<T>(value: T) -> T {
    return value;
}

fun main() {
    var integer: i32 = identity<i32>(10);
    var decimal: f64 = identity<f64>(3.14);
}
```

Generic calls **require explicit type arguments**. Calling a generic template as `identity(10)` without type arguments produces an error.

## Generic structs

```wave
struct Pair<A, B> {
    first: A;
    second: B;
}

fun main() {
    var pair: Pair<i32, f64> = Pair<i32, f64> {
        first: 1,
        second: 2.5
    };
}
```

Concrete type combinations are monomorphized during compilation.

## Default parameters

The function parser supports default parameter values represented by integer, floating-point, and string literals. Omitted arguments are filled during the compiler's generic-rewrite stage. Do not assume arbitrary expression defaults are supported.

## Current generic limitations

- Generic function calls require explicit type arguments.
- Generic methods are not currently supported.
- Functions exposed through `export(...)` cannot be generic.
- `ptr<T>` and `array<T, N>` are built-in type forms handled specially by the type parser, not user-defined generic templates.
