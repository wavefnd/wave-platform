---
translation_set_id: data-types
path: language/structures-enums-and-aliases
locale: en
group: language
group_order: 2
order: 6
title: Structures, enums, and aliases
summary: Define composite data, named alternatives, and reusable type names.
---

## Structures

```wave
struct Point {
    x: f64;
    y: f64;
}

var point: Point = Point { x: 1.0, y: 2.0 };
```

## Enums

```wave
enum State -> i32 {
    Ready = 0,
    Running,
    Stopped,
}
```

Structures group fields under one type. Enums define a closed set of named alternatives. Type aliases give an existing type another domain-specific name.

## Protocol implementations

```wave
proto Point {
    fun length_squared(self: Point) -> f64 {
        return self.x * self.x + self.y * self.y;
    }
}

var distance: f64 = point.length_squared();
```

A proto block attaches typed behavior to a structure. The self parameter states the receiver type explicitly.

> **Layout**
> 
> Do not assume a native layout at an FFI or binary-data boundary unless the compiler and the external ABI both guarantee it.

