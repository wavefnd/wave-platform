---
translation_set_id: whale-ir-reference
path: whale/ir-reference
locale: en
group: whale
group_order: 1
order: 6
title: Whale IR reference
summary: Describes types, identifiers, function validity, evaluation order, and exchange format.
---

## Modules and identities

A module contains target information, global definitions, and functions. Values have explicit types. Frontends resolve source-language names, types, overloads, and generics before producing typed IR.

Functions and global variables occupy separate internal namespaces. A function and a variable may therefore share a name. Internal identity is distinct from an external `link_name`; external names are supplied explicitly by the frontend. Whale does not resolve external collisions by inventing new names. See [symbols and linking](assembler-linker).

Each value definition has an identity. Definitions must be unique, and their type metadata must agree with their declared types. A name is not a substitute for identity when declarations shadow each other.

## Constructing and reading IR

This complete Rust example uses the `ir` crate to construct a function, verify it, and print its typed IR:

```rust
use ir::{ModuleBuilder, Target, Type};

fn main() {
    let target = Target::lookup("x86_64-whale-linux").unwrap();
    let mut module = ModuleBuilder::new(target.name(), target.data_layout());
    let mut function = module.begin_function("answer", vec![], Type::I32);
    let left = function.const_i32(40);
    let right = function.const_i32(2);
    let answer = function.add(Type::I32, left, right);
    function.ret(Some(answer));
    function.finish();
    let module = module.finish();
    ir::verify_module(&module).unwrap();
    print!("{}", ir::print_module(&module));
}
```

The printer produces:

```text
module {
  format_version 4
  semantics_version 1
  target "x86_64-whale-linux"
  datalayout { ptr=64, endian=little }

  declare @f0 "answer": whale () -> i32, linkage internal

  fn @f0 "answer"() -> i32, entry %b0 {
  %b0 "entry":
    %v0: i32 = const i32 40
    %v1: i32 = const i32 2
    %v2: i32 = add i32 %v0, %v1
    ret i32 %v2
  }

}
```

Save the output as `answer.wir` to read and verify it with the text parser or CLI. Scalar integer execution is described below.

Save the output as `answer.wir` to read and verify it with the text parser and CLI. IR execution is not yet available.

## Types

| Type | Meaning |
| --- | --- |
| `bool` | Logical false or true |
| `i1`, `i8`, `i16`, `i32`, `i64`, `i128` | Signed integers of the indicated bit width |
| `u1`, `u8`, `u16`, `u32`, `u64`, `u128` | Unsigned integers of the indicated bit width |
| `f16`, `f32`, `f64` | Floating-point values of the indicated bit width |
| `ptr<T>` | Pointer to a value of type T |
| `fnptr<signature>` | Callable pointer with exact parameter/result types and calling convention |
| `array<T, N>` | N elements of the same type |
| `struct{T, ...}` | Ordered structure fields |
| `tuple<T, ...>` | Ordered tuple elements |
| `void` | Absence of a result |

`bool`, `i1`, and `u1` are different types. Signed `i1` represents −1 and 0; unsigned `u1` represents 0 and 1. An integer 1 is not implicitly a Boolean condition. Branches, Select conditions, and `trap_if` require Bool operands.

Storage size is determined by the [target layout](memory-model), not solely by the number of value bits. For example, `i1` has one value bit but occupies at least one byte in memory.

## Functions and calls

A function specifies its complete parameter and result types, calling convention, and linkage. Direct and indirect calls must match the callable signature. A void call has no result ID. A nonvoid call defines a result even when that result is unused at O0.

A return must agree with the function result type. Void returns carry no value; nonvoid returns carry a value of the declared result type.

### Declarations, identities and calls

`Module.declarations` records each function's `FunctionId`, name, complete signature, linkage and external link name. A definition refers to this identity; parameter and return types must match its declaration. Identical repeated declarations resolve to the same ID through `declare_function`; conflicts and duplicate definitions are errors. An internal declaration needs a body in the module. An external declaration may be unresolved until linking, or have an exported body. Internal functions have no `link_name`; external functions require an explicit nonempty name without NUL. Two distinct function declarations cannot claim the same external name. Globals and functions still use separate internal namespaces.

