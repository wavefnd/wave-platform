---
translation_set_id: functions
path: language/functions-and-generics
locale: en
group: language
group_order: 2
order: 5
title: Functions and generics
summary: Function declarations, parameters, return values, and generic code.
---

## Functions

```wave
fun add(left: i32, right: i32) -> i32 {
    return left + right;
}

fun log(message: str) {
    println(message);
}
```

Parameter and return types are explicit. return transfers a value to the caller; a function without a result omits the return type.

## Generics

```wave
fun identity<T>(value: T) -> T {
    return value;
}
```

```wave
var integer: i32 = identity<i32>(10);
var decimal: f64 = identity<f64>(3.14);

struct Pair<A, B> {
    first: A;
    second: B;
}

var pair: Pair<i32, f64> = Pair<i32, f64> { first: 1, second: 2.5 };
```

Generic parameters allow functions and structures to operate on explicitly supplied concrete types. Calls do not infer omitted type arguments in this release. Keep ABI-facing declarations concrete unless the boundary explicitly defines instantiation.

> **Not a pointer model**
> 
> The angle brackets in ptr<T> belong to the Wave Explicit Memory Type Model. ptr<T> is a built-in memory type and is not an instantiation of the general generic system described here.

