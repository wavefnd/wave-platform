---
translation_set_id: whale-numeric-operations
path: whale/numeric-operations
locale: zh
group: whale
group_order: 1
order: 7
title: 数值运算
summary: 描述整数 wrap、checked 运算、移位、类型转换错误和浮点结果。
---

## 整数表示

N 位整数具有 N 值位。无符号整数的范围是 0 到 2^N − 1，有符号整数的范围是 −2^(N−1) 到 2^(N−1) − 1。运算的 signedness 决定了位串的解释。

下表解释了计算结果。 IR 代码示例使用当前打印机的表示形式。

## 加法、减法、乘法

基本整数 add·sub·mul 保存结果的低 N 位。 overflow 不会导致 trap。 checked运算返回相同的wrap结果，以及Bool，指示数学结果是否超出相应的signed或unsigned的范围。

|操作|Wrap 结果| Checked overflow |
| --- | --- | --- |
| u8: 255 + 1 | 0 | true |
| i8: 127 + 1 | −128 | true |
| u8: 0 − 1 | 255 | true |
| i8: 12 × 3 | 36 | false |

在 overflow 处中断执行的语言的前端必须对 checked 操作的 overflow 结果使用显式 `trap_if`。默认操作不会隐式应用源语言的overflow策略。

### Wrap 和 IR 表示显式检查

将此完整格式 3 模块保存为 `integer-operations.wir`。文本读取器接受它，标量解释器可以执行它。

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

`add_u8` 返回 256 的低 8 位，即 0。`minimum_division` 返回 −128，`minimum_remainder` 返回 0。`negative_shift` 将 −1 解释为 unsigned count 255，再取 255 mod 8 = 7。用显式函数 ID 执行：

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

`require_no_overflow` 从索引 0 提取回绕结果，从索引 1 提取 Bool overflow 标志。显式 trap_if 在返回前停止。以下命令以状态 1 退出，并将诊断写到 stderr：

```shell
whale ir run integer-operations.wir --function @f1
```

```text
Error: integer-operations.wir: trap at @f1 %b1 instruction 5: "integer overflow" (after 6 steps)
```

## 除法和余数

整数除以零或余数会导致陷阱。将最小有符号值除以 -1 即可得到最小值。在这种情况下余数为零。

|操作|结果|
| --- | --- |
| i8: −128 / −1 | −128 |
| i8: −128 % −1 | 0 |
|除以整数 0| trap |
|整数 0 的余数| trap |

## 转变

在移位N位值时，count的位串被解释为unsigned，并且使用除以N的余数。它不会仅仅因为 count 超出 0 到 N−1 的范围而生成 trap。

在 8 位值中，count 0·8·16 均为 0 位移位。 8 位count 的位串`11111111` 移位了 7 位。即使该位串表示signed-1，这也是相同的。 unsigned 分析在其余计算之前进行。

拒绝负数或过多count的源语言必须在转换前进行显式检查。

## 类型转换

|转换|意义|
| --- | --- |
| Zero extension |通过用 0 填充高位来增加宽度|
| Sign extension |通过复制符号位来增加宽度|
|钻头切削|仅保留目标宽度对应的低位|
|节拍重新诠释|将相同的位串解释为不同的类型|
|无损数值转换|如果无法在保留数值的情况下表达，trap|

例如，将 8 位 `11111111` 转换为 16 位 zero extension 变为 `0000000011111111`，sign extension 变为 `1111111111111111`。即使输入位相同，也是不同的操作。

float 到 int 的转换会截断为零，然后检查整数范围。 NaN 和无穷大会导致陷阱。当转换为i8时，127.9变成127，而128.0陷入困境。 Bool 转换为整数 0 或 1。有符号 i1 不能表示 1，因此它不能作为此转换的目标。

将地址转换为整数不会恢复对该整数的指针访问。请参阅[指针有效性](memory-model)。

### cast 与 checked 运算的验证形式

验证先检查实际操作数类型是否与 `src_ty` 一致，再检查 opcode 允许的源类型、目标类型和位宽。结果的类型注解也必须与定义一致。不存在隐式转换。

