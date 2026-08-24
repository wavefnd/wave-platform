---
translation_set_id: overview
path: getting-started/overview
locale: ja
group: getting-started
group_order: 1
order: 1
title: Wave 言語の概要
summary: Wave の基本構造、変数宣言、コンパイラの使い方を簡潔に説明します。
---

## Wave について

Wave は、ネイティブコード生成と明示的な低レベル制御を目的とした静的型付けシステムプログラミング言語です。

## 最初のプログラム

```wave
fun main() {
    println("Hello, Wave!");
}
```

`main.wave` として保存し、次のコマンドで実行します。

```shell
wavec run main.wave
```

## 変数宣言

ローカル変数は `var` で宣言します。

```wave
var count: i32 = 1;
count += 1;
```

以前の `let` と `let mut` は廃止され、現在は構文エラーです。トップレベルの定数と静的ストレージには、それぞれ `const` と `static` を使用します。

## 翻訳状況

日本語に未翻訳のページは、英語原文であることを明示して表示します。
