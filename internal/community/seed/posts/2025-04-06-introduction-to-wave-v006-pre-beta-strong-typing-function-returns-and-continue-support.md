---
title: "Introduction to Wave v0.0.6-pre-beta: Strong Typing, Function Returns, and continue Support"
date: "2025-04-06 13:47:21"
description: "Hello! I'm Lunastev, the developer of Wave.
I'm very happy to introduce Wave v0.0.6-pre-beta, which is an important step forward in the evolution of language.
The release focuses on expanding the type system, enhancing feature support, and introducin..."
tags: []
cover: "https://cdn.hashnode.com/res/hashnode/image/upload/v1754270125855/a1a07851-a9d3-4bd5-b1b5-167630c535b9.webp"
---

# Introduction to Wave v0.0.6-pre-beta: Strong Typing, Function Returns, and continue Support

Hello! I'm Lunastev, the developer of Wave.

I'm very happy to introduce Wave `v0.0.6-pre-beta`, which is an important step forward in the evolution of language.

The release focuses on expanding the type system, enhancing feature support, and introducing powerful new features such as 'continue' doors and floating arithmetic. With the structured 'WaveType' Enum now replacing all string-based types, Wave has taken a strong step towards becoming a statically typed system language.

The function return type is now fully supported, enabling expressive and reusable logic. You can define the function as `-> i32` and return the value using the `return` keyword. LLVM IR generation logic has also been upgraded to be fully type-aware, ensuring safer and more accurate lower-level output.

The waves are growing fast, and we are very excited to share our future plans.
Thank you for supporting me on this journey 💙

---

## ✅ Added Features

### 💬 Comment Support
* Supports single-line comments using `//`
* Supports multi-line comment blocks using `/* */`

### ✅ continue statement Support
* Possible to skip to the next iteration depending on the condition within the 'while' loop
* Syntax supported for 'if (condition) {continue; }'
* In LLVM IR, 'continue' is treated as a condition check block for the corresponding loop

### 🧠 Strong Typing for Variables and Parameters
* Replaced string-based types with structured `WaveType` enums in the AST
* Fully supports types like `i32`, `u64`, `f32` for both variables and parameters
* Enables static type checking and safer LLVM IR generation

### 🔢 Float Type Support (`f32`)
* Supports `f32` literals (e.g., `12.34`)
* Allows declaration, initialization, and reassignment of `f32` variables
* Enables use of `f32` values in arithmetic and comparison operations (e.g., `if`, `while`)
* Promotes `float` to `double` using `fpext` when passing to `printf` to conform to the C ABI

### 🖨️ Formatted Print (`println("...", value)`)
* Automatically maps Wave types to proper C format specifiers (`%d`, `%f`, `%s`)
* Correctly handles printing of `f32`, `i32`, and string types with type inference

### 🌀 Type-Aware LLVM IR Generation
* LLVM `alloca`, `store`, and `load` instructions are generated based on `WaveType`
* Supports both integer and float value initialization
* Uses `BasicTypeEnum` and `BasicValueEnum` to unify value handling
* Ensures correctness in mixed-type binary expressions and conditionals

### 🧩 Function Definition and Calls
* Supports user-defined functions with multiple parameters
* Parameters support explicit typing (e.g., `i32`, `str`)
* Functions can be called with literals or variables as arguments
* LLVM IR correctly handles parameter passing using `%0`, `%1`, ... style
* Parameter values are properly `store`d and `load`ed from the stack
* Enables composable and reusable logic with full IR-level integration

### 🧠 Function Return Type Support
* Functions can now specify return types using `->` syntax (e.g., `-> i32`)
* Supported return types: `i32`, `f32`, `str` (more planned)
* Return statements (`return expr;`) emit the appropriate LLVM `ret` instruction
* If no return is specified in a void function, `ret void` is automatically inserted
* Ensures matching return types between Wave AST and LLVM IR generation

---

## Showcase

![Image des2cription](https://cdn.hashnode.com/res/hashnode/image/upload/v1754270122795/d94ab670-6e73-42e4-8452-9a37370a1fb5.png)

![Image descri3ption](https://cdn.hashnode.com/res/hashnode/image/upload/v1754270123598/3c9e4856-0808-46c1-98b1-bd8a1a608556.png)

---

![Image descr2iption](https://cdn.hashnode.com/res/hashnode/image/upload/v1754270124368/be0640ca-aea8-46fb-92a2-f648f261b898.png)

![Image descr3iption](https://cdn.hashnode.com/res/hashnode/image/upload/v1754270125208/fa32e70b-aa32-4901-9630-9be5e82a260d.png)

---

Thank you for using Wave! Stay tuned for future updates and enhancements.

---

## Installation Guide

### For Linux:

1. **Download and Extract:**
   - Download the `wave-v0.0.6-pre-beta-linux.tar.gz` file from the official source.
   - Use the wget command:
     ```bash
     wget https://github.com/LunaStev/Wave/releases/download/v0.0.6-pre-beta/wave-v0.0.6-pre-beta-linux.tar.gz
     ```
   - Extract the archive:
     ```bash
     sudo tar -xvzf wave-v0.0.6-pre-beta-linux.tar.gz -C /usr/local/bin
     ```

3. Setting up LLVMs
   - Open a terminal and type:
     ```bash
     sudo apt-get update
     sudo apt-get install llvm-14 llvm-14-dev clang-14 libclang-14-dev lld-14 clang
     sudo ln -s /usr/lib/llvm-14/lib/libLLVM-14.so /usr/lib/libllvm-14.so
     export LLVM_SYS_140_PREFIX=/usr/lib/llvm-14
     source ~/.bashrc
     ```

4. **Verify Installation:**
   - Open a terminal and type:
     ```bash
     wave --version
     ```
   - If the version number displays, the installation was successful.


---

## Contributor

@LunaStev | 🇰🇷

---

## Website

[Website](https://wave-lang.dev)
[GitHub](https://github.com/LunaStev/Wave)