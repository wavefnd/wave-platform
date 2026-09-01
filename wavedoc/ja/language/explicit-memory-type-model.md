---
translation_set_id: memory-model
path: language/explicit-memory-type-model
locale: ja
group: language
group_order: 2
order: 7
title: ポインタと明示的なメモリアクセス
summary: Wave Explicit Memory Type Model、ptr<T>、配列、アドレス取得、deref、null、手動メモリの契約を説明します。
---

## Wave Explicit Memory Type Model

Wave のポインタ設計は、**Wave Explicit Memory Type Model** に基づいています。このモデルは、ポインタと配列を構文上の技巧やライブラリ抽象化ではなく、言語レベルの明示的なメモリ型として定義します。`ptr<T>` は、値が `T` を格納しているものとして解釈されるメモリアドレスであることを直接示します。`array<T, N>` は、要素型と固定の要素数をともに示します。

## ptr<T>

`ptr<T>` は、型 `T` の値を指すメモリアドレスの型です。

```wave
var value: i32 = 7;
var address: ptr<i32> = &value;
```

あるメモリアドレスに別のポインタが格納されている場合、ポインタ型を入れ子にできます。

```wave
var value: i32 = 7;
var first: ptr<i32> = &value;
var second: ptr<ptr<i32>> = &first;
```

型は、ポインタが指すメモリをどのように解釈するかを表します。その用途に対してアドレスが有効であることは、プログラムが保証します。

## null

```wave
var buffer: ptr<u8> = null;
if (buffer == null) {
    println("no buffer");
}
```

`null` は、値を指していないポインタを表します。`ptr<T>` にだけ代入できます。API が失敗または値の不在を `null` で表す場合は、`deref` を使う前にポインタを `null` と比較してください。

```wave
if (buffer != null) {
    deref buffer = 0;
}
```

## 明示的なデリファレンス

```wave
var value: i32 = 7;
var p: ptr<i32> = &value;
var copy: i32 = deref p;
deref p = 42;
```

`deref` は、ポインタのアドレスに格納された値を読み書きします。書き込み先は代入可能でなければなりません。

インデックス付きポインタも明示的にデリファレンスできます。

```wave
deref bytes[index] = 0;
```

## ポインタのキャスト

低レベル境界でポインタとアドレスサイズ整数、または異なるポインタ型の間を明示的に変換する必要がある場合は `as` を使います。

```wave
var raw: i64 = 0;
var p: ptr<u8> = raw as ptr<u8>;
```

ポインタ変換では、ターゲットのアドレス幅、アラインメント、ABI を考慮しなければなりません。

## 配列とポインタ

配列とポインタは異なるメモリ形状を表します。`array<T, N>` は `T` の値を `N` 個含み、`ptr<T>` はアドレスを格納します。両者を明示的に組み合わせることもできます。

```wave
var left: i32 = 10;
var right: i32 = 20;
var pointers: array<ptr<i32>, 2> = [&left, &right];
var block: ptr<array<i32, 3>> = &[1, 2, 3];
```

`array<ptr<i32>, 2>` は 2 個のポインタを含む配列です。`ptr<array<i32, 3>>` は 3 個の整数を含む配列ひとつへのポインタです。

## ポインタ演算と比較

整数を加減すると、`ptr<T>` はその個数分の `T` 要素だけ移動します。2 つのポインタを減算すると、バイト単位の差を `i64` で返します。

```wave
var base: ptr<i32> = 0x1000 as ptr<i32>;
var third: ptr<i32> = base + 3; // i32 の大きさの 3 倍だけ進む
var bytes: i64 = third - base;  // 12
```

ポインタは `==` と `!=` で比較でき、`null` との比較も可能です。

## ptr<T> が自動的に追跡しないもの

ポインタ型自体は、次の情報を自動的には追跡しません。

- 割り当ての所有権や解放の責任
- 割り当てサイズやインデックス境界
- アドレスが有効であり続けるライフタイム
- アラインメントや初期化状態
- 並行または重複アクセスにおけるエイリアス規則

FFI と手動割り当て API では、誰がメモリを作成するか、その大きさ、有効期間、アクセスできる型、解放方法を文書化してください。
