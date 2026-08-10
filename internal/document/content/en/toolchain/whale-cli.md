---
translation_set_id: whale-cli
path: toolchain/whale-cli
locale: en
group: toolchain
group_order: 4
order: 4
title: Whale command reference
summary: Current Whale assembler, object, linker, and optional IR commands with their implementation limits.
---

## Build Whale

From the Whale repository:

```shell
cargo build --release
```

The top-level executable exposes four command families:

```text
whale asm [--amd64 | --aarch64] <input> -o <output>
whale object <input> -o <output>
whale link <...>
whale ir <subcommand> [options]
```

These commands are under development. Treat the status below as part of the command contract.

## AMD64 assembler

```shell
whale asm --amd64 input.asm -o output.o
```

The implemented assembler path is AMD64 and currently requires an `.o` output. It produces an ELF64 relocatable object, preserving assembled sections, symbols, and supported relocations. Although `--aarch64` appears in the top-level command shape, the assembler currently rejects it as unsupported.

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

The current `object` command reads raw bytes, places them in an ELF64 `.text` section, and adds a global `start` symbol at offset zero. It is a focused object-construction path, not yet a general object editor or IR-to-object frontend.

## Linker status

```shell
whale link object.o -o executable
```

The `link` command is a placeholder in the current CLI and does not produce an executable. Do not use it in a build pipeline yet.

## Optional IR socket

The `ir` command is compiled only with Whale's `socket-cli` feature:

```shell
cargo run -p whale --features socket-cli -- ir lower program.json
cargo run -p whale --features socket-cli -- ir lower program.json -o program.wir
```

`ir lower` reads socket JSON matching Whale's current frontend schema, lowers it to Whale IR, verifies the module by default, and writes textual IR to stdout or `-o`. `--target <triple>` changes the target string and `--no-verify` skips verification.

Without `socket-cli`, `whale ir` exits with an error that explains how to rebuild it. Socket JSON is versioned internal interchange; producers must match the socket version used by the Whale build.
