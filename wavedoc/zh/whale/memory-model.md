---
translation_set_id: whale-memory-model
path: whale/memory-model
locale: zh
group: whale
group_order: 1
order: 8
title: 内存模型
summary: 分配跟踪、读取初始化值、指针算术、布局和字符串存储规则。
---

## 分配跟踪

跟踪指针将地址与分配 ID、生成、边界、偏移和访问权限相关联。内存模型还跟踪分配的生命周期和初始化状态。访问必须满足这些条件；违规会导致陷阱。

即使重复使用相同的物理地址，各代的生命周期也是分开的。仅存在地址并不能确定指针是否有效或调用者可以访问该存储空间。

native 地址保持 64 位。单独的shadowmetadata与指针一起通过复制/保存/调用/返回传递。初始的native内存范围是可跟踪的堆栈和全局分配。 C 边界需要显式适配器，并且任意外部存储器的所有权转移不包括在此范围内。

## 初始化和读取

声明存储空间并不会初始化该值。检查实际读取的字节范围是否已初始化。如果在未初始化状态下读取该值的一部分，则为trap。它不会用 0 或未指定的值替换读取结果。

对读取值的初始化检查不包括字节padding。例如，如果一个结构体的所有字段都已初始化，则读取值不会仅仅因为字段对齐产生的空字节未初始化而错误。

内存副本携带初始化状态和字节。复制未初始化的存储空间不会将其更改为已初始化的存储空间。稍后从目标读取值时，将应用与读取源时相同的检查。

仅BSS的物理字节为0这一事实不允许对IR变量进行初始化。

### IR 在初始化的标量存储空间中

以下模块配置为builder并通过了验证。 `store` 在读取值之前，所有三个内存指令都指定非零的 2 次方排序。

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

将这个完整模块保存为 `initialized.wir`，执行输出为 `i32 42`。删除store后，load因未初始化而trap。对齐3是验证错误。alloca的物理零字节不代表初始化。

## 与指针运算的比较

地址计算overflow为trap。您可以创建一个 one-past 指针，直接指向分配范围之后，但无法通过它访问内存。

指针相等使用分配标识，而不仅仅是数字地址。从不同的分配中排序或减去指针会导致陷阱。从整数重建地址不会恢复访问权限。

### GEP

GEP 基于元素和字段计算地址。这不是读取值的命令。

第一个索引是基指针指向的类型的元素单元offset。后续索引选择数组元素或结构/元组字段。结构体/元组的字段索引是编译时的字段顺序号，而不是字节offset。所选类型决定了生成的指针类型。

当基本类型为 `ptr<array<i32, 4>>` 时，索引 `[0, 2]` 选择数组的第三个 i32 元素，从而得到 `ptr<i32>`。将第一个索引指定为 1 将 i32 移动一个元素，而不是数组中的一个元素。此示例说明了索引语义，而不是文本命令语法。

初始 native 目标拒绝对大小为 0 的元素进行指针算术。计算地址并不能消除后续访问所需的生存期、范围、初始化和权限检查。

## 数据布局

输出目标确定大小·对齐·字段offset·数组stride。结构和元组中字段的顺序保留其声明顺序。运行编译器的主机的布局不应被假定为输出目标的布局。

|值|保存规则|
| --- | --- |
| Bool, signed i1, unsigned u1 |至少1字节|
|空结构/元组|尺寸 0，对齐方式 1|
|数组|目标布局决定元素stride|
|结构/元组|保留声明顺序和目标的一致性要求|

完成的IR对齐是2的幂，而不是0。在生成这个IR之前必须确定自动对齐。此配置文件不支持布局 packed·union·bitfield，应拒绝。

### 输出布局查询

Rust API 计算独立于构建主机的存储布局。在以下示例中，u64字段之前有 7 个字节padding，末尾有 6 个字节padding。

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

`layout_of` 返回的大小和数组的步长包括尾部填充。结构体和元组使用相同的字段顺序规则。 Bool、i1、u1各占1个字节。空结构体和元组的大小为 0，对齐方式为 1；零长度数组保留其元素的自然对齐方式。 `void`没有存储布局，而`ptr<void>`占用8个字节。

对字段和数组元素使用自然排序。 `allocation_align` 将 SysV AMD64 规则应用于 16 字节或以上的独立本地/全局数组，该规则需要至少 16 字节对齐。它不会增加数组字段或元素stride的对齐。 AST lowering 使用此批量排序查找进行本地存储。

幅度乘法，字段offset加法，padding计算中的overflow返回`LayoutError::Overflow`。 `layout_of` 的复杂类型嵌套限制为 128 层，`layout_of_with_limit` 允许调用者指定限制。它不分配等于数组元素数量的存储空间。 `pointer_stride` 在 native 指针算术中拒绝大小为 0 的 pointee，但该类型本身的存储布局是有效的。 Packed·union·bitfield 没有受支持的类型表达式。此存储布局查找不实现复杂类型调用约定或运行时边界检查。

## 字符串和C边界

字符串是具有指定长度的不可变字节字符串。默认编码是UTF-8。允许内部 NUL 并且不会自动附加结尾 NUL。 O0 不会自动组合具有相同内容的字符串对象。

因此，由`A`、NUL和`B`组成的字节字符串的长度为3。不能用于显式C字符串转换，这会拒绝内部NUL。前端不应该默默地将其截断为`A`。

外部C·原始地址·内联汇编是一个单独的合约边界。跟踪内存的运行时检查不能保证检测到外部代码中的所有不正确行为。

## 跟踪栈内存执行

默认解释器执行整数、Bool、控制流、栈分配、数据指针存取、typed GEP、memcpy和memset。checked二元组读取不检查填充字节。地址是合成的64位值，不会解引用主机内存。参数和返回仍限于整数、Bool或void；float、调用、函数指针、通用聚合值、全局地址及native执行不受支持。栈分配保持有效直到函数返回。词法生命周期结束、调用和返回中的指针传递、外部内存适配器及native shadow metadata仍需实现。

`--max-memory`设置逻辑分配字节预算，默认64 MiB。`InterpreterOptions::memory_limits`还限制分配数为16384、指针字节元数据为262144片段、字节和元数据工作量为256 Mi单位。超限返回带IR位置的`MemoryLimit`错误，与程序trap区分。验证器也在执行前拒绝输出目标已知存储大小的overflow。

```shell
whale ir run initialized.wir --function @f0
```

```text
i32 42
```

## 字节复制与初始化重置

将以下完整模块保存为 `tracked-memory.wir`。@f0复制指针字节和独立元数据，然后读取42。@f1的uninit将类型存储范围标为未初始化并清除指针元数据，但不修改原有字节。memcpy可以复制未初始化字节；之后读取目标值时检查状态。分段复制只有集齐一致的8字节元数据才保留访问权限。用整数或memset写入相同位模式不会恢复权限。

零长度memcpy和memset不进行访问或对齐检查就成功，包括null和one-past指针。非空重叠memcpy会trap。memset初始化写入的字节并清除其指针元数据。Bool存储必须为0或1；读取其他表示会trap。

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

## 验证、lowering与执行职责

| 阶段 | 职责 |
| --- | --- |
| Verifier | 检查操作数与指针类型、对齐形式和输出目标存储大小；拒绝undef。 |
| O0 lowering | 在实际执行的声明位置保留uninit，用store写入初始值，并保留读取。 |
| Runtime | 检查分配标识、代际与生命周期、范围、权限、对齐、地址overflow和初始化；传递复制状态。 |
