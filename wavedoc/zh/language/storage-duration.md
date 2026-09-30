---
translation_set_id: storage-duration
path: language/storage-duration
locale: zh
group: language
group_order: 2
order: 18
title: 存储期限和可变性
summary: 区分var、const和static的范围和可写性。
---

## 每个声明的含义

|格式|允许的位置|重新赋值|使用|
| --- | --- | --- | --- |
| `var` |功能/块|可能|常见的可变局部变量|
| `const` |顶部|不可能|全局常量声明|
| `static` |顶部|可能|在程序生命周期中存在的静态存储声明|

```wave
const PAGE_SIZE: i32 = 4096;
static request_count: i64 = 0;

fun main() {
    var limit: i32 = 4;
    var current: i32 = 0;
    var retries: i32 = 0;

    current += 1;
    retries += 1;
    println("{} {} {}", limit, current, retries);
}
```

## 局部声明规则

```wave
var value: i32 = 1;
value = 2;
```

可以将新值分配给声明为 `var` 的局部变量。整个程序中使用的常量在顶层声明为`const`。

## const 和 static 的局部使用

`const` 和 `static` 是顶级声明。函数体和`for`初始化使用局部声明`var`。

## 寿命和指针

您可以获取局部变量的地址`&`，但类型`ptr<T>`不会跟踪指针所指向的存储的实际生命周期。当从函数中传递本地存储地址时，程序结构必须直接确保该地址保持有效。

## 学习和示例范围

[练习完整的程序](/docs/zh/getting-started/overview) · [标准库](/docs/zh/stdlib)
