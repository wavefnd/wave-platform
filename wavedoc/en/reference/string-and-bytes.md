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

The current `std/string` tree contains these source units:

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
import("std::string::len");

fun main() {
    var size: i32 = len("Wave");
    var empty: bool = is_empty("");
    println("{} {}", size, empty);
}
```

The current `len` implementation indexes the string until it encounters a zero value and returns an `i32` count. Code using this API should therefore follow the release's `str` representation contract.

## Comparison, searching, and trimming

```wave
import("std::string::cmp");
import("std::string::find");
import("std::string::trim");
```

Use the installed standard-library source for exact function names and return contracts. `wavec print std-path` prints the current `std` location.

## ASCII helpers

`std::string::ascii` provides helpers for byte-level ASCII classification and conversion. Do not assume that these APIs provide full Unicode text processing.

## Byte order

`std::bytes` provides modules for reading/writing binary values and handling endianness. When implementing a file format or network protocol, distinguish host byte order from the byte order required by the external format.

Keeping integer widths and offsets explicit makes binary-format code easier to review across targets.
