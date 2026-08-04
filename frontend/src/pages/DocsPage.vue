<script setup lang="ts">
import { Search } from '@lucide/vue'
import { computed, ref, watch, watchEffect } from 'vue'
import { useRoute } from 'vue-router'

import MarkdownContent from '../components/MarkdownContent.vue'
import { useI18n } from '../i18n'
import { getDocument, getDocuments, type DocumentSummary, type DocumentView } from '../services/http'
import { applyPageSEO } from '../services/seo'
import UiInlineState from '../ui/UiInlineState.vue'
import UiSkeletonRows from '../ui/UiSkeletonRows.vue'

const route = useRoute()
const { locale, t } = useI18n()
const documents = ref<DocumentSummary[]>([])
const document = ref<DocumentView | null>(null)
const query = ref('')
const loading = ref(true)
const failed = ref(false)

const currentPath = computed(() => {
  const value = route.params.pathMatch
  return Array.isArray(value) ? value.join('/') : String(value ?? '')
})
const filtered = computed(() => {
  const needle = query.value.trim().toLocaleLowerCase(locale.value)
  if (!needle) return documents.value
  return documents.value.filter((item) => `${item.title} ${item.summary}`.toLocaleLowerCase(locale.value).includes(needle))
})
const groups = computed(() => {
  const result = new Map<string, DocumentSummary[]>()
  for (const item of filtered.value) result.set(item.group, [...(result.get(item.group) ?? []), item])
  return Array.from(result, ([id, items]) => ({ id, title: groupName(id), items }))
})
const headings = computed(() => document.value?.blocks.filter((block) => block.kind === 'heading' && block.anchor) ?? [])
const currentIndex = computed(() => documents.value.findIndex((item) => item.path === currentPath.value))
const previous = computed(() => currentIndex.value > 0 ? documents.value[currentIndex.value - 1] : null)
const next = computed(() => currentIndex.value >= 0 && currentIndex.value < documents.value.length - 1 ? documents.value[currentIndex.value + 1] : null)

function groupName(group: string) {
  const names: Record<string, string> = {
    'getting-started': t('docs.gettingStarted'), language: t('docs.language'),
    reference: t('docs.reference'), toolchain: t('docs.tools'),
  }
  return names[group] ?? group
}

async function load() {
  loading.value = true
  failed.value = false
  try {
    documents.value = await getDocuments(locale.value)
    document.value = currentPath.value ? await getDocument(currentPath.value, locale.value) : null
  } catch {
    failed.value = true
    document.value = null
  } finally {
    loading.value = false
  }
}

watch([currentPath, locale], load, { immediate: true })
watchEffect(() => {
  if (!document.value) return
  applyPageSEO({
    title: `${document.value.title} · Wave Documentation`,
    description: document.value.summary,
    locale: locale.value,
    path: route.path,
    schema: {
      '@type': 'TechArticle',
      headline: document.value.title,
      abstract: document.value.summary,
      dateModified: document.value.updatedAt,
      version: document.value.version,
    },
  })
})
</script>

<template>
  <main class="docs-service">
    <header class="docs-service-header">
      <div class="docs-width docs-service-header-inner">
        <div><h1>{{ t('docs.title') }}</h1><span>v{{ document?.version ?? documents[0]?.version ?? '0.2.0-pre-beta' }}</span></div>
        <label class="docs-search"><Search :size="16" aria-hidden="true" /><input v-model="query" :placeholder="t('docs.search')" /></label>
      </div>
    </header>

    <div v-if="loading" class="docs-width docs-loading"><UiSkeletonRows :rows="8" /></div>
    <div v-else-if="failed" class="docs-width docs-loading"><UiInlineState :message="t('common.loadError')" :action="t('common.retry')" @action="load" /></div>

    <div v-else-if="document" class="docs-width docs-reader-layout">
      <aside class="docs-tree" :aria-label="t('docs.navigation')">
        <section v-for="group in groups" :key="group.id">
          <strong>{{ group.title }}</strong>
          <RouterLink v-for="item in group.items" :key="item.path" :to="`/docs/${item.path}`" :class="{ active: item.path === document.path }">{{ item.title }}</RouterLink>
        </section>
      </aside>

      <article class="document-page">
        <nav class="document-breadcrumb"><RouterLink to="/docs">{{ t('docs.title') }}</RouterLink><span>/</span><span>{{ groupName(document.group) }}</span></nav>
        <header><h1>{{ document.title }}</h1><p>{{ document.summary }}</p><div><span>v{{ document.version }}</span></div></header>
        <div class="document-content"><MarkdownContent :source="document.markdown" /></div>
        <nav class="document-pagination">
          <RouterLink v-if="previous" :to="`/docs/${previous.path}`"><small>{{ t('docs.previous') }}</small><span>← {{ previous.title }}</span></RouterLink><span v-else />
          <RouterLink v-if="next" :to="`/docs/${next.path}`"><small>{{ t('docs.next') }}</small><span>{{ next.title }} →</span></RouterLink>
        </nav>
      </article>

      <aside class="docs-toc" :aria-label="t('docs.contents')"><strong>{{ t('docs.contents') }}</strong><a v-for="heading in headings" :key="heading.anchor" :href="`#${heading.anchor}`" :class="`level-${heading.level}`">{{ heading.text }}</a></aside>
    </div>

    <div v-else class="docs-width docs-catalog-page">
      <header class="docs-titlebar"><h1>{{ t('docs.title') }}</h1><p>{{ t('docs.lead') }}</p></header>
      <div class="docs-catalog">
        <section v-for="group in groups" :key="group.id" class="docs-catalog-group">
          <h2>{{ group.title }}</h2>
          <ul><li v-for="item in group.items" :key="item.path"><RouterLink :to="`/docs/${item.path}`"><strong>{{ item.title }}</strong><span>{{ item.summary }}</span></RouterLink></li></ul>
        </section>
      </div>
      <p v-if="groups.length === 0" class="docs-empty">{{ t('docs.noResults') }}</p>
    </div>
  </main>
</template>
