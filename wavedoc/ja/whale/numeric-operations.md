---
translation_set_id: whale-numeric-operations
path: whale/numeric-operations
locale: ja
group: whale
group_order: 1
order: 7
title: 数値演算
summary: 整数 wrap, checked 演算、シフト、形変換エラー、浮動小数点の結果を説明します。
---

## 整数表現

Nビット整数はNの値ビットを持ちます。符号なし整数の範囲は 0 から 2^N − 1 までで、符号付き整数の範囲は −2^(N−1) から 2^(N−1) − 1 までです。演算のsignednessがビット列の解釈を決定します。

以下の表は演算結果を示しています。 IRコード例は現在のプリンタの表現を使用しています。

## 加算・減算・乗算

基本整数add・sub・mulは結果の下位Nビットを保持します。 overflowはtrapを発生させません。 checked演算は、同じwrap結果とともに、数学的結果が対応するsignedまたはunsignedの範囲外であるかどうかを示すBoolを返します。

|演算|Wrap結果| Checked overflow |
| --- | --- | --- |
| u8: 255 + 1 | 0 | true |
| i8: 127 + 1 | −128 | true |
| u8: 0 − 1 | 255 | true |
| i8: 12 × 3 | 36 | false |

overflowで実行を中断する言語のフロントエンドは、checked操作のoverflow結果に明示的な`trap_if`を使用する必要があります。基本操作はソース言語のoverflowポリシーを暗黙的に適用しません。

### Wrapと明示的検査を表現したIR

この完全な形式 3 モジュールを `integer-operations.wir` に保存してください。テキスト reader が受理し、スカラーインタプリタで実行できます。

```text
module {
  format_version 4
  semantics_version 1
  target "x86_64-whale-linux"
  datalayout { ptr=64, endian=little }

  declare @f0 "add_u8": whale () -> u8, linkage internal
  declare @f1 "require_no_overflow": whale () -> u8, linkage internal

  declare @f2 "minimum_division": whale () -> i8, linkage internal
  declare @f3 "minimum_remainder": whale () -> i8, linkage internal
  declare @f4 "negative_shift": whale () -> i8, linkage internal

  fn @f0 "add_u8"() -> u8, entry %b0 {
  %b0 "entry":
    %v0: u8 = const u8 255
    %v1: u8 = const u8 1
    %v2: u8 = add u8 %v0, %v1
    ret u8 %v2
  }

  fn @f1 "require_no_overflow"() -> u8, entry %b1 {
  %b1 "entry":
    %v3: u8 = const u8 255
    %v4: u8 = const u8 1
    %v5: tuple<u8, bool> = uadd_chk u8 %v3, %v4
    %v6: u8 = extract %v5, 0
    %v7: bool = extract %v5, 1
    trap_if bool %v7, reason="integer overflow"
    ret u8 %v6
  }

  fn @f2 "minimum_division"() -> i8, entry %b0 {
  %b0 "entry":
    %v0: i8 = const i8 -128
    %v1: i8 = const i8 -1
    %v2: i8 = sdiv i8 %v0, %v1
    ret i8 %v2
  }

  fn @f3 "minimum_remainder"() -> i8, entry %b0 {
  %b0 "entry":
    %v0: i8 = const i8 -128
    %v1: i8 = const i8 -1
    %v2: i8 = srem i8 %v0, %v1
    ret i8 %v2
  }

  fn @f4 "negative_shift"() -> i8, entry %b0 {
  %b0 "entry":
    %v0: i8 = const i8 -1
    %v1: i8 = const i8 -1
    %v2: i8 = shl i8 %v0, %v1
    ret i8 %v2
  }

}
```

`add_u8` は 256 の下位 8 ビットである 0 を返します。`minimum_division` は −128、`minimum_remainder` は 0 を返します。`negative_shift` は −1 を unsigned count 255 と解釈し、255 mod 8 = 7 を使います。明示的な関数 ID で実行してください。

```shell
whale ir run integer-operations.wir --function @f0
whale ir run integer-operations.wir --function @f2
whale ir run integer-operations.wir --function @f3
whale ir run integer-operations.wir --function @f4
```

```text
u8 0
i8 -128
i8 0
i8 -128
```

`require_no_overflow` はインデックス 0 のラップ結果と 1 の Bool overflow フラグを取り出します。明示的 trap_if が return の前に停止します。次のコマンドは終了状態 1 で診断を stderr に出力します。

```shell
whale ir run integer-operations.wir --function @f1
```

```text
Error: integer-operations.wir: trap at @f1 %b1 instruction 5: "integer overflow" (after 6 steps)
```

