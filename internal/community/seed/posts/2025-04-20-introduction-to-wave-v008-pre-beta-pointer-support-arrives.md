---
title: "Introduction to Wave v0.0.8-pre-beta: Pointer Support Arrives"
date: "2025-04-20 05:10:42"
description: "Hello! I'm LunaStev, the developer of Wave.
We are very happy to introduce Wave v0.0.8-pre-beta — a version that officially brings first-class pointer support to the language.
Wave was designed with low-level capabilities in mind, and in this version..."
tags: []
cover: "https://cdn.hashnode.com/res/hashnode/image/upload/v1754270367874/ee3376fd-ec48-4fae-afab-7160c74e27ea.webp"
---

# Introduction to Wave v0.0.8-pre-beta: Pointer Support Arrives

Hello! I'm LunaStev, the developer of Wave.

We are very happy to introduce Wave `v0.0.8-pre-beta` — a version that officially brings first-class pointer support to the language.

Wave was designed with low-level capabilities in mind, and in this version, we’re making a major leap in that direction.

---

## ✅ Added Features

### 🧠 Pointer System: First-Class Pointer Support in Wave

* Introduced `ptr<T>` type syntax for defining typed pointers  
    → Example: `var p: ptr<i32>;`
    
* Implemented `&x` address-of operator  
    → Compiles to LLVM IR as `store i32* %x`
    
* Implemented `deref p` dereference operator  
    → Generates IR as `load i32, i32* %p`
    
* Supported pointer-based initialization  
    → `var p: ptr<i32> = &x;` is now fully parsed and compiled
    
* Enabled dereferencing for both expression and assignment  
    → Example: `deref p = 42;` is valid and stored directly via IR
    
* Address values can be printed as integers  
    → `%ld` used for pointer-to-int cast in formatted output  
    → `println("address = {}", p);` prints memory address
    

## 🔧 Bug Fixes

### 🐛 Fixed Pointer Initialization Parsing Issue

* Changed `VariableNode.initial_value` from `Option<Literal>` to `Option<Expression>`
    
* Allowed `&x` to be accepted as a valid initializer expression
    

### 🐛 Fixed LLVM IR Crash on AddressOf Expression

* Added support for `Expression::AddressOf` in IR generation
    
* Prevented crash by checking for variable reference inside address-of
    

### 🐛 Fixed printf format mismatch for pointers

* `%s` → `%ld` for pointer values
    
* Ensured correct casting of `i32*` to `i64` before printing
    

## ✨ Other Changes

### 🧠 Improved Format String Handling in IR

* Added dynamic format string generation based on argument types
    
* Format strings now automatically adapt for int, float, and pointer types
    

---

## Showcase

![Imag3e description](https://cdn.hashnode.com/res/hashnode/image/upload/v1754270366188/032659ca-dbc7-40dc-9d0c-2403ad2b6cd0.png align="left")

![Ima3ge description](https://cdn.hashnode.com/res/hashnode/image/upload/v1754270367057/0735b5ae-2de5-48fd-be40-46923c3c8c85.png align="left")

---

Thank you for using Wave! Stay tuned for future updates and enhancements.

---

## Installation Guide

### For Linux:

1. **Download and Extract:**
    
    * Download the `wave-v0.0.8-pre-beta-x86_64-linux-gnu.tar.gz` file from the official source.
        
    * Use the wget command:
        
        ```bash
        wget https://github.com/LunaStev/Wave/releases/download/v0.0.8-pre-beta/wave-v0.0.8-pre-beta-x86_64-linux-gnu.tar.gz
        ```
        
    * Extract the archive:
        
        ```bash
        sudo tar -xvzf wave-v0.0.8-pre-beta-x86_64-linux-gnu.tar.gz -C /usr/local/bin
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