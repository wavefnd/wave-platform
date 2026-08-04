---
title: "Introduction to Wave v0.1.2-pre-beta: Assignment Operators, Added Remainder and Generalized Indexing Support"
date: "2025-06-21 14:55:18"
description: "Hello! I'm LunaStev, the developer of Wave.
We are pleased to announce Wave v0.1.2-pre-beta — This update supports indexing, supports the rest of the operators. And I added an assignment operator. And I fixed a number of bugs.

✅ Added Features
🔍 Ge..."
tags: []
cover: "https://cdn.hashnode.com/res/hashnode/image/upload/v1754270104684/ddb9ba40-d06b-488b-85b9-1803e65b80bd.webp"
---

# Introduction to Wave v0.1.2-pre-beta: Assignment Operators, Added Remainder and Generalized Indexing Support

Hello! I'm LunaStev, the developer of Wave.

We are pleased to announce Wave `v0.1.2-pre-beta` — This update supports indexing, supports the rest of the operators. And I added an assignment operator. And I fixed a number of bugs.

---

## ✅ Added Features

### 🔍 Generalized Indexing Support

* Extended `arr[index]` syntax to support **variable indexing** in any context.
    
* Now supports dynamic indexing like `arr[i]`, `ptr[j]`, `get_array()[k]` inside `if`, `while`, `return`, and assignments.
    
* Internally handles pointer types (e.g., `i8*`) using `GEP` and performs automatic type casting (e.g., `i8` vs `i64`) for comparisons and assignments.
    

### ➕ Added `%` (Remainder) Operator Support

* Wave now supports the modulo operator `%` for integer types.
    
* Expressions like `a % b`, `10 % 3`, and `var x: i32 = a % b` are now fully parsed and compiled to LLVM IR using srem (signed remainder).
    
* Parser was updated to recognize `%` as `TokenType::Remainder`, and IR generation uses `build_int_signed_rem` for correct runtime behavior.
    
* `%` is fully supported inside expressions, assignments, conditionals, and return statements.
    

### ⚙️ Assignment Operators (`+=`, `-=`, `*=`, `/=`, `%=`)

* Supports compound assignment operators for both integers (`i32`) and floating-point numbers (`f32`).
    
* Operators implemented: `+=`, `-=`, `*=`, `/=`, `%=`.
    
* Type-safe operation with proper distinction between integer and float IR instructions:
    
    * add, sub, mul, sdiv, srem for integers.
        
    * `fadd`, `fsub`, `fmul`, `fdiv`, `frem` (uses `fmodf`) for floats.
        
* Implicit type casting during assignment (e.g., assigning an int to a float variable triggers `int` → `float` conversion).
    
* Proper LLVM IR generation for all supported operations, including float remainder via external `fmodf` call (linked with `-lm`).
    

## 🐛 Bug Fixes

### 🔢 Accurate Token Differentiation Between Integers and Floats

* Number parsing logic has been overhauled to properly distinguish between integer literals (e.g., `42`) and floating-point literals (e.g., `3.14`, `42.0`).
    
* Previously, all numeric literals were parsed as `Float(f64)` tokens, even when no decimal point was present. → Now, tokens like `123` are correctly parsed as `Number(i64)`, and `123.45` as `Float(f64)`.
    
* Introduced internal flag `is_float` to detect presence of `.` during scanning. If found, the number is parsed as a float; otherwise, as an integer.
    
* Implemented type-safe error handling:
    
    * Fallbacks to `TokenType::Number(0)` or `TokenType::Float(0.0)` on parse failure, ensuring the lexer remains stable on malformed input.
        
* This fix improves downstream type checking, IR generation, and expression evaluation, especially for typed languages like Wave where `i64` and `f64` must be handled distinctly.
    

## ✨ Other Changes

### 🧠 Library and Binary 2 Coexist

* Add lib.rs for easy package manager creation, development, and easy access.
    

---

## Showcase

The showcase is available at [Wave-Test](https://github.com/LunaStev/wave-testing).

---

Thank you for using Wave! Stay tuned for future updates and enhancements.

---

## Installation Guide

1. **Download:**
    
    * Download to Curl.
        
        ```bash
        curl -fsSL https://wave-lang.dev/install.sh | bash -s -- --version v0.1.2-pre-beta
        ```
        
2. **Verify Installation:**
    
    * Open a terminal and type:
        
        ```bash
        wavec --version
        ```
        
    * If the version number displays, the installation was successful.
        

---

## PR and Commits

* [Switch to Github Flow It's still before the 0.1.2-pre-beta release.](https://github.com/LunaStev/Wave/pull/172)
    
* [Add simple algorithm examples for bug hunting (issue #178)](https://github.com/LunaStev/Wave/pull/182)
    
* [Add simple algorithm examples for bug hunting (issue #178)](https://github.com/LunaStev/Wave/pull/184)
    
* [primary expression assignment function fix](https://github.com/LunaStev/Wave/pull/185)
    
* [Fix IR generation for if-return and function call assignments](https://github.com/LunaStev/Wave/pull/187)
    
* [Fix inline assembly code generation for statement-level asm blocks (issue #173)](https://github.com/LunaStev/Wave/pull/194)
    
* [Fix expression parser false positives for statement tokens](https://github.com/LunaStev/Wave/pull/195)
    

---

## Contributor

@LunaStev | 🇰🇷

---

## Website

[Website](https://wave-lang.dev)

[GitHub](https://github.com/LunaStev/Wave)

[Discord](https://discord.com/invite/3nev5nHqq9)

[Ko-fi](https://ko-fi.com/lunasev)