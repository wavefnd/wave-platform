---
translation_set_id: whale-overview
path: whale/overview
locale: zh
group: whale
group_order: 1
order: 1
title: Whale 文档
summary: 工具链组件、可用命令和参考文档指南。
---

## 简介

Whale是一个用于编程语言实现和编译工具的通用编译器工具链。提供类型化中间表达式 (IR)、AMD64 汇编器、目标文件库和基于链接器的函数。每个组件都通过Rust库和`whale`命令使用。

IR 独立于机器语言编码表达计算的语义。汇编器对机器指令进行编码以创建可重定位对象。对象库代表部分、符号和重新排列。链接器解析对象之间的引用并放置可执行文件。

## 构建和使用工具

|工作|文件|
| --- | --- |
|每个工具的作用| [工具链组件](/docs/zh/whale/ecosystem) |
|Wave 程序构建/链接/目标选择| [构建和链接](/docs/zh/whale/build-link-targets) |
|包/依赖管理| [Vex](/docs/zh/whale/vex-package-manager) |
|Whale 执行命令| [Whale CLI](/docs/zh/whale/whale-cli) |

## 文件指南

|参考文件|内容|
| --- | --- |
| [参见IR](ir-reference) |类型、值、函数、控制流、验证、交换格式|
| [数值运算](numeric-operations) |整数运算、移位、类型转换、浮点|
| [内存模型](memory-model) |初始化、指针验证、地址计算、布局、字符串|
| [使用 O0 进行调试](o0-debugging) |计算保留、变量存储、调试信息|
| [AMD64 目标](amd64-target) |目标标识符、调用约定、native函数范围|
| [汇编器和链接器](assembler-linker) |汇编操作数、节、符号、静态链接|

参考文档定义了Whale的语义规则。可用功能如下表所示。参考文档中描述的操作可能不适用于所有版本。

## 支持状态

|组件|提供接口|限制|
| --- | --- | --- |
|汇编器|从程序集 AMD64 创建ELF64 可重定位对象|命令和指令的覆盖不完整|
|对象库|对象组合、payload（不含 BSS）、目标验证、AMD64 ELF64 序列化和可选择的 Wave 记录实现以检查大小。|可重定位对象不是可执行文件|
| IR | 构造、打印、签名检查直接/间接调用、版本化 AST 降低、目标验证和检查类型布局 | 支持 AST format 2、typed IR format 4、精确位常量以及文本解析、验证和往返打印；尚不支持机器调用生成 |
|连接器|符号解释、输入目标验证、文件/内存部分放置检查|不支持应用完全重定位和输出可执行文件。|
| 执行与调试 | 整数、Bool、控制流、跟踪栈内存和初始化检查 | float、调用、全局地址、native代码生成及DWARF仍不支持 |

禁止未定义行为的规则适用于经过验证的IR 和跟踪内存。实验实现尚未实现内存/执行参考文档中的所有运行时检查。

## 构建和组装

使用 Rust 1.86.0 或更高版本构建。

```sh
git clone https://github.com/wavefnd/Whale.git
cd Whale
cargo build --release --locked
```

将以下程序集另存为`answer.asm`。

```asm
section .text
global answer

answer:
    mov eax, 42
    ret
```

创建一个对象ELF64。

```sh
./target/release/whale asm --amd64 answer.asm -o answer.o
```

它使用Whale自己的汇编器，因此不需要外部汇编器。输出包含可调用函数，但不包含任何进程启动代码。

实验性的 AST→IR 命令可通过使用 `--features socket-cli` 进行构建来使用。 CLI的命令请参考[命令参考](/docs/zh/whale/whale-cli)。

## 库的使用和诊断

IR 必须在执行或代码生成之前进行验证。输入错误和误用builder将返回结构错误。库中的输入错误不得终止主机进程或覆盖现有内容。接受不受信任输入的工具应该能够调整资源限制。

相同的工具链版本、输入、目标和配置应该产生确定性的输出。该发行版的元数据标识了版本·commit·以及可用功能。 CI 检查良好输入、拒绝输入、trap、O0 保留、round-trip、native 每个接口的执行语义。开发/验证命令请参考[Whale 贡献信息](https://github.com/wavefnd/Whale/blob/master/CONTRIBUTING.md)。
