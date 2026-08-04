---
translation_set_id: memory-model
path: language/explicit-memory-type-model
locale: en
group: language
group_order: 2
order: 7
title: Wave Explicit Memory Type Model
summary: Pointers, explicit dereferencing, pointer operations, and native-memory safety rules.
---

Wave makes low-level memory access visible in source code. A pointer is not a value of the pointee type: pointer creation, address movement, conversion, and dereferencing are distinct operations. This separation is the Wave Explicit Memory Type Model.

> **Dedicated memory type syntax**
> 
> ptr<T> is defined directly by the Wave Explicit Memory Type Model. Its T slot describes the memory element type; it is not a general generic type argument and ptr is not a user-defined generic.

## Pointer types

```wave
var raw: u64 = 0;
var typed: ptr<i32> = raw as ptr<i32>;
```

ptr<T> records the intended pointee type for a native address. A typed pointer still does not establish allocation lifetime, alignment, initialization, or ownership.

```wave
var value: i32 = 7;
var address: ptr<i32> = &value;
```

The address-of operator & creates a pointer to existing storage. The pointer must not outlive that storage.

## Explicit dereference

```wave
var value: i32 = deref typed;
deref typed = 42;
```

deref is the explicit transition from an address-bearing value to a memory access. Before dereferencing, the program must ensure the pointer is non-null, aligned for T, points to initialized storage, remains live for the access, and permits the requested read or write.

## Pointer movement

Pointer arithmetic operates on addresses and must remain within the storage described by the external allocation contract. Do not infer ownership or bounds merely from ptr<T>.

> **Native memory**
> 
> The compiler cannot recover an allocation contract that was lost across FFI. Keep allocation, size, ownership, and release rules together in a narrow wrapper.

## Boundary pattern

- Receive or allocate memory through a documented native function.
- Check failure and null results before conversion.
- Convert to the narrowest typed pointer needed.
- Perform explicit dereference only while the lifetime is valid.
- Release memory with the allocator that created it.
