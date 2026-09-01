---
translation_set_id: program-structure
path: language/program-structure
locale: ja
group: language
group_order: 2
order: 10
title: プログラム構造
summary: トップレベル宣言、main エントリーポイント、インポートの構成、フリースタンディングのエントリーシンボルを説明します。
---

## トップレベルのソース項目

Wave のソースファイルには、次のようなトップレベル項目を記述できます。

- `import(...)`
- `const` と `static`
- `type`
- `struct`、`enum`、`proto`
- `extern(...)` と `export(...)`
- `fun`
- 対応する項目の前に置く `#[target(...)]` 条件

ローカルの `var` 宣言は関数またはブロック内に記述します。

別の Wave モジュールへ公開する宣言には `pub` を付けます。`pub` は関数、構造体、列挙型、型エイリアス、定数、静的変数、選択的再エクスポート、個別の ABI エクスポートに使用できます。`pub` のない宣言はソースモジュール内でのみ使用できます。

## ホスト環境向け実行ファイル

```wave
import("std::string::len")::{len};

const EXIT_OK: i32 = 0;

fun main() -> i32 {
    var message: str = "Wave";
    println("{} {}", message, len(message));
    return EXIT_OK;
}
```

通常のホスト環境向け実行ファイルでは、`main` がプログラムのエントリー関数になります。

## 宣言の構成とインポート

インポートは、別の Wave ソース単位にある公開宣言を名前空間または選択的インポートを通して利用可能にします。標準ライブラリ、外部パッケージ、ローカルインポートでは、それぞれ「モジュール、インポート、FFI」ガイドに記載された異なるパス形式を使います。

ソースファイルごとに役割を絞り、そのファイルが直接依存するモジュールをインポートすると、依存関係の流れを読み取りやすくなります。

## フリースタンディングプログラム

カーネル、ブートコード、ホストランタイムを持たないターゲットでは、フリースタンディングのビルド制御を使用できます。

```shell
wavec build kernel.wave \
  --freestanding \
  --entry=_start \
  --linker-script=linker.ld \
  --no-start-files
```

`--freestanding` は既定ライブラリを使わないビルド計画に切り替え、`--entry` はリンカーのエントリーシンボルを設定します。起動可能な成果物を作るには、ターゲットに適したリンカースクリプト、オブジェクト形式、プラットフォームの起動設計も必要です。
