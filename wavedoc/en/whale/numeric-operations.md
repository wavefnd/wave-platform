---
translation_set_id: whale-numeric-operations
path: whale/numeric-operations
locale: en
group: whale
group_order: 1
order: 7
title: Numeric operations
summary: Describes integer wrap, checked operations, shift, type conversion errors, and floating point results.
---

## Integer representation

An N-bit integer has N value bits. Unsigned integers range from 0 to 2^N − 1; signed integers range from −2^(N−1) to 2^(N−1) − 1. The signedness of an operation determines how the bit pattern is interpreted.

The tables below describe operation results. The IR code example shows the current printer representation.

## Addition, subtraction, and multiplication

Basic integer add, sub, and mul keep the low N bits of the result. Overflow does not trap. Checked arithmetic returns the same wrapped result together with a Bool indicating whether the mathematical result exceeded the signed or unsigned range of the operation.

| Operation | Wrapped result | Checked overflow |
| --- | --- | --- |
| u8: 255 + 1 | 0 | true |
| i8: 127 + 1 | −128 | true |
| u8: 0 − 1 | 255 | true |
| i8: 12 × 3 | 36 | false |

A frontend that requires overflow to terminate execution must use checked arithmetic and an explicit `trap_if` on the overflow result. Basic arithmetic does not inherit the source language's overflow policy implicitly.

### Wrapping and explicitly checked IR

Save this complete format 3 module as `integer-operations.wir`; it is accepted by the text reader and executable by the scalar interpreter.

```text
module {
  format_version 4
  semantics_version 1
  target "x86_64-whale-linux"
  datalayout { ptr=64, endian=little }

  declare @f0 "add_u8": whale () -> u8, linkage internal
  declare @f1 "require_no_overflow": whale () -> u8, linkage internal

  declare @f2 "minimum_division": whale () -> i8, linkage internal
  declare @f3 "minimum_remainder": whale () -> i8, linkage internal
  declare @f4 "negative_shift": whale () -> i8, linkage internal

  fn @f0 "add_u8"() -> u8, entry %b0 {
  %b0 "entry":
    %v0: u8 = const u8 255
    %v1: u8 = const u8 1
    %v2: u8 = add u8 %v0, %v1
    ret u8 %v2
  }

  fn @f1 "require_no_overflow"() -> u8, entry %b1 {
  %b1 "entry":
    %v3: u8 = const u8 255
    %v4: u8 = const u8 1
    %v5: tuple<u8, bool> = uadd_chk u8 %v3, %v4
    %v6: u8 = extract %v5, 0
    %v7: bool = extract %v5, 1
    trap_if bool %v7, reason="integer overflow"
    ret u8 %v6
  }

  fn @f2 "minimum_division"() -> i8, entry %b0 {
  %b0 "entry":
    %v0: i8 = const i8 -128
    %v1: i8 = const i8 -1
    %v2: i8 = sdiv i8 %v0, %v1
    ret i8 %v2
  }

  fn @f3 "minimum_remainder"() -> i8, entry %b0 {
  %b0 "entry":
    %v0: i8 = const i8 -128
    %v1: i8 = const i8 -1
    %v2: i8 = srem i8 %v0, %v1
    ret i8 %v2
  }

  fn @f4 "negative_shift"() -> i8, entry %b0 {
  %b0 "entry":
    %v0: i8 = const i8 -1
    %v1: i8 = const i8 -1
    %v2: i8 = shl i8 %v0, %v1
    ret i8 %v2
  }

}
```

`add_u8` returns 0: the low eight bits of 256. `minimum_division` returns −128, `minimum_remainder` returns 0, and `negative_shift` interprets −1 as the unsigned count 255, then uses 255 mod 8 = 7. Execute the functions by explicit ID:

```shell
whale ir run integer-operations.wir --function @f0
whale ir run integer-operations.wir --function @f2
whale ir run integer-operations.wir --function @f3
whale ir run integer-operations.wir --function @f4
```

```text
u8 0
i8 -128
i8 0
i8 -128
```

`require_no_overflow` extracts the wrapped result at index 0 and the Bool overflow flag at index 1. Its explicit trap_if stops before return. The following command exits with status 1 and writes the diagnostic to stderr:

