---
translation_set_id: language-variants
path: language/variants
locale: ja
group: language
group_order: 2
order: 17
title: バリアントとパターンマッチング
summary: ケースごとのデータをまとめて保存し、matchから安全に分離します。
---

## enumと区別

enumは名前付き整数値を表し、variantは場合ごとに異なるpayloadを含みます。下のResultは、成功した場合のi32の値、失敗した場合のi32エラー番号を格納します。エラーと結果が同じ整数型であっても、名前で意味を区別できます。

## 宣言・生成・検査

`main.wave`に保存して実行します。

<!-- wave-example: variant-api -->
```wave playground
variant Result {
    Value(i32), Error(i32)
}

fun calculate(valid: bool) -> Result {
    if (!valid) {
        return Result::Error(1);
    }

    return Result::Value(42);
}

fun main() {
    var result: Result = calculate(true);
    match (result) {
        Result::Value(value) => {
            println("value={}", value);
        }
        Result::Error(code) => {
            println("error={}", code);
        }
    }
}
```

実行結果：

```text
value=42
```

`Result::Value(42)`がpayloadの場合を作成します。 `match`の対応するパターン内でのみvalueという名前でpayloadを使用します。他の場合のpayloadを強制的に読みません。 arm本文はブロックで作成します。すべてのケースを処理するか、`_`で残りを処理します。新しいケースを追加したときに処理が必要かどうかを明らかにするには、各ケースを明示的に分割することをお勧めします。

## ジェネリックと寿命

型パラメータは`variant Optional<T> { Some(T), None }`のように使用できます。 `Optional<i32>` など、ローカル変数の具体的な型を指定します。ポインタを含むバリアントは、メモリの所有権を自動的に管理しません。値をコピーしても、ポイントされた割り当ては複製されません。

variantのメモリ表現を任意のCunionと同じと仮定しないでください。外部ABIに送信するデータは別々の表現を決め、許可されたFFIタイプに渡します。

## 練習

calculate（false）に変更すると、`error=1`が出なければなりません。 payloadのないEmptyケースを追加した後、matchでもそのケースを処理してみてください。

[構造体とenum](/docs/ja/language/structures-enums-and-aliases) · [エラー処理クラス](/docs/ja/language/errors)