## 割り算と余り

整数の除算またはゼロによる剰余はトラップを引き起こします。符号付きの最小値を -1 で割ると、最小値にラップされます。この場合の余りはゼロです。

|演算|結果|
| --- | --- |
| i8: −128 / −1 | −128 |
| i8: −128 % −1 | 0 |
|整数0除算| trap |
|整数0の残り| trap |

## シフト

Nビット値のシフトでは、countのビット列をunsignedと解釈してからNで割った余りを使います。 countが0からN−1までの範囲外という理由だけでtrapを発生させません。

8ビット値では、count0・8・16はすべて0ビットシフトです。 8ビットcountのビット列`11111111`は7ビットシフトになります。このビット列がsigned−1を表しても同じです。 unsigned解釈は残りの計算の前に行われます。

負の数や過度のcountを拒否するソース言語は、シフトの前に明示的なチェックを入れる必要があります。

## 形変換

|変換|意味|
| --- | --- |
| Zero extension |上位ビットをゼロで埋めて幅を増やす|
| Sign extension |符号ビットを複製して幅を増やす|
|ビットカット|宛先幅に対応する下位ビットのみを保持|
|ビット再解析|同じビット列を異なるタイプとして解釈する|
|損失のない数値変換|数値を保存して表現できない場合はtrap|

たとえば、8ビットの「11111111」を16ビットにゼロ拡張すると、「00000000111111111」、サイン拡張すると「11111111111111111」となります。入力ビットが同じであっても異なる演算です。

float から int への変換では、ゼロに向かって切り捨てられてから、整数の範囲がチェックされます。 NaN と無限大はトラップを引き起こします。 i8 に変換すると、127.9 は 127 になり、128.0 はトラップになります。 Bool は整数 0 または 1 に変換されます。符号付き i1 は 1 を表すことができないため、この変換の宛先にはなりません。

アドレスを整数に変換しても、その整数でポインタアクセス権を回復できるわけではありません。 [ポインタの有効性](memory-model)を参照してください。

### cast と checked 演算の検証形式

検証では実際のオペランド型と `src_ty` の一致を確認した後、opcode が許可する入力・出力の種類と幅を確認します。結果の型注釈も定義と一致する必要があります。暗黙の変換はありません。

| Opcode | 許可される型 |
| --- | --- |
| `zext`, `sext` | Bool 以外の整数。出力幅が厳密に大きい |
| `zext` (Bool) | Bool の整数 0/1 変換。符号付き `i1` 以外の全整数型、`u1` を含む |
| `trunc` | Bool 以外の整数。出力幅が厳密に小さい |
| `fext`, `ftrunc` | 浮動小数点。出力幅が厳密に大きい / 小さい |
| `itof_s`, `itof_u` | 符号付き / 符号なし整数から浮動小数点 |
| `ftoi_s`, `ftoi_u` | 浮動小数点から符号付き / 符号なし整数 |
| `bitcast` | 同じ幅の整数・浮動小数点スカラー、またはデータポインタ間 |
| `ptrtoint`, `inttoptr` | データポインタから整数 / 整数からデータポインタ |

整数の拡張と切り詰めはビット演算なので、整数オペランドの符号性は異なっても構いません。Bool は独立した論理型であり、符号拡張、切り詰め、`i1`/`u1` への bitcast を許可しません。集約型と関数ポインタの cast は拒否します。ポインタ・整数の種類の検証は、有効な割り当てやアクセス権を復元しません。実行時の変換検査と機械語 lowering は別の未実装範囲です。

次の完全な Rust プログラムは有効な変換を2つ出力し、`zext` を `fext` に変えると拒否されることを確認します。

```rust
use ir::*;

fn main() {
    let mut builder = ModuleBuilder::new("x86_64-whale-linux", DataLayout::default_64bit_le());
    let mut function = builder.begin_function("widen", vec![("byte".into(), Type::U8)], Type::Void);
    function.ret(None);
    function.finish();
    let mut module = builder.finish();
    let function = &mut module.functions[0];
    function.value_types.extend([(ValueId(1), Type::U32), (ValueId(2), Type::I32)]);
    function.blocks[0].instructions.extend([
        Instruction::Cast {
            dst: ValueId(1), op: CastOp::ZExt,
            src_ty: Type::U8, src: ValueId(0), dst_ty: Type::U32,
        },
        Instruction::Cast {
            dst: ValueId(2), op: CastOp::Bitcast,
            src_ty: Type::U32, src: ValueId(1), dst_ty: Type::I32,
        },
    ]);
    verify_module(&module).unwrap();
    print!("{}", print_module(&module));
    // An integer source cannot be annotated as a floating widening operation.
    if let Instruction::Cast { op, .. } = &mut module.functions[0].blocks[0].instructions[0] {
        *op = CastOp::FExt;
    }
    assert!(matches!(verify_module(&module), Err(VerifyError::InvalidCast { .. })));
}
```