```shell
whale ir run integer-operations.wir --function @f1
```

```text
Error: integer-operations.wir: trap at @f1 %b1 instruction 5: "integer overflow" (after 6 steps)
```

## Division and remainder

Integer division and remainder by zero trap. Signed division of the minimum representable value by −1 produces that minimum value by wrapping. The corresponding remainder is zero.

| Operation | Result |
| --- | --- |
| i8: −128 / −1 | −128 |
| i8: −128 % −1 | 0 |
| Integer division by 0 | trap |
| Integer remainder by 0 | trap |

## Shifts

For an N-bit value, interpret the shift-count bit pattern as unsigned and reduce it modulo N. A count outside the range 0 through N−1 does not itself trap.

For an 8-bit value, counts 0, 8, and 16 all select a shift of zero. An 8-bit count with bit pattern `11111111` selects a shift of 7, including when that pattern represents signed −1. This unsigned interpretation occurs before the modulo operation.

A source language that rejects excessive or negative counts must express that policy with explicit checks before the shift.

## Conversions

| Conversion | Meaning |
| --- | --- |
| Zero extension | Increase width by adding zero high bits |
| Sign extension | Increase width by replicating the sign bit |
| Bit truncation | Retain the low bits at the destination width |
| Bit reinterpretation | Interpret the same bits using another type |
| Lossless numerical conversion | Preserve the numerical value; trap when it cannot be represented |

For example, extending the bit pattern `11111111` from 8 to 16 bits by zero extension produces `0000000011111111`. Sign extension produces `1111111111111111`. These are different operations even when the source bits are identical.

Float-to-integer conversion truncates toward zero, then checks the integer range. NaN and infinity trap. For conversion to i8, 127.9 produces 127, while 128.0 traps. Bool converts to integer 0 or 1, except that signed i1 is not a permitted destination for this conversion because it cannot represent 1.

Converting an address to an integer does not preserve a right to recover pointer access permissions from that integer. See [pointer validity](memory-model).

### Verified cast and checked-operation shapes

Verification checks the actual operand type against `src_ty`, then checks the opcode's allowed source/destination categories and widths. A result annotation must also match its definition. There are no implicit conversions.

| Opcode | Accepted types |
| --- | --- |
| `zext`, `sext` | Non-Bool integers; destination width strictly larger |
| `zext` from Bool | Integer 0/1 conversion; every integer destination except signed `i1`, including `u1` |
| `trunc` | Non-Bool integers; destination width strictly smaller |
| `fext`, `ftrunc` | Floats; strictly larger / strictly smaller destination width |
| `itof_s`, `itof_u` | Signed / unsigned integer to float |
| `ftoi_s`, `ftoi_u` | Float to signed / unsigned integer |
| `bitcast` | Equal-width integer/float scalars, or data pointer to data pointer |
| `ptrtoint`, `inttoptr` | Data pointer to integer / integer to data pointer |

Integer extension and truncation are bit operations; their integer operands may differ in signedness. Bool is a separate logical type: it cannot be sign-extended, truncated or bitcast as `i1`/`u1`. Aggregate and function-pointer casts are rejected. Pointer/integer category validation does not establish a valid allocation or restore permissions. Runtime conversion checks and machine lowering remain separate work.

This complete Rust program emits two valid conversions, then confirms that replacing `zext` with `fext` is rejected:

```rust
use ir::*;

fn main() {
    let mut builder = ModuleBuilder::new("x86_64-whale-linux", DataLayout::default_64bit_le());
    let mut function = builder.begin_function("widen", vec![("byte".into(), Type::U8)], Type::Void);
    function.ret(None);
    function.finish();
    let mut module = builder.finish();
    let function = &mut module.functions[0];
    function.value_types.extend([(ValueId(1), Type::U32), (ValueId(2), Type::I32)]);
    function.blocks[0].instructions.extend([
        Instruction::Cast {
            dst: ValueId(1), op: CastOp::ZExt,
            src_ty: Type::U8, src: ValueId(0), dst_ty: Type::U32,
        },
        Instruction::Cast {
            dst: ValueId(2), op: CastOp::Bitcast,
            src_ty: Type::U32, src: ValueId(1), dst_ty: Type::I32,
        },
    ]);
    verify_module(&module).unwrap();
    print!("{}", print_module(&module));
    // An integer source cannot be annotated as a floating widening operation.
    if let Instruction::Cast { op, .. } = &mut module.functions[0].blocks[0].instructions[0] {
        *op = CastOp::FExt;
    }
    assert!(matches!(verify_module(&module), Err(VerifyError::InvalidCast { .. })));
}
```

