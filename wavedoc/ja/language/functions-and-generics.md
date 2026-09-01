---
translation_set_id: functions
path: language/functions-and-generics
locale: ja
group: language
group_order: 2
order: 5
title: 関数とジェネリクス
summary: 関数宣言、戻り値、デフォルト引数、明示的なジェネリック実体化を説明します。
---

## 関数宣言

```wave
fun add(left: i32, right: i32) -> i32 {
    return left + right;
}

fun log(message: str) {
    println("{}", message);
}
```

引数は `name: type` の形式です。値を返す関数は `-> type` を使い、戻り値の型を省略すると値を返さない関数になります。

## return

```wave
fun choose(flag: bool) -> i32 {
    if (flag) {
        return 1;
    }
    return 0;
}
```

値を返さない関数では `return;` を使用できます。

## ジェネリック関数

```wave
fun identity<T>(value: T) -> T {
    return value;
}

fun main() {
    var integer: i32 = identity<i32>(10);
    var decimal: f64 = identity<f64>(3.14);
}
```

ジェネリック関数の呼び出しには、**明示的な型引数が必要です**。型引数を指定せずジェネリックテンプレートを `identity(10)` のように呼び出すとエラーになります。

## ジェネリック構造体

```wave
struct Pair<A, B> {
    first: A;
    second: B;
}

fun main() {
    var pair: Pair<i32, f64> = Pair<i32, f64> {
        first: 1,
        second: 2.5
    };
}
```

具体的な型の組み合わせごとに、専用の関数または構造体定義が生成されます。

## デフォルト引数

デフォルト引数には整数、浮動小数点数、文字列のリテラルを使います。末尾の引数にデフォルト値が宣言されている場合、呼び出し側はその引数を省略できます。

```wave
fun repeat(value: i32, count: i32 = 1) -> i32 {
    return value * count;
}

var result: i32 = repeat(7);
```

## ジェネリクスの規則

- ジェネリック関数の呼び出しには明示的な型引数が必要です。
- `export(...)` で公開する関数をジェネリックにはできません。
- `ptr<T>` と `array<T, N>` は組み込みのメモリ型であり、ユーザー定義のジェネリックテンプレートではありません。
