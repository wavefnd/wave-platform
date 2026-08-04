---
title: "Introduction to Wave v0.1.5-pre-beta: Structs, Proto methods, and System-wide Build System"
date: "2025-11-27 14:14:18"
description: "Link: Wave v0.1.5-pre-beta
Hello! I'm LunaStev, the developer of Wave.
We are pleased to announce Wave v0.1.5-pre-beta. In this release, we have laid the foundation for Object-Oriented Programming by introducing Structs and Proto methods. Additionall..."
tags: ["wave", "compiler", "programming languages"]
cover: "https://cdn.hashnode.com/res/hashnode/image/upload/v1764252340035/7dd95086-f969-43aa-8f29-96a2cfdcb488.png"
---

# Introduction to Wave v0.1.5-pre-beta: Structs, Proto methods, and System-wide Build System

Link: [Wave v0.1.5-pre-beta](https://github.com/wavefnd/Wave/releases/tag/v0.1.5-pre-beta)

Hello! I'm LunaStev, the developer of Wave.

We are pleased to announce [**Wave v0.1.5-pre-beta**](https://github.com/wavefnd/Wave/releases/tag/v0.1.5-pre-beta). In this release, we have laid the foundation for Object-Oriented Programming by introducing **Structs** and **Proto methods**. Additionally, we have completely overhauled the build system to improve stability and added a new automation script to streamline the development workflow.

Here are the key changes in v0.1.5-pre-beta:

#### 1\. Structs: Custom Data Structures

You can now define your own data types using the `struct` keyword. This allows you to group related data fields together and pass them around as a single unit.

```kotlin
struct Box {
    size: i32;
}

fun main() {
    // Initialize a struct
    var b: Box = Box { size: 42 };
    
    // Access fields using dot notation
    println("Box size is {}", b.size);
}
```

#### 2\. Proto: Attaching Methods

The `proto` keyword has been redefined. It is now used to attach methods to a specific struct. This enables the `object.method()` syntax, making your code more expressive and organized. You can access the instance data using the `self` parameter.

```kotlin
proto Box {
    fun double_size(self: Box) -> i32 {
        return self.size * 2;
    }
}

fun main() {
    var b: Box = Box { size: 10 };
    // Call the method defined in proto
    println("Doubled size: {}", b.double_size());
}
```

#### 3\. Build System Overhaul

We have significantly improved how Wave is built. previously, the compiler relied on downloading custom LLVM binaries, which caused compatibility issues.

* **System LLVM:** Wave now detects and links against your system's installed LLVM 14 libraries (Linux, macOS via Homebrew, and Windows).
    
* [`x.py`](https://github.com/wavefnd/Wave/blob/master/x.py) Script: We added a Python script ([`x.py`](https://github.com/wavefnd/Wave/blob/master/x.py)) to automate tasks like installing targets, building for cross-platform, and packaging releases.
    
* **macOS Support:** We added a macOS build job to our CI pipeline, ensuring better support for Apple users.
    

This update makes Wave more robust and easier to develop on different operating systems.

Thank you for your continued interest in Wave!

### Installation Guide

1. **Download:**
    
    * Download to Curl.
        
        ```bash
        curl -fsSL https://wave-lang.dev/install.sh | bash -s -- --version v0.1.5-pre-beta
        ```
        
2. **Verify Installation:**
    
    * Open a terminal and type:
        
        ```bash
        wavec --version
        ```
        
    * If the version number displays, the installation was successful.