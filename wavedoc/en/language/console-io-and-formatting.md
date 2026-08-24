---
translation_set_id: console-io-formatting
path: language/console-io-and-formatting
locale: en
group: language
group_order: 2
order: 13
title: Console I/O and formatting
summary: The print, println, and input statements and their placeholder rules.
---

## I/O statements

`print`, `println`, and `input` are Wave language statements for formatted console I/O.

```wave
fun main() {
    var count: i32 = 0;
    input("{}", count);
    print("count = ");
    println("{}", count);
}
```

Each statement ends with `;`. The first argument must be a string literal; a variable or computed string cannot be used as the format argument.

## Placeholders

Only the exact two-character sequence `{}` is a placeholder.

```wave
println("name = {}, score = {}", name, score);
```

The number of placeholders must equal the number of following expressions. A mismatch is a compile-time error.

```wave
println("{} {}", one);       // error: two placeholders, one value
println("plain text", one); // error: no placeholder, extra value
```

Other braces remain literal text; there are no named or indexed placeholders in this syntax.

## print and println

`print` writes the formatted text as-is. `println` appends a newline.

```wave
print("loading...");
println("done");
```

Formatting accepts scalar values such as integers, floating-point values, strings, and pointers. Arrays and structs are not formatting arguments.

## input destinations

Every expression after the `input` format must identify writable storage for the parsed value.

```wave
var number: i32 = 0;
input("{}", number);
```

Variables, supported field accesses, and supported dereference forms can be destinations. Literals and computed rvalues cannot.

If `input` cannot convert every requested value, the program exits with a failure status.

## Runtime boundary

These console statements require a hosted runtime. Freestanding kernels and embedded programs provide target-specific input and output through explicit functions or FFI boundaries.