Register declarations before constructing bodies with `begin_declared_function` to support forward calls and recursion. `begin_function` remains a convenience for a new internal Whale function. The checked `declare_function`, `begin_declared_function`, `function_addr`, `null_function` and `call` APIs return `Result`; a rejected call does not append an instruction or allocate its result ID.

The following complete Rust program declares an external function, takes its typed address, and emits both direct and indirect calls:

```rust
use ir::{Callee, CallingConvention, DataLayout, FunctionSignature, Linkage, ModuleBuilder, Type};

fn main() {
    let mut module = ModuleBuilder::new("x86_64-whale-linux", DataLayout::default_64bit_le());
    let signature = FunctionSignature {
        params: vec![Type::I32], ret: Type::I32,
        convention: CallingConvention::SysV64, variadic: false,
    };
    let identity = module.declare_function(
        "identity", signature, Linkage::External, Some("identity_i32".into()),
    ).unwrap();
    let mut function = module.begin_function("answer", vec![], Type::I32);
    let input = function.const_i32(42);
    let callback = function.function_addr(identity).unwrap();
    // The direct call's result remains defined even though it is unused.
    function.call(Callee::Direct(identity), vec![input]).unwrap();
    let result = function.call(Callee::Indirect(callback), vec![input]).unwrap().unwrap();
    function.ret(Some(result));
    function.finish();
    let module = module.finish();
    ir::verify_module(&module).unwrap();
    print!("{}", ir::print_module(&module));
}
```

```text
module {
  format_version 4
  semantics_version 1
  target "x86_64-whale-linux"
  datalayout { ptr=64, endian=little }

  declare @f0 "identity": sysv64 (i32) -> i32, linkage external, link_name "identity_i32"
  declare @f1 "answer": whale () -> i32, linkage internal

  fn @f1 "answer"() -> i32, entry %b0 {
  %b0 "entry":
    %v0: i32 = const i32 42
    %v1: fnptr<sysv64 (i32) -> i32> = function_addr @f0
    %v2: i32 = call sysv64 i32 @f0(%v0)
    %v3: i32 = call sysv64 i32 indirect %v1(%v0)
    ret i32 %v3
  }

}
```

`Callee::Direct(FunctionId)` resolves through the declaration table; `Callee::Indirect(ValueId)` requires a `Type::FnPtr(FunctionSignature)` value. The signature includes all parameter types, the result type and `CallingConvention::{Whale, SysV64}`. It is retained through copies, storage, parameters, returns, phi and select. Data pointers and integer values are not callable. Casts involving function-pointer types are rejected; changing a type annotation cannot change a callable signature. A function pointer has 64-bit address storage on this target; this does not itself implement runtime shadow metadata.

Arity, exact argument/result types, result-ID presence, and calling convention must match. There are no implicit conversions. The indirect callee must dominate the call just like its arguments. `variadic: true`, void parameters and SysV64 aggregate parameter/result signatures are rejected. Whale aggregate signatures can be represented in IR; native ABI classification and machine call emission are not yet available for either convention.

`null_function(signature)` represents a typed null function pointer. Calling it is well-typed IR with a required runtime trap before entering a callee. A nonnull target that is invalid, expired or incompatible with the checked signature must also trap. These runtime checks and foreign callback lifetime management await the interpreter/native execution layer; verifier success does not mean arbitrary external addresses are safe.

### AST call forms

These are expression fragments inside an AST format 2 program:

```json
{"Call":{"callee":{"Direct":"increment"},"args":[{"Lit":{"Int":{"bits":32,"signed":true,"value":"41"}}}]}}
```

```json
{"Call":{"callee":{"Indirect":{"FunctionRef":"increment"}},"args":[{"Lit":{"Int":{"bits":32,"signed":true,"value":"41"}}}]}}
```

`Direct` and `FunctionRef` use the function namespace even when a variable has the same name. `Indirect` evaluates its expression first, then evaluates arguments from left to right. A void call is valid as an `ExprStmt`, but not as a variable initializer, argument, operand or returned value. Calls and function references are not compile-time numeric constant expressions. `NullFunction` takes a signature object with `params`, `ret`, `convention` and `variadic` fields.

