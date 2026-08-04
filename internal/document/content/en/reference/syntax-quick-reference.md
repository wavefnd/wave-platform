---
translation_set_id: quick-reference
path: reference/syntax-quick-reference
locale: en
group: reference
group_order: 3
order: 3
title: Syntax quick reference
summary: A compact reminder of the Wave constructs covered by this manual.
---

## Declarations

```wave
var value: i32 = 1;
var mutable: i32 = 2;
let fixed: i32 = 3;       // constrained immutable binding
let mut guarded: i32 = 4; // explicit mutable binding for strict code
const LIMIT: u64 = 64;
type Identifier = u64;
```

## Function

```wave
fun max(left: i32, right: i32) -> i32 {
    if left > right { return left; }
    return right;
}
```

## Composite types

```wave
struct Pair { left: i32; right: i32; }
enum Result -> i32 { Ok = 0, Error }
```

## Memory and native boundary

```wave
var address: ptr<i32> = raw as ptr<i32>;
var item: i32 = deref address;
extern(c) fun native_call(value: i32) -> i32;
```

