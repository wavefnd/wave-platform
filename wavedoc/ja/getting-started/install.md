---
translation_set_id: install
path: getting-started/install
locale: ja
group: getting-started
group_order: 1
order: 2
title: Wave のインストール
summary: 公式スクリプトで Wave コンパイラと Vex パッケージマネージャーをインストールし、ツールチェーンを確認する方法を説明します。
---

## 公式インストールスクリプト

Linux と macOS では、公式シェルインストーラーによって `wavec` コンパイラと `vex` パッケージマネージャーの最新リリースをまとめてインストールできます。

```shell
curl -fsSL https://wave-lang.dev/install.sh | bash -s -- latest
```

Windows x86-64 では PowerShell インストーラーを実行します。Windows GNU ビルドの `wavec` と Windows MSVC ビルドの `vex` が、同じツールチェーンディレクトリにインストールされます。

```powershell
irm https://wave-lang.dev/install.ps1 -OutFile install.ps1
powershell -ExecutionPolicy Bypass -File .\install.ps1 -Latest
```

実行中のシェル環境をインストーラーが更新できない場合は、表示された PATH 設定に従うか、新しいシェルを開いてください。Windows インストーラーは既定で `%LOCALAPPDATA%\Wave\bin` を使用し、ユーザー PATH に追加します。どちらのインストーラーも、既存のインストールを置き換える前に公開済み SHA-256 チェックサムを検証します。

Wave と Vex のリリースバージョンは互いに独立しています。最新の Vex リリースを自動選択せず、両方のバージョンを固定する場合は次のように指定します。

```shell
curl -fsSL https://wave-lang.dev/install.sh | bash -s -- --version <wave-release> --vex-version <vex-release>
```

```powershell
powershell -ExecutionPolicy Bypass -File .\install.ps1 -Version <wave-release> -VexVersion <vex-release>
```

## インストールの確認

```shell
wavec --version
wavec --help
vex --version
vex --help
```

インストールや文書の問題を報告するときは、この出力を含めてください。

## リリースアーカイブ

リリースアセット名には選択したバージョンが含まれます。正確なバージョン文字列と利用可能なプラットフォームはリリースページで確認できます。

| プラットフォーム | アーカイブ名の形式 |
| --- | --- |
| Linux x86-64 GNU | `wave-<version>-x86_64-linux-gnu.tar.gz` |
| Windows x86-64 GNU | `wave-<version>-x86_64-pc-windows-gnu.zip` |
| macOS Apple Silicon | `wave-<version>-aarch64-apple-darwin.tar.gz` |

手動でダウンロードしたアーカイブを検証するための `SHA256SUMS` アセットも公開されます。

## ソースからのビルド

コンパイラリポジトリのビルドには Rust ツールチェーンが必要です。

```shell
git clone https://github.com/wavefnd/Wave.git
cd Wave
cargo build
```

このコマンドはデフォルトブランチをビルドします。公開リリースを再現する場合は、Cargo を実行する前に対象タグをチェックアウトしてください。

```shell
git checkout <release-tag>
```

最適化されたコンパイラバイナリを作成する場合は次を実行します。

```shell
cargo build --release
```

生成される実行ファイルは通常 `target/debug/wavec` または `target/release/wavec` です。

## インストール後の確認事項

- `wavec --version` で想定したバージョンか確認します。
- `vex --version` でパッケージマネージャーのバージョンを確認します。
- `wavec print supported-targets` で認識されるターゲットを確認します。
- `wavec print std-path` で標準ライブラリの場所を確認します。
- シェルが `wavec` を見つけられない場合は、まずインストール先と PATH を確認します。

> **リモートスクリプトの確認**
>
> 利用環境のセキュリティポリシーで必要とされる場合は、実行前に `install.sh` または `install.ps1` をダウンロードして内容を確認してください。
