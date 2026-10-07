---
translation_set_id: whale-amd64-target
path: whale/amd64-target
locale: zh
group: whale
group_order: 1
order: 10
title: AMD64 目标和 ABI
summary: Linux AMD64 描述目标属性、调用边界和 native 函数范围。
---

## 目标标识符

配置文件 native 的目标标识符是 `x86_64-whale-linux`。

|属性|值|
| --- | --- |
|操作系统| Linux |
|指令集| AMD64 |
|字节顺序| Little endian |
|Native 地址宽度|64位|
|对象格式| ELF64 |
|C 调用协议| SysV AMD64 ABI |
|静态可执行文件格式| ELF ET_EXEC |

构建主机和输出目标是不同的概念。仅仅因为您可以在另一台主机上运行Whale，并不意味着它可以输出该主机的命令集或对象格式。实现的编译路径请参考[支持状态](overview)。

## 目标选择与验证

实验性 `ir lower` 命令使用 `x86_64-whale-linux` 作为默认且唯一受支持的目标。要使用此命令，请构建为`--features socket-cli`。

```sh
whale ir lower program.json --target x86_64-whale-linux
```

无论构建主机如何，输出目标都提供 64 位 little-endian 数据布局。未知标识符或不受支持的组合（例如 `aarch64-whale-linux`、`x86_64-whale-windows`）将失败，并在读取输入或替换输出文件之前指导受支持的目标出现错误。 `--no-verify` 不禁用目标选择检查。

输入空白AST(`{"format_version":2,"semantics_version":1,"features":[],"program":{"declarations":[],"globals":[],"functions":[]}}`)将输出以下标头和模块：

```text
module {
  format_version 4
  semantics_version 1
  target "x86_64-whale-linux"
  datalayout { ptr=64, endian=little }

}
```

如果指定不受支持的目标，它将以失败状态结束。

```sh
whale ir lower program.json --target banana
```

```text
Error: unsupported target "banana"; supported targets: x86_64-whale-linux
```

在Rust中，您可以通过`ir::Target::lookup("x86_64-whale-linux")`选择目标并将`name()`和`data_layout()`传输到`lower_o0`。如果交付的布局与所选目标不同，则Lowering 会拒绝。检查自配置的IR以及`verify_module`中是否存在不支持的目标名称和布局不一致。

对象模型将`format`、`machine`、`endian`和`address_bits`存储在`ObjectTarget`中。 `ObjectFile::with_target`保留指定的标识信息，序列化只允许AMD64·little-endian·64位ELF64的组合。不同架构的 machine 标识符的存在并不意味着支持该编码器。 ELF writer 入口点都会检查元数据，并且在符号解析或链接之前检查链接器输入。支持对象的 ELF 标头保留 `EM_X86_64`。

`ObjectFile::new(ObjectFormat::ELF64)` 是一个便捷构造函数，它像以前一样使用此 AMD64 标识符。之前访问过 `object.format` 的代码必须使用 `object.target.format`。

该检查提供目标选择和对象识别。结构体、数组等的大小，字段offset、stride可以通过IR布局API进行搜索。尚不支持读取目标文件，native ABI lowering，链接到可执行文件。标量lowering继续指定显式对齐，完整的复杂类型布局规则在[内存参考文档](memory-model)中定义。

## 致电并签字

IR声明和调用验证器接受显式的`Whale`和`SysV64`约定。该约定是 `fnptr` 类型的一部分，并且必须在调用站点匹配。可变参数签名和SysV64聚合签名被拒绝。请参阅[可执行调用构造示例](ir-reference)。这将验证IR合同；机器 ABI 分类、参数寄存器/堆栈放置和本机调用发射尚未实现。

支持的调用需要显式签名和调用约定。不支持的签名是一个错误。后端不得通过缺少参数或用不同的表达式替换它们来进行类似的调用。

签名支持分为基本整数/指针、f32/f64、结构/宽值和变量参数。仅仅因为 IR 类型系统中有一个类型并不意味着ABI 支持传递或返回该类型的参数。选择后端时，您应该检查每个后端是否支持您需要的类别。

例如，处理返回的后端不支持的 i32 结构返回签名应该被拒绝。不能仅因为结构的一部分进入标量寄存器而应用标量返回约定。

## 内部调用和C边界

native 指针地址为64位。即使在复制、保存、参数和返回之后，跟踪指针的shadowmetadata也必须一起传递。

C ABI 和内部元数据传输协议是不同的。 C 跨界调用需要显式适配器。仅仅因为您将数字地址传递给 C ABI 不应被视为保留分配标识、生存期、范围或访问权限。

## 栈帧

保留帧指针并且不使用redzone。在O0中，不同局部变量的存储空间没有被复用。调用帧调试信息描述了native帧与原始程序的对应关系。

此规则仅用于调试目的，并不旨在支持异常unwinding。请参阅[使用O0进行调试](o0-debugging)。

## 个人资料限制

第一个 O0 native 配置文件不包含以下功能：

- O1 优化、矢量化、LTO.
- 共享内存和atomic操作。
- 异常 unwinding、async 函数、coroutine.
- GC 和特定于语言的所有权检查。
- 任意外部存储器所有权的转移。
- 完整的内联装配约束支持。
- 动态链接，TLS，附加命令集，附加对象类型。

超出支持范围的请求应该被拒绝，而不是悄悄地用另一个功能替换。此限制是针对native编译路径的，并不意味着删除其他独立提供的工具链组件。
