---
title: "Introduction to Wave v0.0.7-pre-beta: Fixed infinite recursive and stack overflow bugs due to previously incorrect ASTNode reuse"
date: "2025-04-14 15:01:22"
description: "Hello! I'm LunaStev, the developer of Wave.
We are very happy to introduce Wave v0.0.7-pre-beta.
This release has previously addressed the issue of infinite recursive and stack overflow due to incorrect ASTNode reuse.
We also changed the name of the ..."
tags: []
cover: "https://cdn.hashnode.com/res/hashnode/image/upload/v1754270120138/4b00b4af-ba59-4ba2-a0d4-e74585b481a3.webp"
---

# Introduction to Wave v0.0.7-pre-beta: Fixed infinite recursive and stack overflow bugs due to previously incorrect ASTNode reuse

Hello! I'm LunaStev, the developer of Wave.

We are very happy to introduce Wave `v0.0.7-pre-beta`.

This release has previously addressed the issue of infinite recursive and stack overflow due to incorrect ASTNode reuse.

We also changed the name of the Wave compiler binary from `wave` to `wavec`.

Wave is growing fast, and we are very excited to share our future plans.

---

## 🔧 Bug Fixes

### 🐛 Fixed else if IR Infinite Loop Bug

* Resolved a critical bug where `stmt` was reused inside the `else if` block IR generation loop
    
* Previously caused infinite recursion and stack overflow due to incorrect ASTNode reuse
    
* Fixed by replacing `stmt` with `else_if` when iterating over `else_if_blocks`
    

## ✨ Other Changes

### 🛠 Renamed `wave` to `wavec`

* Renamed the Wave compiler binary from `wave` to `wavec`
    
* This change was made to prepare for the integration of Vex, the official package manager for Wave
    
* Separating the compiler tool (`wavec`) from the language name (`Wave`) aligns with future plans for Vex integration
    

---

## Showcase

![12](https://cdn.hashnode.com/res/hashnode/image/upload/v1754270118332/3b6bf633-82c3-44ff-9e06-71798c2b104e.png align="left")

![34](https://cdn.hashnode.com/res/hashnode/image/upload/v1754270119275/1b27235d-bcb8-442e-ae65-cda61989e8b0.png align="left")

---

Thank you for using Wave! Stay tuned for future updates and enhancements.

---

## Installation Guide

### For Linux:

1. **Download and Extract:**
    
    * Download the `wave-v0.0.7-pre-beta-linux.tar.gz` file from the official source.
        
    * Use the wget command:
        
        ```bash
        wget https://github.com/LunaStev/Wave/releases/download/v0.0.7-pre-beta/wave-v0.0.7-pre-beta-linux.tar.gz
        ```
        
    * Extract the archive:
        
        ```bash
        sudo tar -xvzf wave-v0.0.7-pre-beta-linux.tar.gz -C /usr/local/bin
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