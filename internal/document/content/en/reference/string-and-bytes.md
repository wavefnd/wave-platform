---
translation_set_id: string-bytes
path: reference/string-and-bytes
locale: en
group: reference
group_order: 3
order: 4
title: Strings and bytes
summary: String length, comparison, searching, hashing, ASCII helpers, and endian byte operations.
---

## String modules

```wave
import("std::string::len");
import("std::string::cmp");
import("std::string::find");
import("std::string::trim");

var size: i32 = len("Wave");
var empty: bool = is_empty("");
```

The release separates string operations by role: len, cmp, find, hash, trim, and ascii. Import the concrete module that declares the function you use.

## Byte order

std::bytes::endian provides endian swap, load, and store helpers for binary formats and protocols. Always pair a byte operation with the width and byte order required by the external format.

