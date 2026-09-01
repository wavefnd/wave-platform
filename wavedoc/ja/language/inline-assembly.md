---
translation_set_id: assembly
path: language/inline-assembly
locale: ja
group: language
group_order: 2
order: 9
title: インラインアセンブリ
summary: asm ブロック内の命令文字列、in/out オペランド、clobber 契約を説明します。
---

## asm ブロック

`asm` は、ターゲットアーキテクチャの命令を挿入するための低レベル機能です。

```wave
fun read_value() -> i64 {
    var result: i64 = 0;
    asm {
        "mov rax, 123"
        out("rax") result
    }
    return result;
}
```

ブロック内の文字列リテラルは、アセンブリ命令の項目になります。

## 入力と出力

```wave
var result: i64 = 0;
asm {
    "mov rax, rdi"
    in("rdi") 123
    out("rax") result
}
```

- `in("reg") expression` は Wave の値を入力オペランドへ関連付けます。
- `out("reg") target` は出力を代入可能な Wave の対象へ格納します。
- レジスタ名は文字列または識別子形式で記述できます。

入力オペランドには、変数、整数または文字列リテラル、`&identifier`、`deref identifier`、負の数値リテラルを使用できます。

## clobber

明示した出力以外のレジスタやメモリ状態をブロックが変更する場合は、その状態を `clobber(...)` に列挙します。

```wave
asm {
    "nop"
    clobber("rax", "rcx", "memory")
}
```

## 確認事項

- 命令構文はターゲットアーキテクチャと LLVM インラインアセンブリの契約に一致させます。
- 呼び出し規約で保存が求められるレジスタを破壊しないでください。
- ブロックが暗黙の状態を読み書きする場合は、必要な clobber を宣言します。
- アーキテクチャ固有の asm は、小さな型付き関数の内側に分離することを推奨します。

インラインアセンブリの正しさと移植性は、Wave の型システムだけでは保証されません。
