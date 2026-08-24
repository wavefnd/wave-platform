---
translation_set_id: comments
path: language/comments
locale: en
group: language
group_order: 2
order: 11
title: Comments
summary: Line comments, nestable block comments, and diagnostics for unterminated comments.
---

## Line comments

Everything after `//` through the end of the line is a comment.

```wave
var count: i32 = 10; // active request count
```

## Block comments

`/*` and `*/` delimit a block comment.

```wave
/* A comment can span
   multiple lines. */
```

Block comments can be nested.

```wave
/* outer comment
   /* inner comment */
   outer comment again
*/
```

## Comment markers inside strings

Inside string and character literals, `//`, `/*`, and `*/` are literal text rather than comment delimiters.

```wave
var text: str = "https://wave-lang.dev";
```

## Unterminated block comments

Every `/*` requires a matching `*/`. An unterminated block comment produces diagnostic `E1002` (`UnterminatedComment`).

When temporarily commenting out a large region, make sure nested comment depth remains balanced.
