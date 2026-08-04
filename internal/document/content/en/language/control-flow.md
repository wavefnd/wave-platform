---
translation_set_id: control-flow
path: language/control-flow
locale: en
group: language
group_order: 2
order: 4
title: Control flow
summary: Conditions, loops, match expressions, and control transfer.
---

## Conditions

```wave
if score >= 90 {
    println("A");
} else if score >= 80 {
    println("B");
} else {
    println("C");
}
```

## Loops

```wave
var index: i32 = 0;
while index < 10 {
    index += 1;
}

for (item: i32 = 0; item < 10; item += 1) {
    println("{}", item);
}
```

break exits the nearest loop and continue begins its next iteration.

## Pattern selection

```wave
match (status) {
    200 => { println("ok"); }
    404 => { println("not found"); }
    _ => { println("other"); }
}
```

