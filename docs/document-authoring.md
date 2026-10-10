# Editing language documentation

The editable source for the official Wave documentation is stored in the root
`wavedoc/{locale}` module. Documentation supports `en`, `ko`, `ja`, `zh`, `es`,
`de`, `ru`, `id`, `vi`, `pt`, `fr`, `pl`, `nl`, `tr`, and `it`. The `zh` locale is Simplified Chinese, and `id`
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

## Wave, standard library, and Whale navigation

The documentation header provides separate Wave, standard library, and Whale navigation. A document
belongs to Whale when its front-matter `path` starts with `whale/`; `stdlib/` paths belong to the standard library. The existing
`reference/standard-library`, `reference/string-and-bytes`,
`reference/memory-and-buffer`, and `reference/system-io-network-process` paths
also belong to the standard library, preserving their published URLs. Other
paths remain in Wave. Group names and titles do not determine the project.
Toolchain documents live under `whale/`. Retired URLs are mapped in `wavedoc/redirects.json` and redirect to the canonical document.

Add Whale Markdown files directly to `wavedoc/{locale}/whale/`:

```yaml
---
translation_set_id: whale-symbols
path: whale/symbols
locale: ko
group: whale
group_order: 5
order: 4
title: 심볼과 외부 연결 이름
summary: 함수와 전역 변수의 식별 및 외부 연결 이름 규칙입니다.
---
```

The example produces `/docs/ko/whale/symbols`. The Whale catalogue is
`/docs/ko/whale`, while the Wave catalogue stays at `/docs/ko`. Use the same path
and translation set in English at `wavedoc/en/whale/symbols.md`.

Sidebars, search, and previous/next links stay inside the selected project.
Language changes retain the project and document path. Missing translations use
the English page with the existing fallback notice; a project without documents
has an empty catalogue. Navigation is sorted by `group_order`, then `order`, then
`path`, after translations replace their corresponding English entries.

No per-document route or frontend registry is needed. Markdown is embedded in
the Go binary, so **rebuild and restart the application after adding or editing
files** (for Docker, `docker compose up --build -d`). The server imports the
embedded files at startup and stores published revisions in the platform
database. Do not edit generated XML or database values by hand.


## Language lessons and examples

Korean language lessons and detailed rules share `ko/language/`, with complete projects in `ko/practice/`. Do not create a separate `learn` section.
Use learning objectives, a runnable example, expected output, explanation, an
exercise, and its answer. Reference pages describe exact contracts; link to them
from lessons rather than assuming the reader will inspect compiler sources.
Standard library pages must identify imports, units, return/error behavior,
ownership, platform requirements, and a complete example where practical.

Mark validated complete examples with `<!-- wave-example: unique-name -->`
immediately before their Wave fence and add the same ID to `wavedoc/examples.json`.
The manifest supplies input, expected output, expected status, and local fixture
files. `check` is for examples requiring an external service; it does not certify
runtime behavior. `reject` examples must specify the expected diagnostic fragment.
Unmarked fences must say whether they are snippets, declarations, or deliberately
invalid examples. Do not label incomplete snippets as runnable programs.

```shell
python3 tools/check-doc-translations.py
python3 tools/check-doc-examples.py --links-only
python3 tools/check-doc-examples.py --compiler /path/to/wavec --std-root /matching/std
```

The checker writes builds and program files to temporary directories, verifies
internal document links in every locale, and never uses an implicit installed std. Run it
on a native host with the OS facilities required by the selected examples. Use
`--case example-name` to check a subset. Node navigation tests and Go SEO tests
cover tab isolation, legacy URLs, translation fallback, and server-rendered pages.

## Readable teaching examples

Use one statement per line, four-space indentation, and multiline function and
control-flow bodies. Separate declarations, checks, computation, and cleanup
with blank lines. Split long calls and imports at meaningful boundaries.
Avoid compact one-line code even when the compiler accepts it.

Installation pages describe installing Wave and running the first program.
Keep implementation history and backend comparisons out of learning material.
Teach each concept with progressive examples, explanations of results, boundary
cases, and exercises with worked answers. Do not replace explanation with a
signature list or repeat generic cautions to increase page length.

## Keeping translations complete

The Korean course is the source for corresponding pages in all other locales.
The six new locales may be translated one document at a time; existing complete
locales retain full coverage. See [translation contribution rules](translation-contributing.md)
for language codes, writing systems, and the single-document workflow. Translate the full explanations, tables, exercises, and worked answers.
Keep fenced programs, commands, input, and output unchanged, including indentation
and blank lines. Preserve API names and language syntax in prose. Use localised
links so that readers stay in their selected language.

Keep the same example IDs in translations; they refer to the canonical Korean
examples and must not create duplicate manifest entries. The example checker
verifies translated code against that source and executes each canonical case
once. The translation checker compares page coverage, navigation metadata,
headings, lists, tables, code fences, and internal links. It also detects leftover
translation placeholders and untranslated Korean prose, while allowing the
Hangul literals used in Unicode examples.

Automated checks do not establish linguistic quality. Review translated prose
for terminology, negation, boundary values, ownership, and error semantics. Do
not remove an explanation or silently change an example to make a check pass.

## Embedded playgrounds

Use `wave playground` for a complete program that can run in the browser. It
renders an editable, highlighted program with Run, Stop, Reset, and output in
place. Opening a document never compiles or runs its examples automatically.
Ordinary `wave` fences remain read-only highlighted code, including fragments,
invalid examples, and programs needing native OS facilities.

````markdown
<!-- wave-example: book-hello -->
```wave playground
fun main() {
    println("Hello, Wave!");
}
```
````

The fence works without metadata, but official documentation must keep its
adjacent `wave-example` marker and set `"playground": true` on the matching entry
in `wavedoc/examples.json`. The existing `stdin`, `stdout`, and `exit` fields
supply initial input and the collapsible expected output. Reset restores the
original source and input and clears the previous execution result. The same ID
in translations reuses this metadata; keep the fence modifier in translations.

Enable only single-file programs that fit the browser execution limits and do
not require files, sockets, external services, or native-only imports. The live
playground CI compiles each enabled canonical example to Wasm and checks its
output and exit code, including alternate `runs` from the manifest:

```shell
python3 tools/test-doc-playgrounds.py
python3 tools/check-doc-examples.py --links-only
cd frontend
npm run test:docs
node --experimental-strip-types --test tests/playground-live.test.ts
```

The live test uses the running local service at `http://localhost:8080` unless
`PLAYGROUND_URL` selects another instance. The document checker also rejects
unregistered playground fences and cases that require file fixtures or have
`check`/`reject` modes. Add new examples only after the live test passes.
