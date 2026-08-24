---
translation_set_id: system-io
path: reference/system-io-network-process
locale: en
group: reference
group_order: 3
order: 6
title: System I/O, files, networking, and processes
summary: FD-based I/O, file systems, sockets/TCP/UDP, process APIs, and their failure contracts.
---

## Opening and closing files

```wave
import("std::fs::file");

fun main() {
    var fd: i64 = fs_open_read("input.txt");
    if (fd >= 0) {
        fs_close(fd);
    }
}
```

`fs_open_read` can preserve a negative failure result, so check it before use. `fs_close` also returns a result; code that cares about close failures should inspect it.

## std::io

`std::io` provides file-descriptor operations for reading, writing, exact-length transfers, seeking, and copying. APIs that accept buffers require the pointer and length to be kept in the same unit and contract.

## std::fs

`std::fs::file` includes helpers for existence checks, opening, closing, size queries, removing files, creating/removing directories, complete reads/writes, and file copying.

```wave
var size: i64 = fs_file_size("input.txt");
if (size < 0) {
    println("file error");
}
```

## Networking

`std::net` is divided into address handling, base sockets, socket options, polling, TCP, and UDP. Network code should explicitly manage:

- Address family and socket type
- Port/integer byte order
- Blocking versus non-blocking state
- Partial reads and writes
- Shutdown and error results

## Processes

`std::process` provides source units for process creation, waiting, constants, and standard-stream redirection. Treat successful process creation and the child's eventual exit status as separate results.

## Platform dependence

These modules are implemented on top of `std::sys` and native calls. Programs that depend on specific error values or flags should define and test their target OS and ABI explicitly.
