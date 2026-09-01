---
translation_set_id: overview
path: getting-started/overview
locale: ja
group: getting-started
group_order: 1
order: 1
title: Wave 言語の概要
summary: Wave の構文、プログラム構造、コンソール入出力、低レベル機能を実践的に紹介します。
---

## Wave について

Wave は、ネイティブコード生成と明示的な低レベル制御のために設計された静的型付きシステムプログラミング言語です。型、メモリアクセス、ネイティブインターフェース、ターゲット設定は、ソースコードとビルドコマンド上に明示されます。

## 最初のプログラム

```wave
fun main() {
    println("Hello, Wave!");
}
```

このソースを `main.wave` として保存し、次のように実行します。

```shell
wavec run main.wave
```

実行せずに実行ファイルをビルドする場合は、次のコマンドを使います。

```shell
wavec build main.wave -o app
```

ホスト環境向けの実行ファイルは通常 `main` から開始します。フリースタンディングビルドでは、適切なリンカー設定とともに別のエントリーシンボルを指定できます。

## 基本的な文の形式

Wave の関数シグネチャは明示的です。ローカル変数では、変数名の後に型を記述します。

```wave
fun add(left: i32, right: i32) -> i32 {
    var result: i32 = left + right;
    return result;
}

fun main() {
    var count: i32 = 1;
    var next: i32 = count + 1;
    count += 1;
    println("count = {}, next = {}", count, next);
}
```

`var` はローカル変数を宣言する構文です。ローカル宣言では `var name: Type = value;` の形式で変数の型を明示します。`const` と `static` はトップレベル宣言です。

## コンソール入出力

`print`、`println`、`input` は Wave のコンソール入出力文です。最初の引数は文字列リテラルであり、各 `{}` プレースホルダーは後続する値ひとつに対応します。

```wave
fun main() {
    var value: i32 = 0;
    input("{}", value);
    print("value = ");
    println("{}", value);
}
```

## 低レベル機能

Wave は `ptr<T>`、アドレス取得の `&`、明示的な `deref`、`extern(c)` と `export(c)` による C ABI 境界、インライン `asm` を提供します。これらの機能を使うプログラムは、各メモリ API やネイティブ API が必要とする所有権、境界、アラインメント、ライフタイムの規則を定義します。

## 推奨学習順序

1. コンパイラをインストールし、`wavec --version` と `wavec --help` を確認します。
2. 宣言、型、式、制御フローを学びます。
3. 関数、ジェネリクス、構造体、列挙型、`proto` を学びます。
4. ポインタ、インポート、FFI、標準ライブラリへ進みます。
5. コンパイラ、Vex、Whale、ターゲット、クイックリファレンスの各ページを参照資料として使います。