```text
module {
  format_version 4
  semantics_version 1
  target "x86_64-whale-linux"
  datalayout { ptr=64, endian=little }

  declare @f0 "widen": whale (u8) -> void, linkage internal

  fn @f0 "widen"(%v0 "byte": u8) -> void, entry %b0 {
  %b0 "entry":
    %v1: u32 = zext u8 %v0 to u32
    %v2: i32 = bitcast u32 %v1 to i32
    ret void
  }

}
```

次の命令断片は意図的に無効な IR です。

```text
%v1: i32 = ftoi_s f64 %v0 to i32   // Invalid when %v0 is actually Bool.
%v2: i16 = zext i32 %v3 to i16     // zext cannot narrow.
%v4: i1 = zext bool %v5 to i1      // Signed i1 cannot represent true as 1.
```

実際の入力型の不一致は `OperandTypeMismatch`、無効な型の組み合わせは opcode と両方の型を含む `InvalidCast` になります。検証エラーは変換を挿入せず、モジュールも変更しません。

`sadd_chk`、`ssub_chk`、`smul_chk` は符号付き整数、`uadd_chk`、`usub_chk`、`umul_chk` は符号なし整数を要求します。両オペランドは演算型 `T` と完全に一致し、結果は `tuple<T, bool>` でなければなりません。抽出には tuple、存在するフィールドインデックス、そのフィールドと同じ型が必要です。特に overflow フィールドを `i1` として抽出すると拒否されます。上記の wrap・overflow の結果は実行契約です。これらのテストは構造検証を確認するもので、実装済みバックエンドによる実行結果ではありません。

## 浮動小数点演算

浮動小数点値は正確なf16・f32・f64ビット列を持ちます。演算は宣言された幅から最も近い値に丸められ、正確に中間の場合は有効数字の最下位ビットが偶数の値を選択するties-to-evenを使用します。

基本演算は、fast-math、暗黙的FMA、小さい値の強制ゼロ処理を許可しません。乗算次の加算はそれぞれの丸めステップを維持し、バックエンドは暗黙的に1つの操作にまとめてはいけません。

数値演算の結果はNaNでも無限大でもかまいません。数値演算と幅変更から出るNaNは幅ごとに1つの固定量のquietNaNに正規化します。保存とコピーは元のNaNビットを保存します。したがって、NaNpayloadを算術演算なしでメモリに渡す場合と計算する場合の動作が異なります。

浮動小数点状態フラグは公開されません。基本浮動小数点算術がNaN・無限大結果を許容しても、float→int変換には上記のtrap規則を適用します。


### 正確な定数を保存

`FloatBits`variantまたは正確な幅の16進ビット文字列を使用します。格納値の等価性は、負のゼロとNaNpayloadを含むビット列で比較されます。検証器はIRタイプとpayloadの幅が異なると拒否します。

```rust
use ir::{FloatBits, ModuleBuilder, Target, Type};
fn main() {
    let bits = FloatBits::parse(32, "0xffc01234").unwrap();
    assert_eq!(bits, FloatBits::F32(0xffc01234));
    let target = Target::X86_64WhaleLinux;
    let mut module = ModuleBuilder::new(target.name(), target.data_layout());
    let mut function = module.begin_function("payload", vec![], Type::F32);
    let value = function.const_float_bits(Type::F32, bits);
    function.ret(Some(value));
    function.finish();
    let module = module.finish();
    ir::verify_module(&module).unwrap();
    assert!(ir::print_module(&module).contains("const f32 0xffc01234"));
    println!("{}", bits);
}
```

```text
0xffc01234
```

f16、f32、または f64 ビット列は、`0x` で始まり、それぞれ 4、8、または 16 桁の 16 進数が続きます。 `0x80000000` は、f32 負のゼロを表します。 `0x7f800000` は正の無限大を表します。 `const_float` はホスト f64 の値を数値に変換します。 `const_float_bits` を使用して元のビットを保存します。正確なストレージ表現は、完全な浮動小数点実行バックエンドを意味するものではありません。コンパイル時の算術演算では依然としてホスト f64 中間値が使用されるため、宣言された各幅の完全な丸め規約はまだ実装されていません。
