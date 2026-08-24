---
translation_set_id: standard-library
path: reference/standard-library
locale: en
group: reference
group_order: 3
order: 1
title: Standard library index
summary: The top-level Wave standard-library modules, import paths, API discovery, and platform boundaries.
---

## Top-level modules

The Wave standard library contains these top-level modules:

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
import("std::string::len")::{len, is_empty};
import("std::fs::file")::{fs_open_read, fs_close};
import("std::mem::alloc")::{mem_alloc, mem_free};
```

The selective form makes the named public declarations available without a namespace prefix. A plain `import("std::string::len");` instead exposes them through the `std::string::len` namespace.

## Finding APIs

The standard library is distributed as Wave source. Locate it with:

```shell
wavec print std-path
```

Open a module's `.wave` files to inspect its public signatures, return contracts, and lower-level imports.

## Platform boundaries

APIs under modules such as `sys`, `libc`, `fs`, `net`, and `process` often preserve platform-style failure results, including negative error values. Do not assume success; inspect and handle each function's return contract.

Use the standard library installed with the compiler so its modules and compiler agree on their language and ABI contracts.
