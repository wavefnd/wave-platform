---
translation_set_id: memory-model
path: language/explicit-memory-type-model
locale: zh
group: language
group_order: 2
order: 9
title: 9. 指针、修改值和生命周期
summary: 学习取地址、解引用、通过指针修改值以及悬空指针。
---

## 区分值和存储位置

整数 42 和存储该整数的地址是两个不同的值。指针指向存储位置。传递地址可以让函数读取或修改调用者的存储。

本章介绍取地址、解引用、修改原始值、指针算术和生命周期。动态分配将在下一章介绍。我们先从局部变量和数组元素的地址开始。

## 获取地址并解引用

<!-- wave-example: book-pointer-first -->

```wave
fun main() {
    var value: i32 = 42;
    var address: ptr<i32> = &value;

    println("value={}", value);
    println("through pointer={}", deref address);

    deref address = 99;
    println("changed={}", value);
}
```

执行结果：

```text
value=42
through pointer=42
changed=99
```

`&value` 获取地址，`deref address` 读取或写入该地址中的值，依此类推。99 并不是存储在 `address` 中，而是写入 `address` 所指向的整数。变量 `address` 本身仍然指向 `value`。

`ptr<i32>` 是用于访问 `i32` 存储的指针类型。该类型不记录长度，也不提供自动释放。

## 重新赋值指针

<!-- wave-example: book-pointer-reassign -->

```wave
fun main() {
    var first: i32 = 10;
    var second: i32 = 20;
    var selected: ptr<i32> = &first;

    selected = &second;
    deref selected = 25;

    println("{} {}", first, second);
}
```

执行结果：

```text
10 25
```

`selected = &second` 将另一个地址存储到指针变量中。`first` 的值不会改变。之后，通过 `deref` 写入值时，`second` 会发生变化。将“更改地址”和“通过地址更改值”分开说明，可以减少混淆。

## 让函数修改原始值

<!-- wave-example: book-pointer-increment -->

```wave
fun increment(value: ptr<i32>) {
    deref value = deref value + 1;
}

fun main() {
    var count: i32 = 4;

    increment(&count);
    increment(&count);

    println("{}", count);
}
```

执行结果：

```text
6
```

该函数接收的是 `count` 的地址，而不是值 4。修改该存储也会修改调用者中的 `count`。该函数需要一个有效且可写的 `i32` 地址。传入 `null` 不符合此要求。

每个函数都会定义自己是否接受 `null`。如果不接受，调用者必须提供有效地址；如果接受，函数就必须包含处理 `null` 的分支。

## 处理 `null` 指针

<!-- wave-example: book-pointer-null -->

```wave
fun try_increment(value: ptr<i32>) -> bool {
    if (value == null) {
        return false;
    }

    deref value = deref value + 1;
    return true;
}

fun main() {
    var count: i32 = 7;

    if (!try_increment(null)) {
        println("no value");
    }

    if (try_increment(&count)) {
        println("count={}", count);
    }
}
```

执行结果：

```text
no value
count=8
```

`null` 检查只能处理地址缺失的情况。将任意非 `null` 数字转换为指针，并不会创建有效内存。读取和写入还要求目标存储处于有效的生命周期内，具有足够的大小、正确的对齐方式以及适当的访问权限。

## 数组地址和逐元素指针算术

<!-- wave-example: book-pointer-array -->

```wave
fun main() {
    var values: array<i32, 3> = [10, 20, 30];
    var first: ptr<i32> = &values[0];
    var second: ptr<i32> = first + 1;

    println("{}", deref first);
    println("{}", deref second);
    println("{}", deref first[2]);
}
```

执行结果：

```text
10
20
30
```

给指针加 1，会使它按目标类型移动一个元素。`i32` 的下一个元素和 `u8` 的下一个元素需要移动的字节数不同。如果再次将 `first + 1` 乘以类型大小并加到它上面，指针就会移到错误的位置。

指针索引也必须在有效范围内进行。`first` 本身不会记住数组长度 3，因此将这段范围传给函数时，需要同时传递指针和长度。

