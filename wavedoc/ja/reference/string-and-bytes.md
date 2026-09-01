---
translation_set_id: string-bytes
path: reference/string-and-bytes
locale: ja
group: reference
group_order: 3
order: 4
title: 文字列とバイト
summary: 文字列サブモジュール、len/is_empty、比較、検索、トリミング、ASCII 補助、エンディアン対応のバイト操作を説明します。
---

## 文字列モジュールの構成

`std::string` には次のソース単位があります。

- `ascii.wave`
- `cmp.wave`
- `find.wave`
- `hash.wave`
- `len.wave`
- `trim.wave`

必要な操作を定義しているソース単位をインポートします。

## 長さと空判定

`std::string::len` は `len` と `is_empty` の両方を定義します。

```wave
import("std::string::len")::{len, is_empty};

fun main() {
    var size: i32 = len("Wave");
    var empty: bool = is_empty("");
    println("{} {}", size, empty);
}
```

`len(s)` は、文字列の終端ゼロより前にある値の数を `i32` で返します。`is_empty(s)` は、最初の値が終端ゼロであるとき `true` になります。

## 比較、検索、トリミング

```wave
import("std::string::cmp")::{eq, cmp, starts_with, ends_with};
import("std::string::find")::{find, contains};
import("std::string::trim")::{trim_range};
```

`wavec print std-path` で標準ライブラリのソースを探し、各モジュールの公開シグネチャを確認してください。

## ASCII 補助機能

`std::string::ascii` は、バイト単位の ASCII 分類と変換を提供します。これらの操作は Unicode テキスト規則ではなく ASCII 規則に従います。

## バイトオーダー

`std::bytes` は、バイナリ値の読み書きとエンディアン処理のためのモジュールを提供します。ファイル形式やネットワークプロトコルを実装するときは、ホストのバイトオーダーと外部形式が要求するバイトオーダーを区別してください。

整数幅とオフセットを明示すると、ターゲットをまたぐバイナリ形式のコードを確認しやすくなります。
