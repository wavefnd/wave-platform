---
translation_set_id: whale-memory-model
path: whale/memory-model
locale: en
group: whale
group_order: 1
order: 8
title: Memory model
summary: Allocation tracking, reading initialized values, pointer arithmetic, layout and string storage rules.
---

## Tracked allocations

A tracked pointer associates an address with an allocation identity, generation, bounds, offset, and access permissions. The memory model also tracks allocation lifetime and initialization state. Access must satisfy the tracked conditions; a violation traps.

The generation distinguishes lifetimes even when a physical address is reused. An address alone does not establish that a pointer is valid or that the caller owns access to its storage.

Native addresses remain 64 bits. Shadow metadata accompanies pointers through copying, storage, calls, and returns. The initial native memory scope consists of trackable stack and global allocations. C boundaries require explicit adapters; arbitrary transfer of external-memory ownership is outside this scope.

## Initialization and reads

Declaring storage does not initialize its value. Check initialization for the byte range actually read. Reading an uninitialized part of a value traps. Do not replace the read with zero or an unspecified value.

Value-read checks exclude padding bytes. For example, after every field of a structure has been initialized, an uninitialized gap inserted for field alignment does not by itself make a value read invalid.

A memory copy propagates initialization state along with the bytes. Copying uninitialized storage does not turn it into initialized storage. A later value read at the destination is checked in the same way as a read at the source.

Physical zeros in BSS do not, on their own, establish that an IR variable has been initialized.

### Initialized scalar storage in IR

The following builder-produced module passes verification. Its `store` precedes the value read, and all three memory instructions specify a nonzero power-of-two alignment.

```text
module {
  format_version 4
  semantics_version 1
  target "x86_64-whale-linux"
  datalayout { ptr=64, endian=little }

  declare @f0 "initialized_local": whale () -> i32, linkage internal

  fn @f0 "initialized_local"() -> i32, entry %b0 {
  %b0 "entry":
    %v0: ptr<i32> = alloca i32, align 4
    %v1: i32 = const i32 42
    store i32 %v1, ptr<i32> %v0, align 4
    %v2: i32 = load i32, ptr<i32> %v0, align 4
    ret i32 %v2
  }

}
```

Save this complete module as `initialized.wir`; execution prints `i32 42`. Removing store traps at load on uninitialized storage. Alignment 3 is a verification error. Physical zero bytes from alloca do not establish initialization.

## Pointer arithmetic and comparison

Address-calculation overflow traps. A pointer one element past an allocation may be formed, but it must not be used to access memory.

Pointer equality uses allocation identity; the numerical address alone does not establish identity. Ordering and subtraction between pointers belonging to different allocations trap. Reconstructing an address from an integer does not restore access permissions.

### GEP

GEP computes an address in element and field units. It does not load a value.

The first index offsets in units of the base pointer's pointee type. Subsequent indices select array elements or structure/tuple fields. A structure or tuple field index is a compile-time field ordinal, not a byte offset. The selected type determines the result pointer type.

For a base of type `ptr<array<i32, 4>>`, indices `[0, 2]` select the third i32 element of that array and produce `ptr<i32>`. A first index of 1 instead advances by one entire four-element array. These examples describe indices, not textual instruction syntax.

Pointer arithmetic on zero-sized elements is rejected in the initial native target. Calculating an address does not remove the lifetime, bounds, initialization, or permission checks on a later access.

## Data layout

The output target determines sizes, alignments, field offsets, and array strides. Structure and tuple fields retain their declared order. Do not derive target layout from the machine on which the compiler happens to run.

| Value | Storage rule |
| --- | --- |
| Bool, signed i1, unsigned u1 | At least one byte |
| Empty structure or tuple | Size zero, alignment one |
| Array | Element stride determined by target layout |
| Structure or tuple | Ordered fields with target-required alignment |

Alignment in completed IR is a nonzero power of two. Resolve automatic alignment before producing that IR. Packed layouts, unions, and bitfields are unsupported in this profile and must be rejected.

### Querying the output layout

The Rust API computes storage layout independently of the build host. This example has seven padding bytes before the u64 field and six trailing padding bytes:

```rust
use ir::{allocation_align, layout_of, Target, Type};

fn main() {
    let target = Target::X86_64WhaleLinux;
    let record = Type::Struct(vec![Type::U8, Type::U64, Type::U16]);
    let layout = layout_of(&record, target).unwrap();
    assert_eq!((layout.size, layout.align), (24, 8));
    assert_eq!(layout.field_offsets, [0, 8, 16]);

    let array = Type::Array(Box::new(record), 3);
    let layout = layout_of(&array, target).unwrap();
    assert_eq!((layout.size, layout.align), (72, 8));
    assert_eq!(layout.element_stride, Some(24));
    assert_eq!(allocation_align(&array, target).unwrap(), 16);
}
```

```text
struct{u8, u64, u16}: size 24, natural alignment 8
field 0: byte 0
field 1: byte 8
field 2: byte 16
array of 3: size 72, element stride 24
standalone array placement alignment: 16
```

`layout_of` includes tail padding in size and array stride. Structs and tuples use the same ordered-field rules. Bool, i1, and u1 each occupy one byte. Empty structs and tuples have size zero and alignment one; a zero-length array retains its element's natural alignment. `void` has no storage layout, while `ptr<void>` occupies eight bytes.

