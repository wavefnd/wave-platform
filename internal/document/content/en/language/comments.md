---
translation_set_id: comments
path: language/comments
locale: en
group: language
group_order: 2
order: 11
title: Comments
summary: Line comments, nested block comments, and diagnostic behavior.
---

## Line comments

```wave
var count: i32 = 10; // ignored through the end of the line
```

## Block comments

```wave
/* outer comment
   /* nested comment */
   outer comment continues
*/
```

Comment markers inside string literals remain string content. An unterminated block comment is a compile-time error whose diagnostic points to the opening delimiter.

