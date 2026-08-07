---
translation_set_id: memory-buffer
path: reference/memory-and-buffer
locale: en
group: reference
group_order: 3
order: 5
title: Memory and buffers
summary: Manual allocation, copying, alignment, page helpers in std::mem, and safe usage patterns around std::buffer.
---

## Basic manual allocation

```wave
import("std::mem::alloc");
import("std::mem::ops");

fun main() {
    let size: i64 = 256;
    let mut memory: ptr<u8> = mem_alloc(size);

    if (memory != null) {
        mem_zero(memory, size);
        mem_free(memory, size);
    }
}
```

`mem_alloc(size)` returns `ptr<u8>` and may return `null` on failure. `mem_free` receives both the pointer and the original allocation size.

## Zeroing and reallocation

`std::mem::alloc` implements families including:

- `mem_alloc`
- `mem_alloc_zeroed`
- `mem_realloc`
- `mem_free`
- Generic item allocation/reallocation/free helpers
- Page-count and page-alignment helpers
- Aligned allocation and free helpers

The current `mem_realloc` implementation allocates new storage, copies the necessary range, and frees the old storage. Check its exact failure behavior before building ownership logic around it.

## Size units

Major memory-size parameters use `i64` byte counts. Make units visible in calling code so element counts are not confused with byte counts.

```wave
let count: i64 = 32;
let elem_size: i64 = 4;
let bytes: i64 = count * elem_size;
```

## Pointer access

A manually allocated `ptr<u8>` does not carry bounds. Before accessing `deref p[index]`, the caller must ensure that the index remains inside the allocation.

## std::buffer

`std::buffer` builds buffer helpers on top of `std::mem`. Its APIs still require callers to respect allocation failure, length/capacity, ownership, and release contracts.

Keeping allocation and release in the same abstraction layer makes leaks and size mismatches easier to prevent and review.
