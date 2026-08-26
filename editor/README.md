# WaveEditor

WaveEditor is a portable editor toolkit maintained by the Wave project. It is
designed to work outside Wave Platform: the package exposes a framework-neutral
document engine, a plain DOM adapter, an optional Vue 3 component, and the Wave
source package that defines the native document contract.

WaveEditor does not own persistence, accounts, uploads, Markdown rendering, or
authorization. Applications provide those boundaries and can replace the
default JavaScript transform adapter with a native Wave, HTTP, or future
WebAssembly adapter.

## Package layout

```text
editor/
├── core/            # framework- and DOM-independent TypeScript API
├── dom/             # mountWaveEditor() for plain browser applications
├── vue/             # Vue 3 package entry
├── web/             # shared Vue view and CSS
├── src/lib.wave     # Vex library entry and public Wave API
├── wave/main.wave   # production C ABI adapter for Wave Platform
├── tests/           # public core contract tests
├── vex.ws            # Vex library manifest
└── package.json      # ESM package and type declaration contract
```

## Web package

The package is not coupled to the Wave Platform frontend. Build and inspect the
same artifact that an external consumer receives:

```sh
cd editor
npm ci
npm test
npm run build
npm pack --dry-run
```

Install the resulting package archive in any ESM project without copying
Wave Platform source files:

```sh
npm install /path/to/wavefnd-editor-0.2.0.tgz
```

The public entries resolve only to built JavaScript and declaration files. The
package also carries the `source/*` exports for Wave Platform's monorepo
development build; external consumers should use the built entries above.

### Framework-independent core

```ts
import {
  WaveEditorDocument,
  WaveEditorEngine,
  analyzeDocument,
  transformSelection,
} from '@wavefnd/editor'

const metrics = analyzeDocument('Wave editor')
const transformed = transformSelection('Wave editor', 5, 11, 'bold')

const document = new WaveEditorDocument('Wave', 0, 4)
await document.apply('italic', new WaveEditorEngine())
document.undo()
```

Browser selections use UTF-16 offsets because they map directly to
`HTMLTextAreaElement.selectionStart` and `selectionEnd`. Conversion helpers are
exported for Wave/native adapters, whose public contract uses Unicode character
offsets.

### Plain DOM

```ts
import { mountWaveEditor } from '@wavefnd/editor/dom'
import '@wavefnd/editor/style.css'

const editor = mountWaveEditor(document.querySelector('#editor')!, {
  value: '# Hello, Wave',
  label: 'Document',
  onChange(value) { console.log(value) },
  onSave(value) { save(value) },
})

editor.focus()
editor.setValue('# Updated')
editor.destroy()
```

### Vue 3

```vue
<script setup lang="ts">
import { ref } from 'vue'
import { WaveEditor } from '@wavefnd/editor/vue'
import '@wavefnd/editor/style.css'

const content = ref('# Hello, Wave')
</script>

<template>
  <WaveEditor v-model="content" label="Document" @save="save(content)" />
</template>
```

### Wave adapter

Every UI accepts a `transform` adapter. This keeps the UI independent from the
runtime that executes Wave:

```ts
import type { WaveEditorTransform } from '@wavefnd/editor'

const transform: WaveEditorTransform = async (content, start, end, command) => {
  const response = await fetch('/editor/transform', {
    method: 'POST',
    body: JSON.stringify({ content, start, end, command }),
  })
  return response.json()
}
```

Adapters must return the transformed content, the new browser selection,
document metrics, and an engine identifier. Wave Platform uses its authenticated
XML endpoint backed by the native Wave module; a future WebAssembly adapter can
implement the same interface without changing consumers.

## Wave and Vex package

`vex.ws` declares a library package named `wave_editor`. The canonical entry is
`src/lib.wave` and exposes the ABI version, document analysis, and selection
transform functions:

```wson
{
    name = "my_editor_host",
    version = 0.1.0,
    dependencies = [
        { name = "wave_editor", path = "../WaveEditor" }
    ]
}
```

```wave
import("wave_editor")::{editor_abi_version, analyze_document, transform_document};
```

The current tagged `wavec 0.2.0-pre-beta` binary predates public module
visibility even though the latest Wave source and Vex contract support `pub`.
Validate the Vex library with a compiler build that includes the public module
contract:

```sh
VEX_WAVEC=/path/to/current/wavec vex check
```

Wave Platform keeps `wave/main.wave` as an ABI-compatible production adapter so
the running site remains buildable with the tagged compiler. The two entry
points share ABI version 1 and are covered by native integration tests.

Vex Git dependencies require `vex.ws` at the dependency repository root. Until
WaveEditor is split into its own repository, use the extracted `editor/`
directory as a path dependency. The directory is deliberately self-contained
so it can later be moved or mirrored without platform code.

## Stable contract

- Markdown commands: bold, italic, inline code, heading, quote, unordered list,
  and link.
- Unicode-safe metrics and explicit browser/native offset conversion.
- Bounded undo and redo history in the framework-independent document model.
- Native C ABI version 1 for Wave Platform.
- No implicit network requests, telemetry, persistence, HTML rendering, or
  image upload behavior.
- Vue remains an optional peer dependency; core and DOM consumers do not need
  Vue at runtime.

WaveEditor is licensed under the repository's MPL-2.0 license.
