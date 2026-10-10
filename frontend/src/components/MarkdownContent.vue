<script setup lang="ts">
import DOMPurify from 'dompurify'
import { computed, nextTick, onMounted, ref, shallowRef, useId, watch } from 'vue'

import { grammarForLanguage, normalizeFenceLanguage, Prism } from './source/syntax'
import { highlightWave } from '../services/waveSyntax'
import { renderDocumentMarkdown, type DocumentExample } from '../services/documentMarkdown'
import WavePlayground from './WavePlayground.vue'

defineOptions({ inheritAttrs: false })
const props = defineProps<{ source: string; repository?: string; path?: string; reference?: string; runnable?: boolean }>()
const article = ref<HTMLElement | null>(null)
const slotPrefix = `wave-example-${useId()}`
const targets = shallowRef<Array<{ element: HTMLElement; example: DocumentExample }>>([])

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
  const result = renderDocumentMarkdown(props.source, {
    playgrounds: props.runnable ?? false,
    slotPrefix,
    renderCode(text, lang) {
      const language = normalizeFenceLanguage(lang)
      const grammar = grammarForLanguage(language)
      const content = language === 'wave' ? highlightWave(text) : grammar ? Prism.highlight(text, grammar, language) : String(Prism.util.encode(text))
      const languageClass = language ? ` class="language-${language}"` : ''
      return `<pre><code${languageClass}>${content}\n</code></pre>`
    },
  })
  return { ...result, html: rewriteRelativeLinks(DOMPurify.sanitize(result.html, { USE_PROFILES: { html: true } })) }
})

let generation = 0
async function mountExamples() {
  const current = ++generation
  const result = rendered.value
  targets.value = []
  await nextTick()
  if (current !== generation || !article.value) return
  const placeholders = Array.from(article.value.querySelectorAll<HTMLElement>('[data-wave-playground]'))
  targets.value = result.slots.flatMap((example, index) => {
    const element = placeholders.find(item => item.dataset.wavePlayground === `${slotPrefix}-${index}`)
    return element ? [{ element, example }] : []
  })
}
onMounted(mountExamples)
watch(rendered, mountExamples, { flush: 'post' })
</script>

<template>
  <article ref="article" v-bind="$attrs" class="markdown-content markdown-body" v-html="rendered.html" />
  <Teleport v-for="(target, index) in targets" :key="index" :to="target.element">
    <WavePlayground v-bind="target.example" embedded />
  </Teleport>
</template>
