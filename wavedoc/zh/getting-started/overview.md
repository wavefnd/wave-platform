---
translation_set_id: overview
path: getting-started/overview
locale: zh
group: getting-started
group_order: 1
order: 1
title: Wave 语言概览
summary: 简要介绍 Wave 的程序结构、变量声明和编译器用法。
---

## 关于 Wave

Wave 是一种静态类型的系统编程语言，面向原生代码生成和显式的底层控制。本语言环境仅提供简体中文文档。

## 第一个程序

```wave
fun main() {
    println("Hello, Wave!");
}
```

保存为 `main.wave`，然后运行：

```shell
wavec run main.wave
```

## 变量声明

局部变量统一使用 `var` 声明。

```wave
var count: i32 = 1;
count += 1;
```

顶层常量和静态存储分别使用 `const` 与 `static` 声明。

## 翻译状态

尚未翻译成简体中文的页面会明确标注，并显示英文原文。
