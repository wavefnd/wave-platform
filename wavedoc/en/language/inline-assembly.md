---
translation_set_id: assembly
path: language/inline-assembly
locale: en
group: language
group_order: 2
order: 9
title: Inline assembly
summary: Instruction strings, in/out operands, and clobber contracts inside asm blocks.
---

## asm blocks

`asm` is a low-level escape hatch for inserting target-architecture instructions.

```wave
fun read_value() -> i64 {
    var result: i64 = 0;
    asm {
        "mov rax, 123"
        out("rax") result
    }
    return result;
}
```

String literals in the block become assembly instruction entries.

## Inputs and outputs

```wave
var result: i64 = 0;
asm {
    "mov rax, rdi"
    in("rdi") 123
    out("rax") result
}
```

- `in("reg") expression` associates a Wave value with an input operand.
- `out("reg") target` stores an output into an assignable Wave target.
- Register names can be parsed from strings or identifier forms.

An input operand can be a variable, an integer or string literal, `&identifier`, `deref identifier`, or a negative numeric literal.

## clobbers

If a block changes registers or memory state beyond explicit outputs, list that state in `clobber(...)`.

```wave
asm {
    "nop"
    clobber("rax", "rcx", "memory")
}
```

## What to verify

- Instruction syntax must match the target architecture and LLVM inline-assembly contract.
- Do not destroy registers that the calling convention requires you to preserve.
- Declare the required clobbers when the block reads or writes hidden state.
- Prefer isolating architecture-specific asm behind small typed functions.

Inline assembly's correctness and portability are not guaranteed by Wave's type system alone.
