---
translation_set_id: lexical
path: language/lexical-structure
locale: ja
group: language
group_order: 2
order: 1
title: 字句構造
summary: 識別子、リテラル、文の区切り、キーワード、組み込み型の表記を説明します。
---

## 識別子

識別子は変数、関数、型、フィールド、その他の宣言に名前を付けます。名前は大文字と小文字を区別します。識別子には Unicode の文字、数字、`_` を使用できますが、数字から始めることはできません。

```wave
var request_count: i64 = 0;
var 名前: str = "Wave";
```

ツールでの扱いや検索性を考慮し、プロジェクト内で一貫した命名規則を採用することもできます。

## 文と区切り文字

多くの宣言と式文は `;` で終わります。関数、条件分岐、ループ、構造体のように本体を持つ構文では `{ ... }` ブロックを使います。

```wave
var answer: i32 = 42;

fun double(value: i32) -> i32 {
    return value * 2;
}
```

## リテラル

```wave
var integer: i32 = 42;
var decimal: f64 = 3.14;
var text: str = "Wave";
var letter: char = 'W';
var enabled: bool = true;
var address: ptr<u8> = null;
```

Wave には整数、浮動小数点数、文字列、文字、真偽値、`null` のリテラルがあります。`null` はポインタ値として使います。

## キーワードと型名

Wave のキーワードには次のものがあります。

`pub`、`fun`、`extern`、`export`、`type`、`enum`、`static`、`var`、`deref`、`const`、`if`、`else`、`proto`、`struct`、`while`、`for`、`in`、`out`、`clobber`、`as`、`asm`、`import`、`return`、`continue`、`print`、`input`、`println`、`match`、`break`、`true`、`false`、`null`。

組み込み型の表記には `bool`、`char`、`byte`、`str`、整数型、浮動小数点型、`ptr<T>`、`array<T, N>` があります。キーワードと組み込み型名は宣言名として使用できません。
