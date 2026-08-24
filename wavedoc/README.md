# Wave documentation sources

This directory is the canonical source for the documentation published at
`/docs`. Documentation locales are independent from the site interface.

Supported locales are `en`, `ko`, `ja`, `zh`, `es`, `de`, `ru`, `id`, and
`vi`. The `zh` locale is Simplified Chinese. The `id` locale covers the
Indonesian–Malay translation; a separate `ms` locale is intentionally not
published.

English is the fallback when a page has not been translated yet. A fallback
page is displayed as English and is never labelled as a completed translation.

Documentation describes the Wave language without attaching the manual to a
product version. Local variables are declared with `var`.

Authoring rules:

- State the language, library, and tool contracts directly from a user's point
  of view.
- Do not narrate compiler internals, feature-development history, or temporary
  development notes in the language manual.
- Keep examples valid Wave source and explain unfamiliar behavior in plain
  language before introducing edge cases.
- Treat the legacy website documentation as background material, not as an
  authoritative syntax reference.
