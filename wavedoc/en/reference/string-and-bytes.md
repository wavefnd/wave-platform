---
translation_set_id: string-bytes
path: reference/string-and-bytes
locale: en
group: reference
group_order: 3
order: 4
title: Strings and bytes
summary: String submodules, len/is_empty, comparison, search, trimming, ASCII helpers, and endian-aware byte operations.
---

## String module layout

`std::string` contains these source units:

- `ascii.wave`
- `cmp.wave`
- `find.wave`
- `hash.wave`
- `len.wave`
- `trim.wave`

Import the source unit that defines the operations you need.

## Length and emptiness

`std::string::len` defines both `len` and `is_empty`.

```wave
import("std::string::len")::{len, is_empty};

fun main() {
    var size: i32 = len("Wave");
    var empty: bool = is_empty("");
    println("{} {}", size, empty);
}
```

`len(s)` returns the number of values before the string's terminating zero as an `i32`. `is_empty(s)` is true when the first value is the terminating zero.

## Comparison, searching, and trimming

```wave
import("std::string::cmp")::{eq, cmp, starts_with, ends_with};
import("std::string::find")::{find, contains};
import("std::string::trim")::{trim_range};
```

Use `wavec print std-path` to locate the standard-library source and read each module's public signatures.

## ASCII helpers

`std::string::ascii` provides byte-level ASCII classification and conversion. These operations follow ASCII rules rather than Unicode text rules.

## Byte order

`std::bytes` provides modules for reading/writing binary values and handling endianness. When implementing a file format or network protocol, distinguish host byte order from the byte order required by the external format.

Keeping integer widths and offsets explicit makes binary-format code easier to review across targets.