## 将可读范围传递给函数

<!-- wave-example: book-pointer-range -->

```wave
fun sum(values: ptr<i32>, count: i32) -> i32 {
    var total: i32 = 0;

    for (var index: i32 = 0; index < count; index += 1) {
        total += deref values[index];
    }

    return total;
}

fun main() {
    var values: array<i32, 4> = [2, 4, 6, 8];

    println("first two={}", sum(&values[0], 2));
    println("all={}", sum(&values[0], 4));
}
```

执行结果：

```text
first two=6
all=20
```

`count` 的单位是元素个数。此函数要求调用者提供包含 `count` 个可读 `i32` 的范围。传入大于实际数组长度的值违反了约定。即使 API 使用 `ptr<u8>` 与 `i64` 的组合，也应在文档中确认长度的单位是字节还是元素。

## 生命周期：地址的有效期有多长？

局部变量只在函数调用和代码块的生命周期内有效。如果返回函数内部局部变量的地址，供调用者稍后读取，那么该存储空间可能已经结束了生命周期。

下面是一个不应实现的糟糕设计：

```wave
fun invalid_address() -> ptr<i32> {
    var local: i32 = 42;

    return &local;
}
```

如果只需要返回一个值，就返回 `i32`。如果需要写入调用者提供的存储空间，就将指针作为输入。如果需要独立的存储空间，以便在函数调用结束后继续保留内容，就显式进行分配，并移交释放它的责任。

## 两个指针指向同一存储空间

复制指针会创建一个指向同一地址的别名，但不会复制内存。

<!-- wave-example: book-pointer-alias -->

```wave
fun main() {
    var value: i32 = 1;
    var first: ptr<i32> = &value;
    var second: ptr<i32> = first;

    deref second = 9;

    println("{} {}", value, deref first);
}
```

执行结果：

```text
9 9
```

通过 `second` 做出的修改也可以通过 `first` 看到。一旦源存储被释放，两个指针都会变得不可用。将 `null` 赋给一个指针变量，不会自动改变其他副本。

## 练习：交换两个整数

编写一个函数，接收两个 `i32` 地址并交换它们的值。第一个值必须先存入临时变量，才能覆盖原值。检查两次传入同一个地址时，该值是否仍能保留。

### 完整解答

<!-- wave-example: book-pointer-swap -->

```wave
fun swap(left: ptr<i32>, right: ptr<i32>) {
    var saved: i32 = deref left;

    deref left = deref right;
    deref right = saved;
}

fun main() {
    var first: i32 = 3;
    var second: i32 = 8;

    swap(&first, &second);
    println("{} {}", first, second);

    swap(&first, &first);
    println("{}", first);
}
```

执行结果：

```text
8 3
8
```

该函数还要求两个地址都指向有效且可写的整数存储。要处理 `null`，请像 `try_increment` 一样添加一个表示成功或失败的返回结果。

## **Wave 显式内存类型模型**

Wave 的指针设计基于 **Wave 显式内存类型模型**。该模型将指针和数组定义为语言层面的显式内存类型，而不是语法技巧或库抽象。

`ptr<T>` 是指向存储 `T` 值的内存地址的类型；`array<T, N>` 是定长内存类型，连续存储 `N` 个 `T` 值。因此，指针和数组的结构会原样显示在函数参数、返回值、结构字段和其他类型中。

## null

```wave
var buffer: ptr<u8> = null;
if (buffer == null) {
    println("no buffer");
}
```

`null` 是一个不指向有效内存地址的指针值。`null` 只能赋给 `ptr<T>` 类型，不能用作整数、布尔值或数组值。

当分配或查找函数没有结果时，可以返回 `null`。在解引用这类结果之前，应先检查 `null`。解引用 `null` 指针不会访问有效存储。

## 指针转换

当需要转换地址或其他指针表示形式时，请使用 `as`。

```wave
var raw: i64 = 0;
var p: ptr<u8> = raw as ptr<u8>;
```

仅在低级边界处使用整数与指针之间的转换，并考虑目标平台的地址宽度和 ABI。
