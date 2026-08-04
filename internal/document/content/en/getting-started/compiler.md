---
translation_set_id: compiler
path: getting-started/compiler
locale: en
group: getting-started
group_order: 1
order: 3
title: Compiler command reference
summary: Compile, check, run, and inspect Wave programs with wavec.
---

## Basic workflow

```shell
wavec build main.wave
wavec run main.wave
wavec build --emit=check main.wave
```

Use wavec --help as the authoritative list of flags in the installed release. Diagnostics include the source location and a description of the rejected construct.

## Outputs

- build compiles an input program.
- run compiles and executes a program.
- --emit=check performs frontend validation without producing a normal executable.

> **Whale status**
> 
> The --whale option is reserved in 0.2.0-pre-beta and reports that the Whale backend is not implemented. Do not depend on it for production builds.

