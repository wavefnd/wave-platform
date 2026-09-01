---
translation_set_id: types
path: language/declarations-and-types
locale: ja
group: language
group_order: 2
order: 2
title: 宣言と型
summary: ローカル変数、トップレベルの定数と静的ストレージ、組み込み型、配列、型エイリアスを説明します。
---

## ローカル変数

```wave
var count: i32 = 1;
var limit: i32 = 10;
var index: i32 = count + 1;
```

| 宣言 | 意味 |
| --- | --- |
| `var` | 値を再代入できるローカル変数を宣言します |

`var` はローカル変数を宣言する構文です。各ローカル宣言では、変数名の後に型を記述します。

```wave
var capacity: i64 = 4096;
var doubled: i64 = capacity * 2;
```

初期値なしでストレージを宣言するときは `var name: Type;`、初期値を与えるときは `var name: Type = expression;` を使います。

## トップレベルのストレージ宣言

```wave
const PAGE_SIZE: i32 = 4096;
static request_count: i64 = 0;
```

`const` と `static` はトップレベル宣言です。関数本体のローカル変数には `var` を使います。

## 整数型と浮動小数点型

Wave は次の整数型と浮動小数点型を提供します。

- 符号付き整数: `i8`、`i16`、`i32`、`i64`、`i128`
- 符号なし整数: `u8`、`u16`、`u32`、`u64`、`u128`
- ポインタサイズ整数: `isz`、`usz`
- 浮動小数点数: `f32`、`f64`

`isz` と `usz` は、ターゲットのアドレス空間に合わせた大きさの符号付き整数と符号なし整数を表します。

## その他の組み込み型

| 型 | 用途 |
| --- | --- |
| `bool` | `true` または `false` |
| `char` | 文字値 |
| `byte` | バイト値 |
| `str` | 文字列値 |
| `ptr<T>` | `T` を指すためのポインタ |
| `array<T, N>` | 要素型 `T`、長さ `N` の固定長配列 |

ユーザー定義の構造体、列挙型、型エイリアスも型として使用できます。

## 配列

```wave
var values: array<i32, 4> = [10, 20, 30, 40];
var first: i32 = values[0];
values[1] = 25;
```

配列リテラルで配列を初期化するときは、宣言した長さと要素数が一致していなければなりません。インデックスアクセスには `[]` を使います。

## 型エイリアス

```wave
type UserId = u64;
var id: UserId = 7;
```

型エイリアスは、別の型に読みやすい代替名を付けます。この例では、`u64` が必要な場所で `UserId` を使用できます。
