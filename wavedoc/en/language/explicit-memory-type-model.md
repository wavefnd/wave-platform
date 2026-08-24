---
translation_set_id: memory-model
path: language/explicit-memory-type-model
locale: en
group: language
group_order: 2
order: 7
title: Pointers and explicit memory access
summary: The Wave Explicit Memory Type Model, ptr<T>, arrays, address-of, deref, null, and manual-memory contracts.
---

## Wave Explicit Memory Type Model

Wave's pointer design is based on the **Wave Explicit Memory Type Model**. This model defines pointers and arrays as explicit, language-level memory types rather than syntactic tricks or library abstractions. A type such as `ptr<T>` states directly that a value is a memory address interpreted as holding a `T`, while `array<T, N>` states both the element type and fixed element count.

## ptr<T>

`ptr<T>` is the type of a memory address that points to a value of type `T`.

```wave
var value: i32 = 7;
var address: ptr<i32> = &value;
```

Pointer types can be nested when a memory address contains another pointer:

```wave
var value: i32 = 7;
var first: ptr<i32> = &value;
var second: ptr<ptr<i32>> = &first;
```

The type describes how the pointed-to memory is interpreted. The program remains responsible for ensuring that the address is valid for that use.

## null

```wave
var buffer: ptr<u8> = null;
if (buffer == null) {
    println("no buffer");
}
```

`null` represents a pointer that does not point to a value. It can be assigned only to `ptr<T>`. When an API uses `null` to report failure or the absence of a value, compare the pointer with `null` before using `deref`.

```wave
if (buffer != null) {
    deref buffer = 0;
}
```

## Explicit dereference

```wave
var value: i32 = 7;
var p: ptr<i32> = &value;
var copy: i32 = deref p;
deref p = 42;
```

`deref` reads or writes the value stored at a pointer's address. The target of a write must be assignable.

Indexed pointers can also be dereferenced explicitly:

```wave
deref bytes[index] = 0;
```

## Pointer casts

Use `as` when a low-level boundary requires an explicit conversion between a pointer and an address-sized integer, or between pointer types.

```wave
var raw: i64 = 0;
var p: ptr<u8> = raw as ptr<u8>;
```

Pointer conversions must account for the target's address width, alignment, and ABI.

## Arrays and pointers

Arrays and pointers express different memory shapes. `array<T, N>` contains `N` values of `T`; `ptr<T>` stores an address. They can be combined explicitly:

```wave
var left: i32 = 10;
var right: i32 = 20;
var pointers: array<ptr<i32>, 2> = [&left, &right];
var block: ptr<array<i32, 3>> = &[1, 2, 3];
```

`array<ptr<i32>, 2>` is an array containing two pointers. `ptr<array<i32, 3>>` is one pointer to an array containing three integers.

## Pointer arithmetic and comparison

Adding or subtracting an integer moves a `ptr<T>` by that many `T` elements. Subtracting two pointers returns their byte difference as an `i64`.

```wave
var base: ptr<i32> = 0x1000 as ptr<i32>;
var third: ptr<i32> = base + 3; // advances by 3 * the size of i32
var bytes: i64 = third - base;  // 12
```

Pointers can be compared with `==` and `!=`, including comparisons with `null`.

## What ptr<T> does not track automatically

The pointer type itself does not automatically track:

- Allocation ownership or deallocation responsibility
- Allocation size or index bounds
- The lifetime for which an address remains valid
- Alignment or initialization state
- Aliasing rules for concurrent or overlapping access

For FFI and manual allocation APIs, document who creates the memory, its size, how long it remains valid, which type may access it, and how it is released.
