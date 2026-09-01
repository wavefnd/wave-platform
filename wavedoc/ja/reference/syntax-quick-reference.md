---
translation_set_id: quick-reference
path: reference/syntax-quick-reference
locale: ja
group: reference
group_order: 3
order: 3
title: 構文クイックリファレンス
summary: よく使う宣言、制御フロー、型、ポインタ、FFI の構文を 1 ページにまとめます。
---

## 宣言

```wave
var value: i32 = 1;
var next: i32 = value + 1;
const LIMIT: i32 = 64;
static total: i64 = 0;
type Identifier = u64;
```

`var` はローカル宣言であり、明示的な型が必要です。`const` と `static` はトップレベル宣言です。

## 関数

```wave
fun max(left: i32, right: i32) -> i32 {
    if (left > right) {
        return left;
    }
    return right;
}
```

## ジェネリクス

```wave
fun identity<T>(value: T) -> T {
    return value;
}

var value: i32 = identity<i32>(10);
```

ジェネリック関数の呼び出しには明示的な型引数が必要です。

## 構造体と列挙型

```wave
struct Pair {
    left: i32;
    right: i32;
}

enum Result -> i32 {
    Ok = 0,
    Error,
}
```

## 条件分岐とループ

```wave
if (ready) {
    println("ready");
}

while (count < 10) {
    count += 1;
}

for (var i: i32 = 0; i < 10; i += 1) {
    println("{}", i);
}

match (status) {
    Ready => { println("ready"); }
    0 => { println("zero"); }
    _ => { println("other"); }
}
```

`if`、`while`、`for`、`match` のヘッダーは括弧で囲みます。

## 配列とポインタ

```wave
var values: array<i32, 4> = [1, 2, 3, 4];
var p: ptr<i32> = &values[0];
var first: i32 = deref p;
```

## コンソール入出力

```wave
print("value = ");
println("{}", value);
input("{}", value);
```

最初の引数は文字列リテラルです。正確な `{}` プレースホルダーひとつにつき後続する式がひとつ必要で、`input` の格納先は代入可能でなければなりません。

## インポートと FFI

```wave
import("std::string::len")::{len};
import("./helpers" as helpers);
import("math")::{add, Point};
extern(c) fun native_call(value: i32) -> i32;

export(c) fun wave_call(value: i32) -> i32 {
    return value + 1;
}
```

別の Wave モジュールからインポートする宣言には `pub` を使い、選択した公開名を再エクスポートするには `pub import("path")::{name};` を使います。

## ターゲット条件付き項目

```wave
#[target(os="linux", arch="riscv64")]
extern(c) fun platform_call(value: i32) -> i32;
```

対応する条件キーは `arch`、`os`、`env`、`abi` です。この属性は次のトップレベル項目を制御します。

## インラインアセンブリ

```wave
var result: i64 = 0;
asm {
    "mv a0, a1"
    in("a1") 7
    out("a0") result
    clobber("memory")
}
```

命令テキストとレジスタ名はターゲット固有です。ブロックに必要なすべての入力、出力、暗黙の clobber を宣言します。

## コンパイラ照会

```shell
wavec build main.wave --emit=check
wavec print supported-targets
wavec print supported-emit-kinds
```
