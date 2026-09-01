---
translation_set_id: build-link-targets
path: toolchain/build-link-targets
locale: ja
group: toolchain
group_order: 4
order: 2
title: ビルド、リンク、ターゲットオプション
summary: アーティファクト出力、入力形式、ネイティブリンク、ターゲット／CPU／ABI 制御、フリースタンディングのビルド計画を説明します。
---

## アーティファクトの出力

```shell
wavec build main.wave --emit=check
wavec build main.wave --emit=ast
wavec build main.wave --emit=ir
wavec build main.wave --emit=bc
wavec build main.wave --emit=asm
wavec build main.wave --emit=obj -o main.o
wavec build main.wave --emit=bin -o app
```

アーティファクトの出力形式は `ast`、`ir`、`bc`、`asm`、`obj`、`bin` です。`check` はアーティファクトを生成せずソースを検証し、単独で使用します。

```shell
wavec print supported-emit-kinds
```

## 入力形式とリンク専用モード

コンパイラは Wave ソース、IR、ビットコード、アセンブリ、オブジェクト、アーカイブ入力を区別します。インストール済みコンパイラへ次のように照会できます。

```shell
wavec print supported-input-types
```

生成済みのオブジェクトまたはアーカイブをリンクするには、`--input-type` と `--link-only` を組み合わせます。

```shell
wavec build module.o --input-type=obj --link-only --emit=bin -o app
```

## ネイティブリンク

```shell
wavec --link=m -L ./lib build main.wave
```

`--link` はライブラリを追加し、`-L` はライブラリ検索パスを追加します。FFI シンボルを宣言しても、そのシンボルを提供するライブラリが自動的にリンクされるわけではありません。

## ターゲットの選択

主な LLVM ターゲット制御は次のとおりです。

- `--target <triple>`
- `--cpu <name>`
- `--features <csv>`
- `--abi <name>`
- `--sysroot <path>`

ホストの既定値と対応ターゲットはコンパイラへ照会します。

```shell
wavec print host-target
wavec print supported-targets
wavec print target-spec --target <triple>
wavec print cpu-list --target <triple>
wavec print target-features --target <triple>
```

## 対応するターゲットファミリー

すべてのターゲットを含めてビルドされたコンパイラは、次のターゲット契約を提供します。特定のコンパイラビルドに含まれるターゲットは `wavec print supported-targets` で確認してください。

| ターゲット | 環境 | オブジェクト形式 |
| --- | --- | --- |
| `x86_64-unknown-linux-gnu` | ホスト環境 Linux GNU | ELF |
| `x86_64-apple-darwin` | ホスト環境 macOS | Mach-O |
| `x86_64-w64-windows-gnu` | ホスト環境 Windows GNU | COFF |
| `x86_64-pc-windows-gnu` | ホスト環境 Windows GNU の別名 | COFF |
| `x86_64-unknown-none-elf` | フリースタンディング | ELF |
| `aarch64-unknown-linux-gnu` | ホスト環境 Linux GNU | ELF |
| `aarch64-apple-darwin` | ホスト環境 macOS | Mach-O |
| `aarch64-unknown-none-elf` | フリースタンディング | ELF |
| `riscv64-unknown-linux-gnu` | ホスト環境 Linux GNU | ELF |
| `riscv64-unknown-none-elf` | フリースタンディング | ELF |

## RISC-V 64 の契約

ホスト環境向け RISC-V ターゲットの既定値は `generic-rv64`、RV64GC、`lp64d` ABI です。フリースタンディングターゲットの既定値は `generic-rv64`、RV64IMAC、`lp64` です。

```shell
wavec print target-spec --target riscv64-unknown-linux-gnu --format=json
wavec print target-spec --target riscv64-unknown-none-elf --format=json
```

対応する RISC-V CPU は `generic`、`generic-rv64`、`rocket-rv64`、`sifive-u74` です。機能の上書きには、`m`、`a`、`f`、`d`、`c`、`zicsr`、`zifencei` に符号を付け、カンマで区切って指定します。

```shell
wavec build main.wave \
  --target riscv64-unknown-linux-gnu \
  --features=+m,+a,+f,-d,+c,+zicsr \
  --abi=lp64f
```

RISC-V の検証では、一貫しない組み合わせが拒否されます。`d` には `f` が必要で、`f` には `zicsr` が必要です。また、`lp64`、`lp64f`、`lp64d` は有効な浮動小数点機能と一致しなければなりません。ABI を指定しない場合、コンパイラはこれらの機能から ABI を決定します。

## フリースタンディングリンク

```shell
wavec build kernel.wave \
  --target riscv64-unknown-none-elf \
  --freestanding \
  --entry=_start \
  --linker-script=linker.ld \
  --no-start-files \
  -o kernel.elf
```

`--freestanding` はビルドを既定ライブラリから切り離します。`--entry` はリンカーのエントリーシンボルを設定し、`--linker-script` はリンカースクリプトを指定し、`--no-start-files` はホスト環境用の起動ファイルを省略します。

実行前に予定されているビルドとリンクの段階を確認するには `--dry-run` を使います。

## ホスト環境向けクロスリンク

あるターゲット向けのコード生成だけでは、そのターゲットの C ランタイム、起動オブジェクト、ライブラリは提供されません。ホスト環境向けクロスビルドには互換性のある sysroot が必要で、場合によってはリンカーも明示します。

```shell
wavec build main.wave \
  --target riscv64-unknown-linux-gnu \
  --sysroot /path/to/riscv64-sysroot \
  -C linker=/path/to/target-linker \
  -o app-riscv64
```

sysroot には選択した ABI 用のファイルが必要です。同じ名前のホストライブラリをターゲットライブラリの代わりにはできません。

## クロスビルドのチェックリスト

- ターゲットトリプルがコンパイラの対応ターゲット一覧に含まれている。
- sysroot とリンカーがターゲット ABI に一致している。
- リンクするライブラリがターゲットアーキテクチャ向けにビルドされている。
- 要求する CPU 機能が選択した CPU で有効である。
- フリースタンディングビルドでは、エントリーシンボルとメモリレイアウトがリンカースクリプトに一致している。
