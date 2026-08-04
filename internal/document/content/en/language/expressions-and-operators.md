---
translation_set_id: expressions
path: language/expressions-and-operators
locale: en
group: language
group_order: 2
order: 3
title: Expressions and operators
summary: Arithmetic, comparison, logic, assignment, casts, and precedence.
---

## Operator families

| Purpose | Operators |
| --- | --- |
| Arithmetic | + - * / % |
| Comparison | == != < <= > >= |
| Logic | && || ! |
| Bitwise | & | ^ ~ << >> |
| Assignment | = and compound assignments |
| Explicit cast | as |

```wave
var width: i32 = 12;
var area: i32 = width * 8;
var large: bool = area >= 64;
var widened: i64 = area as i64;
```

Parenthesize mixed operator families when the intended evaluation order is not immediately obvious. Explicit casts use as.

> **Reserved operations**
> 
> Some tokens, including is and xnand, are reserved by the lexer but are not verified expression operators in 0.2.0-pre-beta.

