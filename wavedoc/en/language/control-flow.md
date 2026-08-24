---
translation_set_id: control-flow
path: language/control-flow
locale: en
group: language
group_order: 2
order: 4
title: Control flow
summary: Parenthesized conditionals and loops, C-style for loops, match, break, and continue.
---

## Conditionals

Conditions for `if` and `else if` **must be parenthesized**.

```wave
if (score >= 90) {
    println("A");
} else if (score >= 80) {
    println("B");
} else {
    println("C");
}
```

Write `if (score >= 90) { ... }`; the unparenthesized form is invalid.

## Conditions cannot mutate state

Assignments, compound assignments, `++`, and `--` are rejected anywhere inside
an `if`, `else if`, `while`, or `for` condition. Use `==` for comparison, or
move the mutation to a statement before the condition.

```wave
value = read_value();
if (value == expected) {
    println("matched");
}
```

Ordinary assignment statements and the increment section of a `for` loop are
still allowed.

## while loops

`while` conditions are also parenthesized.

```wave
var index: i32 = 0;
while (index < 10) {
    index += 1;
}
```

## for loops

`for` uses an initializer, condition, and update expression.

```wave
for (var i: i32 = 0; i < 10; i += 1) {
    println("{}", i);
}
```

The initializer can declare a local with `var` or evaluate an expression. `const` and `static` are top-level declarations and cannot be used as for-loop initializers.

## break and continue

```wave
var i: i32 = 0;
while (i < 20) {
    i += 1;
    if (i == 5) {
        continue;
    }
    if (i == 10) {
        break;
    }
}
```

`break` exits the nearest loop, while `continue` proceeds to the next iteration.

## match

The value matched by `match` is parenthesized. A pattern can be an integer literal, an enum variant name, or the `_` wildcard.

```wave
match (status) {
    200 => { println("ok"); }
    404 => { println("not found"); }
    _ => { println("other"); }
}
```

A `match` cannot contain duplicate `_` wildcard arms. Every arm body is a `{ ... }` block.
