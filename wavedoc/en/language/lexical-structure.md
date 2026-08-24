---
translation_set_id: lexical
path: language/lexical-structure
locale: en
group: language
group_order: 2
order: 1
title: Lexical structure
summary: Identifiers, literals, statement delimiters, keywords, and built-in type spellings.
---

## Identifiers

Identifiers name variables, functions, types, fields, and other declarations. Names are case-sensitive. Identifiers can use Unicode letters, digits, and `_`; a digit cannot begin an identifier.

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

## Keywords and type spellings

Wave keywords include:

`pub`, `fun`, `extern`, `export`, `type`, `enum`, `static`, `var`, `deref`, `const`, `if`, `else`, `proto`, `struct`, `while`, `for`, `in`, `out`, `clobber`, `as`, `asm`, `import`, `return`, `continue`, `print`, `input`, `println`, `match`, `break`, `true`, `false`, and `null`.

Built-in type spellings include `bool`, `char`, `byte`, `str`, the integer and floating-point types, `ptr<T>`, and `array<T, N>`. Keywords and built-in type names cannot be used as declaration names.
