---
title: "Introduction to Wave v0.1.1-pre-beta: Inline Assembly, Pointer Chain, and Array Support"
date: "2025-05-05 11:46:33"
description: "Hello! I'm LunaStev, the developer of Wave.
We are excited to announce Wave v0.1.1-pre-beta — This update introduces inline assembly (asm {}) support, enabling you to write low-level system code directly in Wave, such as making syscalls with direct r..."
tags: []
cover: "https://cdn.hashnode.com/res/hashnode/image/upload/v1754270115332/6b7b8340-d3eb-4c3c-a714-e319515eace2.webp"
---

# Introduction to Wave v0.1.1-pre-beta: Inline Assembly, Pointer Chain, and Array Support

Hello! I'm LunaStev, the developer of Wave.

We are excited to announce Wave `v0.1.1-pre-beta` — This update introduces inline assembly (`asm {}`) support, enabling you to write low-level system code directly in Wave, such as making syscalls with direct register manipulation.

Additionally, Wave now fully supports pointer chaining (`ptr<ptr<i32>>`) and array types (`array<T, N>`), including index access, address-of operations, and validation of literal lengths — expanding Wave's capability for systems-level and memory-safe programming.

These improvements bring Wave closer to its vision as a low-level but expressive programming language.

---

## ✅ Added Features

### ⚙️ Inline Assembly (`asm { ... }`) Support

* Introduced `asm { ... }` block syntax to embed raw assembly instructions directly within Wave code.
    
* Supports instruction strings (e.g., `"syscall"`) and explicit register constraints via `in("reg") var` and `out("reg") var`.
    
* Variables used in `in(...)` are passed into specified registers; variables in `out(...)` receive output from registers.
    
* Supports passing literal constants directly to registers (e.g., `in("rax") 60`).
    
* Pointer values (e.g., `ptr<i8>`) are correctly passed to registers such as `rsi`, enabling low-level syscalls like `write`.
    
* Internally leverages LLVM's inline assembly mechanism using Intel syntax.
    
* Currently supports single-output only; multiple `out(...)` constraints will overwrite each other.
    
* Does not yet support clobber lists or advanced constraint combinations.
    
* Provides essential capability for system-level programming (e.g., making direct syscalls, writing device-level code).
    

> ⚠️ This is not a fully general-purpose inline ASM facility yet, but it enables practical low-level operations within Wave. Full support is planned for later phases.

### ⚙️ Make pointer chain explicit

* Nested parsing like `ptr<i32>`, `ptr<ptr<i32>>`
    
* Can create `ptr<T>` for any type (no restrictions on `T`)
    
* Support for consecutive `deref` operations (e.g., `deref deref deref`)
    

### ⚙️ Array type complete

* IndexAccess (`numbers[0]`) handling
    
* ArrayLiteral → Parse into AST and validate length
    
* AddressOf → Support array literals with address-of values (e.g., `[&a, &b]`)
    
* Confirmed that `array<T, N>` supports any type as T
    

## ✨ Other Changes

### 🧠 Library and Binary 2 Coexist

* Add lib.rs for easy package manager creation, development, and easy access.
    

---

## Showcase

![Image1description](https://cdn.hashnode.com/res/hashnode/image/upload/v1754270106823/d333ea14-16f8-4ffd-aec7-e7179ffc424b.png align="left")

![Image2description](https://cdn.hashnode.com/res/hashnode/image/upload/v1754270107747/ff489e36-2f67-4b53-a9e9-4db817140e38.png align="left")

---

![Image3description](https://cdn.hashnode.com/res/hashnode/image/upload/v1754270108480/60ae0d18-4054-445e-b021-434f6571b56e.png align="left")

![Image4description](https://cdn.hashnode.com/res/hashnode/image/upload/v1754270109260/49c75e00-4bf2-4aa3-a305-6337cf1e4951.png align="left")

---

![Image5description](https://cdn.hashnode.com/res/hashnode/image/upload/v1754270110178/7c4d4428-edb4-45fd-84ac-f07d523ca9ff.png align="left")

![Image6description](https://cdn.hashnode.com/res/hashnode/image/upload/v1754270111118/e47370b2-c636-4b19-9193-488e69b406a3.png align="left")

---

![Image7description](https://cdn.hashnode.com/res/hashnode/image/upload/v1754270111992/2f0a41b0-b47f-447f-8085-53684c5f57ac.png align="left")

![Image8description](https://cdn.hashnode.com/res/hashnode/image/upload/v1754270112958/b09a702a-90aa-40f9-89a3-b4bb1e2a7e53.png align="left")

---

![Image9description](https://cdn.hashnode.com/res/hashnode/image/upload/v1754270113726/f10143b2-a9a7-40ae-a16e-d27457bd94dc.png align="left")

![Image10description](https://cdn.hashnode.com/res/hashnode/image/upload/v1754270114590/dc6bd0b4-32a9-4630-a767-a5b97a56d9ea.png align="left")

---

Thank you for using Wave! Stay tuned for future updates and enhancements.

---

## Installation Guide

### For Linux:

1. **Download and Extract:**
    
    * Download the `wave-v0.1.1-pre-beta-x86_64-linux-gnu.tar.gz` file from the official source.
        
    * Use the wget command:
        
        ```bash
        wget https://github.com/LunaStev/Wave/releases/download/v0.1.1-pre-beta/wave-v0.1.1-pre-beta-x86_64-linux-gnu.tar.gz
        ```
        
    * Extract the archive:
        
        ```bash
        sudo tar -xvzf wave-v0.1.1-pre-beta-x86_64-linux-gnu.tar.gz -C /usr/local/bin
        ```
        
2. Setting up LLVMs
    
    * Open a terminal and type:
        
        ```bash
        sudo apt-get update
        sudo apt-get install llvm-14 llvm-14-dev clang-14 libclang-14-dev lld-14 clang
        sudo ln -s /usr/lib/llvm-14/lib/libLLVM-14.so /usr/lib/libllvm-14.so
        export LLVM_SYS_140_PREFIX=/usr/lib/llvm-14
        source ~/.bashrc
        ```
        
3. **Verify Installation:**
    
    * Open a terminal and type:
        
        ```bash
        wavec --version
        ```
        
    * If the version number displays, the installation was successful.
        

---

## Contributor

@LunaStev | 🇰🇷

---

## Website

[Website](https://wave-lang.dev)

[GitHub](https://github.com/LunaStev/Wave)

[Discord](https://discord.com/invite/3nev5nHqq9)

[Ko-fi](https://ko-fi.com/lunasev)