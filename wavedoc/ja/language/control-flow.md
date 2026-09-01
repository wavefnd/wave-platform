---
translation_set_id: control-flow
path: language/control-flow
locale: ja
group: language
group_order: 2
order: 4
title: 制御フロー
summary: 括弧付き条件式とループ、C 形式の for ループ、match、break、continue を説明します。
---

## 条件分岐

`if` と `else if` の条件は、**必ず括弧で囲みます**。

```wave
if (score >= 90) {
    println("A");
} else if (score >= 80) {
    println("B");
} else {
    println("C");
}
```

`if (score >= 90) { ... }` のように記述します。括弧のない形式は無効です。

## 条件式では状態を変更できない

`if`、`else if`、`while`、`for` の条件内では、代入、複合代入、`++`、`--` は使用できません。比較には `==` を使うか、状態の変更を条件式より前の文へ移動してください。

```wave
value = read_value();
if (value == expected) {
    println("matched");
}
```

通常の代入文と `for` ループの更新部分では、引き続き代入を使用できます。

## while ループ

`while` の条件も括弧で囲みます。

```wave
var index: i32 = 0;
while (index < 10) {
    index += 1;
}
```

## for ループ

`for` は初期化、条件、更新式を使います。

```wave
for (var i: i32 = 0; i < 10; i += 1) {
    println("{}", i);
}
```

初期化部分では `var` によるローカル変数宣言、または式の評価ができます。`const` と `static` はトップレベル宣言であり、`for` ループの初期化には使用できません。

## break と continue

```wave
var i: i32 = 0;
while (i < 20) {
    i += 1;
    if (i == 5) {
        continue;
    }
    if (i == 10) {
        break;
    }
}
```

`break` は最も内側のループを終了し、`continue` は次の反復へ進みます。

## match

`match` の対象値は括弧で囲みます。パターンには整数リテラル、列挙型のバリアント名、`_` ワイルドカードを使用できます。

```wave
match (status) {
    200 => { println("ok"); }
    404 => { println("not found"); }
    _ => { println("other"); }
}
```

ひとつの `match` に重複した `_` ワイルドカードアームを含めることはできません。各アームの本体は `{ ... }` ブロックです。
