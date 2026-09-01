---
translation_set_id: compiler
path: getting-started/compiler
locale: ja
group: getting-started
group_order: 1
order: 3
title: コンパイラコマンドリファレンス
summary: ビルド、検査、実行、リンク、ターゲット照会、ビルドツール連携に使う wavec コマンドを説明します。
---

## コマンドモデル

`wavec` はコンパイラの CLI です。個別の入力を直接コンパイルし、ツールにコンパイラ機能を公開し、インストール済み標準ライブラリのソースを管理できます。

```text
wavec [global-options] <command> [command-options]
```

| コマンド | 用途 |
| --- | --- |
| `wavec build <input...>` | フラグで選択された検査、コード生成、リンク、実行のパイプラインを処理します。 |
| `wavec check <file>` | `build <file> --emit=check` の別名です。 |
| `wavec run <file> [-- <args...>]` | `build <file> --run` の別名です。`--` 以降の引数はプログラムへ渡されます。 |
| `wavec print <item>` | ターゲットとツールチェーンの機能を照会します。 |
| `wavec install std` | 標準ライブラリをインストールします。 |
| `wavec update std` | インストール済み標準ライブラリを更新します。 |
| `wavec --version` | コンパイラと LLVM バックエンドのバージョンを表示します。 |

インストール済みコンパイラが提供する完全なオプション一覧は `wavec --help` で確認できます。

## ビルド、検査、実行

```shell
wavec build main.wave
wavec check main.wave
wavec run main.wave -- first-argument second-argument
```

`build` は既定で実行ファイルを生成します。`check` はフロントエンド検証後に停止します。`run` にはバイナリ出力が必要で、共有ライブラリのビルドとは併用できません。

コンパイル、リンク、実行を行わずに要求を検証し、予定されている段階を確認するには `--dry-run` を使います。

```shell
wavec build main.wave --target riscv64-unknown-linux-gnu --dry-run
wavec build main.wave --dry-run --error-format=json
```

JSON 形式は Vex などのビルドツールが使用する安定した連携インターフェースです。

## 出力形式と入力形式

```shell
wavec build main.wave --emit=ast
wavec build main.wave --emit=ir
wavec build main.wave --emit=bc
wavec build main.wave --emit=asm
wavec build main.wave --emit=obj -o main.o
wavec build main.wave --emit=bin -o app
```

アーティファクトの出力形式は `ast`、`ir`、`bc`、`asm`、`obj`、`bin` です。`check` は制御モードであり、単独で使用します。パイプラインが許可する場合は、複数のアーティファクト形式をカンマ区切りで指定できます。

受け付ける入力形式は `wave`、`ir`、`bc`、`asm`、`obj`、`archive` です。`--input-type=<kind>` はすべての入力形式をひとつに固定します。オブジェクトまたはアーカイブ入力をバイナリへリンクするときは `--link-only` を使います。

```shell
wavec build module.o --input-type=obj --link-only --emit=bin -o app
```

## 出力先

| オプション | 効果 |
| --- | --- |
| `-o <file>` | 主出力のパスを設定します。 |
| `--out-dir <dir>` | 生成されたアーティファクトを指定ディレクトリに配置します。 |
| `--target-dir <dir>` | 中間生成物と既定アーティファクトのルートを指定します。 |

## 最適化とコンパイラの調査

```shell
wavec -O2 build main.wave
wavec --debug-wave=tokens,ast build main.wave
```

最適化レベルは `-O0`、`-O1`、`-O2`、`-O3`、`-Os`、`-Oz`、`-Ofast` です。`--debug-wave` には `tokens`、`ast`、`ir`、`mc`、`hex`、`all` を指定でき、複数の段階はカンマで組み合わせられます。

## ネイティブリンク

```shell
wavec --link=m -L ./lib build main.wave
wavec build main.wave --shared -o libexample.so
wavec build main.wave --static -o app
wavec build main.wave --pie -o app
```

`--link=<lib>` はネイティブライブラリを追加し、`-L <path>` は検索パスを追加します。リンクモードには、それぞれの互換性規則に従う `--shared`、`--static`、`--pie`、`--no-pie` があります。

バックエンドとリンカーの主な制御オプションは次のとおりです。

- `--target`、`--cpu`、`--features`、`--abi`、`--sysroot`
- `-C linker=<path>` と `-C link-arg=<arg>`
- `-C link-sysroot=<path>` と `-C relocation-model=<model>`
- `-C no-default-libs`

カーネルなどのフリースタンディング出力では、適切な `--entry`、`--linker-script`、`--no-start-files` とともに `--freestanding` を使います。

## 外部パッケージの解決

```shell
wavec --dep-root .vex/deps build main.wave
wavec --dep math=/opt/wave-deps/math build main.wave
```

`--dep-root <dir>` は外部の `package::module` インポートを解決するためのルートを追加します。`--dep <name>=<path>` はパッケージ名をひとつのディレクトリに固定します。これらはコンパイラとの連携点であり、プロジェクトマニフェスト、依存関係の取得、ロックファイルは Vex が管理します。

## 機能の照会

ツール側にコンパイラ機能をハードコードしないでください。インストール済みコンパイラへ照会します。

```shell
wavec print host-target
wavec print target-spec --format=json
wavec print supported-targets
wavec print supported-input-types
wavec print supported-emit-kinds
wavec print supported-print-items
wavec print cpu-list --target riscv64-unknown-linux-gnu
wavec print target-features --target riscv64-unknown-linux-gnu
wavec print default-linker
wavec print sysroot
wavec print std-path
wavec print dep-search-paths
```

ほかに照会できる項目として `host`、`default-target`、`target-list` があります。構造化出力に対応する項目では `--format=json` を指定できます。

## コンパイラとツールチェーンの境界

`wavec` は Wave ソースをコンパイルし、生成アーティファクト、ターゲット、リンクを制御します。Vex はパッケージマニフェスト、依存グラフ、ロックファイル、再現可能なパッケージビルドを管理します。Whale は独自のコマンドを持つ別の低レベルツールチェーンです。
