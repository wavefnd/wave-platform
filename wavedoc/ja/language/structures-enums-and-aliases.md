---
translation_set_id: data-types
path: language/structures-enums-and-aliases
locale: ja
group: language
group_order: 2
order: 6
title: 構造体、列挙型、型エイリアス
summary: 構造体のフィールドとメソッド、proto 拡張、表現型を明示する列挙型、型エイリアスを説明します。
---

## 構造体

```wave
struct Point {
    x: f64;
    y: f64;
}

fun main() {
    var point: Point = Point { x: 1.0, y: 2.0 };
    println("{}", point.x);
}
```

構造体のフィールドは `name: type;` の形式です。フィールドへは `value.field` でアクセスします。

## メソッドと proto ブロック

構造体の本体内で `fun` を使ってメソッドを宣言できます。`proto` ブロックを使うと、既に宣言された構造体へメソッドを追加できます。

```wave
struct Counter {
    value: i32;
}

proto Counter {
    fun read(self: Counter) -> i32 {
        return self.value;
    }
}
```

通常のメソッド構文で呼び出します。

```wave
var counter: Counter = Counter { value: 3 };
var value: i32 = counter.read();
```

## 列挙型

列挙型では `->` の後に整数の表現型を宣言します。

```wave
enum State -> i32 {
    Ready = 0,
    Running,
    Stopped,
}
```

バリアントには整数値を明示できます。最初に省略された値は `0` で、それ以降に省略された値は直前のバリアントより 1 大きい値になります。

## 型エイリアス

```wave
type FileHandle = i64;
```

型エイリアスを使うと、同じ基になる型を別の名前で参照できます。

```wave
var handle: FileHandle = 4;
var raw: i64 = handle;
```

`FileHandle` と `i64` は同じ型です。エイリアスは変換を必要とせず、値の用途を伝えます。

## ABI とレイアウト

Wave の構造体を外部 ABI やバイナリファイルの境界で直接共有するときは、フィールドの順序だけから C 互換レイアウトを推測しないでください。ターゲット ABI とコンパイラが保証する内容に従って表現を設計します。
