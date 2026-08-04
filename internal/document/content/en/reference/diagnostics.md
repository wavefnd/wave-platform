---
translation_set_id: diagnostics
path: reference/diagnostics
locale: en
group: reference
group_order: 3
order: 2
title: Diagnostics and troubleshooting
summary: Interpret compiler failures and reduce a program to a useful report.
---

## Diagnostic workflow

- Read the first diagnostic and its source span before secondary errors.
- Confirm that wavec --version matches the documentation version.
- Reduce the failure to the smallest complete source file.
- Use --emit=check to separate frontend validation from linking and execution.
- For FFI failures, verify the symbol, ABI, library path, and target architecture independently.

## Useful bug reports

Include the exact compiler version, operating system, target, command, minimal source, complete diagnostic text, expected result, and actual result. Remove secrets and private paths.

