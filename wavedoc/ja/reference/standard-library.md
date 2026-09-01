---
translation_set_id: standard-library
path: reference/standard-library
locale: ja
group: reference
group_order: 3
order: 1
title: 標準ライブラリ索引
summary: Wave 標準ライブラリのトップレベルモジュール、インポートパス、API の調べ方、プラットフォーム境界を説明します。
---

## トップレベルモジュール

Wave 標準ライブラリには、次のトップレベルモジュールがあります。

| 分野 | モジュール | 主な用途 |
| --- | --- | --- |
| 文字列とデータ | `string`、`bytes`、`buffer` | 文字列操作、エンディアン／バイト補助、可変長バッファ |
| 数学 | `math` | 整数および数学演算の補助 |
| メモリと C 境界 | `mem`、`libc` | 手動メモリと C ランタイムのバインディング |
| ファイルと入出力 | `io`、`fs` | ファイルディスクリプタ、ファイル、ディレクトリ |
| ネットワーク | `net` | アドレス、ソケット、ポーリング、TCP、UDP |
| 環境、パス、時刻 | `env`、`path`、`time` | 環境変数、パス、時刻操作 |
| システムとプロセス | `sys`、`process` | OS 境界、プロセスの作成と待機 |

## インポートの粒度

通常はトップレベルモジュール名だけでなく、必要な API を定義している具体的な `.wave` ソース単位をインポートします。

```wave
import("std::string::len")::{len, is_empty};
import("std::fs::file")::{fs_open_read, fs_close};
import("std::mem::alloc")::{mem_alloc, mem_free};
```

選択的インポートでは、指定した公開宣言を名前空間の接頭辞なしで使用できます。通常の `import("std::string::len");` では、代わりに `std::string::len` 名前空間を通してアクセスします。

## API の調べ方

標準ライブラリは Wave ソースとして配布されます。次のコマンドで場所を確認できます。

```shell
wavec print std-path
```

モジュールの `.wave` ファイルを開き、公開シグネチャ、戻り値の契約、下位モジュールのインポートを確認してください。

## プラットフォーム境界

`sys`、`libc`、`fs`、`net`、`process` などのモジュールにある API は、負のエラー値を含むプラットフォーム形式の失敗結果をそのまま返すことがあります。成功を前提にせず、各関数の戻り値契約を確認して処理してください。

標準ライブラリのモジュールとコンパイラの言語・ABI 契約を一致させるため、コンパイラとともにインストールされた標準ライブラリを使用します。
