---
translation_set_id: whale-overview
path: whale/overview
locale: en
group: whale
group_order: 1
order: 1
title: Whale documentation
summary: Guide to toolchain components, available commands and reference documentation.
---

## Introduction

Whale is a general-purpose compiler toolchain for programming language implementations and compiler tools. Provides typed intermediate expressions (IR), AMD64 assembler, object file library, and linker-based functions. Each component is used through the Rust library and the `whale` command.

IR expresses the semantics of a computation independently of the machine language encoding. The assembler encodes machine instructions to create relocatable objects. Object libraries represent sections, symbols, and relocations. The linker resolves references between objects and places the executable file.

## Build and use tools

|work|document|
| --- | --- |
|What each tool does| [Toolchain components](/docs/en/whale/ecosystem) |
|Wave Program build/link/target selection| [Build and Link](/docs/en/whale/build-link-targets) |
|Package/dependency management| [Vex](/docs/en/whale/vex-package-manager) |
|Whale Execute command| [Whale CLI](/docs/en/whale/whale-cli) |

## Document Guide

|reference document|content|
| --- | --- |
| [See IR](ir-reference) |Types, values, functions, control flow, validation, exchange formats|
| [Numeric operations](numeric-operations) |Integer arithmetic, shift, type conversion, floating point|
| [Memory model](memory-model) |Initialization, pointer validation, address calculation, layout, strings|
| [Debugging with O0](o0-debugging) |Computation retention, variable storage, debug information|
| [AMD64 Target](amd64-target) |Target identifier, calling convention, native scope of function|
| [Assembler and Linker](assembler-linker) |Assembly operands, sections, symbols, static links|

The reference document defines the semantic rules for Whale. Available features follow the table below. The operations described in the reference documentation may not be available in all builds.

## Support status

|component|Interface provided|limit|
| --- | --- | --- |
|assembler|Create ELF64 relocatable object from assembly AMD64|Incomplete coverage of commands and directives|
|object library|Object composition, payload without BSS, target validation, AMD64 ELF64 serialization and selectable Wave record implementation to check size.|Relocatable objects are not executables|
| IR | Construction, printing, signature-checked direct/indirect calls, versioned AST lowering, target validation and checked type layouts | AST format 2 / typed IR format 4, bit-exact constants and text parsing, verification and round trips are available; machine call emission remains unavailable |
|linker|Symbol interpretation, input target verification, file/memory section placement inspection|Applying full relocation and outputting executable files are not supported.|
| Running and debugging | Integer/Bool, control flow, tracked stack memory and initialization checks | Float, calls, global addresses, native code generation and DWARF remain unsupported |

Rules prohibiting undefined behavior apply to verified IR and trace memories. The experimental implementation does not yet implement all of the runtime checks in the memory/execution reference documentation.

## Build and Assemble

Build with Rust 1.86.0 or higher.

```sh
git clone https://github.com/wavefnd/Whale.git
cd Whale
cargo build --release --locked
```

Save the following assembly as `answer.asm`.

```asm
section .text
global answer

answer:
    mov eax, 42
    ret
```

Create an object ELF64.

```sh
./target/release/whale asm --amd64 answer.asm -o answer.o
```

It uses Whale's own assembler, so no external assembler is required. The output contains a callable function and does not contain any process startup code.

The experimental AST→IR commands are available by building with `--features socket-cli`. For the command of CLI, refer to [Command Reference](/docs/en/whale/whale-cli).

## Library usage and diagnosis

IR must be verified before execution or code generation. Input errors and misuse of builder will return structural errors. Input errors in the library must not terminate the host process or overwrite existing content. Tools that accept untrusted input should be able to adjust resource limits.

The same toolchain version, input, target, and configuration should produce deterministic output. The distribution's metadata identifies the version·commit·and the available features. CI checks for good input, input to reject, trap, O0 retention, round-trip, native execution semantics for each interface. For development/verification commands, please refer to [Whale Contribution Information](https://github.com/wavefnd/Whale/blob/master/CONTRIBUTING.md).
