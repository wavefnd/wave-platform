# Wave documentation sources

This directory is the canonical source for the documentation published at
`/docs`. Documentation locales are independent from the site interface.

Supported locales are `en`, `ko`, `ja`, `zh`, `es`, `de`, `ru`, `id`, and
`vi`. The `zh` locale is Simplified Chinese. The `id` locale covers the
Indonesian–Malay translation; a separate `ms` locale is intentionally not
published.

English is the fallback when a page has not been translated yet. A fallback
page is displayed as English and is never labelled as a completed translation.

Documentation describes the current compiler contract without attaching the
whole manual to a product version. Local declarations use `var`; removed
`let` and `let mut` syntax must not be added to examples.
