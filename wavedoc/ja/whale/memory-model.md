---
translation_set_id: whale-memory-model
path: whale/memory-model
locale: ja
group: whale
group_order: 1
order: 8
title: メモリモデル
summary: 割り当てトレース、初期化された値の読み取り、ポインター算術、レイアウト、および文字列の保存規則。
---

## 割り当て追跡

追跡されるポインタは、アドレスを割り当て ID、世代、境界、オフセット、およびアクセス許可に関連付けます。メモリ モデルは、割り当ての有効期間と初期化状態も追跡します。アクセスは次の条件を満たす必要があります。違反するとトラップが発生します。

同じ物理アドレスが再使用されても、世代は各寿命を区別します。アドレスがあるという事実だけでは、ポインタが有効であるか、呼び出し元にその記憶領域へのアクセス権があるとは判断しません。

nativeアドレスは64ビットを保持します。別途のshadow metadataがコピー・保存・呼出・返還を通じてポインタとともに渡されます。初期 native メモリ範囲は、追跡可能なスタックとグローバル割り当てです。 C境界には明示的なアダプタが必要であり、任意の外部メモリの所有権の譲渡はこの範囲に含まれていません。

## 初期化と読み取り

記憶領域を宣言したとしても値が初期化されるわけではありません。実際に読み取るバイト範囲の初期化を確認します。値の一部でも未初期化状態で読み取るとtrapです。読み取り結果をゼロまたは不特定の値に置き換えません。

値読み取りの初期化チェックは、paddingバイトを除外します。たとえば、構造体内のすべてのフィールドが初期化されている場合、フィールドのソートのために発生した空のバイトが未初期化状態であるという理由だけで、値の読み取りは誤りではありません。

メモリコピーはバイトとともに初期化状態を渡します。未初期化ストレージスペースをコピーしても、初期化されたストレージスペースに変わりません。後で宛先から値を読み取るときにも、原稿を読み取るときと同じチェックを適用します。

BSSの物理バイトが0であるという事実だけで、IR変数の初期化は成立しません。

### 初期化されたスカラー記憶領域のIR

次のモジュールはbuilderで構成され、検証器を通過しました。値を読み取る前に`store`し、3つのメモリコマンドはすべてゼロではなく2の重なり合いを指定します。

```text
module {
  format_version 4
  semantics_version 1
  target "x86_64-whale-linux"
  datalayout { ptr=64, endian=little }

  declare @f0 "initialized_local": whale () -> i32, linkage internal

  fn @f0 "initialized_local"() -> i32, entry %b0 {
  %b0 "entry":
    %v0: ptr<i32> = alloca i32, align 4
    %v1: i32 = const i32 42
    store i32 %v1, ptr<i32> %v0, align 4
    %v2: i32 = load i32, ptr<i32> %v0, align 4
    ret i32 %v2
  }

}
```

この完全なモジュールを `initialized.wir` に保存すると実行結果は `i32 42` です。storeを削除するとloadで未初期化trapが発生します。整列3は検証エラーです。allocaの物理的な0は初期化を意味しません。

## ポインタ算術との比較

アドレス計算overflowはtrapです。割り当て範囲のすぐ次を指す one-past ポインタは作成できますが、これによりメモリにアクセスすることはできません。

ポインタの等価性では、単なる数値アドレスではなく、割り当て ID が使用されます。異なる割り当てからポインタを順序付けたり減算すると、トラップが発生します。整数からアドレスを再構築しても、アクセス許可は復元されません。

### GEP

GEPは、要素とフィールドでアドレスを計算します。値を読み取るコマンドではありません。

最初のインデックスは、参照ポインタが指す型の要素単位offsetです。以降、インデックスは配列要素または構造体・タプルフィールドを選択します。構造体・タプルのフィールドインデックスはコンパイル時点のフィールド順番であり、バイトoffsetではありません。選択されたタイプが結果ポインタタイプを決定します。

基準タイプが`ptr<array<i32, 4>>`の場合、インデックス`[0, 2]`は配列の3番目のi32要素を選択し、結果は`ptr<i32>`です。最初のインデックスを1に指定すると、i32要素は1つではなく、4つの要素を持つ1つの配列だけ移動します。この例ではインデックスの意味を説明し、テキストコマンド文法ではありません。

