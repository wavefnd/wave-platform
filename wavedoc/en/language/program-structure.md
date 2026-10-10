---
translation_set_id: program-structure
path: language/program-structure
locale: en
group: language
group_order: 2
order: 1
title: 1. From source file to running program
summary: Learn about source files, functions, output, inspection, and execution.
---

## Before starting this chapter

Prepare compiler and standard library according to [Installation instructions](/docs/en/getting-started/install). If you can run `wavec --version` in your terminal, you can get started. Any editor that can save plain text files will do.

In this chapter, we start with creating a single line of output and learn the relationships between source files, functions, compilation, execution, and exit code. The goal is not just to copy instructions, but to be able to explain what happens at what stage.

## Create a working directory

Using a separate directory for each program makes it easier to find the source and generated executable files. Create and navigate to a directory in the terminal.

```shell
mkdir wave-study
cd wave-study
```

Create `main.wave` in this directory with your editor. Check the extension to make sure the file name is not `main.wave.txt`. Below is the entire file, not just a fragment inside the function.

<!-- wave-example: book-hello -->
```wave playground
fun main() {
    println("Hello, Wave!");
}
```

Execution result:

```text
Hello, Wave!
```

Run it with the following command:

```shell
wavec run main.wave
```

Do not mix terminal commands with the code Wave. Enter `wavec run` into the terminal, and write `fun main` into the source file. It is not necessary to paste the `Hello, Wave!` output by the program back into the source.

## read line by line

`fun` is a keyword that declares a function. A function is a named set of operations, and `main` is the entry point of the executable. You will define other functions yourself later.

`()` after `main` is where you write the parameters. main in this program takes no parameters and is therefore empty. Write down the function's operations between `{` and `}`. Indentation makes blocks easier for humans to read, and the block boundaries themselves are indicated by curly braces.

`println("Hello, Wave!");` is a sentence that outputs a string. Double quotation marks indicate the beginning and end of a string literal. The quotation marks themselves are not included in the output. The semicolon indicates the end of this statement.

## Statements are executed in the order they are written

Try changing it to print three times. Do not combine with the previous program, but replace the contents of main.wave with the entire program below.

<!-- wave-example: book-sequence -->
```wave playground
fun main() {
    println("start");
    println("working");
    println("done");
}
```

Execution result:

```text
start
working
done
```

After finishing the first sentence, move on to the next sentence. There are no tasks running concurrently here. If you want to change the output order, just change the order of the sentences. You can control this sequence by learning conditional statements and loops later.

## print and println

`println` adds a line break at the end. `print` does not automatically change lines. The difference is evident in connecting small pieces to create a single line.

<!-- wave-example: book-print-lines -->
```wave playground
fun main() {
    print("Wave");
    print(" ");
    println("study");
    println("second line");
}
```

Execution result:

```text
Wave study
second line
```

Printing three times does not always result in three lines. Distinguish between number of output calls and number of lines. When writing a line break directly in a string, use `\n` escape. The string escape and line breaks in the source are covered in detail in [string sheet](/docs/en/language/strings).

## Insert value into string

To put the calculation result into a string, pass the value corresponding to `{}`.

<!-- wave-example: book-format-first -->
```wave playground
fun main() {
    println("{} + {} = {}", 2, 3, 2 + 3);
}
```

Execution result:

```text
2 + 3 = 5
```

The first `{}` contains 2, the second contains 3, and the third contains 5. The format string and value are separated by commas. If you change the number of values, the number of placeholder must also match. This is a different output syntax than the string addition operation.

## Divide inspection, build, and execution

run used so far carries out build and execution sequentially. If you want to know which step failed, you can break it down like this:

```shell
wavec check main.wave
wavec build main.wave -o hello
```

check checks grammar, types, etc., but does not test all input to the program. For example, source inspection alone cannot determine whether a file exists at runtime. build creates an executable file. In Linux/macOS, execute as follows.

```shell
./hello
```

In Windows, name the output file `hello.exe` and run in PowerShell as `.\hello.exe`. If you modify the source after creating an executable file, you must rebuild it for the changes to be reflected.

## The exit code is also a result

Humans read the output statement, but a shell or other program can determine success by the exit code. The following specifies that main returns i32.

<!-- wave-example: book-exit-success -->
```wave playground
fun main() -> i32 {
    println("completed");
    return 0;
}
```

Execution result:

```text
completed
```

0 is a convention to indicate a normal shutdown. When signaling a failure directly, return a non-zero code. In Linux/macOS shell, the code is checked as `echo $?` immediately after execution, and in PowerShell, it is checked as `$LASTEXITCODE`. If you run other commands in the meantime, what you check may change.

Executing `return` ends the function. main does not declare any parameters and even if they have default values, they are not allowed. Omit the return type of main or use i32.

## How to read the first error

The following code is intentionally incorrect: Unlike the running example, it should fail at check.

```wave
fun main() {
    println("hello")
}
```

There is no semicolon at the end of the sentence. Look at the line the diagnosis displays and the sentence immediately preceding it. The location marked by the compiler may not be where the error originated, but rather where the incorrect structure can no longer be interpreted.

First, fix the first error and then recheck. If the preceding parentheses or quotes are not closed, various errors may ensue even in the normal code that follows. If you try to fix all lines at the same time, it's easy to miss the original cause.

## practice problems

1. Create a program that prints three lines of self-introduction.
2. Pass 12 and 8 as format arguments to print `12 * 8 = 96`.
3. Write it to print a success message and return an exit code of 0.
4. Predict before execution what the number of output lines will be if you use print only three times.

### Solution: Calculation process output

<!-- wave-example: book-first-solution -->
```wave playground
fun main() -> i32 {
    println("Learning Wave");
    println("My first program");
    println("Ready to calculate");
    println("{} * {} = {}", 12, 8, 12 * 8);
    return 0;
}
```

Execution result:

```text
Learning Wave
My first program
Ready to calculate
12 * 8 = 96
```

Instead of writing the calculation result directly in the string as `96`, I passed it as an expression. This is the first step to ensure that the calculation results and display do not change even if the input value changes. In the next chapter, we will give variables names to avoid writing the same value multiple times.


## Top-level item in source file

The Wave source can consist of the following top-level items:

- `import(...)`
- `const`, `static`
- `type`
- `struct`, `enum`, `proto`
- `extern(...)`, `export(...)`
- `fun`
- `#[target(...)]` condition before the supported item

Prefix importable declarations with `pub`. Place the local `var` declaration inside a function or block.

## Freestanding Program

Targets without a kernel, boot code, or runtime can use the freestanding build option.

```shell
wavec build kernel.wave \
  --freestanding \
  --entry=_start \
  --linker-script=linker.ld \
  --no-start-files
```

`--freestanding` links to a build plan that turns off default library dependencies, and `--entry` sets the linker entry symbol. To create an actual bootable output, you need to design the target architecture, linker script, and even object format.

## Entry points that intentionally fail

You cannot have a parameter in main even if it has a default value. It is an error to copy the following file check.

<!-- wave-example: reject-main-argument -->
```wave
fun main(value: i32 = 0) -> i32 {
    return value;
}
```
