---
translation_set_id: whale-o0-debugging
path: whale/o0-debugging
locale: zh
group: whale
group_order: 1
order: 9
title: O0 与调试
summary: 保存计算、常量表达式、存储空间和调试响应信息的规则。
---

## 保护模型

O0 保留原始 typed IR 的计算、变量和控制流，以供调试之用。它还保留未使用的结果和结构上无法访问的块。验证诊断故障IR，并且不删除块或简化操作。

生成机器语言所需的转换在单独的子子IR中执行。即使在转换后，也必须保持与原始ID的对应关系。 typed IR 守恒并不意味着所有 IR 操作和机器指令都一一对应。

## 转换未由 O0 执行

|转换|O0 操作|
| --- | --- |
|内联|维护调用和函数边界|
|转换 Tail-call|维护一般调用/返回结构|
|删除死代码|维护未使用的计算和无法访问的块|
|折叠运行时常量|保持原来的操作|
|结合常见的计算|保持单独的计算|
|局部变量存储空间的复用|维护各个存储空间|
|省略帧指针|维护帧指针|
|自动合并字符串|维护单独的字符串对象|

例如，将两个常量相加的运行时操作即使不使用结果，仍然保持加法。条件不变的分支也保持原来的控制流结构。

## 编译时常量

编译时常量声明保留 typed 初始化表达式和计算结果。这与常规运行时指令中的常量折叠不同。

初始化表达式为 `1 + 2` 的声明留下加法表达式和结果 `3`。此示例说明了表达式的含义，而不是特定源语言的声明性语法。结果可以在构造静态数据时使用，并且不会用运行时添加替换初始化表达式。

还必须识别未使用和无法访问的声明。即使同名的声明相互混淆，也必须用名称·声明ID·引用分隔。验证会拒绝具有无效引用、依赖循环、无效类型或与原始表达式不匹配的存储结果。

## 无法访问源语句

return·break·continue之后的句子也保持在不连接的块中。由于此语句，您不应更改前面的terminator或创建新的执行路径。它还可以诊断无法到达的句子中的无效表达式。

该保存规则允许检查原始程序结构。这并不意味着terminator之后的语句将实际执行。

## 保留 IR 示例

以下是通过验证程序的模块的当前打印机输出。一起显示未使用的计算、编译时声明和未链接的块。

```text
module {
  format_version 4
  semantics_version 1
  target "x86_64-whale-linux"
  datalayout { ptr=64, endian=little }

  declare @f0 "preserved": whale () -> void, linkage internal

  fn @f0 "preserved"() -> void, entry %b0 {
  %b0 "entry":
    %v0: i32 = const i32 1
    %v1: i32 = const i32 2
    %v2: i32 = add i32 %v0, %v1
    %v3: i32 = const_decl "count" add(i32 1, i32 2) => const i32 3
    ret void
  %b1 "unreachable.cont":
    %v4: i32 = const i32 4
    %v5: i32 = const i32 5
    %v6: i32 = add i32 %v4, %v5
    ret void
  }

}
```

即使没有用，`%v2` 仍然是 `add`。 `%v3`是一个单独的`const_decl`，它将表达式`add(i32 1, i32 2)`和评估结果3一起保留。由于`unreachable.cont`没有传入边，我们可以检查`%v6`的添加，而无需在`ret void`之后添加执行路径。

验证保留这些指令和块。此示例显示 IR 配置/验证/输出，并不意味着提供 native 执行或 DWARF 输出。

## 源和堆栈信息

调试接口使用DWARF 5中的函数信息、源代码行、默认局部变量和调用帧信息。即使转换为子表达式，也必须保持函数/局部变量信息与原始IR标识符之间的对应关系。

AMD64 配置文件保留帧指针，并且不使用 red zone。调用帧信息用于堆栈检查，并不意味着支持异常unwinding。 trap 终止执行而不保证析构函数或 unwinding。

有关DWARF输出和native执行的可用性，请参阅[工具链概述](overview)。 O1 及以上优化的行为超出了本 O0 参考的范围。

## 循环中的重复声明

将这个完整模块保存为 `initialization-loop.wir`。即使第一次迭代存储42，第二次声明的uninit也会重置初始化状态，所以读取会trap。O0 lowering即使把alloca放在入口块，也将uninit保留在实际声明位置。执行的未使用读取仍进行检查；保留的不可达块不执行。

```text
module {
  format_version 4
  semantics_version 1
  target "x86_64-whale-linux"
  datalayout { ptr=64, endian=little }

  declare @f0 "redeclaration": whale () -> u32, linkage internal

  fn @f0 "redeclaration"() -> u32, entry %b0 {
  %b0 "entry":
    %v0: ptr<u32> = alloca u32, align 4
    %v1: u32 = const u32 0
    %v2: u32 = const u32 1
    %v3: u32 = const u32 42
    br label %b1
  %b1 "declaration":
    %v4: u32 = phi u32 [ %v1, %b0 ], [ %v6, %b2 ]
    uninit u32, ptr<u32> %v0, align 4
    %v5: bool = icmp eq u32 %v4, %v1
    cbr bool %v5, label %b2, label %b3
  %b2 "first_iteration":
    store u32 %v3, ptr<u32> %v0, align 4
    %v6: u32 = add u32 %v4, %v2
    br label %b1
  %b3 "second_iteration":
    %v7: u32 = load u32, ptr<u32> %v0, align 4
    ret u32 %v7
  }

}
```

```shell
whale ir run initialization-loop.wir --function @f0
```

```text
Error: initialization-loop.wir: trap at @f0 %b3 instruction 0 (%v7): uninitialized byte at allocation offset 0 (after 17 steps)
```
