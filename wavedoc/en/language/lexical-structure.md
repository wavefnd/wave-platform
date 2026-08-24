---
translation_set_id: lexical
path: language/lexical-structure
locale: en
group: language
group_order: 2
order: 1
title: Lexical structure
summary: Distinguish identifiers, literals, delimiters, type spellings, and lexer-reserved tokens.
---

## Identifiers

Identifiers name variables, functions, types, fields, and other declarations. Names are case-sensitive. The lexer accepts alphabetic characters, numeric characters, and `_` while scanning an identifier and uses Unicode character classification.

```wave
var request_count: i64 = 0;
var 이름: str = "Wave";
```

Projects may still prefer a consistent naming convention for tooling and searchability.

## Statements and delimiters

Most declarations and expression statements end with `;`. Constructs with bodies, such as functions, conditionals, loops, and structs, use `{ ... }` blocks.

```wave
var answer: i32 = 42;

fun double(value: i32) -> i32 {
    return value * 2;
}
```

## Literals

```wave
var integer: i32 = 42;
var decimal: f64 = 3.14;
var text: str = "Wave";
var letter: char = 'W';
var enabled: bool = true;
var address: ptr<u8> = null;
```

Wave has integer, floating-point, string, character, Boolean, and `null` literals. Use `null` as a pointer value.

## Reserved keywords and type spellings

Major spellings tokenized specially by the lexer include:

`fun`, `extern`, `export`, `type`, `enum`, `static`, `var`, `deref`, `const`, `if`, `else`, `proto`, `struct`, `while`, `for`, `module`, `class`, `in`, `out`, `clobber`, `is`, `as`, `asm`, `xnand`, `import`, `return`, `continue`, `print`, `input`, `println`, `match`, `break`, `true`, `false`, and `null`.

`char`, `byte`, and the built-in integer and floating-point spellings are also tokenized as types. `ptr` and `array` remain identifier tokens and are interpreted by the type parser when written as `ptr<T>` and `array<T, N>`.

> **Reserved does not mean implemented**
>
> Lexer-reserved spellings such as `module`, `class`, `is`, and `xnand` do not by themselves imply a completed language feature. Treat a feature as supported when its usable syntax is documented and implemented through parsing and code generation.
