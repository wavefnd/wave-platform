<script setup lang="ts">
import { computed } from 'vue'
import type { Token as PrismToken, TokenStream } from 'prismjs'

import type { SourceBlob } from './types'
import { grammarForLanguage, languageForPath, Prism } from './syntax'

type Token = NonNullable<SourceBlob['waveHighlight']>['tokens'][number]
type Segment = { classes: string[]; text: string }

const props = defineProps<{ content: string; path: string; tokens?: Token[] }>()

function flattenPrism(stream: TokenStream, inherited: string[] = []): Segment[] {
  const result: Segment[] = []
  const values = Array.isArray(stream) ? stream : [stream]
  for (const value of values) {
    if (typeof value === 'string') {
      if (value) result.push({ classes: inherited, text: value })
      continue
    }
    const token = value as PrismToken
    const aliases = Array.isArray(token.alias) ? token.alias : token.alias ? [token.alias] : []
    result.push(...flattenPrism(token.content, ['token', token.type, ...aliases]))
  }
  return result
}

function waveSegments(content: string, tokens: Token[]): Segment[] {
  const bytes = new TextEncoder().encode(content)
  const decoder = new TextDecoder()
  const segments: Segment[] = []
  let cursor = 0

  for (const token of [...tokens].sort((left, right) => left.start - right.start)) {
    const start = Math.max(cursor, Math.min(token.start, bytes.length))
    const end = Math.max(start, Math.min(token.end, bytes.length))
    if (start > cursor) segments.push({ classes: [], text: decoder.decode(bytes.slice(cursor, start)) })
    if (end > start) segments.push({ classes: [`source-token-${token.kind}`], text: decoder.decode(bytes.slice(start, end)) })
    cursor = end
  }
  if (cursor < bytes.length) segments.push({ classes: [], text: decoder.decode(bytes.slice(cursor)) })
  return segments
}

const lines = computed<Segment[][]>(() => {
  let segments: Segment[]
  if (props.tokens?.length) {
    segments = waveSegments(props.content, props.tokens)
  } else {
    const language = languageForPath(props.path)
    const grammar = grammarForLanguage(language)
    segments = grammar ? flattenPrism(Prism.tokenize(props.content, grammar)) : [{ classes: [], text: props.content }]
  }
  if (!segments.length) segments = [{ classes: [], text: props.content }]

  const result: Segment[][] = [[]]
  for (const segment of segments) {
    const parts = segment.text.split('\n')
    parts.forEach((text, index) => {
      if (text) result.at(-1)?.push({ classes: segment.classes, text })
      if (index < parts.length - 1) result.push([])
    })
  }
  return result
})
</script>

<template>
  <pre><code><span v-for="(line, lineIndex) in lines" :id="`L${lineIndex + 1}`" :key="lineIndex"><a :href="`#L${lineIndex + 1}`" :aria-label="`Line ${lineIndex + 1}`">{{ lineIndex + 1 }}</a><span class="source-code-line"><span
      v-for="(segment, segmentIndex) in line"
      :key="segmentIndex"
      :class="segment.classes"
    >{{ segment.text }}</span></span>
</span></code></pre>
</template>
