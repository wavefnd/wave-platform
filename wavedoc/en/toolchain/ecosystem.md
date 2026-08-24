---
translation_set_id: ecosystem
path: toolchain/ecosystem
locale: en
group: toolchain
group_order: 4
order: 1
title: Whale toolchain
summary: The role, component boundaries, and current maturity of the separate Whale low-level toolchain.
---

## What Whale is

Whale is a separate low-level toolchain written in Rust. Its long-term role is to provide reusable assembly, object, linking, and intermediate-representation components for Wave and other native-code producers.

Whale is not the name of the entire Wave developer experience. The projects have distinct responsibilities:

| Project | Responsibility |
| --- | --- |
| `wavec` | Parse, validate, compile, and currently generate native code through LLVM. |
| Vex | Manage Wave packages, manifests, dependency graphs, lockfiles, and package builds. |
| Whale | Develop independent low-level assembler, object, linker, and IR components. |
| Wave `std` | Provide Wave source modules for runtime and system APIs. |

## Components

The Whale workspace currently contains four main library areas:

- `assembler`: tokenization, AMD64 parsing and encoding, sections, symbols, and relocations
- `object`: an object-file model and ELF64 writer
- `linker`: the developing link layer
- `ir`: Whale IR types, builders, printing, verification, and an optional frontend socket

The `whale` executable exposes these areas as `asm`, `object`, `link`, and `ir` commands.

## Current integration status

The current `wavec` uses its LLVM backend for normal builds. Whale is developed as an independent toolchain and is not selected by a `wavec --whale` option.

This separation matters for both users and tool authors: do not assume that installing Whale changes `wavec`, and do not pass Whale options through Vex. Invoke each tool through its own command-line contract.

## Maturity boundary

Whale is under active development. The AMD64 assembler and ELF64 object paths are usable for focused experiments; the AArch64 assembler path and linker CLI are not complete. The IR socket command is opt-in at build time. Check the Whale command-reference page before depending on a subcommand in automation.

Until a component declares a stable contract, keep generated artifacts reproducible at the source level and validate object format, architecture, symbols, and relocations with independent tools.