| Opcode | 允许的类型 |
| --- | --- |
| `zext`, `sext` | 非 Bool 整数；目标位宽必须更大 |
| `zext` (Bool) | Bool 到整数 0/1；允许除有符号 `i1` 外的所有整数目标，包括 `u1` |
| `trunc` | 非 Bool 整数；目标位宽必须更小 |
| `fext`, `ftrunc` | 浮点数；目标位宽必须更大 / 更小 |
| `itof_s`, `itof_u` | 有符号 / 无符号整数到浮点数 |
| `ftoi_s`, `ftoi_u` | 浮点数到有符号 / 无符号整数 |
| `bitcast` | 相同位宽的整数或浮点标量，或数据指针到数据指针 |
| `ptrtoint`, `inttoptr` | 数据指针到整数 / 整数到数据指针 |

整数扩展和截断是位运算，因此整数操作数的有符号性可以不同。Bool 是独立的逻辑类型，不允许符号扩展、截断或到 `i1`/`u1` 的 bitcast。聚合类型和函数指针的 cast 会被拒绝。检查指针和整数的类别不会建立有效分配，也不会恢复访问权限。运行时转换检查和机器码 lowering 仍是单独的未完成工作。

下面的完整 Rust 程序输出两个有效转换，并确认将 `zext` 替换为 `fext` 后会被拒绝。

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

以下指令片段是故意无效的 IR。

```text
%v1: i32 = ftoi_s f64 %v0 to i32   // Invalid when %v0 is actually Bool.
%v2: i16 = zext i32 %v3 to i16     // zext cannot narrow.
%v4: i1 = zext bool %v5 to i1      // Signed i1 cannot represent true as 1.
```

实际源类型不匹配会产生 `OperandTypeMismatch`；非法类型组合会产生包含 opcode 和两种类型的 `InvalidCast`。验证错误不会插入转换，也不会修改模块。

`sadd_chk`、`ssub_chk`、`smul_chk` 要求有符号整数；`uadd_chk`、`usub_chk`、`umul_chk` 要求无符号整数。两个操作数都必须与运算类型 `T` 完全相同，结果必须是 `tuple<T, bool>`。提取要求 tuple、存在的字段索引以及该字段的精确类型。尤其不能将溢出字段提取为 `i1`。上文的回绕和溢出结果是执行契约；这些测试验证结构正确性，并不表示已有后端执行这些运算。

## 浮点运算

浮点值具有精确的位串f16·f32·f64。该操作四舍五入到声明宽度中最接近的值，如果正好在中间，则使用ties-to-even，它选择具有有效数字的最低有效位的值。

默认操作是fast-math，隐式FMA，不允许对小值进行强制清零处理。乘法后加法保留它们各自的舍入步骤，并且后端不应隐式地将它们组合成一个运算。

数值运算的结果可以是 NaN 或无穷大。来自数学运算和宽度变化的 NaN 被标准化为每个宽度一个固定数量的安静 NaN。保存并复制保留原始 NaN 位。因此，在不进行算术运算的情况下将 NaN 有效负载传递到内存时和计算它时，行为是不同的。

它不公开浮点状态标志。尽管基本浮点运算允许 NaN·无穷大结果，但 float→int 转换应用上述 trap 规则。


### 存储精确的常量

使用 `FloatBits` variant 或任何精确宽度的十六进制位字符串。存储值的相等性与包含负0和NaNpayload的位串进行比较。如果类型IR和payload的宽度不同，验证者会拒绝。

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

f16、f32 或f64 位串以 `0x` 开头，后面分别紧跟 4、8 或 16 个十六进制数字。 `0x80000000`代表f32负零； `0x7f800000` 代表正无穷大。 `const_float` 将主机f64值转换为数字；使用`const_float_bits`保留原始位。精确的存储表示并不意味着完整的浮点执行后端。编译时算术仍然使用主机f64中间体，因此每个声明宽度的完整舍入契约尚未实现。
