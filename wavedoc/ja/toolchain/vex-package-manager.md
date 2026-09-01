---
translation_set_id: vex-package-manager
path: toolchain/vex-package-manager
locale: ja
group: toolchain
group_order: 4
order: 3
title: Vex パッケージマネージャー
summary: マニフェスト形式の Wave プロジェクト、Git とパス依存関係、ロックファイル、オフラインビルド、wavec との境界を説明します。
---

## 役割

Vex は Wave のパッケージマネージャー兼ビルドツールです。Vex は `wavec` の上位に位置し、プロジェクト構造と依存関係の解決を管理します。`wavec` はコンパイラフラグとコンパイルパイプラインを管理します。

Vex のコマンドはマニフェストに基づきます。生の `wavec` フラグを `vex build`、`vex check`、`vex run` へ直接渡すことはできません。

## パッケージの作成

```shell
vex init
vex init --lib
```

アプリケーションは `src/main.wave`、ライブラリは `src/lib.wave` を使います。パッケージルートの構成は次のとおりです。

```text
my_project/
├── src/
│   └── main.wave
├── vex.ws
├── vex.lock
└── .vex/
    └── deps/
```

`vex.ws` がマニフェストです。Vex は `.wson` 拡張子のマニフェストを使用しません。

```wson
{
    name = "my_project",
    version = 0.1.0,
    lib = false,
    description = "my_project Project",
    author = "unknown",
    license = "Unknown",
    dependencies = []
}
```

## ビルドコマンド

```shell
vex build [--target <triple>] [--release] [--dry-run] [--locked] [--offline]
vex check [--target <triple>] [--release] [--dry-run] [--locked] [--offline]
vex run   [--target <triple>] [--release] [--dry-run] [--locked] [--offline] [-- <args...>]
```

Vex はコマンド面を小さく保ちます。コンパイラ固有の出力、リンカー、CPU、ABI、デバッグ制御が必要な場合は `wavec` を直接使用します。Vex に特定のコンパイラバイナリを使わせる場合は `VEX_WAVEC=/path/to/wavec` を設定します。

`Resolving`、`Fetching`、`Compiling`、`Checking`、`Running`、`Finished` などの進行段階は標準エラー出力に書き込まれます。プログラム出力は標準出力に保持されます。

## Git を中心とした依存関係

Vex はローカルパスまたは Git URL から依存関係を解決します。各依存関係では、ソース形式をひとつだけ指定します。

```wson
{
    name = "app",
    version = 0.1.0,
    dependencies = [
        { name = "local_math", path = "../local_math" },
        { name = "remote_math", git = "https://github.com/example/math.git", tag = "v0.1.0" }
    ]
}
```

Git 依存関係では `branch`、`tag`、`rev` のいずれかひとつだけを指定できます。すべての依存関係ルートには固有の `vex.ws` が必要です。Vex は依存先のマニフェストを再帰的に解決し、競合するパッケージ識別情報を拒否し、管理対象の Git チェックアウトを `.vex/deps/<name>` に格納します。

## ロックファイルの契約

スキーマ v2 の `vex.lock` は、完全な推移的依存グラフと正確な Git コミットを記録します。マニフェストとともにコミットしてください。同じマニフェストと有効なロックファイルを使うと、Vex はブランチやタグを再追跡せず、同じ依存グラフを選択します。

依存関係が必要なコマンドは自動的に解決を行います。次のコマンドでグラフを明示的に準備することもできます。

```shell
vex fetch
vex update
vex update math shared_core
```

`vex update` はすべての Git パッケージ、または指定したパッケージと影響を受ける推移的グラフだけを更新します。関係のないロック済みパッケージは選択済みコミットを維持します。

## ロック済み・オフラインのワークフロー

`--locked` は `vex.lock` の作成と変更を禁止します。ファイルがない、未知のスキーマを使っている、マニフェストグラフと一致しない場合に失敗します。ロックファイルに固定済みのコミットは取得することがあります。

`--offline` は Git のネットワーク操作を禁止します。必要なチェックアウトとコミットがローカルに存在していなければなりません。

```shell
vex fetch --locked
vex build --locked --offline
```

この組み合わせが厳密な CI ワークフローです。ネットワークを使用できる間にロックされた正確なコミットを取得し、その後はネットワークアクセスやロックファイル変更なしでコンパイルします。ドライランでは依存関係を取得せず、ロックファイルも書き換えません。

## コンパイラのセットアップと調査

```shell
vex info
vex setup wavec
vex setup wavec --version <version>
vex --version
```

Vex は、コンパイルを開始する前に `wavec --dry-run` が返す JSON ビルド計画を検証します。スキーマバージョンが一致しない場合、ビルドを開始せず互換性エラーを報告します。
