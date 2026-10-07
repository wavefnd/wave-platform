---
translation_set_id: whale-o0-debugging
path: whale/o0-debugging
locale: ja
group: whale
group_order: 1
order: 9
title: O0 とデバッグ
summary: 計算、定数式、記憶領域、デバッグ対応情報を保存するルールです。
---

## 保存モデル

O0はデバッグのためにオリジナルtypedIRの計算・変数・制御フローを保存します。未使用の結果と構造的に到達不能なブロックも維持します。検証は誤った IR を診断し、ブロックの削除や操作を簡素化しません。

機械語を生成するために必要な変換は、別々のサブIRで行われます。変換後もオリジナルIDとの対応を維持しなければなりません。 typed IR 保存がすべて IR 演算と機械命令の一対一対応を意味するわけではありません。

## O0で行わない変換

|変換|O0動作|
| --- | --- |
|インライニング|呼び出しと関数の境界を維持|
|Tail-call変換|一般呼び出し・返還構造を維持|
|死んだコードを削除|未使用の計算と到達不能なブロックを維持|
|ランタイム定数の折りたたみ|元の操作を維持|
|共通計算を合わせる|別々の計算を維持|
|ローカル変数ストレージスペースの再利用|各ストレージスペースを維持|
|フレームポインタを省略|フレームポインタを保持|
|文字列の自動マージ|別個の文字列オブジェクトを保持|

たとえば、2 つの定数を加算するランタイム演算は、結果を使用しなくても加算として残ります。条件が定数の分岐も元の制御フロー構造を維持します。

## コンパイル時定数

コンパイル時定数宣言は、typed初期化式と評価結果を一緒に保存します。これは、通常のランタイムコマンドの定数折りたたみとは異なります。

初期化式が`1 + 2`の宣言には加算式と結果`3`が一緒に残ります。この例は式の意味を説明し、特定のソース言語の宣言文法ではありません。結果は静的データを構成するときに使用でき、初期化式をランタイム加算に置き換えることはありません。

未使用・到達不可能な宣言も識別できなければなりません。同じ名前の宣言がお互いを隠しても、名前・宣言ID・参照に区分されなければなりません。検証は、誤った参照、依存性循環、誤った型、元の式と一致しない保存結果を拒否します。

## 到達不可能なソース文

return・break・continue 後の文章も連結されていないブロックに残ります。この文のため、前のterminatorを変更したり、新しい実行パスを作成したりしないでください。到達不可能な文章内の誤った式も診断します。

この保存規則により、元のプログラム構造を調べることができます。 terminator 後の文が実際に実行されるという意味ではありません。

## 保存されるIRの例

以下は、バリデータを通過したモジュールの現在のプリンタ出力です。未使用の計算、コンパイル時の宣言、リンクされていないブロックが一緒に表示されます。

```text
module {
  format_version 4
  semantics_version 1
  target "x86_64-whale-linux"
  datalayout { ptr=64, endian=little }

  declare @f0 "preserved": whale () -> void, linkage internal

  fn @f0 "preserved"() -> void, entry %b0 {
  %b0 "entry":
    %v0: i32 = const i32 1
    %v1: i32 = const i32 2
    %v2: i32 = add i32 %v0, %v1
    %v3: i32 = const_decl "count" add(i32 1, i32 2) => const i32 3
    ret void
  %b1 "unreachable.cont":
    %v4: i32 = const i32 4
    %v5: i32 = const i32 5
    %v6: i32 = add i32 %v4, %v5
    ret void
  }

}
```

`%v2`は使用先がなくても`add`のままです。 `%v3`は別個の`const_decl`であり、`add(i32 1, i32 2)`式と評価結果3を一緒に保存します。 `unreachable.cont`には入ってくる幹線がないので、`ret void`の後に実行パスを追加せずに`%v6`加算を確認できます。

検証はこのコマンドとブロックを保存します。この例は、IR構成、検証、出力を示しており、native実行またはDWARF出力が提供されるという意味ではありません。

## ソースとスタック情報

デバッグインタフェースは、DWARF5の関数情報、ソース行、デフォルトのローカル変数、呼び出しフレーム情報を使用します。下位表現に変換しても、関数・地域変数情報と元のIR識別子の対応を維持しなければなりません。

AMD64プロファイルはフレームポインタを保持し、redzoneを使用しません。呼び出しフレーム情報はスタック調査に使用され、例外unwindingサポートを意味しません。 trapはデストラクタやunwindingを保証せずに実行を終了します。

DWARF出力とnative実行の提供可否は[ツールチェーンの概要](overview)を参照してください。 O1以上最適化の動作は、このO0参照の範囲外です。

## ループ内で繰り返される宣言

この完全なモジュールを `initialization-loop.wir` に保存してください。最初の反復で42を保存しても、次の宣言のuninitで初期化状態が消えるため読取りはtrapです。O0 loweringはallocaを入口に置いてもuninitを実行される宣言位置に残します。未使用の読取りも実行時に検査し、保存された到達不能ブロックは実行しません。

```text
module {
  format_version 4
  semantics_version 1
  target "x86_64-whale-linux"
  datalayout { ptr=64, endian=little }

  declare @f0 "redeclaration": whale () -> u32, linkage internal

  fn @f0 "redeclaration"() -> u32, entry %b0 {
  %b0 "entry":
    %v0: ptr<u32> = alloca u32, align 4
    %v1: u32 = const u32 0
    %v2: u32 = const u32 1
    %v3: u32 = const u32 42
    br label %b1
  %b1 "declaration":
    %v4: u32 = phi u32 [ %v1, %b0 ], [ %v6, %b2 ]
    uninit u32, ptr<u32> %v0, align 4
    %v5: bool = icmp eq u32 %v4, %v1
    cbr bool %v5, label %b2, label %b3
  %b2 "first_iteration":
    store u32 %v3, ptr<u32> %v0, align 4
    %v6: u32 = add u32 %v4, %v2
    br label %b1
  %b3 "second_iteration":
    %v7: u32 = load u32, ptr<u32> %v0, align 4
    ret u32 %v7
  }

}
```

```shell
whale ir run initialization-loop.wir --function @f0
```

```text
Error: initialization-loop.wir: trap at @f0 %b3 instruction 0 (%v7): uninitialized byte at allocation offset 0 (after 17 steps)
```
