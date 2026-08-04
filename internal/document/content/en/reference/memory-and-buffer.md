---
translation_set_id: memory-buffer
path: reference/memory-and-buffer
locale: en
group: reference
group_order: 3
order: 5
title: Memory and buffers
summary: Manual allocation, copying, alignment, pages, and growable byte buffers.
---

## Manual allocation

```wave
import("std::mem::alloc");
import("std::mem::ops");

var memory: ptr<u8> = mem_alloc(256);
if (memory != null) {
    mem_zero(memory, 256);
    mem_free(memory, 256);
}
```

mem_alloc, mem_alloc_zeroed, mem_realloc, mem_free, page helpers, and aligned allocation expose explicit native memory. The caller retains the allocation size and must release storage with the matching function.

## Growable buffers

std::buffer is built on std::mem and separates allocation, read, write, and buffer types. Check null and error results before using returned storage.

> **Memory model**
> 
> ptr<T> in these APIs is the dedicated Wave Explicit Memory Type Model. It does not imply generic ownership, automatic bounds, or garbage collection.

