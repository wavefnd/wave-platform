---
translation_set_id: whale-cli
path: whale/whale-cli
locale: en
group: whale
group_order: 1
order: 3
title: Whale command reference
summary: Describes the Whale assembler, object wrapper, diagnostic output, and the optional IR commands.
---

## Whale Build

In the Whale repository, run:

```shell
cargo build --release
```

A top-level executable file has four families of commands:

```text
whale asm [--amd64 | --aarch64] <input> -o <output>
whale object <input> -o <output>
whale ir <subcommand> [options]
```

## AMD64 assembler

```shell
whale asm --amd64 input.asm -o output.o
```

AMD64 assembler receives the `.o` path as output, and ELF64 relocatable containing section, symbol and relocation Create object.

Turn on detailed diagnostic output with `--debug-whale`.

```shell
whale asm --amd64 input.asm -o output.o \
  --debug-whale --token --ast --bytes --dump-hex --stats
```

Diagnostic flags include `--token`, `--ast`, `--bytes`, `--dump-hex`, `--dump-bin`, `--dump-json`, and `--stats`. `--trace` prints the processing progress.

## Object wrapper

```shell
whale object input.bin -o output.o
```

The `object` command places raw bytes in an ELF64 `.text` section and adds a global `start` symbol at offset 0. It wraps raw machine code in an ELF object file.

## Typed text IR verification and printing

The default build reads and verifies format 4 typed IR. Save the complete example in the [IR reference](ir-reference) as `answer.wir`. `print` verifies before canonical printing and preserves existing output on failure. It does not execute IR or generate native code.

```shell
whale ir verify answer.wir
whale ir print answer.wir -o canonical.wir
```

## Optional IR socket

AST JSON `ir lower` requires the `socket-cli` feature. Text IR `verify` and `print` do not.

```shell
cargo run -p whale --features socket-cli -- ir lower program.json
cargo run -p whale --features socket-cli -- ir lower program.json -o program.wir
```

`ir lower` reads JSON of Whale socket schema, converts it to Whale IR, and verifies the module. The text IR is output to the path stdout or `-o`. `--target <triple>` replaces the target string and `--no-verify` omits validation.

Build with `socket-cli` to use `ir lower`. Socket JSON producers and Whale must use the same AST schema version.


## Scalar integer interpreter

The default build also runs scalar integer/bool IR. `--function @fN` is required; repeat `--arg` for exact decimal arguments, or true/false for Bool. `--max-steps` counts instructions and terminators and defaults to 1,000,000. run accepts neither -o nor --no-verify. Save the complete loop from [IR reference](ir-reference) as `swap-loop.wir`:

```shell
whale ir run swap-loop.wir --function @f7 --arg 3 --max-steps 100
```

```text
i32 22
```

## Tracked stack memory execution

The default interpreter executes integer/Bool values, control flow, stack allocations, data-pointer storage and loads, typed GEP, memcpy and memset. Checked-pair reads exclude padding. Addresses are synthetic 64-bit values, never host-memory dereferences. Function arguments and returns remain integer/Bool or void; float, calls, function pointers, general aggregate values, global addresses and native execution are unsupported. Stack allocations live until function return. Lexical lifetime-end operations, pointer transfer through calls/returns, foreign-memory adapters and native shadow metadata require further implementation.

`--max-memory` sets the logical allocation-byte budget, default 64 MiB. `InterpreterOptions::memory_limits` also limits allocation count to 16384, pointer-byte metadata to 262144 fragments, and byte/metadata work to 256 Mi units. Exceeding a budget returns an IR-located `MemoryLimit` error, separately from a program trap. Verification also rejects known output-target storage-size overflow before execution.

[Memory model](memory-model): `tracked-memory.wir`.

```shell
whale ir run tracked-memory.wir --function @f0 --max-memory 20
whale ir run tracked-memory.wir --function @f0 --max-memory 3
```

```text
u32 42
Error: tracked-memory.wir: interpreter memory Bytes limit 3 reached at @f0 %b0 instruction 0 (%v0)
```
