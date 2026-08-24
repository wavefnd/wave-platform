---
translation_set_id: storage-duration
path: language/storage-duration
locale: en
group: language
group_order: 2
order: 12
title: Storage duration and mutability
summary: Distinguish the scope and write semantics of var, const, and static.
---

## Declaration semantics

| Form | Allowed location | Reassignable | Purpose |
| --- | --- | --- | --- |
| `var` | Function/block | Yes | Ordinary mutable local variable |
| `const` | Top level | No | Global constant declaration |
| `static` | Top level | Yes | Static storage that exists for program lifetime |

```wave
const PAGE_SIZE: i32 = 4096;
static request_count: i64 = 0;

fun main() {
    var limit: i32 = 4;
    var current: i32 = 0;
    var retries: i32 = 0;

    current += 1;
    retries += 1;
    println("{} {} {}", limit, current, retries);
}
```

## Local declaration contract

```wave
var value: i32 = 1;
value = 2;
```

The reassignment above is allowed. Wave no longer has separate immutable local
binding syntax: `let` and `let mut` have been removed. Use `const` at the top
level when a compile-time constant is required.

## Local const and static

The function parser explicitly rejects `const` and `static` inside function bodies. The same restriction applies to for-loop initializers.

## Lifetimes and pointers

You can take the address of a local with `&`, but `ptr<T>` does not track how long the referenced storage remains valid. If a local address escapes its scope, the surrounding program must ensure that the address is not used after the storage becomes invalid.
