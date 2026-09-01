---
translation_set_id: modules-ffi
path: language/modules-imports-and-ffi
locale: ja
group: language
group_order: 2
order: 8
title: モジュール、インポート、FFI
summary: ローカル、標準ライブラリ、外部パッケージのインポート、名前空間、可視性、再エクスポート、C ABI 宣言を説明します。
---

## import 構文

```wave
import("std::string::len");
```

`import` は文字列リテラルを受け取り、`;` で終わります。通常のインポートは、モジュールの公開宣言にアクセスするための名前空間を作成します。

## 標準ライブラリのインポート

`std::` で始まるパスは、インストール済みの Wave 標準ライブラリを基準に解決されます。

```wave
import("std::fs::file");
import("std::io::fd");
```

インストール済み標準ライブラリの場所は `wavec print std-path` で確認できます。

## ローカルモジュールのインポート

ローカルモジュールには `./` で始まるパスを使います。パスはインポートするソースファイルからの相対パスで、`.wave` は省略できます。

```wave
import("./math");
```

この形式は `math.wave` を読み込み、`math` 名前空間を作成します。ローカルインポートのパスはモジュールディレクトリ内に留まり、`..` やバックスラッシュは使用しません。

## 外部パッケージのインポート

修飾のない名前は外部パッケージのルートを表します。追加の `::` 要素は、そのパッケージ内のモジュールを表します。

```wave
import("math");
import("math::vector::ops");
```

パッケージルートは `src/lib.wave` または `lib.wave` を読み込みます。`math::vector::ops` のようなパッケージモジュールは `src/vector/ops.wave` または `vector/ops.wave` を読み込みます。公開宣言にはインポート名前空間を通してアクセスします。

```wave
var sum: i32 = math::add(1, 2);
var unit: Vector = math::vector::ops::normalize(value);
```

外部パッケージの場所は依存関係オプションで指定します。

```shell
wavec --dep-root .vex/deps build main.wave
wavec --dep math=/absolute/path/to/math build main.wave
```

同じパッケージが複数の依存ルートに存在すると解決が曖昧になります。`--dep name=path` でひとつに固定してください。

## エイリアスと選択的インポート

短い名前や曖昧さのない名前空間を選ぶには `as` を使います。

```wave
import("./geometry_helpers" as geometry);
var area: f64 = geometry::area(width, height);
```

選択的インポートを使うと、指定した公開宣言をインポート先のモジュールへ直接取り込めます。

```wave
import("math")::{add, Point};

var sum: i32 = add(1, 2);
var origin: Point = Point { x: 0, y: 0 };
```

インポートエイリアスと選択的インポートは別の形式であり、同時には使用できません。

## 公開宣言と再エクスポート

`pub` を付けない宣言は、そのモジュール内でのみ使用できます。

```wave
pub fun add(left: i32, right: i32) -> i32 {
    return left + right;
}

pub struct Point {
    x: i32;
    y: i32;
}
```

`pub` は `enum`、`type`、`const`、`static` の宣言にも適用できます。モジュールは、別のモジュールから選択した公開名を再エクスポートできます。

```wave
pub import("./arithmetic")::{add, subtract};
```

`pub` は Wave モジュールの可視性を制御します。ネイティブ ABI シンボルを公開する `export(c)` とは別のものです。エントリー関数 `main` は非公開のままです。

## C 関数のインポート

```wave
extern(c) fun puts(text: ptr<i8>) -> i32;
```

Wave 側の名前とネイティブシンボル名が異なる場合は、ABI の後に外部シンボル名を明示します。

```wave
extern(c, "native_symbol") fun local_name(value: i32) -> i32;
```

## Wave 関数のエクスポート

```wave
export(c) fun wave_add(left: i32, right: i32) -> i32 {
    return left + right;
}
```

`extern` と `export` は、個別の関数形式とブロック形式の両方に対応します。エクスポートする関数をジェネリックにはできません。別の Wave モジュールからもエクスポート関数をインポートする必要がある場合は、別途 `pub` を付けます。

## 明示的に確認する ABI の詳細

- 整数幅とポインタ幅
- 呼び出し規約とターゲット ABI 名
- 外部シンボル名
- 実際の文字列表現
- ポインタのライフタイムと所有権
- 必要なライブラリとリンカー検索パス

リンクに成功しただけでは、関数シグネチャとメモリ契約が一致していることの証明にはなりません。

## ターゲット条件属性

トップレベル宣言は `#[target(...)]` を使ってターゲットごとに選択できます。

```wave
#[target(os="linux", arch="x86_64")]
extern(c) fun platform_call(value: i32) -> i32;
```

対応する条件キーは `arch`、`os`、`env`、`abi` です。この属性は次のトップレベル項目に適用されます。
