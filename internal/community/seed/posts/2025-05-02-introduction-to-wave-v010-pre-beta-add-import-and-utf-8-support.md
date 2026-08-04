---
title: "Introduction to Wave v0.1.0-pre-beta: Add Import and UTF-8 Support"
date: "2025-05-02 08:58:02"
description: "Hello! I'm LunaStev, the developer of Wave.
We are very pleased to introduce Wave 'v0.1.0-pre-beta' — This update supports the import function and UTF-8, allowing you to output other characters, unlike previous versions that only supported ASCII.

✅ ..."
tags: []
cover: "https://cdn.hashnode.com/res/hashnode/image/upload/v1754270362649/042f67e0-22ba-4e52-809d-23b9260acdd3.webp"
---

# Introduction to Wave v0.1.0-pre-beta: Add Import and UTF-8 Support

Hello! I'm LunaStev, the developer of Wave.

We are very pleased to introduce Wave 'v0.1.0-pre-beta' — This update supports the import function and UTF-8, allowing you to output other characters, unlike previous versions that only supported ASCII.

---

## ✅ Added Features

### 📦 Local File Import Support

* Introduced `import("...");` statement in Wave syntax.
    
* Supports importing `.wave` source files relative to the current file's directory.
    
* Prevents duplicate imports automatically using an internal `HashSet`.
    
* Imported files are parsed, converted to AST, and merged into the main program at compile time.
    
* Enables modular project structure by allowing multi-file composition.
    

## 🔧 Bug Fixes

### 🐞 UTF-8 Handling in Lexer

* Fixed tokenizer crash on non-ASCII characters.
    
* Lexer now correctly processes UTF-8 multi-byte characters, enabling support for Korean and other languages in source code.
    

### 🐞 Underscore (`_`) Support in Identifiers

* Variable and function names can now contain underscores.
    
* Lexer now treats identifiers like `my_var` or `some_function` as valid.
    

---

## Showcase

![Image1description](https://cdn.hashnode.com/res/hashnode/image/upload/v1754270358343/f88308b1-c4ec-4a1b-8de1-cff4eb6f3893.png align="left")

![Image2description](https://cdn.hashnode.com/res/hashnode/image/upload/v1754270359325/8a0d91cf-6d81-4295-be91-387179da5ebb.png align="left")

![Image3description](https://cdn.hashnode.com/res/hashnode/image/upload/v1754270360299/f81102b1-fbb4-46e5-9c11-a8e3e36bf45c.png align="left")

---

![Image4description](https://cdn.hashnode.com/res/hashnode/image/upload/v1754270361200/6dc81239-086b-45b3-b172-5fc6966f263a.png align="left")

![Image5description](https://cdn.hashnode.com/res/hashnode/image/upload/v1754270361905/f06da6b6-7de0-4345-b43f-c2bdefb02ea9.png align="left")

---

Thank you for using Wave! Stay tuned for future updates and enhancements.

---

## Installation Guide

### For Linux:

1. **Download and Extract:**
    
    * Download the `wave-v0.1.0-pre-beta-x86_64-linux-gnu.tar.gz` file from the official source.
        
    * Use the wget command:
        
        ```bash
        wget https://github.com/LunaStev/Wave/releases/download/v0.1.0-pre-beta/wave-v0.1.0-pre-beta-x86_64-linux-gnu.tar.gz
        ```
        
    * Extract the archive:
        
        ```bash
        sudo tar -xvzf wave-v0.1.0-pre-beta-x86_64-linux-gnu.tar.gz -C /usr/local/bin
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