初期 native ターゲットはサイズ 0 要素へのポインタ算術を拒否します。住所を計算しても、以降のアクセスに必要な寿命・範囲・初期化・権限検査はなくなりません。

## データレイアウト

出力ターゲットがサイズ・整列・フィールド offset・配列 strideを決定します。構造体とタプルのフィールド順序は宣言順序を保持します。コンパイラを実行しているホストのレイアウトを出力ターゲットのレイアウトと仮定してはいけません。

|値|保存ルール|
| --- | --- |
| Bool, signed i1, unsigned u1 |最小1バイト|
|空の構造体・タプル|サイズ 0、アライメント 1|
|配列|ターゲットレイアウトが要素strideを決定|
|構造体・タプル|宣言の順序とターゲットの整合要件を維持する|

完成したIRのソートは、ゼロではなく2の累乗です。自動ソートは、このIRを作成する前に決定する必要があります。このプロファイルでは、packedレイアウト・union・bitfieldはサポートされておらず、拒否する必要があります。

### 出力レイアウトの照会

Rust APIは、ビルドホストとは無関係にストレージレイアウトを計算します。次の例には、u64フィールドの前のpadding7バイトと最後のpadding6バイトがあります。

```rust
use ir::{allocation_align, layout_of, Target, Type};

fn main() {
    let target = Target::X86_64WhaleLinux;
    let record = Type::Struct(vec![Type::U8, Type::U64, Type::U16]);
    let layout = layout_of(&record, target).unwrap();
    assert_eq!((layout.size, layout.align), (24, 8));
    assert_eq!(layout.field_offsets, [0, 8, 16]);

    let array = Type::Array(Box::new(record), 3);
    let layout = layout_of(&array, target).unwrap();
    assert_eq!((layout.size, layout.align), (72, 8));
    assert_eq!(layout.element_stride, Some(24));
    assert_eq!(allocation_align(&array, target).unwrap(), 16);
}
```

```text
struct{u8, u64, u16}: size 24, natural alignment 8
field 0: byte 0
field 1: byte 8
field 2: byte 16
array of 3: size 72, element stride 24
standalone array placement alignment: 16
```

`layout_of` によって返されるサイズと配列のストライドには、末尾のパディングが含まれます。構造体とタプルは同じフィールド順序ルールを使用します。 Bool、i1、u1 はそれぞれ 1 バイトを占めます。空の構造体とタプルのサイズは 0、アライメントは 1 です。長さ 0 の配列は、その要素の自然な位置合わせを保持します。 `void` にはストレージ レイアウトがありませんが、`ptr<void>` は 8 バイトを占有します。

フィールドと配列要素には自然な配置を使用します。 `allocation_align`は、16バイト以上の独立した地域・グローバル配列に最小16バイトのソートを要求するSysVAMD64ルールを適用します。配列フィールドのソートや要素strideを増やすことはありません。 ASTloweringは、ローカルストレージスペースにこのバッチソート照会を使用します。

サイズ乗算、フィールドoffset加算、padding計算のoverflowは`LayoutError::Overflow`として返されます。 `layout_of`の複合タイプのネスト限度はステップ128であり、`layout_of_with_limit`で発信者は制限を指定できます。配列要素の数だけ記憶領域を割り当てません。 `pointer_stride`はnativeポインタ演算でサイズ0のpointeeを拒否しますが、そのタイプの保存レイアウト自体は有効です。 Packed・union・bitfieldにはサポートされるタイプ表現はありません。この保管レイアウト照会は、複合型呼び出し規約または実行時範囲検査を実装しません。

## 文字列とC境界

文字列は、長さが指定された不変バイト列です。デフォルトのエンコーディングはUTF-8です。内部NULを許可し、自動的に終了NULを貼り付けません。 O0は、同じ内容の文字列オブジェクトを自動的に結合しません。

したがって、`A`、NUL、`B`からなるバイト列の長さは3です。内部NULを拒否する明示的なC文字列変換には使用できません。フロントエンドはこれを静かに`A`に切ってはいけません。

外部C・生アドレス・インラインアセンブリは別途契約境界です。トレースメモリのランタイムチェックが外部コードのすべての誤動作まで検出されることを保証するものではありません。

## 追跡スタックメモリの実行

