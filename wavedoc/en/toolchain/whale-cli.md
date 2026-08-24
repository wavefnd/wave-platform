---
translation_set_id: whale-cli
path: toolchain/whale-cli
locale: en
group: toolchain
group_order: 4
order: 4
title: Whale command reference
summary: AMD64 assembly, ELF64 object construction, developer inspection, and the optional Whale IR socket command.
---

## Build Whale

From the Whale repository:

```shell
cargo build --release
```

The command-line workflow provides assembly, object construction, and an optional IR interface:

```text
whale asm --amd64 <input> -o <output>
whale object <input> -o <output>
whale ir <subcommand> [options]
```

## AMD64 assembler

```shell
whale asm --amd64 input.asm -o output.o
```

The AMD64 assembler writes an ELF64 relocatable `.o` file containing the assembled sections, symbols, and relocations represented by the input.

Developer inspection begins with `--debug-whale`:

```shell
whale asm --amd64 input.asm -o output.o \
  --debug-whale --token --ast --bytes --dump-hex --stats
```

Available inspection flags include `--token`, `--ast`, `--bytes`, `--dump-hex`, `--dump-bin`, `--dump-json`, and `--stats`. `--trace` prints pipeline progress.

## Object wrapper

```shell
whale object input.bin -o output.o
```

`object` reads raw bytes, places them in an ELF64 `.text` section, and adds a global `start` symbol at offset zero. Use the target platform's linker when an executable is required.

## Optional IR socket

The `ir` command is compiled only with Whale's `socket-cli` feature:

```shell
cargo run -p whale --features socket-cli -- ir lower program.json
cargo run -p whale --features socket-cli -- ir lower program.json -o program.wir
```

`ir lower` reads socket JSON matching the Whale frontend schema, converts it to Whale IR, verifies the module by default, and writes textual IR to stdout or `-o`. `--target <triple>` selects the target string and `--no-verify` skips verification.

Without `socket-cli`, `whale ir` exits with an error that explains how to rebuild it. Socket JSON is versioned internal interchange; producers must match the socket version used by the Whale build.
