<script setup lang="ts">
import DOMPurify from 'dompurify'
import { Marked, Renderer } from 'marked'
import { gfmHeadingId } from 'marked-gfm-heading-id'
import { computed } from 'vue'

import { grammarForLanguage, normalizeFenceLanguage, Prism } from './source/syntax'

const props = defineProps<{ source: string; repository?: string; path?: string; reference?: string }>()

function rewriteRelativeLinks(html: string) {
  if (!props.repository || typeof document === 'undefined') return html

  const template = document.createElement('template')
  template.innerHTML = html
  const baseParts = (props.path ?? '').split('/').slice(0, -1)

  const resolvePath = (value: string) => {
    const resolved: string[] = [...baseParts]
    for (const part of value.split('/')) {
      if (!part || part === '.') continue
      if (part === '..') resolved.pop()
      else resolved.push(part)
    }
    return resolved.join('/')
  }

  template.content.querySelectorAll<HTMLAnchorElement>('a[href]').forEach((anchor) => {
    const href = anchor.getAttribute('href') ?? ''
    if (!href || href.startsWith('#') || /^[a-z][a-z\d+.-]*:/i.test(href) || href.startsWith('//')) return

    const [relativePath, fragment] = href.split('#', 2)
    const query = new URLSearchParams({ path: resolvePath(relativePath) })
    if (props.reference) query.set('ref', props.reference)
    anchor.href = `/source/${encodeURIComponent(props.repository ?? '')}?${query.toString()}${fragment ? `#${fragment}` : ''}`
  })

  template.content.querySelectorAll<HTMLImageElement>('img[src]').forEach((image) => {
    const source = image.getAttribute('src') ?? ''
    if (!source || source.startsWith('data:') || /^[a-z][a-z\d+.-]*:/i.test(source) || source.startsWith('//')) return
    const query = new URLSearchParams({ path: resolvePath(source) })
    if (props.reference) query.set('ref', props.reference)
    image.src = `/api/v1/source/repositories/${encodeURIComponent(props.repository ?? '')}/raw?${query.toString()}`
  })
  return template.innerHTML
}

const rendered = computed(() => {
  const renderer = new Renderer()
  renderer.code = ({ text, lang }) => {
    const language = normalizeFenceLanguage(lang)
    const grammar = grammarForLanguage(language)
    const content = grammar ? Prism.highlight(text, grammar, language) : String(Prism.util.encode(text))
    const languageClass = language ? ` class="language-${language}"` : ''
    return `<pre><code${languageClass}>${content}\n</code></pre>`
  }
  const parser = new Marked(gfmHeadingId(), { async: false, gfm: true, breaks: false, renderer })
  const html = parser.parse(props.source) as string
  const sanitized = DOMPurify.sanitize(html, { USE_PROFILES: { html: true } })
  return rewriteRelativeLinks(sanitized)
})
</script>

<template>
  <article class="markdown-content markdown-body" v-html="rendered" />
</template>
