---
title: "Introduction to Wave v0.1.4-pre-beta: Add optimization, method chaining, and image file build"
date: "2025-08-01 03:28:54"
description: "Hello! I'm LunaStev, the developer of Wave.
We are pleased to announce Wave v0.1.4-pre-beta — We've added CLI improvements, LLVMO2 optimization, and we've added commands to help build the image file, and we've added the most important feature, Method..."
tags: []
cover: "https://cdn.hashnode.com/res/hashnode/image/upload/v1754270096252/b768a649-a82e-4b94-ab0e-89de88f680e0.webp"
---

# Introduction to Wave v0.1.4-pre-beta: Add optimization, method chaining, and image file build

Hello! I'm LunaStev, the developer of Wave.

We are pleased to announce Wave v0.1.4-pre-beta — We've added CLI improvements, LLVMO2 optimization, and we've added commands to help build the image file, and we've added the most important feature, Method Chaining.

## PR and Commits

* [\[#212\]CLI improvements](https://github.com/LunaStev/Wave/pull/212)
    
* [\[#213\]Optimized to pass and -O2](https://github.com/LunaStev/Wave/pull/213)
    
* [\[#215\]Image caption Agregar comandos CLI y palabras clave de Proto para construir archivos](https://github.com/LunaStev/Wave/pull/215)
    
* [\[#216\]Add Method Chaining](https://github.com/LunaStev/Wave/pull/216)
    

## Showcase

The showcase is available at [Wave-Test](https://github.com/LunaStev/wave-testing).

---

Thank you for using Wave! Stay tuned for future updates and enhancements.

---

## Features

CLI:

```bash
wavec run --img main.wave
```

Method Chaining:

```kotlin
fun len(s: str) -> i32 {
    var count: i32 = 0;
    while (s[count] != 0) {
        count += 1;
    }
    return count;
}

fun main() {
    var my_string: str = "Hello World";
    var length: i32 = my_string.len();
    println("Result of my_string.len(): {}", length);
}
```

---

## Installation Guide

1. **Download:**
    
    * Download to Curl.
        
        ```bash
        curl -fsSL https://wave-lang.dev/install.sh | bash -s -- --version v0.1.4-pre-beta
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

[Community](https://discord.com/invite/3nev5nHqq9)