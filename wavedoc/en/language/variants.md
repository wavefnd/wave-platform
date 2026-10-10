---
translation_set_id: language-variants
path: language/variants
locale: en
group: language
group_order: 2
order: 17
title: Variants and pattern matching
summary: Store case-specific data together and safely separate from match.
---

## Distinguished from enum

enum represents a named integer value, and variant holds payload, which is different in each case. Result below stores the value i32 in case of success and the error number i32 in case of failure. Even if the error and result are of the same integer type, their meaning can be differentiated by name.

## Declaration, creation, inspection

Save it to `main.wave` and run it.

<!-- wave-example: variant-api -->
```wave playground
variant Result {
    Value(i32), Error(i32)
}

fun calculate(valid: bool) -> Result {
    if (!valid) {
        return Result::Error(1);
    }

    return Result::Value(42);
}

fun main() {
    var result: Result = calculate(true);
    match (result) {
        Result::Value(value) => {
            println("value={}", value);
        }
        Result::Error(code) => {
            println("error={}", code);
        }
    }
}
```

Execution result:

```text
value=42
```

Makes the case that `Result::Value(42)` has payload. Use payload under the name value only within the corresponding pattern of `match`. It does not force reading of payload in other cases. arm The text is written in blocks. Handle all cases or the rest with `_`. It is better to explicitly break down each case to reveal whether processing is needed when a new case is added.

## Generics and longevity

Type parameters can be used as in `variant Optional<T> { Some(T), None }`. Specify a concrete type for local variables, such as `Optional<i32>`. A variant containing a pointer does not automatically manage memory ownership. Copying the value does not duplicate the pointed-to allocation.

Do not assume that the memory representation of variant is identical to any arbitrary C union. For data to be sent externally ABI, a separate expression is determined and delivered as the permitted FFI type.

## practice

If you change it to calculate(false), you should get `error=1`. Try adding the case Empty without payload, and then handle that case in match as well.

[Structure and enum](/docs/en/language/structures-enums-and-aliases) · [Error handling class](/docs/en/language/errors)
