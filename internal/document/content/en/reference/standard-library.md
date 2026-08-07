---
translation_set_id: standard-library
path: reference/standard-library
locale: en
group: reference
group_order: 3
order: 1
title: Standard library index
summary: The 14 top-level modules declared by the v0.2.0-pre-beta std manifest and how to navigate them.
---

## Top-level modules in this release

`std/manifest.json` for v0.2.0-pre-beta declares 14 top-level modules:

| Area | Modules | Representative purpose |
| --- | --- | --- |
| Strings and data | `string`, `bytes`, `buffer` | String operations, endian/byte helpers, growable buffers |
| Math | `math` | Integer and mathematical helpers |
| Memory and C boundary | `mem`, `libc` | Manual memory and C runtime bindings |
| Files and I/O | `io`, `fs` | File descriptors, files, and directories |
| Networking | `net` | Addresses, sockets, polling, TCP, UDP |
| Environment, paths, time | `env`, `path`, `time` | Environment variables, paths, time operations |
| System and processes | `sys`, `process` | OS boundaries, process creation and waiting |

## Import granularity

Imports normally identify the specific `.wave` source unit that defines the API you need, rather than only the top-level module name.

```wave
import("std::string::len");
import("std::fs::file");
import("std::mem::alloc");
```

For example, the `std::string::len` source unit defines both `len` and `is_empty`.

## Finding APIs

The standard library is shipped as Wave source, so the installed release is the most precise API reference.

```shell
wavec print std-path
```

Open the module's `.wave` files at that location to inspect signatures, return contracts, and lower-level imports.

## Platform boundaries

APIs under modules such as `sys`, `libc`, `fs`, `net`, and `process` often preserve platform-style failure results, including negative error values. Do not assume success; inspect and handle each function's return contract.

The standard library evolves separately from language syntax, so avoid mixing `std` sources from a different Wave release with the compiler you are documenting.