標準インタープリタは整数・Bool、制御フロー、スタック割り当て、データポインタの保存と読取り、typed GEP、memcpy・memsetを実行します。checked ペアの読取りではパディングを除外します。アドレスは合成64ビット値で、ホストメモリを参照しません。引数と戻り値は整数・Boolまたはvoidに限定され、float、呼出し、関数ポインタ、一般aggregate値、グローバルアドレス、native実行は未対応です。割り当ては関数のreturnまで有効です。字句スコープの寿命終了、呼出し・returnでのポインタ転送、外部メモリアダプタ、native shadow metadataは今後の実装です。

`--max-memory`は論理的な割り当てバイトの上限で、標準は64 MiBです。`InterpreterOptions::memory_limits`は割り当て数16384、ポインタのバイトメタデータ262144個、バイト・メタデータ処理256 Mi単位も制限します。超過はプログラムtrapとは別のIR位置付き`MemoryLimit`エラーです。検証器は出力ターゲットの既知の保存サイズoverflowも実行前に拒否します。

```shell
whale ir run initialized.wir --function @f0
```

```text
i32 42
```

## バイトコピーと初期化のリセット

次の完全なモジュールを `tracked-memory.wir` に保存してください。@f0はポインタのバイトと別のメタデータをコピーして42を読みます。@f1のuninitは型の保存範囲を未初期化にし、ポインタメタデータを消去しますが、既存のバイトは変更しません。memcpyは未初期化バイトもコピーでき、後の宛先読取りで状態を検査します。分割コピーは一貫した8バイト分のメタデータが揃った場合のみアクセス権限を保持します。整数やmemsetで同じビットを書いても権限は復元しません。

長さ0のmemcpy・memsetはアクセスや整列を検査せず成功し、nullやone-pastも使用できます。空でない重複memcpyはtrapです。memsetは書いたバイトを初期化し、そのポインタメタデータを消去します。Boolの保存値は0または1で、他の表現の読取りはtrapです。

```text
module {
  format_version 4
  semantics_version 1
  target "x86_64-whale-linux"
  datalayout { ptr=64, endian=little }

  declare @f0 "copied_pointer": whale () -> u32, linkage internal
  declare @f1 "uninitialized": whale () -> u32, linkage internal

  fn @f0 "copied_pointer"() -> u32, entry %b0 {
  %b0 "entry":
    %v0: ptr<u32> = alloca u32, align 4
    %v1: u32 = const u32 42
    store u32 %v1, ptr<u32> %v0, align 4
    %v2: ptr<ptr<u32>> = alloca ptr<u32>, align 8
    %v3: ptr<ptr<u32>> = alloca ptr<u32>, align 8
    store ptr<u32> %v0, ptr<ptr<u32>> %v2, align 8
    %v4: ptr<u8> = bitcast ptr<ptr<u32>> %v2 to ptr<u8>
    %v5: ptr<u8> = bitcast ptr<ptr<u32>> %v3 to ptr<u8>
    %v6: u64 = const u64 8
    memcpy ptr<u8> %v5, ptr<u8> %v4, u64 %v6, align 8
    %v7: ptr<u32> = load ptr<u32>, ptr<ptr<u32>> %v3, align 8
    %v8: u32 = load u32, ptr<u32> %v7, align 4
    ret u32 %v8
  }

  fn @f1 "uninitialized"() -> u32, entry %b0 {
  %b0 "entry":
    %v0: ptr<u32> = alloca u32, align 4
    uninit u32, ptr<u32> %v0, align 4
    %v1: u32 = load u32, ptr<u32> %v0, align 4
    ret u32 %v1
  }

}
```

```shell
whale ir run tracked-memory.wir --function @f0
whale ir run tracked-memory.wir --function @f1
```

```text
u32 42
Error: tracked-memory.wir: trap at @f1 %b0 instruction 2 (%v1): uninitialized byte at allocation offset 0 (after 3 steps)
```

## 検証・lowering・実行の責任

| 段階 | 責任 |
| --- | --- |
| Verifier | オペランド・ポインタ型、整列の形式、出力ターゲットの保存サイズを検査し、undefを拒否します。 |
| O0 lowering | 実行される宣言位置にuninit、初期値にstoreを置き、読取りを保持します。 |
| Runtime | 割り当てID・世代・寿命・境界・権限・整列・アドレスoverflow・初期化を検査し、コピー状態を渡します。 |
