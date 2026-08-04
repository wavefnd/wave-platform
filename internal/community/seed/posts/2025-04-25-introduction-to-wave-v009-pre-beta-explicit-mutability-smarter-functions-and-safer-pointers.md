---
title: "Introduction to Wave v0.0.9-pre-beta: Explicit Mutability, Smarter Functions, and Safer Pointers"
date: "2025-04-25 15:26:07"
description: "Hello! I'm Lunastev, the developer of Wave.
We are very excited to introduce Wave 'v0.0.9-Free Beta' — A version that offers Explicit Mutability, Smarter Functions, and Safer Pointers.
Wave is designed with low-level features in mind, and in this ver..."
tags: []
cover: "https://cdn.hashnode.com/res/hashnode/image/upload/v1754270368495/7403b4d4-fbd4-40f9-a850-994adac0383d.webp"
---

# Introduction to Wave v0.0.9-pre-beta: Explicit Mutability, Smarter Functions, and Safer Pointers

Hello! I'm Lunastev, the developer of Wave.

We are very excited to introduce Wave 'v0.0.9-Free Beta' — A version that offers Explicit Mutability, Smarter Functions, and Safer Pointers.

Wave is designed with low-level features in mind, and in this version, We are making a big leap in that direction.

---

## 📐 Language Specification Updates

### 🧠 Introduction of `let`, `let mut`, and var for Explicit Mutability

* Wave now supports three types of variable declarations to express mutability explicitly:
    
    * `var`: fully mutable, intended for general-purpose variables
        
    * `let`: immutable, reassignment is forbidden
        
    * `let mut`: mutable under immutable declaration context (safe controlled mutability)
        
* This design introduces clearer ownership intent and improves safety in low-level and system-oriented programming.
    

### 🧠 Default Parameter Values in Function Declarations

* Wave functions can now define parameters with default values:
    
    ```plaintext
    fun main(name: str = "World") {
        println("Hello {}", name);
    }
    ```
    
* If an argument is not provided at runtime, the default value will be inserted automatically by the compiler.
    
* This enables more expressive and flexible function declarations.
    

## ✅ Added Features

### 🧠 Parser and IR support for explicit mutability

* Introduced internal Mutability enum: `Var`, `Let`, `LetMut`
    
* Implemented `parse_let()` with optional mut keyword for parsing
    
* Wave's IR generation now restricts `let` variables from being reassigned
    

### 🧠 IR handling of function parameter defaults

* When default values are present in function parameters, they are now correctly recognized and handled during LLVM IR generation
    
* If an argument is not passed at runtime, the default value is inserted directly into the stack-allocated variable
    

## 🔧 Bug Fixes

### 🐛 Incorrect string output in `println()` format

* Fixed an issue where `str` values (`i8*`) were printed as raw addresses
    
* The format translation now maps `i8*` to `%s` correctly
    
* Values are passed directly to `printf` as string pointers, avoiding `ptr_to_int` conversion
    

### 🐛 Incorrect handling of `deref` assignment

* Fixed an issue where dereferencing a pointer and assigning its value caused type mismatches in the generated IR.
    
* The IR now properly handles dereferencing a pointer (`deref p1 = deref p2;`) and assigning the values correctly without causing `i32**` mismatches.
    

## ✨ Other Changes

### 🧠 IR-level enforcement of immutability

* Reassignment attempts to `let` variables now cause a compile-time panic
    
* All memory operations (store/load) respect mutability constraints
    

### 🧠 IR-level enforcement of pointer dereferencing

* Introduced a fix to ensure that pointer dereferencing (`deref p1 = deref p2;`) is handled correctly in the IR.
    
* Adjusted the `generate_address_ir()` function to properly dereference pointers and load/store values without causing pointer type mismatches.
    

---

## Showcase

![Image1description](https://cdn.hashnode.com/res/hashnode/image/upload/v1754270365525/ec473068-392d-4bc1-a6f1-3f1cf05231e8.png align="left")

![Image3description](https://cdn.hashnode.com/res/hashnode/image/upload/v1754270366298/20648347-0dd6-4820-9522-8d2108dfd417.png align="left")

---

![Image2description](https://cdn.hashnode.com/res/hashnode/image/upload/v1754270367126/74f553fd-89ea-4159-9f7b-1fa5427abddf.png align="left")

![Image4description](https://cdn.hashnode.com/res/hashnode/image/upload/v1754270367816/68c70de2-96fe-4830-94b2-c0b821adf543.png align="left")

---

Thank you for using Wave! Stay tuned for future updates and enhancements.

---

## Installation Guide

### For Linux:

1. **Download and Extract:**
    
    * Download the `wave-v0.0.9-pre-beta-x86_64-linux-gnu.tar.gz` file from the official source.
        
    * Use the wget command:
        
        ```bash
        wget https://github.com/LunaStev/Wave/releases/download/v0.0.9-pre-beta/wave-v0.0.9-pre-beta-x86_64-linux-gnu.tar.gz
        ```
        
    * Extract the archive:
        
        ```bash
        sudo tar -xvzf wave-v0.0.9-pre-beta-x86_64-linux-gnu.tar.gz -C /usr/local/bin
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

[Ko-fi](https://ko-fi.com/lunasev)