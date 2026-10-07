---
translation_set_id: whale-cli
path: whale/whale-cli
locale: ja
group: whale
group_order: 1
order: 3
title: Whale コマンドリファレンス
summary: Whaleassembler、objectwrapper、診断出力とオプションのIRコマンドについて説明します。
---

## Whaleビルド

Whaleリポジトリで以下を実行します。

```shell
cargo build --release
```

最上位の実行可能ファイルには4つのコマンドシリーズがあります。

```text
whale asm [--amd64 | --aarch64] <input> -o <output>
whale object <input> -o <output>
whale ir <subcommand> [options]
```

## AMD64 assembler

```shell
whale asm --amd64 input.asm -o output.o
```

AMD64 assembler は `.o` パスを出力として受け取り、section、symbol、および relocation を含む ELF64 relocatable オブジェクトを作成します。

詳細診断出力は`--debug-whale`でオンにします。

```shell
whale asm --amd64 input.asm -o output.o \
  --debug-whale --token --ast --bytes --dump-hex --stats
```

診断フラグには`--token`、`--ast`、`--bytes`、`--dump-hex`、`--dump-bin`、`--dump-json`、`--stats`があります。 `--trace`は処理プロセスを出力します。

## Object wrapper

```shell
whale object input.bin -o output.o
```

`object` コマンドは、生のバイトを ELF64 `.text` セクションに配置し、グローバル `start` シンボルをオフセット 0 に追加します。生のマシン コードを ELF オブジェクト ファイルにラップします。

## テキスト IR の検証と出力

標準ビルドで format 4 typed IR を読み取り・検証できます。[IR リファレンス](ir-reference)の完全な例を `answer.wir` に保存してください。`print` は検証後に標準形式で出力し、失敗時は既存ファイルを保護します。IR 実行や native コード生成は行いません。

```shell
whale ir verify answer.wir
whale ir print answer.wir -o canonical.wir
```

## オプション IR socket

AST JSON の `ir lower` には `socket-cli` feature が必要です。テキスト IR の `verify` と `print` には不要です。

```shell
cargo run -p whale --features socket-cli -- ir lower program.json
cargo run -p whale --features socket-cli -- ir lower program.json -o program.wir
```

`ir lower`はWhalesocketschemaのJSONを読み、WhaleIRテキストIRはstdoutまたは`-o`パスに出力されます。 `--target <triple>`はターゲット文字列を置き換え、`--no-verify`は検証を省略します。

`ir lower` を使うには `socket-cli` でビルドしてください。Socket JSON の生成側と Whale は同じ AST schema version を使う必要があります。


## スカラー整数インタプリタ

デフォルトビルドはスカラー整数・Bool IR も実行します。`--function @fN` は必須で、`--arg` を繰り返して正確な十進引数を渡します。Bool は true/false です。`--max-steps` は命令と terminator を数え、既定値は 1,000,000 です。run では -o と --no-verify は使えません。[IR 参照](ir-reference)の完全なループを `swap-loop.wir` に保存してください。

```shell
whale ir run swap-loop.wir --function @f7 --arg 3 --max-steps 100
```

```text
i32 22
```

## 追跡スタックメモリの実行

標準インタープリタは整数・Bool、制御フロー、スタック割り当て、データポインタの保存と読取り、typed GEP、memcpy・memsetを実行します。checked ペアの読取りではパディングを除外します。アドレスは合成64ビット値で、ホストメモリを参照しません。引数と戻り値は整数・Boolまたはvoidに限定され、float、呼出し、関数ポインタ、一般aggregate値、グローバルアドレス、native実行は未対応です。割り当ては関数のreturnまで有効です。字句スコープの寿命終了、呼出し・returnでのポインタ転送、外部メモリアダプタ、native shadow metadataは今後の実装です。

`--max-memory`は論理的な割り当てバイトの上限で、標準は64 MiBです。`InterpreterOptions::memory_limits`は割り当て数16384、ポインタのバイトメタデータ262144個、バイト・メタデータ処理256 Mi単位も制限します。超過はプログラムtrapとは別のIR位置付き`MemoryLimit`エラーです。検証器は出力ターゲットの既知の保存サイズoverflowも実行前に拒否します。

[メモリモデル](memory-model): `tracked-memory.wir`.

```shell
whale ir run tracked-memory.wir --function @f0 --max-memory 20
whale ir run tracked-memory.wir --function @f0 --max-memory 3
```

```text
u32 42
Error: tracked-memory.wir: interpreter memory Bytes limit 3 reached at @f0 %b0 instruction 0 (%v0)
```
