---
translation_set_id: overview
path: getting-started/overview
locale: en
group: getting-started
group_order: 1
order: 1
title: Wave language overview
summary: What Wave is, what this reference covers, and which release it describes.
---

## Scope

Wave is a statically typed systems programming language designed for explicit control, readable low-level code, and interoperability with native libraries. This manual describes Wave 0.2.0-pre-beta as implemented by source revision bd5549b.

> **Release baseline**
> 
> Every syntax example in this manual targets Wave 0.2.0-pre-beta. Later development builds may differ.

## First program

```wave
fun main() {
    println("Hello, Wave!");
}
```

Save the source as main.wave, then compile it with wavec. A Wave program starts in main.

## Console input and output

```wave
var name: str;
input("{}", name);
print("Hello, ");
println("{}", name);
```

print writes without completing a line, println writes a formatted line, and input reads formatted input into a variable.

## Recommended reading path

- Install the compiler and verify its version.
- Learn declarations, types, expressions, and control flow.
- Continue with functions, structures, enums, modules, and the explicit memory type model.
- Use the syntax reference when you need a compact reminder.

