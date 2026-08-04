---
translation_set_id: lexical
path: language/lexical-structure
locale: en
group: language
group_order: 2
order: 1
title: Lexical structure
summary: Identifiers, literals, comments, separators, and reserved words.
---

## Identifiers and statements

Identifiers name declarations and are case-sensitive. Statements normally end with a semicolon. Braces delimit blocks.

```wave
var answer: i32 = 42;
var greeting: str = "hello";
var enabled: bool = true;
```

## Literals

- Integer and floating-point literals represent numeric values.
- String literals use double quotes and character literals use single quotes.
- true, false, and null are built-in literals.

## Comments

```wave
// line comment
/* block comment */
```

## Reserved words

fun, extern, export, type, enum, static, var, deref, let, mut, const, if, else, proto, struct, while, for, module, class, in, out, clobber, is, as, asm, xnand, import, return, continue, print, input, println, match, break, true, false, and null are reserved by this release.

