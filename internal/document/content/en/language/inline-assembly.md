---
translation_set_id: assembly
path: language/inline-assembly
locale: en
group: language
group_order: 2
order: 9
title: Inline assembly
summary: Use architecture-specific instructions at an explicit low-level boundary.
---

## Assembly block

asm provides an architecture-specific escape hatch. Inputs, outputs, and clobbered state must describe the instruction sequence accurately.

```wave
var result: i64;
asm {
    "mov rax, 123"
    in("rdi") 1
    out("rax") result
}
```

in binds a Wave value to an input register and out writes an output register to a Wave variable. clobber declarations must name additional state changed by the block when required.

> **Portability**
> 
> Inline assembly is tied to a target architecture, ABI, and compiler contract. Isolate it behind a typed Wave function and provide a non-assembly implementation where practical.

