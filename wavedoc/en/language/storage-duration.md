---
translation_set_id: storage-duration
path: language/storage-duration
locale: en
group: language
group_order: 2
order: 18
title: Storage duration and mutability
summary: Distinguish between the scope and writability of var, const and static.
---

## Meaning of each declaration

|format|Allowed Location|reassignment|Use|
| --- | --- | --- | --- |
| `var` |Function/block|possible|Common mutable local variables|
| `const` |top|Impossible|Global constant declaration|
| `static` |top|possible|Static stored declarations that exist for the life of the program|

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

## Local declaration rules

```wave
var value: i32 = 1;
value = 2;
```

A new value can be assigned to a local variable declared as `var`. Constants to be used throughout the program are declared at the top level as `const`.

## Local use of const and static

`const` and `static` are top-level declarations. The function body and `for` initialization use the local declaration `var`.

## Lifespan and Pointers

You can get the address of a local variable as `&`, but the type `ptr<T>` does not track the actual lifetime of the storage pointed to by the pointer. When passing a local storage address out of a function, the program structure must directly ensure that the address remains valid.

## Learning and Example Range

[Practice with the full program](/docs/en/getting-started/overview) · [Standard library](/docs/en/stdlib)
