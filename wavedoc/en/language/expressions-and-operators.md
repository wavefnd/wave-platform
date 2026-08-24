---
translation_set_id: expressions
path: language/expressions-and-operators
locale: en
group: language
group_order: 2
order: 3
title: Expressions and operators
summary: Assignment, arithmetic, bitwise and logical operations, casts, increment/decrement, and operator precedence.
---

## Main operators

| Category | Operators |
| --- | --- |
| Arithmetic | `+`, `-`, `*`, `/`, `%` |
| Comparison | `==`, `!=`, `<`, `<=`, `>`, `>=` |
| Logical | `&&`, `||`, `!` |
| Bitwise | `&`, `\|`, `^`, `~`, `<<`, `>>` |
| Assignment | `=`, `+=`, `-=`, `*=`, `/=`, `%=` |
| Increment/decrement | `++`, `--` |
| Address/dereference | `&value`, `deref pointer` |
| Explicit cast | `as` |

```wave
var width: i32 = 12;
var area: i32 = width * 8;
var large: bool = area >= 64;
var widened: i64 = area as i64;
area += 8;
```

## Precedence

Operators bind from tighter to looser in this order:

1. Primary and postfix operations: calls, field access, indexing, postfix `++` and `--`
2. Unary operations: `!`, `~`, `&`, `deref`, prefix `++` and `--`, unary `+` and `-`
3. `as` casts
4. `*`, `/`, `%`
5. `+`, `-`
6. `<<`, `>>`
7. `<`, `<=`, `>`, `>=`
8. `==`, `!=`
9. Bitwise `&`, `^`, `|`
10. `&&`, `||`
11. Assignment and compound assignment

Assignment is right-associative, so chained assignments are evaluated from right to left. Use parentheses when mixing operators would make intent unclear.

## Assignable targets

Assignment and `++`/`--` operate on storage locations such as variables, fields, indexed elements, and dereferenced pointers. Writes to `const` values are not allowed.
