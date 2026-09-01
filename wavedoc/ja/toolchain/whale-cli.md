---
translation_set_id: whale-cli
path: toolchain/whale-cli
locale: ja
group: toolchain
group_order: 4
order: 4
title: Whale コマンドリファレンス
summary: AMD64 アセンブリ、ELF64 オブジェクト構築、開発者向け調査、任意の Whale IR ソケットコマンドを説明します。
---

## Whale のビルド

Whale リポジトリで次を実行します。

```shell
cargo build --release
```

コマンドラインのワークフローは、アセンブリ、オブジェクト構築、任意の IR インターフェースを提供します。

```text
whale asm --amd64 <input> -o <output>
whale object <input> -o <output>
whale ir <subcommand> [options]
```

## AMD64 アセンブラ

```shell
whale asm --amd64 input.asm -o output.o
```

AMD64 アセンブラは、入力で表現されたセクション、シンボル、リロケーションを含む ELF64 再配置可能 `.o` ファイルを書き出します。

開発者向けの調査には、まず `--debug-whale` を使います。

```shell
whale asm --amd64 input.asm -o output.o \
  --debug-whale --token --ast --bytes --dump-hex --stats
```

利用できる調査フラグには `--token`、`--ast`、`--bytes`、`--dump-hex`、`--dump-bin`、`--dump-json`、`--stats` があります。`--trace` はパイプラインの進行を表示します。

## オブジェクトラッパー

```shell
whale object input.bin -o output.o
```

`object` は生のバイト列を読み込み、ELF64 の `.text` セクションに配置し、オフセット 0 にグローバルな `start` シンボルを追加します。実行ファイルが必要な場合は、ターゲットプラットフォームのリンカーを使います。

## 任意の IR ソケット

`ir` コマンドは、Whale の `socket-cli` 機能を有効にした場合にだけコンパイルされます。

```shell
cargo run -p whale --features socket-cli -- ir lower program.json
cargo run -p whale --features socket-cli -- ir lower program.json -o program.wir
```

`ir lower` は Whale フロントエンドスキーマに合うソケット JSON を読み込み、Whale IR へ変換し、既定でモジュールを検証して、テキスト IR を標準出力または `-o` へ書き出します。`--target <triple>` はターゲット文字列を選択し、`--no-verify` は検証を省略します。

`socket-cli` がない場合、`whale ir` は再ビルド方法を説明するエラーで終了します。ソケット JSON はバージョン付きの内部交換形式であり、生成側は Whale ビルドが使用するソケットバージョンに合わせる必要があります。
