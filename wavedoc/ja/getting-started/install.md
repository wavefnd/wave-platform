---
translation_set_id: install
path: getting-started/install
locale: ja
group: getting-started
group_order: 1
order: 2
title: Waveのインストール
summary: Linux、macOS、WindowsにWaveをインストールし、最初のプログラムを実行します。
---

## インストール方針

インストーラーは、バージョン番号付きのプレリリースを含む最新の公開リリースだけをインストールします。Draft と Nightly は対象外です。旧バージョンと Nightly は下記の手動手順を使用してください。最新リリースに対応するパッケージがなければ、旧版を選ばず終了します。

| OS | アーキテクチャ |
| --- | --- |
| Linux | amd64, arm64, riscv64, loong64 |
| macOS | amd64, arm64 |
| Windows | amd64, arm64 |
| FreeBSD | amd64 |

## Linux、macOS、FreeBSD

Bash、curl、jq、tar、SHA-256 ツールが必要です。FreeBSD では先に bash、curl、jq をインストールしてください。Linux には互換性のある glibc とシステムライブラリ、macOS には Apple Command Line Tools、FreeBSD には互換性のある基本システムが必要です。

```shell
curl -fsSL https://wave-lang.dev/install.sh | bash
```

## Windows

PowerShell で実行します。Windows の amd64 または arm64 を検出し、MSVC パッケージを選択します。Visual C++ ランタイム、Windows SDK、MSVC/UCRT ライブラリが必要です。コンパイル検証用の環境は Visual Studio 開発者シェルでも設定できます。

```powershell
irm https://wave-lang.dev/install.ps1 -OutFile install.ps1
powershell -ExecutionPolicy Bypass -File .\install.ps1
```

## インストールオプション

既定では Wave と、対象プラットフォーム用のパッケージがある場合に最新の Vex をインストールします。Vex パッケージがなければ Wave だけをインストールして通知します。Vex のダウンロードや検証の失敗は省略せずエラーにします。

| | Bash オプション | PowerShell オプション |
| --- | --- | --- |
| Vex を必須にする | `--with-vex` | `-WithVex` |
| Wave だけをインストール | `--without-vex` | `-WithoutVex` |
| PATH を変更しない | `--no-modify-path` | `-NoModifyPath` |

更新にも同じコマンドを使用します。新しいコンパイラーと同梱 std の実行検証が成功するまで以前のインストールを保持し、失敗時は復元します。既定の場所は Unix では ~/.wave/bin、Windows では %LOCALAPPDATA%\Wave\bin です。WAVE_INSTALL_DIR で専用ディレクトリを指定できます。PATH の自動設定後は新しい端末を開いてください。

## 旧バージョンと Nightly の手動インストール

[GitHub Releases](https://github.com/wavefnd/Wave/releases) から目的のバージョン、OS、アーキテクチャの .tar.gz または .zip を取得します。wavec、std、llvm などの同梱ファイルの相対位置を保って展開し、wavec のあるディレクトリを PATH に追加します。旧版の構成や要件はそのリリースの案内に従ってください。インストーラーの --version、--vex-version、-Version、-VexVersion は使用できません。

## 最初の実行

インストーラーは同梱 std を使う小さなプログラムをコンパイルして実行します。最新の std を別途取得する必要はありません。インストール後に次の例を実行してください。Vex もインストールされた場合は vex --version でも確認できます。

```shell
wavec --version
```

任意のディレクトリにプレーンテキストファイル `main.wave` を作成し、以下のプログラム全体を保存してください。ファイル名が `main.wave.txt` になっていないことを確認してください。

<!-- wave-example: install-stdlib -->
```wave
import("std::string::len")::{
    len
};

fun main() {
    println("Wave: {} bytes", len("Wave"));
}
```

`main.wave` のあるディレクトリで端末を開くか、`cd` で移動してから実行してください:

```shell
wavec run main.wave
```

期待される出力:

```text
Wave: 4 bytes
```
