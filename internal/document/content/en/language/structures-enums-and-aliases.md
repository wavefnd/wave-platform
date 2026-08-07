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

The v0.2.0-pre-beta parser accepts `fun` methods directly inside a struct body. `proto` also provides a separate block for attaching methods to an already declared struct.

```wave
struct Counter {
    value: i32;
}

proto Counter {
    fun current(self: Counter) -> i32 {
        return self.value;
    }
}
```

Call a method with ordinary method syntax:

```wave
var counter: Counter = Counter { value: 3 };
var value: i32 = counter.current();
```

## Enums

This release's enum syntax writes an underlying representation type after `->`.

```wave
enum State -> i32 {
    Ready = 0,
    Running,
    Stopped,
}
```

A variant can specify an integer value explicitly; omitted values continue from preceding values according to compiler enum processing.

## Type aliases

```wave
type FileHandle = i64;
```

Aliases let the same underlying type be referenced by another name.

## ABI and layout

When sharing a Wave struct directly across an external ABI or binary-file boundary, do not infer C-compatible layout from field order alone. Design the representation according to guarantees provided by the target ABI and compiler.
