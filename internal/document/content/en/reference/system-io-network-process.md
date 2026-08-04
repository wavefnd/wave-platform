---
translation_set_id: system-io
path: reference/system-io-network-process
locale: en
group: reference
group_order: 3
order: 6
title: System I/O, files, networking, and processes
summary: Use the release's fd, filesystem, socket, TCP, UDP, and process modules.
---

## File descriptors and files

```wave
import("std::fs::file");
import("std::io::fd");

var fd: i64 = fs_open_read("input.txt");
if (fd >= 0) {
    fs_close(fd);
}
```

std::io provides descriptor-level read, write, seek, and copy helpers. std::fs provides existence, open, size, read-all, write-all, copy, metadata, directory, and removal operations.

## Networking

std::net separates addresses, polling, base sockets, socket options, TCP, and UDP. Network addresses and ports have explicit host/network byte-order conversions.

## Processes

std::process provides core process operations, constants, spawn, wait, and standard-stream redirection. Return values preserve platform-style error reporting and must be checked.

