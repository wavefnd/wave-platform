---
translation_set_id: whale-amd64-target
path: whale/amd64-target
locale: en
group: whale
group_order: 1
order: 10
title: AMD64 targets and ABI
summary: Linux AMD64 Describes target properties, call boundaries, and native function scope.
---

## Target identity

The native profile uses the target identifier `x86_64-whale-linux`.

| Property | Value |
| --- | --- |
| Operating system | Linux |
| Instruction set | AMD64 |
| Byte order | Little endian |
| Native address width | 64 bits |
| Object format | ELF64 |
| C calling convention | SysV AMD64 ABI |
| Static executable format | ELF ET_EXEC |

A build host and an output target are different concepts. Running Whale on another host does not imply support for that host's instruction set or object format. See [feature availability](overview) for the implemented compilation paths.

## Selecting and validating a target

The experimental `ir lower` command accepts `x86_64-whale-linux` as its default and only supported target. Build with `--features socket-cli` to use the command:

```sh
whale ir lower program.json --target x86_64-whale-linux
```

The target supplies a 64-bit, little-endian data layout independently of the build host. An unknown identifier or unsupported combination, such as `aarch64-whale-linux` or `x86_64-whale-windows`, fails with the supported choice before reading the input or replacing an output file. `--no-verify` does not disable target selection checks.

For an empty AST (`{"format_version":2,"semantics_version":1,"features":[],"program":{"declarations":[],"globals":[],"functions":[]}}`), the command prints the following IR header and module:

```text
module {
  format_version 4
  semantics_version 1
  target "x86_64-whale-linux"
  datalayout { ptr=64, endian=little }

}
```

For an unsupported target, the command exits with failure:

```sh
whale ir lower program.json --target banana
```

```text
Error: unsupported target "banana"; supported targets: x86_64-whale-linux
```

Rust clients can select `ir::Target::lookup("x86_64-whale-linux")` and pass its `name()` and `data_layout()` to `lower_o0`. Lowering rejects a supplied layout that disagrees with the selected target. `verify_module` also rejects unsupported target names and mismatched layouts in manually constructed IR.

The object model stores `ObjectTarget` with `format`, `machine`, `endian`, and `address_bits`. `ObjectFile::with_target` retains explicit identity; serialization accepts only AMD64, little-endian, 64-bit ELF64. Machine identifiers for other architectures do not enable their encoders. Both ELF writer entry points validate metadata, and linker inputs are checked before symbol resolution or linking. The ELF header retains `EM_X86_64` for supported objects.

`ObjectFile::new(ObjectFormat::ELF64)` remains a convenience constructor for that AMD64 identity. Code that previously accessed `object.format` must use `object.target.format`.

These checks provide target selection and object identity. Aggregate size, field-offset, and stride queries are available through the IR layout API. Object-file reading, native ABI lowering, and executable linking remain unavailable. The scalar lowering path continues to assign explicit alignment; complete aggregate layout rules are described in the [memory reference](memory-model).

## Calls and signatures

The IR declaration and call verifier accepts explicit `Whale` and `SysV64` conventions. The convention is part of a `fnptr` type and must match at the call site. Variadic signatures and SysV64 aggregate signatures are rejected. See the [executable call-construction example](ir-reference). This validates IR contracts; machine ABI classification, argument registers/stack placement and native call emission are not yet implemented.

A supported call must have an explicit signature and calling convention. Unsupported signatures are errors; a backend must not approximate them by dropping arguments or substituting a different representation.

Signature support distinguishes basic integers and pointers, f32/f64, aggregates and wide values, and variadic arguments. A type appearing in the IR type system does not by itself imply ABI support for passing or returning that type. Check each required category before selecting a backend.

For example, a backend accepting an i32 return type must still reject an unsupported aggregate-return signature. It cannot use the scalar return convention merely because part of the aggregate fits in a scalar register.

## Internal calls and C boundaries

Native pointer addresses occupy 64 bits. Tracked pointers also carry shadow metadata that must survive copies, storage, arguments, and return values.

The C ABI and the internal metadata-passing convention are distinct. Calls across a C boundary require explicit adapters. Passing a numerical address through the C ABI must not silently claim that allocation identity, lifetime, bounds, or access permissions were preserved.

## Stack frames

Keep the frame pointer and do not use the red zone. At O0, local-variable storage is not reused between distinct variables. Call-frame debug information describes the native frame and its relation to the original program.

These rules support debugging and do not establish exception-unwinding support. See [O0 and debugging](o0-debugging).

## Profile limits

The first O0 native profile does not include:

- O1 and higher optimization, vectorization, or LTO.
- Shared memory and atomic operations.
- Exception unwinding, async functions, or coroutines.
- Garbage collection or language-specific ownership enforcement.
- Arbitrary ownership transfer of external memory.
- Full inline-assembly constraint support.
- Dynamic linking, TLS, additional instruction sets, or additional object formats.

A request outside the supported profile must be rejected rather than silently replaced with another feature. These limits describe the native compilation path; they do not remove independently available toolchain components.
