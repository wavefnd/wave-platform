---
translation_set_id: language-variants
path: language/variants
locale: zh
group: language
group_order: 2
order: 17
title: 变体和模式匹配
summary: 将特定于案例的数据存储在一起并与match安全地分开。
---

## 区别于enum

enum 表示一个命名的整数值，variant 保存payload，这在每种情况下都是不同的。如果成功，下面的Result存储值i32，如果失败则存储错误号i32。即使错误和结果是相同的整数类型，也可以通过名称来区分它们的含义。

## 申报、创建、检查

保存到`main.wave`并运行。

<!-- wave-example: variant-api -->
```wave playground
variant Result {
    Value(i32), Error(i32)
}

fun calculate(valid: bool) -> Result {
    if (!valid) {
        return Result::Error(1);
    }

    return Result::Value(42);
}

fun main() {
    var result: Result = calculate(true);
    match (result) {
        Result::Value(value) => {
            println("value={}", value);
        }
        Result::Error(code) => {
            println("error={}", code);
        }
    }
}
```

执行结果：

```text
value=42
```

使得`Result::Value(42)`有payload。仅在 `match` 的相应模式中使用名称 value 下的 payload。在其他情况下，它不会强制读取payload。 arm 文字是分块书写的。使用 `_` 处理所有情况或其余情况。最好明确地分解每个案例，以揭示添加新案例时是否需要处理。

## 仿制药和长寿

类型参数可以像`variant Optional<T> { Some(T), None }` 中那样使用。为局部变量指定具体类型，例如`Optional<i32>`。包含指针的变体不会自动管理内存所有权。复制该值不会复制指向的分配。

不要假设 variant 的内存表示与任何任意 C union 相同。对于要外部发送的数据ABI，确定单独的表达式并将其作为允许的FFI类型传递。

## 练习

如果将其更改为calculate(false)，您应该得到`error=1`。尝试添加不带 payload 的情况 Empty，然后也在 match 中处理该情况。

[结构及enum](/docs/zh/language/structures-enums-and-aliases) · [错误处理类](/docs/zh/language/errors)
