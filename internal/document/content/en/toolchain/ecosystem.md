---
translation_set_id: ecosystem
path: toolchain/ecosystem
locale: en
group: toolchain
group_order: 4
order: 1
title: Wave toolchain
summary: The compiler, Vex package manager, Whale toolchain work, and native bindings.
---

## Components

| Component | Role in this release |
| --- | --- |
| wavec | Wave compiler and command-line frontend |
| Vex | Wave package-management project |
| Whale | Toolchain/backend work; wavec --whale is not implemented in 0.2.0-pre-beta |
| native | Narrow C/C++ bindings used where Wave's standard library is not yet sufficient |

Treat each component's own release and repository as authoritative for commands it implements. The platform Source service provides read-only mirrors for code exploration.