The [complete JSON example](https://github.com/wavefnd/Whale/blob/master/ir/tests/fixtures/ast-v2-calls.json) stores a callback and invokes it before an external call. Lower it with:

```sh
cargo run --locked --features socket-cli -- ir lower ir/tests/fixtures/ast-v2-calls.json
```

Its [expected IR](https://github.com/wavefnd/Whale/blob/master/ir/tests/fixtures/calls-v3.wir) is checked in the lowering tests. Function identity and link names are represented at the IR boundary; preserving them through native object generation and linking is still separate work.

## Blocks and value availability

Every block has a unique identity and exactly one terminator. Branch destinations must belong to the same function. The entry block must exist and has no incoming edges or phi instructions. To form a loop, branch from entry to a separate loop header.

On executable paths, a value definition must dominate its ordinary uses: every path from entry to the use must pass through the definition. Within one block, the definition precedes the use. Physical block storage order does not establish dominance.

Consider entry branching to either left or right, followed by a join. A value defined only in left cannot be used as an ordinary value at the join, because the path through right has no definition. Use a phi with an input from each predecessor instead.

Unreachable blocks remain in the module. Verification still checks their identities, types, operands, and branch structure. A definition in an unreachable block cannot supply an ordinary value on a reachable path.

## Phi instructions

Phi instructions appear before all non-phi instructions in a block. A phi has exactly one input per distinct predecessor block. Each input value must have the phi's type and be available at the end of its corresponding predecessor.

Repeated edges from one predecessor require one input, not one input per edge. A loop phi can refer to a value computed on its backedge even when that block is stored later in the module. Missing, duplicate, unrelated, or incorrectly typed inputs are verification errors.

## Evaluation and selection

The Whale AST evaluates the call target and subexpressions in the specified left-to-right order. Frontends express short-circuit operations with control-flow branches.

Select chooses between values that have already been computed. It does not suppress evaluation of either input. For example, selecting a safe value does not prevent a trap in the computation of the other input. Put a potentially trapping computation inside a conditional block when it must not execute on the other path.

## Validation and traps

Malformed IR is a verification error. The verifier rejects legacy `undef`; regenerate existing AST lowering. An uninitialized declaration uses format 4 `uninit` and checked actual reads, without zero initialization or arbitrary values. Runtime condition violations produce defined traps with IR locations.

The current `InterpreterTrap` reports a reason, executed step count and `ExecutionSite`: function ID, block ID, zero-based instruction index and optional result value ID. The terminator index follows the instructions. CLI diagnostics also name the input file. Typed IR does not yet carry source spans, so these are IR locations, not source line numbers. A trap returns an error and stops subsequent execution; the library does not abort the host process.

The guarantee covers verified IR and tracked memory. External C, raw addresses, and inline assembly have separate contracts; violations beyond those boundaries are not guaranteed to be detected. See the [memory model](memory-model).

## Interchange and textual representation

AST and typed IR use separate format versions and a common semantics version. Readers reject missing or unknown versions, unknown fields or features, and duplicate JSON keys. Producers must not rely on a reader silently ignoring an unsupported property.

Integers carry a bit width, signedness, and a textual numeric value. Floating-point constants carry a width and an exact bit pattern. A text IR round trip must preserve names, IDs, types, constants, ordering, attributes, and metadata. Whitespace and comment placement need not survive the round trip.

The AST JSON contract below and typed IR format 4 reading, verification and round-trip printing are available.

### Printed identities and quoted names

Typed IR format 4 prints function identities as `@fN`, globals as `@gN`, values as `%vN`, and blocks as `%bN`. Function and global IDs belong to the module; value and block IDs belong to the containing function. Preserve the supplied IDs, including gaps. Quoted names are descriptive annotations and are not used to resolve references. A function explicitly records `entry %bN`, independent of the order of its stored blocks.

This complete module was verified and printed through the Rust IR API. Both branch blocks are named `"branch"`; the IDs distinguish their definitions and the phi inputs:

```text
module {
  format_version 4
  semantics_version 1
  target "x86_64-whale-linux"
  datalayout { ptr=64, endian=little }

  declare @f0 "choose": whale (bool) -> i32, linkage internal

  fn @f0 "choose"(%v0 "condition": bool) -> i32, entry %b0 {
  %b0 "entry":
    cbr bool %v0, label %b1, label %b2
  %b1 "branch":
    %v1: i32 = const i32 1
    br label %b3
  %b2 "branch":
    %v2: i32 = const i32 2
    br label %b3
  %b3 "join":
    %v99: i32 = phi i32 [ %v1, %b1 ], [ %v2, %b2 ]
    ret i32 %v99
  }

}
```

`%v0` is defined by the parameter list, `%b1` and `%b2` are distinct despite their equal names, and the phi identifies each predecessor by ID. Branches and switch destinations use the same block-ID syntax. The printer does not renumber the explicitly supplied `%v99`.

All name and string fields use double quotes: target, function/global/parameter/block names, external link names, constant declaration names and trap reasons. Printable Unicode remains literal. The escapes are `\"`, `\\`, `\n`, `\r`, `\t`, `\0`, and `\u{hex}` with lowercase hexadecimal for other control characters and U+2028/U+2029. For example, a name containing a newline, tab, quote, backslash and Korean text prints on one record:

```text
"line\ncolumn\tquote\"slash\\한글"
```

The reader accepts typed IR formats 3 and 4 with semantics version 1 and always prints format 4. `uninit` requires format 4. Valid existing format 3 instructions remain readable, but legacy `undef` is a verification error in either format and must be regenerated from AST. Format 2 text requires manual migration to explicit IDs, quoted names and entry references. AST JSON has its separate format 2.

### Reading and verifying text IR

`ir::parse_module` reads typed IR, verifies it and returns a `Module`. It accepts the current printer’s scalar, control-flow, memory, direct/indirect call and constant-expression syntax. It preserves IDs, names, types, exact integer and float bits, expression trees and evaluated results, block order, entry, alignment, signatures and link_name. Whitespace and `//` line comments become canonical output. Save the complete IR above as `answer.wir` and run these commands in the default build.

```sh
cargo run --locked -- ir verify answer.wir
cargo run --locked -- ir print answer.wir -o canonical.wir
```

Unknown versions, fields, instructions and escapes, out-of-range literals, duplicate IDs, conflicting type annotations and trailing input are rejected. `ParseError` reports a byte offset and one-based line/Unicode-scalar column. Verification errors are attached to the containing function or global declaration when available. `print` also verifies input and preserves an existing output on failure. New syntax or semantics requires the corresponding format/semantics version change; unknown versions are errors.

### Verification resource limits

`IrLimits` defaults to 8 MiB of input, 1,000,000 tokens, 1,000,000 traversal nodes, and type/expression depths of 128 each. Root depth is 0; each child adds 1. Depths may be lowered or raised up to `MAX_IR_NESTING`, 256. Excessive depth/work returns `LimitError`, `VerifyError::ResourceLimit`, `ConstEvalError::ResourceLimit` or `CallError::ResourceLimit`. Nodes count traversal at each input boundary, including type annotations and expression nodes, rather than elapsed time.

```rust
use ir::{parse_module_with_limits, print_module, IrLimits};

fn main() -> Result<(), Box<dyn std::error::Error>> {
    let source = std::fs::read_to_string("answer.wir")?;
    let limits = IrLimits {
        max_type_depth: 32,
        max_const_depth: 32,
        max_nodes: 10_000,
        ..IrLimits::default()
    };
    let module = parse_module_with_limits(&source, limits)?;
    let canonical = print_module(&module);
    let reread = parse_module_with_limits(&canonical, limits)?;
    assert_eq!(print_module(&reread), canonical);
    let invalid = source.replacen("format_version 4", "format_version 99", 1);
    assert!(parse_module_with_limits(&invalid, limits).is_err());
    Ok(())
}
```

Pass limits to `verify_module_with_limits`, `ConstExpr::evaluate_with_limits`, `validate_signature_with_limits` and `ModuleBuilder::declare_function_with_limits` as well. Iterative type traversal precedes recursive clone, equality and diagnostics; constant evaluation uses a work stack. Borrowed Rust trees remain caller-owned, including Drop. Arbitrary unverified trees still have recursive clone/Drop; a rejected owned signature in the checked declaration API is disposed of iteratively. Malformed IR is a verification error. The verifier rejects legacy `undef`; regenerate existing AST lowering. An uninitialized declaration uses format 4 `uninit` and checked actual reads, without zero initialization or arbitrary values. Runtime condition violations produce defined traps with IR locations.

### Versioned AST JSON

Save the following as `program.json`. All four envelope fields are required. `program` contains required `declarations`, `globals` and `functions` arrays, which may be empty. Function name, parameters, return type, body, `convention` and `linkage` are required. `link_name` may be absent/null for internal functions and must be a nonempty string without NUL for external functions. Each enum uses either its unit name or a single variant-key object. Unit variants also accept a null-valued object, such as `{"Void":null}`; the encoder emits the unit name `"Void"`. `VarDecl.init` may be absent or null; other required fields must be present.

```json
{
  "format_version": 2,
  "semantics_version": 1,
  "features": [],
  "program": {
    "globals": [],
    "functions": [
      {
        "name": "answer",
        "parameters": [],
        "return_type": {
          "Int": {
            "bits": 128,
            "signed": false
          }
        },
        "body": [
          {
            "Return": {
              "Lit": {
                "Int": {
                  "bits": 128,
                  "signed": false,
                  "value": "340282366920938463463374607431768211455"
                }
              }
            }
          }
        ],
        "convention": "Whale",
        "linkage": "Internal",
        "link_name": null
      }
    ],
    "declarations": []
  }
}
```

```sh
cargo run --locked --features socket-cli -- ir lower program.json
```

```text
module {
  format_version 4
  semantics_version 1
  target "x86_64-whale-linux"
  datalayout { ptr=64, endian=little }

  declare @f0 "answer": whale () -> u128, linkage internal

  fn @f0 "answer"() -> u128, entry %b0 {
  %b0 "entry":
    %v0: u128 = const u128 340282366920938463463374607431768211455
    ret u128 %v0
  }

}
```

Integer `value` is a decimal string: optional minus for signed integers, followed by decimal digits, without whitespace, plus, exponent or separators. The declared width/signedness determines the accepted range. `u128::MAX` above survives JSON and lowering exactly. A negative unsigned value or an out-of-range value fails instead of wrapping. Floating values use exact-width hexadecimal storage strings, described in [numeric operations](numeric-operations).

`format_version` is 2 for this AST format; `semantics_version` is 1. `features` must be an empty array. Unknown fields, versions, features, duplicate raw JSON keys (including escaped equivalent keys), and trailing values are errors, even with `--no-verify`. The library entry point is `ir::lower_ast::interchange::decode`; `encode` emits the envelope. `decode` defaults to an 8 MiB source-byte limit; `decode_with_limit` accepts a caller limit. JSON nesting is bounded. Use this raw decoder rather than parsing into a generic map that could already discard duplicate keys.

[The complete JSON Schema](https://github.com/wavefnd/Whale/blob/master/ir/schema/ast-v2.schema.json) specifies shapes, required fields and variants. Range/type checks and duplicate-key detection additionally apply. The scalar lowering subset includes literals, variables/constants, add/sub/mul, comparisons, assignment, if/while, return and break/continue. Function references, direct calls and indirect calls are supported; aggregate expressions are unsupported. `Opaque` is representable in the schema but unsupported by lowering.

Migration requires wrapping old bare Program payloads and replacing numeric JSON literals with decimal integer strings or float bit strings. Old unversioned payloads are rejected. Format 1 payloads must be migrated to format 2: add `program.declarations` (an empty array when unused) and explicit `convention`/`linkage` on definitions. AST and typed IR version numbers are independent: AST format 2, typed IR format 4, and semantics version 1.

### Rejected input and CLI recovery

Save this complete input as `invalid.json`:

```json
{
  "format_version": 99,
  "semantics_version": 1,
  "features": [],
  "program": {
    "globals": [],
    "functions": [],
    "declarations": []
  }
}
```

```sh
cargo run --locked --features socket-cli -- ir lower invalid.json -o rejected.wir
```

```text
Failed to parse socket JSON: unsupported AST format_version 99; expected 2
```

The command exits nonzero without creating an output or replacing an existing file. A type mismatch also fails before output publication. A binary built without `socket-cli` exits with status 2 and prints a recovery command containing `--features socket-cli`.


## Scalar integer interpreter

The default interpreter executes integer/Bool values, control flow, stack allocations, data-pointer storage and loads, typed GEP, memcpy and memset. Checked-pair reads exclude padding. Addresses are synthetic 64-bit values, never host-memory dereferences. Function arguments and returns remain integer/Bool or void; float, calls, function pointers, general aggregate values, global addresses and native execution are unsupported. Stack allocations live until function return. Lexical lifetime-end operations, pointer transfer through calls/returns, foreign-memory adapters and native shadow metadata require further implementation.

`InterpreterOptions::max_steps` defaults to 1,000,000. Each executed instruction, including a phi, and each terminator consumes one step. A zero limit prevents the first operation; an infinite branch loop returns `InterpreterError::StepLimit`. `ir_limits` independently bounds verification. Unused arithmetic still executes and can trap. Overflow from checked arithmetic is a Bool result; only an explicit trap_if makes it a trap.

Save this complete module as `swap-loop.wir`. The entry block is explicit although the exit block is stored first. On entry to a block, all phi inputs are read from the previous block before any phi result is written. Three iterations swap 11 and 22 three times and return 22.

```text
module {
  format_version 4
  semantics_version 1
  target "x86_64-whale-linux"
  datalayout { ptr=64, endian=little }

  declare @f7 "swap_loop": whale (u32) -> i32, linkage internal

  fn @f7 "swap_loop"(%v0 "iterations": u32) -> i32, entry %b11 {
  %b90 "exit":
    ret i32 %v5
  %b11 "entry":
    %v1: i32 = const i32 11
    %v2: i32 = const i32 22
    %v3: u32 = const u32 0
    %v4: u32 = const u32 1
    br label %b20
  %b20 "loop":
    %v5: i32 = phi i32 [ %v1, %b11 ], [ %v6, %b30 ]
    %v6: i32 = phi i32 [ %v2, %b11 ], [ %v5, %b30 ]
    %v7: u32 = phi u32 [ %v3, %b11 ], [ %v9, %b30 ]
    %v8: bool = icmp ult u32 %v7, %v0
    cbr bool %v8, label %b30, label %b90
  %b30 "next":
    %v9: u32 = add u32 %v7, %v4
    br label %b20
  }

}
```

```shell
whale ir run swap-loop.wir --function @f7 --arg 3 --max-steps 100
```

```text
i32 22
```

The Rust API returns a value and step count, or a structured verification, unsupported-operation, argument, step-limit or trap error. This complete program reads the same `swap-loop.wir` file:

```rust
use ir::{interpret_with_options, parse_module, ConstValue, FunctionId,
         InterpreterError, InterpreterOptions};

fn main() -> Result<(), Box<dyn std::error::Error>> {
    let source = std::fs::read_to_string("swap-loop.wir")?;
    let module = parse_module(&source)?;
    let options = InterpreterOptions {
        max_steps: 100,
        ..InterpreterOptions::default()
    };
    let result = interpret_with_options(
        &module, FunctionId(7), &[ConstValue::U(3)], options,
    )?;
    assert_eq!(result.value, Some(ConstValue::I(22)));
    assert_eq!(result.steps, 32);
    let stopped = interpret_with_options(
        &module, FunctionId(7), &[ConstValue::U(3)],
        InterpreterOptions { max_steps: 0, ..options },
    );
    assert!(matches!(stopped, Err(InterpreterError::StepLimit { .. })));
    println!("i32 22");
    Ok(())
}
```
