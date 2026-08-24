# Editing language documentation

The editable source for the official Wave documentation is stored in the root
`wavedoc/{locale}` module. Documentation supports `en`, `ko`, `ja`, `zh`, `es`,
`de`, `ru`, `id`, and `vi`. The `zh` locale is Simplified Chinese, and `id`
covers the Indonesian–Malay documentation without a separate `ms` locale.

```text
wavedoc/
├── en/language/explicit-memory-type-model.md
└── ko/language/explicit-memory-type-model.md
```

Each file contains front matter followed by Markdown:

````markdown
---
translation_set_id: memory-model
path: language/explicit-memory-type-model
locale: en
group: language
group_order: 2
order: 7
title: Wave Explicit Memory Type Model
summary: Pointer types and explicit memory access in Wave.
---

## Pointer types

`ptr<T>` is dedicated syntax in the Wave Explicit Memory Type Model. It is not a general-purpose generic type.

```wave
var address: ptr<i32> = raw as ptr<i32>;
```
````

Use `##` for the first heading because the page title is rendered from the front matter. GFM tables, lists, block quotes, links, and fenced code blocks are supported.

Keep `path`, `group`, and ordering fields consistent between translations. Use the same `translation_set_id` for pages that represent the same document in different languages.

English is the explicit fallback for a path that has not been translated. Do
not copy English text into another locale merely to make the translation look
complete. Documentation describes the current compiler contract without a
manual-wide Wave version label. Local variables use `var`; `let` and `let mut`
are removed syntax.

The server imports these files at startup and stores published revisions in the platform database. Do not edit generated XML or database values by hand.
