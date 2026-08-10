---
translation_set_id: console-io-formatting
path: language/console-io-and-formatting
locale: en
group: language
group_order: 2
order: 13
title: Console I/O and formatting
summary: Parser-recognized print, println, and input statements and their placeholder rules.
---

## I/O statements

Wave currently recognizes `print`, `println`, and `input` as language statements. They look like calls, but they are parsed directly rather than resolved as ordinary functions.

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

The parser counts placeholders and requires exactly the same number of following expressions. A mismatch is a compile-time error.

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

Formatting is type-directed during code generation. Integer, floating-point, pointer, string-like, and aggregate values are lowered according to their Wave type and the hosted C runtime interface.

## input destinations

Every expression after the `input` format must be a writable lvalue because the runtime stores parsed data through its address.

```wave
var number: i32 = 0;
input("{}", number);
```

Variables, supported field accesses, and supported dereference forms can be destinations. Literals and computed rvalues cannot.

The generated hosted implementation checks how many fields were converted. If the runtime does not convert every requested value, the program exits with a failure status.

## Runtime boundary

These statements currently lower through hosted C `printf` and `scanf`-style facilities. They therefore require the normal hosted runtime and are not a portable freestanding I/O mechanism. Kernels and embedded programs should define target-specific output and input behind explicit functions or FFI boundaries.
