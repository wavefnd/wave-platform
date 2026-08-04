---
title: "Wave v0.1.9-pre-beta: Portable Toolchains, Freestanding Systems, and Release Hardening"
date: "2026-06-20 18:00:00"
description: "Wave v0.1.9-pre-beta strengthens the compiler as a portable systems toolchain with standalone packages, a redesigned build CLI, Windows support, freestanding code generation, explicit exports, safer inline assembly, and stricter release validation."
tags: ["wave-lang", "release", "compiler", "systems-programming"]
pinned: true
cover: "release-19.png"
---

# Wave v0.1.9-pre-beta: Portable Toolchains, Freestanding Systems, and Release Hardening

Wave v0.1.9-pre-beta is focused on making the compiler usable as a real systems toolchain rather than only as a language frontend. This release advances the raw compiler interface needed by package managers such as Vex, improves standalone distribution on Linux and Windows, and hardens the low-level code generation paths used by kernels and freestanding software.

This remains a pre-beta release. The supported surface is growing quickly, but target guarantees and language compatibility are not yet final.

## A Compiler CLI Designed for Build Tools

The `wavec` command line now centers on a single flag-driven build pipeline:

```bash
wavec build main.wave --emit=bin -o app
wavec build main.wave --emit=obj,ir --out-dir artifacts
wavec build start.o kernel.o --link-only --freestanding -o kernel.elf
wavec run main.wave -- argument-one argument-two
```

The compiler can infer Wave source, LLVM IR, LLVM bitcode, assembly, and object inputs from file extensions. Mixed inputs can move through one compile-and-link plan, while `--input-type` remains available when a build system needs to override inference.

New controls cover emitted artifacts, output directories, target triples, CPUs and target features, sysroots, linker selection, raw linker arguments, relocation models, static or shared output, PIE policy, freestanding builds, entry symbols, linker scripts, and start-file handling. `--dry-run` exposes the planned commands without executing them, which gives Vex and other build orchestrators a stable way to inspect compiler behavior.

Ambiguous combinations now have explicit rules. `--emit=check` is a standalone frontend mode, `--link-only` requires object-compatible input and a final binary, and `--run` requires exactly one executable output. These constraints make failures predictable for both people and automation.

## Standalone Linux and Windows Packages

Release packaging now bundles the LLVM 21 runtime and the LLVM tools that `wavec` needs. A user should not need to install a matching system LLVM merely to run a published compiler package.

The Linux x86_64 package keeps `wavec` as a native 64-bit ELF executable and resolves its bundled LLVM libraries through an embedded runtime search path. Its package also includes LLD and the compiler tools required by the build pipeline.

The Windows x86_64 package contains a native `wavec.exe`, Windows LLVM runtime libraries, LLD tools, and the MinGW-compatible objects and import libraries needed by the bundled Windows link path. The compiler can therefore build and run ordinary Wave programs on supported Windows 10 and Windows 11 systems without requiring a separate GCC installation from the end user.

Windows remains an experimental target while its standard-library and CI coverage catches up with the primary platforms. The standalone package is an important step, but it is not a declaration that every platform API is complete.

## Freestanding and Cross-Compilation Work

`wavec` now has explicit freestanding controls for kernels, bootloaders, firmware, and other environments that do not provide a hosted C runtime:

```bash
wavec build boot.wave \
  --target x86_64-unknown-none-elf \
  --freestanding \
  --emit=obj \
  -o boot.o
```

Freestanding and bare-metal targets disable the x86_64 red zone and mark generated Wave functions as non-unwinding. The linker path can omit default system libraries and start files, select a custom entry point, and consume a linker script. RISC-V target support and target-aware inline assembly validation also broaden the cross-compilation surface.

This release does not embed a WaveOS-hosted compiler. The current supported development model is a host `wavec` running on Linux, macOS, or Windows and producing objects or binaries for the target environment. A dedicated WaveOS target specification and, later, a WaveOS userland target remain separate milestones.

## Explicit ABI Exports