Natural alignment governs fields and array elements. `allocation_align` applies the SysV AMD64 requirement that a standalone local or global array of at least 16 bytes has at least 16-byte alignment. It does not increase an array field's alignment or element stride. AST lowering uses this allocation query for local storage.

Size multiplication, field offset addition, and padding overflow return `LayoutError::Overflow`. `layout_of` limits aggregate nesting to 128 levels; `layout_of_with_limit` accepts a caller-selected budget. Neither query allocates storage for array elements. `pointer_stride` rejects a zero-sized pointee for native pointer arithmetic, although its storage layout remains valid. Packed layouts, unions, and bitfields have no supported type representation. These storage queries do not implement aggregate calling conventions or runtime bounds checks.

## Strings and C boundaries

A string is an immutable byte sequence with an explicit length. UTF-8 is the default encoding. Embedded NUL bytes are allowed; there is no implicit terminating NUL. O0 does not automatically merge equal string objects.

The three-byte sequence `A`, NUL, `B` therefore has length three. It cannot be passed through the explicit C-string conversion, which rejects embedded NUL. A frontend must not silently truncate it to `A`.

External C, raw addresses, and inline assembly are separate contract boundaries. Runtime checking of tracked memory does not guarantee detection of every invalid action performed by external code.

## Tracked stack memory execution

The default interpreter executes integer/Bool values, control flow, stack allocations, data-pointer storage and loads, typed GEP, memcpy and memset. Checked-pair reads exclude padding. Addresses are synthetic 64-bit values, never host-memory dereferences. Function arguments and returns remain integer/Bool or void; float, calls, function pointers, general aggregate values, global addresses and native execution are unsupported. Stack allocations live until function return. Lexical lifetime-end operations, pointer transfer through calls/returns, foreign-memory adapters and native shadow metadata require further implementation.

`--max-memory` sets the logical allocation-byte budget, default 64 MiB. `InterpreterOptions::memory_limits` also limits allocation count to 16384, pointer-byte metadata to 262144 fragments, and byte/metadata work to 256 Mi units. Exceeding a budget returns an IR-located `MemoryLimit` error, separately from a program trap. Verification also rejects known output-target storage-size overflow before execution.

```shell
whale ir run initialized.wir --function @f0
```

```text
i32 42
```

## Byte copies and initialization reset

Save the following complete module as `tracked-memory.wir`. @f0 copies pointer bytes and separate metadata, then reads 42. In @f1, uninit marks the type storage range uninitialized and clears pointer metadata without changing existing bytes. memcpy can copy uninitialized bytes; a later destination value read checks their state. Split pointer copies preserve authority only when all eight consistent byte metadata fragments are present. Writing identical bits with integers or memset does not recover authority.

Zero-length memcpy/memset succeeds without access or alignment checks, including null and one-past pointers. Nonempty overlapping memcpy traps. memset initializes written bytes and clears their pointer metadata. Bool storage must be 0 or 1; reading another representation traps.

```text
module {
  format_version 4
  semantics_version 1
  target "x86_64-whale-linux"
  datalayout { ptr=64, endian=little }

  declare @f0 "copied_pointer": whale () -> u32, linkage internal
  declare @f1 "uninitialized": whale () -> u32, linkage internal

  fn @f0 "copied_pointer"() -> u32, entry %b0 {
  %b0 "entry":
    %v0: ptr<u32> = alloca u32, align 4
    %v1: u32 = const u32 42
    store u32 %v1, ptr<u32> %v0, align 4
    %v2: ptr<ptr<u32>> = alloca ptr<u32>, align 8
    %v3: ptr<ptr<u32>> = alloca ptr<u32>, align 8
    store ptr<u32> %v0, ptr<ptr<u32>> %v2, align 8
    %v4: ptr<u8> = bitcast ptr<ptr<u32>> %v2 to ptr<u8>
    %v5: ptr<u8> = bitcast ptr<ptr<u32>> %v3 to ptr<u8>
    %v6: u64 = const u64 8
    memcpy ptr<u8> %v5, ptr<u8> %v4, u64 %v6, align 8
    %v7: ptr<u32> = load ptr<u32>, ptr<ptr<u32>> %v3, align 8
    %v8: u32 = load u32, ptr<u32> %v7, align 4
    ret u32 %v8
  }

  fn @f1 "uninitialized"() -> u32, entry %b0 {
  %b0 "entry":
    %v0: ptr<u32> = alloca u32, align 4
    uninit u32, ptr<u32> %v0, align 4
    %v1: u32 = load u32, ptr<u32> %v0, align 4
    ret u32 %v1
  }

}
```

```shell
whale ir run tracked-memory.wir --function @f0
whale ir run tracked-memory.wir --function @f1
```

```text
u32 42
Error: tracked-memory.wir: trap at @f1 %b0 instruction 2 (%v1): uninitialized byte at allocation offset 0 (after 3 steps)
```

## Verification, lowering and execution obligations

| Stage | Obligation |
| --- | --- |
| Verifier | Check operands/pointee types, alignment shape and output-target storage sizes; reject undef. |
| O0 lowering | Keep uninit at the executed declaration, store initializers and preserve reads. |
| Runtime | Check allocation identity, generation/lifetime, range, permissions, alignment, address overflow and initialization; transfer copy state. |
