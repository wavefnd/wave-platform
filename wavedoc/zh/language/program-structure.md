---
translation_set_id: program-structure
path: language/program-structure
locale: zh
group: language
group_order: 2
order: 1
title: 1.从源文件到运行程序
summary: 了解源文件、函数、输出、检查和执行。
---

## 在开始本章之前

按照[安装说明](/docs/zh/getting-started/install)准备编译器和标准库。如果您可以在终端中运行`wavec --version`，则可以开始。任何可以保存纯文本文件的编辑器都可以。

在本章中，我们从创建单行输出开始，了解源文件、函数、编译、执行和退出代码之间的关系。目标不仅仅是复制指令，而是能够解释在哪个阶段发生了什么。

## 创建工作目录

为每个程序使用单独的目录可以更轻松地查找源文件和生成的可执行文件。在终端中创建并导航到目录。

```shell
mkdir wave-study
cd wave-study
```

使用编辑器在此目录中创建`main.wave`。检查扩展名以确保文件名不是`main.wave.txt`。下面是整个文件，而不仅仅是函数内的片段。

<!-- wave-example: book-hello -->
```wave playground
fun main() {
    println("Hello, Wave!");
}
```

执行结果：

```text
Hello, Wave!
```

使用以下命令运行它：

```shell
wavec run main.wave
```

请勿将终端命令与代码 Wave 混合使用。在终端中输入`wavec run`，并将`fun main`写入源文件中。无需将程序输出的`Hello, Wave!`粘贴回源中。

## 逐行读取

`fun` 是声明函数的关键字。函数是一组命名的操作，`main` 是可执行文件的入口点。稍后您将自己定义其他函数。

`main`之后的`()`是写入参数的位置。该程序中的main不带参数，因此为空。记下`{`和`}`之间的函数操作。缩进使块更易于人们阅读，并且块边界本身由花括号表示。

`println("Hello, Wave!");`是一个输出字符串的句子。双引号表示字符串文字的开头和结尾。引号本身不包含在输出中。分号表示该语句的结束。

## 语句按照写入的顺序执行

尝试将其更改为打印三遍。不要与之前的程序合并，而是将main.wave的内容替换为下面的整个程序。

<!-- wave-example: book-sequence -->
```wave playground
fun main() {
    println("start");
    println("working");
    println("done");
}
```

执行结果：

```text
start
working
done
```

说完第一句后，继续说下一句。这里没有同时运行的任务。如果要更改输出顺序，只需更改句子的顺序即可。您可以通过稍后学习条件语句和循环来控制此序列。

## print 和 println

`println` 在末尾添加换行符。 `print` 不会自动换行。连接小块以创建一条线时，差异很明显。

<!-- wave-example: book-print-lines -->
```wave playground
fun main() {
    print("Wave");
    print(" ");
    println("study");
    println("second line");
}
```

执行结果：

```text
Wave study
second line
```

打印三次并不总是产生三行。区分输出调用数和行数。直接在字符串中写入换行符时，请使用 `\n` escape。源代码中的字符串 escape 和换行符在 [串表](/docs/zh/language/strings) 中有详细介绍。

## 将值插入字符串

要将计算结果放入字符串中，请传递`{}`对应的值。

<!-- wave-example: book-format-first -->
```wave playground
fun main() {
    println("{} + {} = {}", 2, 3, 2 + 3);
}
```

执行结果：

```text
2 + 3 = 5
```

第一个`{}`包含2，第二个包含3，第三个包含5。格式字符串和值以逗号分隔。如果更改值的数量，placeholder 的数量也必须匹配。这是与字符串加法运算不同的输出语法。

## 划分检查、构建和执行

到目前为止使用的run按顺序进行构建和执行。如果你想知道哪一步失败了，你可以这样分解：

```shell
wavec check main.wave
wavec build main.wave -o hello
```

check 检查语法、类型等，但不测试程序的所有输入。例如，仅靠源检查无法确定文件在运行时是否存在。 build 创建一个可执行文件。在Linux/macOS中，执行如下。

```shell
./hello
```

在Windows中，将输出文件命名为`hello.exe`，并在PowerShell中运行为`.\hello.exe`。如果在创建可执行文件后修改源，则必须重新构建它才能反映更改。

## 退出代码也是一个结果

人类读取输出语句，但 shell 或其他程序可以通过退出代码确定成功。以下指定 main 返回 i32。

<!-- wave-example: book-exit-success -->
```wave playground
fun main() -> i32 {
    println("completed");
    return 0;
}
```

执行结果：

```text
completed
```

0 是表示正常关闭的约定。当直接发出故障信号时，返回非零代码。在Linux/macOS shell 中，代码执行后立即检查为`echo $?`，在PowerShell 中，检查为`$LASTEXITCODE`。如果您同时运行其他命令，您检查的内容可能会发生变化。

执行`return`结束该函数。 main 没有声明任何参数，即使有默认值也是不允许的。省略 main 的返回类型或使用 i32。

## 如何读取第一个错误

以下代码是故意不正确的：与正在运行的示例不同，它应该在check处失败。

```wave
fun main() {
    println("hello")
}
```

句子末尾没有分号。查看诊断显示的行及其前面的句子。编译器标记的位置可能不是错误起源的位置，而是无法再解释错误结构的位置。

首先，修复第一个错误，然后重新检查。如果前面的括号或引号没有关闭，即使在后面的正常代码中也可能会出现各种错误。如果您尝试同时修复所有线路，很容易错过最初的原因。

## 练习题

1. 创建一个打印三行自我介绍的程序。
2. 传递 12 和 8 作为格式参数来打印 `12 * 8 = 96`。
3. 编写它以打印成功消息并返回退出代码 0。
4. 在执行前预测如果仅使用 print 三次，输出行数将会是多少。

### 解决方案：计算过程输出

<!-- wave-example: book-first-solution -->
```wave playground
fun main() -> i32 {
    println("Learning Wave");
    println("My first program");
    println("Ready to calculate");
    println("{} * {} = {}", 12, 8, 12 * 8);
    return 0;
}
```

执行结果：

```text
Learning Wave
My first program
Ready to calculate
12 * 8 = 96
```

我没有将计算结果直接写在字符串中作为`96`，而是将其作为表达式传递。这是确保即使输入值发生变化计算结果和显示也不会改变的第一步。在下一章中，我们将为变量命名，以避免多次写入相同的值。


## 源文件中的顶级项目

Wave 源可以包含以下顶级项目：

- `import(...)`
- `const`, `static`
- `type`
- `struct`, `enum`, `proto`
- `extern(...)`, `export(...)`
- `fun`
- `#[target(...)]` 支持项目之前的条件

在可导入声明前加上 `pub` 前缀。将本地 `var` 声明放置在函数或块内。

## 独立程序

没有内核、启动代码或运行时的目标可以使用独立构建选项。

```shell
wavec build kernel.wave \
  --freestanding \
  --entry=_start \
  --linker-script=linker.ld \
  --no-start-files
```

`--freestanding` 链接到关闭默认库依赖项的构建计划，`--entry` 设置链接器入口符号。要创建实际的可引导输出，您需要设计目标体系结构、链接器脚本，甚至对象格式。

## 故意失败的切入点

即使参数有默认值，也不能在 main 中包含参数。复制以下文件check时出错。

<!-- wave-example: reject-main-argument -->
```wave
fun main(value: i32 = 0) -> i32 {
    return value;
}
```
