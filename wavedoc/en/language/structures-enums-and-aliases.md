---
translation_set_id: data-types
path: language/structures-enums-and-aliases
locale: en
group: language
group_order: 2
order: 6
title: Structs, enums, and type aliases
summary: Struct fields and methods, proto extensions, enums with explicit representation types, and type aliases.
---

## Structs

```wave
struct Point {
    x: f64;
    y: f64;
}

fun main() {
    var point: Point = Point { x: 1.0, y: 2.0 };
    println("{}", point.x);
}
```

Struct fields use `name: type;`. Field access uses `value.field`.

## Methods and proto blocks

Methods can be declared with `fun` inside a struct body. A `proto` block can attach methods to an already declared struct.

```wave
struct Counter {
    value: i32;
}

proto Counter {
    fun read(self: Counter) -> i32 {
        return self.value;
    }
}
```

Call a method with ordinary method syntax:

```wave
var counter: Counter = Counter { value: 3 };
var value: i32 = counter.read();
```

## Enums

An enum declares its integer representation type after `->`.

```wave
enum State -> i32 {
    Ready = 0,
    Running,
    Stopped,
}
```

A variant can specify an integer value explicitly. The first omitted value is `0`; each later omitted value is one greater than the preceding variant.

## Type aliases

```wave
type FileHandle = i64;
```

A type alias allows the same underlying type to be referenced by another name.

```wave
var handle: FileHandle = 4;
var raw: i64 = handle;
```

`FileHandle` and `i64` are the same type. The alias communicates the value's purpose without requiring a conversion.

## ABI and layout

When sharing a Wave struct directly across an external ABI or binary-file boundary, do not infer C-compatible layout from field order alone. Design the representation according to guarantees provided by the target ABI and compiler.
