---
title: "Announcing Wave v0.1.7-pre-beta Release: The Foundation for System-Level Excellence"
date: "2026-02-09 08:54:58"
description: "We are proud to announce a major milestone in the evolution of the Wave programming language. This release is not just a collection of patches; it represents a fundamental shift in Wave’s capabilities. From a completely overhauled CLI to sophisticate..."
tags: []
cover: "https://cdn.hashnode.com/res/hashnode/image/upload/v1770627265960/01a96a8c-fe7d-4837-b7ad-36eaeea1524f.png"
---

# Announcing Wave v0.1.7-pre-beta Release: The Foundation for System-Level Excellence

We are proud to announce a major milestone in the evolution of the Wave programming language. This release is not just a collection of patches; it represents a fundamental shift in Wave’s capabilities. From a completely overhauled CLI to sophisticated C ABI interoperability and a comprehensive Linux system interface, Wave is now equipped for serious systems engineering.

---

### 1. Mastering Interoperability: System V ABI & C FFI
The implementation of a robust **C Calling Convention (ABI) lowering** mechanism allows Wave to handle complex data structures with the same precision as a native C compiler.

*   **ABI Compliance**: Our LLVM backend now strictly follows standard ABI rules, including **SRet** (Structured Return) for large aggregates, **ByVal** for passing structures by value, and **Split/HFA** (Homogeneous Floating-point Aggregates) for high-performance vector passing.
*   **The `extern` Keyword**: Declare external C functions with ease, including support for block syntax and symbol redirection.
*   **Safety in Packing**: We’ve refactored aggregate packing to use `build_memcpy` instead of simple bit-casts, ensuring memory alignment requirements are strictly respected during FFI calls.

### 2. Direct Kernel Access: Linux x86_64 Syscall Suite
Wave now speaks the language of the Linux kernel natively via high-precision inline assembly. This foundation has allowed us to build the initial **Wave Standard Library (`std`)**:

*   **`std::sys::linux`**: Native, zero-overhead interfaces for FS, Memory Management (`mmap`), Process Control, and Time.
*   **`std::net`**: Synchronous networking with `TcpListener`, `TcpStream`, and `UdpSocket`.
*   **`std::libc`**: Ready-to-use bindings for essential C functions like `malloc`, `free`, `stdio`, and `unistd`.

> *Note: The standard library is in early development. While we are stabilizing the internal coupling, we recommend using `std/libc` for critical external bindings.*

### 3. Hardened Inline Assembly: Clobbers and Normalization
Low-level programming requires total control. We’ve upgraded our `asm` blocks to support **clobber clauses** (e.g., `clobber("rax", "memory")`). This prevents subtle optimization bugs by explicitly informing the compiler about register and memory trashing.

The new **AsmPlan** engine automatically handles register normalization and sign-aware operand extension, making inline assembly both safer and more expressive.

### 4. Advanced Type System: Enums and Type Aliases
Wave now provides better tools for data modeling and code readability:

*   **Enums**: Define named constants with a specific underlying representation. Enum variants are automatically treated as global constants.
    ```rust
    enum ShaderUniformType -> i32 { FLOAT = 0, VEC2, VEC3 }
    ```
*   **Type Aliases**: Simplify complex type signatures and improve code reuse.
    ```rust
    enum ShaderUniformType -> i32 {
        FLOAT = 0,
        VEC2,
        VEC3,
        VEC4
    }

    type UniformType = ShaderUniformType;
    ```
*   **Type Resolution Pass**: A new compiler pass flattens aliases and resolves enum types across the entire program before code generation begins.

### 5. Iterative Constant Evaluation
Our constant evaluator is now significantly more powerful. It supports **multi-round resolution**, allowing constants to depend on other constants defined later in your code.

*   **Aggregate Support**: You can now define constants that are **struct literals** or **array literals**.
*   **Built-in Keywords**: `true`, `false`, and `null` are now fully supported in constant contexts.
*   **Frontend Validation**: The compiler now catches undeclared identifiers and type mismatches much earlier in the verification phase.

### 6. A Modern, Structured CLI
We have completely refactored the `wavec` toolchain with a command-dispatch system:

*   **Commands**: `run`, `build`, `install std`, and `update std`.
*   **Enhanced Linking**: Use `--link` and `-L` to link against external system libraries directly.
*   **Improved Diagnostics**: Refined RGB color palette and accurate source-pointing for a better developer experience.

### 7. Formalizing the Project
*   **Wave Foundation**: Core maintenance and assets are now officially managed by the Wave Foundation.
*   **MPL-2.0 License**: All source files are now licensed under the Mozilla Public License v2.0.
*   **Rich Examples**: Explore `examples/` for everything from a **mini-game** to a **TCP HTTP server** and **Graph algorithms (DFS/BFS)**.

---

### Looking Ahead
Wave v0.1.7-pre-beta transforms the language into a tool capable of building its own ecosystem. With C FFI, enums, and direct syscall support, Wave is ready for the next level of system programming.

## Get started:
```bash
# Install the compiler
# Set up the standard library
wavec install std

# Run the new Enum/Type Alias example
wavec run examples/type_enum.wave
```

## Install

```bash
curl -fsSL https://wave-lang.dev/install.sh | bash -s -- --version v0.1.7-pre-beta
```

*Wave: Build fast, stay lean, and control the machine.*

*Build fast, stay lean, and keep waving!*

---

### Link

* [Release Note](https://github.com/wavefnd/Wave/releases/tag/v0.1.7-pre-beta)
    
* [GitHub](https://github.com/wavefnd/Wave)
    
* [Community](https://discord.gg/3nev5nHqq9)