Wave functions can now be exported with an explicit ABI and optional symbol name:

```wave
export(c, "wave_add_i32") fun add_i32(a: i32, b: i32) -> i32 {
    return a + b;
}
```

This makes Wave code directly callable from C, C++, assembly, boot firmware, or another language that can consume the selected ABI. It also allows low-level software to define precise entry symbols such as firmware or kernel entry points without relying on accidental symbol naming.

The frontend shares the structural parsing logic used by `extern` declarations, while preserving the important direction difference: `extern` imports a foreign symbol and `export` publishes a Wave function.

## Safer Inline Assembly Contracts

Inline assembly now has explicit contracts for behavior the compiler cannot safely infer from operands alone.

`clobber("stack")` declares intentional stack modification. The compiler rejects undeclared stack-pointer changes and rejects unbalanced stack deltas in the supported instruction patterns. `clobber("noreturn")` identifies assembly that transfers control permanently, allowing the backend to terminate the LLVM block with `unreachable`. Conflicting contracts and invalid uses, such as `noreturn` on an expression that must produce a value, are diagnosed before object generation.

The generated LLVM inline assembly is treated as side-effecting, and target-specific register validation occurs against the selected target rather than only the host architecture. These rules are deliberately strict because silent inline assembly mistakes are especially destructive in kernels and boot code.

## Code Generation and Language Correctness

Several bugs exposed by real WaveOS and compiler tests have been fixed:

- Stores through `deref` now update the referenced memory instead of losing the assignment target.
- Pointer-to-array indexing and stores through pointer-valued structure fields generate the correct addresses.
- Structure field and array lvalues have dedicated regression coverage.
- Default function arguments are applied at call sites, including calls exercised by the end-to-end suite.
- Pointer-to-integer coercion errors now include the operation context instead of failing with an unqualified backend panic.
- Architecture-specific tests declare their target requirements instead of accidentally exercising host-only assumptions.

The Python end-to-end runner is also stricter. An unexpected nonzero process exit is now a failure, not a passing test. Tests can declare check-only, build-only, freestanding, target, emitted-artifact, and expected-exit metadata, so low-level compiler behavior can be verified without requiring an unavailable host linker.

## Standard Library Foundation

The standard library has expanded its libc-like foundation with byte handling, I/O, filesystem, process, memory, string, math, and system-facing abstractions. Higher-level protocols such as HTTP and game-engine APIs remain outside the core standard library. The goal is to provide the low-level, reusable pieces needed to build those libraries without forcing application policy into `std`.

## Release Validation and Reproducibility

Wave now tracks `Cargo.lock` for reproducible compiler builds and runs both Rust tests and Wave end-to-end programs in CI. A manual GitHub Actions release workflow performs the full validation sequence before it can create a tag or GitHub release:

- Rust formatting and Clippy with warnings denied
- all Rust targets and code generation regression tests
- an optimized compiler build and version check
- the complete Wave end-to-end suite
- Linux, macOS, and Windows package builds
- isolated package smoke tests and SHA-256 verification

The release job creates a draft by default. If any validation, packaging, architecture, runtime dependency, or smoke test fails, no release is created.

## What Comes Next

The next systems milestone is a named WaveOS target specification. Initially it can map to the existing bare-metal backend while fixing the default ABI, relocation model, code model, red-zone policy, runtime assumptions, and linker behavior in one target definition. A separate WaveOS userland target can follow when the operating system provides stable process, filesystem, memory, and syscall interfaces.

Longer term, running `wavec` inside WaveOS is a hosted-compiler project of its own. It requires a sufficient WaveOS userland and a port of the current Rust and LLVM-based compiler dependencies. Cross-compiling WaveOS from an existing host remains the correct path for this release.

## Links

- [Wave on GitHub](https://github.com/wavefnd/Wave)
- [Wave releases](https://github.com/wavefnd/Wave/releases)
- [Wave documentation](https://wave-lang.dev/docs/intro/)
- [Wave community](https://discord.gg/3nev5nHqq9)
