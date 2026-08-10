---
translation_set_id: diagnostics
path: reference/diagnostics
locale: en
group: reference
group_order: 3
order: 2
title: Diagnostics and troubleshooting
summary: Human and JSON diagnostics, check mode, debug output, and a reproducible bug-report workflow.
---

## Check the version first

Before investigating syntax behavior, record the exact installed compiler version.

```shell
wavec --version
```

Also compare the command surface with `wavec --help`; an older compiler may not implement the current documented contract.

## Check the front end only

To separate Wave source errors from linking or execution:

```shell
wavec build main.wave --emit=check
```

This checks Wave input without producing a normal executable.

## JSON diagnostics

For IDEs, CI, and build tools that need structured diagnostics:

```shell
wavec --error-format=json build main.wave --emit=check
```

Keep the default human-readable format for terminal use and JSON for automated consumers.

## Inspect compiler stages

```shell
wavec --debug-wave=tokens build main.wave --emit=check
wavec --debug-wave=ast build main.wave --emit=check
```

`--debug-wave` can expose selected stages such as lexer tokens, AST, or IR. For ordinary source errors, read the first actionable diagnostic and its source location before reaching for internal dumps.

## Separate common failure classes

1. **Parsing/type errors** also fail under `--emit=check`.
2. **Import errors** require checking `std-path`, `--dep-root`, `--dep`, and the actual filesystem layout.
3. **Link errors** require checking `--link`, `-L`, the target ABI, and symbol names.
4. **Runtime errors** occur after a successful build and should be separated by exit status and runtime environment.
5. **FFI errors** require rechecking widths, string representation, pointer lifetime, and calling convention against the native declaration.

## A useful bug report

Include:

- `wavec --version`
- Host OS and target triple
- The complete command you ran
- A minimal `.wave` source that reproduces the issue
- Full diagnostic output
- Expected and actual behavior

Remove unrelated secrets, tokens, and private paths before posting logs publicly.
