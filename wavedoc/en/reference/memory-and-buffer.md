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
import("std::mem::alloc")::{mem_alloc, mem_free};
import("std::mem::ops")::{mem_zero};

fun main() {
    var size: i64 = 256;
    var memory: ptr<u8> = mem_alloc(size);

    if (memory != null) {
        mem_zero(memory, size);
        mem_free(memory, size);
    }
}
```

`mem_alloc(size)` returns `ptr<u8>` and may return `null` on failure. `mem_free` receives both the pointer and the original allocation size.

## Zeroing and reallocation

`std::mem::alloc` provides:

- `mem_alloc`
- `mem_alloc_zeroed`
- `mem_realloc`
- `mem_free`
- Generic item allocation/reallocation/free helpers
- Page-count and page-alignment helpers
- Aligned allocation and free helpers

`mem_realloc(old_ptr, old_size, new_size)` returns storage sized to `new_size` and preserves up to the smaller of the old and new sizes. A non-positive new size frees valid old storage and returns `null`. If a new allocation fails, it returns `null` and leaves the old allocation available to its owner.

## Size units

Major memory-size parameters use `i64` byte counts. Make units visible in calling code so element counts are not confused with byte counts.

```wave
var count: i64 = 32;
var elem_size: i64 = 4;
var bytes: i64 = count * elem_size;
```

## Pointer access

A manually allocated `ptr<u8>` does not carry bounds. Before accessing `deref p[index]`, the caller must ensure that the index remains inside the allocation.

## std::buffer

`std::buffer` builds buffer helpers on top of `std::mem`. Its APIs still require callers to respect allocation failure, length/capacity, ownership, and release contracts.

Keeping allocation and release in the same abstraction layer makes leaks and size mismatches easier to prevent and review.
