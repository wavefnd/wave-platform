---
translation_set_id: whale-cli
path: whale/whale-cli
locale: zh
group: whale
group_order: 1
order: 3
title: Whale 命令参考
summary: 描述 Whale assembler、object wrapper、诊断输出和可选 IR 命令。
---

## Whale 构建

在 Whale 存储库中，运行：

```shell
cargo build --release
```

顶级可执行文件有四个命令系列：

```text
whale asm [--amd64 | --aarch64] <input> -o <output>
whale object <input> -o <output>
whale ir <subcommand> [options]
```

## AMD64 assembler

```shell
whale asm --amd64 input.asm -o output.o
```

AMD64 assembler 接收 `.o` 路径作为输出，并且 ELF64 relocatable 包含 section、symbol 和 relocation 创建object。

使用 `--debug-whale` 打开详细诊断输出。

```shell
whale asm --amd64 input.asm -o output.o \
  --debug-whale --token --ast --bytes --dump-hex --stats
```

诊断标志包括`--token`、`--ast`、`--bytes`、`--dump-hex`、`--dump-bin`、`--dump-json`和`--stats`。 `--trace` 打印处理进度。

## Object wrapper

```shell
whale object input.bin -o output.o
```

`object` 命令将原始字节放入 ELF64 `.text` 部分，并在偏移量 0 处添加全局 `start` 符号。它将原始机器代码包装在 ELF 对象文件中。

## 文本 IR 验证与打印

默认构建可以读取并验证 format 4 typed IR。将 [IR 参考](ir-reference)中的完整示例保存为 `answer.wir`。`print` 验证后输出规范文本，失败时保留已有文件。它不执行 IR 或生成 native 代码。

```shell
whale ir verify answer.wir
whale ir print answer.wir -o canonical.wir
```

## 可选 IR socket

AST JSON 的 `ir lower` 需要 `socket-cli` feature。文本 IR 的 `verify` 和 `print` 不需要。

```shell
cargo run -p whale --features socket-cli -- ir lower program.json
cargo run -p whale --features socket-cli -- ir lower program.json -o program.wir
```

`ir lower`读取Whalesocketschema的JSON，将其转换为WhaleIR，并验证模块。文本IR输出到路径stdout或`-o`。 `--target <triple>` 替换目标字符串，`--no-verify` 省略验证。

使用 `ir lower` 时须启用 `socket-cli` 构建。Socket JSON 生产者与 Whale 必须使用相同的 AST schema version。


## 标量整数解释器

默认构建也执行标量整数与 Bool IR。必须提供 `--function @fN`；重复 `--arg` 传入精确十进制参数，Bool 用 true/false。`--max-steps` 统计指令与 terminator，默认为 1,000,000。run 不接受 -o 或 --no-verify。把 [IR 参考](ir-reference)的完整循环保存为 `swap-loop.wir`：

```shell
whale ir run swap-loop.wir --function @f7 --arg 3 --max-steps 100
```

```text
i32 22
```

## 跟踪栈内存执行

默认解释器执行整数、Bool、控制流、栈分配、数据指针存取、typed GEP、memcpy和memset。checked二元组读取不检查填充字节。地址是合成的64位值，不会解引用主机内存。参数和返回仍限于整数、Bool或void；float、调用、函数指针、通用聚合值、全局地址及native执行不受支持。栈分配保持有效直到函数返回。词法生命周期结束、调用和返回中的指针传递、外部内存适配器及native shadow metadata仍需实现。

`--max-memory`设置逻辑分配字节预算，默认64 MiB。`InterpreterOptions::memory_limits`还限制分配数为16384、指针字节元数据为262144片段、字节和元数据工作量为256 Mi单位。超限返回带IR位置的`MemoryLimit`错误，与程序trap区分。验证器也在执行前拒绝输出目标已知存储大小的overflow。

[内存模型](memory-model): `tracked-memory.wir`.

```shell
whale ir run tracked-memory.wir --function @f0 --max-memory 20
whale ir run tracked-memory.wir --function @f0 --max-memory 3
```

```text
u32 42
Error: tracked-memory.wir: interpreter memory Bytes limit 3 reached at @f0 %b0 instruction 0 (%v0)
```
