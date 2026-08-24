---
translation_set_id: ecosystem
path: toolchain/ecosystem
locale: en
group: toolchain
group_order: 4
order: 1
title: Whale toolchain
summary: The role and component boundaries of Wave's separate low-level assembler, object, linker, and IR toolchain.
---

## What Whale is

Whale is a separate low-level toolchain written in Rust. It provides reusable assembly, object, linking, and intermediate-representation components for Wave and other native-code producers.

Whale is not the name of the entire Wave developer experience. The projects have distinct responsibilities:

| Project | Responsibility |
| --- | --- |
| `wavec` | Validate and compile Wave source, generate artifacts, and coordinate native linking. |
| Vex | Manage Wave packages, manifests, dependency graphs, lockfiles, and package builds. |
| Whale | Develop independent low-level assembler, object, linker, and IR components. |
| Wave `std` | Provide Wave source modules for runtime and system APIs. |

## Components

The Whale workspace is organized into four main library areas:

- `assembler`: tokenization, AMD64 parsing and encoding, sections, symbols, and relocations
- `object`: an object-file model and ELF64 writer
- `linker`: link planning and executable-construction components
- `ir`: Whale IR types, builders, printing, verification, and an optional frontend socket

The `whale` executable exposes these areas as `asm`, `object`, `link`, and `ir` commands.

## Tool boundaries

`wavec` and Whale are independent command-line tools. Installing or invoking Whale does not change a `wavec` build.

Vex sends package build plans to `wavec`; it does not forward Whale options. Invoke each tool through its own command-line interface.

## Choosing the command

Use `whale asm --amd64` to assemble AMD64 source into an ELF64 relocatable object. Use `whale object` to wrap raw bytes in an ELF64 object. The IR socket command is available in builds that enable `socket-cli`. Use the platform linker or `wavec` for an executable build, and validate generated objects with the object-format tools used by your target platform.
