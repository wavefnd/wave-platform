---
translation_set_id: storage-duration
path: language/storage-duration
locale: en
group: language
group_order: 2
order: 12
title: Storage duration and mutability
summary: Choose var, let, const, and static according to scope and state requirements.
---

| Form | Scope and purpose |
| --- | --- |
| var | Normal local variable; Wave's default variable syntax |
| let | Constrained immutable local binding for strict OS/security code |
| let mut | Constrained explicitly mutable local binding for strict OS/security code |
| const | Top-level compile-time constant |
| static | Top-level stored variable |

```wave
const PAGE_SIZE: i32 = 4096;
static request_count: i64 = 0;

fun main() {
    var current: i32 = 1;
    current += 1;
}
```

> **Scope rule**
> 
> Use var and let forms inside functions or blocks. Use const and static for top-level storage declarations.

