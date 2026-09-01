---
translation_set_id: storage-duration
path: language/storage-duration
locale: ja
group: language
group_order: 2
order: 12
title: ストレージ期間と変更可能性
summary: var、const、static のスコープと書き込み規則の違いを説明します。
---

## 宣言の意味

| 形式 | 使用できる場所 | 再代入 | 用途 |
| --- | --- | --- | --- |
| `var` | 関数／ブロック | 可 | 通常の変更可能なローカル変数 |
| `const` | トップレベル | 不可 | グローバル定数宣言 |
| `static` | トップレベル | 可 | プログラムの実行期間中に存在する静的ストレージ |

```wave
const PAGE_SIZE: i32 = 4096;
static request_count: i64 = 0;

fun main() {
    var limit: i32 = 4;
    var active: i32 = 0;
    var retries: i32 = 0;

    active += 1;
    retries += 1;
    println("{} {} {}", limit, active, retries);
}
```

## ローカル変数

```wave
var value: i32 = 1;
value = 2;
```

`var` はローカルストレージを宣言し、その値は再代入できます。名前付き定数にはトップレベルの `const` を使います。

## ローカルの const と static

`const` と `static` はトップレベル宣言です。関数本体の内側や `for` ループの初期化部分には宣言できません。

## ライフタイムとポインタ

`&` を使うとローカル変数のアドレスを取得できますが、`ptr<T>` は参照先ストレージの有効期間を追跡しません。ローカル変数のアドレスがスコープ外へ渡される場合、そのストレージが無効になった後にアドレスを使用しないことをプログラム側で保証します。
