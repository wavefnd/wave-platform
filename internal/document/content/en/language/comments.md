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
var count: i32 = 10; // current request count
```

## Block comments

`/*` and `*/` delimit a block comment.

```wave
/* A comment can span
   multiple lines. */
```

The v0.2.0-pre-beta lexer tracks block-comment depth, so **nested block comments are supported**.

```wave
/* outer comment
   /* inner comment */
   outer comment again
*/
```

## Comment markers inside strings

String and character literals are tokenized as literals, so `//`, `/*`, and `*/` inside them are not interpreted as comment delimiters.

```wave
var text: str = "https://wave-lang.dev";
```

## Unterminated block comments

If the lexer reaches the end of the file before finding the final `*/`, it emits an `UnterminatedComment` diagnostic. v0.2.0-pre-beta uses diagnostic code `E1002` for this case.

When temporarily commenting out a large region, make sure nested comment depth remains balanced.
