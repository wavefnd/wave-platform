---
translation_set_id: functions
path: language/functions-and-generics
locale: en
group: language
group_order: 2
order: 5
title: Functions and generics
summary: Function declarations, return values, default parameters, and explicit generic instantiation.
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

Each concrete type combination produces a specialized function or struct definition.

## Default parameters

Default parameters use integer, floating-point, or string literals. A call can omit a trailing argument when its parameter declares a default value.

```wave
fun repeat(value: i32, count: i32 = 1) -> i32 {
    return value * count;
}

var result: i32 = repeat(7);
```

## Generic rules

- Generic function calls require explicit type arguments.
- Functions exposed through `export(...)` cannot be generic.
- `ptr<T>` and `array<T, N>` are built-in memory types, not user-defined generic templates.
