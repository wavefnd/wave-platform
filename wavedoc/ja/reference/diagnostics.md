---
translation_set_id: diagnostics
path: reference/diagnostics
locale: ja
group: reference
group_order: 3
order: 2
title: 診断とトラブルシューティング
summary: 人間向け・JSON 診断、検査モード、デバッグ出力、再現可能なバグ報告の手順を説明します。
---

## コンパイラバージョンを記録する

構文の動作を調査する前に、インストール済みコンパイラの正確なバージョンを記録します。

```shell
wavec --version
```

そのインストールで受け付けるコマンドとオプションの表記は `wavec --help` で確認してください。

## リンクせずにソースを検査する

Wave ソースのエラーをリンクや実行の問題から切り分けるには、次のコマンドを使います。

```shell
wavec build main.wave --emit=check
```

通常の実行ファイルを生成せずに Wave 入力を検査します。

## JSON 診断

IDE、CI、ビルドツールが構造化された診断を必要とする場合は、次の形式を使います。

```shell
wavec --error-format=json build main.wave --emit=check
```

ターミナルでは既定の人間向け形式を使い、自動処理するツールでは JSON を使用します。

## コンパイラ出力を調べる

```shell
wavec --debug-wave=tokens build main.wave --emit=check
wavec --debug-wave=ast build main.wave --emit=check
```

`--debug-wave` はトークン、AST、IR などの選択した表現を出力します。通常のソースエラーでは、最初の診断とソース位置から確認してください。コンパイラの動作を調査するときや開発ツールを作るときにデバッグ出力を使います。

## よくある失敗の分類

1. **構文解析／型エラー**は `--emit=check` でも失敗します。
2. **インポートエラー**では `std-path`、`--dep-root`、`--dep`、実際のファイルシステム構成を確認します。
3. **リンクエラー**では `--link`、`-L`、ターゲット ABI、シンボル名を確認します。
4. **実行時エラー**はビルド成功後に発生するため、終了ステータスと実行環境によって切り分けます。
5. **FFI エラー**では、幅、文字列表現、ポインタのライフタイム、呼び出し規約をネイティブ宣言と照合します。

## 有用なバグ報告

次の情報を含めてください。

- `wavec --version`
- ホスト OS とターゲットトリプル
- 実行した完全なコマンド
- 問題を再現する最小の `.wave` ソース
- 診断出力の全文
- 期待した動作と実際の動作

ログを公開する前に、関係のない秘密情報、トークン、非公開パスを削除してください。