```text
module {
  format_version 4
  semantics_version 1
  target "x86_64-whale-linux"
  datalayout { ptr=64, endian=little }

  declare @f0 "widen": whale (u8) -> void, linkage internal

  fn @f0 "widen"(%v0 "byte": u8) -> void, entry %b0 {
  %b0 "entry":
    %v1: u32 = zext u8 %v0 to u32
    %v2: i32 = bitcast u32 %v1 to i32
    ret void
  }

}
```

The following instruction fragments are deliberately invalid:

```text
%v1: i32 = ftoi_s f64 %v0 to i32   // Invalid when %v0 is actually Bool.
%v2: i16 = zext i32 %v3 to i16     // zext cannot narrow.
%v4: i1 = zext bool %v5 to i1      // Signed i1 cannot represent true as 1.
```

A mismatched actual source yields `OperandTypeMismatch`; an illegal pair yields `InvalidCast` with the opcode and both types. Neither error inserts a conversion or changes the module.

`sadd_chk`, `ssub_chk`, and `smul_chk` require signed integer operands; `uadd_chk`, `usub_chk`, and `umul_chk` require unsigned integer operands. Both operands must exactly match the operation type `T`, and the result must be `tuple<T, bool>`. Extraction requires a tuple, an existing field index, and that field's exact type. In particular, extracting the overflow field as `i1` is rejected. The wrapping/overflow results described above remain the execution contract; these tests establish structural verification, not execution by an implemented backend.

## Floating-point arithmetic

Floating-point values have exact f16, f32, or f64 bit patterns. Arithmetic rounds at the declared width using round-to-nearest, ties-to-even: a halfway result selects the representable value with an even least-significant significand bit.

Default arithmetic does not permit fast math, implicit fused multiply-add, or flushing small values to zero. A multiply followed by an add retains its separate rounding steps; the backend must not combine them implicitly.

Numerical operations may produce NaN or infinity. NaNs produced by numerical operations or width conversions are normalized to one fixed positive quiet NaN per width. Storage and copying instead preserve the original NaN bits. This distinction matters when moving a NaN payload through memory without performing arithmetic on it.

Floating-point status flags are not exposed. Float-to-integer conversion has the trapping rules above even though default floating-point arithmetic permits NaN and infinity results.


### Exact constant storage

Use `FloatBits` variants or parse an exact-width hexadecimal bit string. Storage equality compares bits, including signed zero and NaN payloads. The verifier rejects a payload width that differs from its IR type.

```rust
use ir::{FloatBits, ModuleBuilder, Target, Type};
fn main() {
    let bits = FloatBits::parse(32, "0xffc01234").unwrap();
    assert_eq!(bits, FloatBits::F32(0xffc01234));
    let target = Target::X86_64WhaleLinux;
    let mut module = ModuleBuilder::new(target.name(), target.data_layout());
    let mut function = module.begin_function("payload", vec![], Type::F32);
    let value = function.const_float_bits(Type::F32, bits);
    function.ret(Some(value));
    function.finish();
    let module = module.finish();
    ir::verify_module(&module).unwrap();
    assert!(ir::print_module(&module).contains("const f32 0xffc01234"));
    println!("{}", bits);
}
```

```text
0xffc01234
```

f16/f32/f64 strings have exactly 4/8/16 hex digits after `0x`. `0x80000000` is f32 negative zero; `0x7f800000` is positive infinity. `const_float` is a numeric convenience conversion from host f64; use `const_float_bits` to preserve original storage. Bit-exact storage does not complete the floating arithmetic backend: compile-time arithmetic still uses host f64 intermediates, so the complete declared-width rounding contract remains unfinished.
