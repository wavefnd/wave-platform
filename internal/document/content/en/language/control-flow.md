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

The unparenthesized form `if score >= 90 { ... }` is not accepted by this release's parser.

## while loops

`while` conditions are also parenthesized.

```wave
var index: i32 = 0;
while (index < 10) {
    index += 1;
}
```

## for loops

`for` uses an initializer, condition, and increment expression.

```wave
for (var i: i32 = 0; i < 10; i += 1) {
    println("{}", i);
}
```

The initializer can use `var`, `let`, `let mut`, a typed binding, or a general expression. `const` and `static` are rejected as local for-loop initializers.

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

The value matched by `match` is also parenthesized. Current patterns handle integer literals, identifier-shaped names, and the `_` wildcard.

```wave
match (status) {
    200 => { println("ok"); }
    404 => { println("not found"); }
    _ => { println("other"); }
}
```

A `match` cannot contain duplicate `_` wildcard arms. Every arm body is a `{ ... }` block.
