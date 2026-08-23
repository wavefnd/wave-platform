# WaveEditor

WaveEditor is Wave Platform's reusable text-editing module. Its document engine is written in Wave, while `web/` is a thin Vue adapter. The platform talks to the engine through a versioned XML/HTTP contract so that a future WebAssembly build can replace the server adapter without changing editor consumers.

The package intentionally owns no blog, mail, community, or account persistence. Consumers keep their own authorization, validation, uploads, and save APIs.

## Current contract

- Markdown selection transforms: bold, italic, inline code, heading, quote, unordered list, and link.
- UTF-8 input at the native boundary and Unicode character offsets at the public API.
- Document line and word metrics.
- ABI version 1, exported from `wave/main.wave`.

The Go reference engine is a compatibility fallback when native Wave modules are disabled. Production builds load `libwave-editor.so` and report `engine=wave` in transform responses.
