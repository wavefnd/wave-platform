---
title: "Introduction to Wave v0.1.3-pre-beta: Syntax tweaks, bug fixes, and enhanced type and import handling."
date: "2025-07-14 06:22:27"
description: "Hello! I'm LunaStev, the developer of Wave.
We are pleased to announce Wave v0.1.3-pre-beta — We changed the function parameter syntax from semicolons to commas, fixed LLVM IR generation for arrays of pointers and index access, and allowed parameters..."
tags: []
cover: "https://cdn.hashnode.com/res/hashnode/image/upload/v1754270102216/2f752c36-0f2b-48b3-8a7b-2cecad258a00.webp"
---

# Introduction to Wave v0.1.3-pre-beta: Syntax tweaks, bug fixes, and enhanced type and import handling.

Hello! I'm LunaStev, the developer of Wave.

We are pleased to announce Wave v0.1.3-pre-beta — We changed the function parameter syntax from semicolons to commas, fixed LLVM IR generation for arrays of pointers and index access, and allowed parameters to have multiple types. We also fixed bugs related to parameter parsing and `if` statements, improved inline assembly support for negative values, and restructured the `import` system.

## PR and Commits

* [\[#197\]Change function parameter syntax from semicolon to comma (issue #196)](https://github.com/LunaStev/Wave/pull/197)
    
* [\[#198\]Fix incorrect LLVM IR generation for array of pointers and IndexAccess dereferencing (issue #198)](https://github.com/LunaStev/Wave/pull/199)
    
* [\[#201\]Parameters can have multiple types](https://github.com/LunaStev/Wave/pull/201)
    
* [\[#204\]Param bug fix](https://github.com/LunaStev/Wave/pull/204)
    
* [\[#206\]Handling Inline Assembly Negative Values (issue #205)](https://github.com/LunaStev/Wave/pull/206)
    
* [\[#208\]Troubleshooting if statement bugs in paser](https://github.com/LunaStev/Wave/pull/208)
    
* [\[#210\]Cambiar la estructura de import](https://github.com/LunaStev/Wave/pull/210)
    

## Showcase

The showcase is available at [Wave-Test](https://github.com/LunaStev/wave-testing).

---

Thank you for using Wave! Stay tuned for future updates and enhancements.

---

## Installation Guide

1. **Download:**
    
    * Download to Curl.
        
        ```bash
        curl -fsSL https://wave-lang.dev/install.sh | bash -s -- --version v0.1.3-pre-beta
        ```
        
2. **Verify Installation:**
    
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