---
translation_set_id: memory-model
path: language/explicit-memory-type-model
locale: en
group: language
group_order: 2
order: 7
title: Pointers and explicit memory access
summary: ptr<T>, address-of, deref, null, pointer casts, and the contracts required around manual memory.
---

## ptr<T>

`ptr<T>` is a pointer type intended to address a `T` value.

```wave
var value: i32 = 7;
var address: ptr<i32> = &value;
```

The pointee type records the intended interpretation of the address; it does not prove that the address is valid.

## null

```wave
var buffer: ptr<u8> = null;
if (buffer == null) {
    println("no buffer");
}
```

Code generation treats `null` as a pointer value. Many APIs use a null pointer to report failure, so check it before dereferencing when required by the API contract.

## Explicit dereference

```wave
var value: i32 = 7;
var p: ptr<i32> = &value;
var copy: i32 = deref p;
deref p = 42;
```

`deref` performs the actual memory read or write through a pointer. A write target must be assignable.

Indexed pointers can also be dereferenced explicitly:

```wave
deref bytes[index] = 0;
```

## Pointer casts

Use `as` when a low-level boundary requires changing the pointer or address representation.

```wave
var raw: i64 = 0;
var p: ptr<u8> = raw as ptr<u8>;
```

Integer/pointer conversions should be limited to low-level boundaries and must account for the target's address width and ABI.

## What ptr<T> does not track automatically

The pointer type itself does not automatically track:

- Allocation ownership or deallocation responsibility
- Allocation size or index bounds
- The lifetime for which an address remains valid
- Alignment or initialization state
- Aliasing rules for concurrent or overlapping access

For FFI and manual allocation APIs, document who creates the memory, its size, how long it remains valid, which type may access it, and how it is released.
