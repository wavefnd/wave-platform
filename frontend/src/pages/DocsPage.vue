<script setup lang="ts">
import { Search } from '@lucide/vue'
import { computed, ref, watch, watchEffect } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import MarkdownContent from '../components/MarkdownContent.vue'
import { useI18n } from '../i18n'
import { documentLocales, isDocumentLocale, saveDocumentLocale } from '../services/documentLocale'
import { documentationCatalog, documentationPath, documentationProject, projectDocuments } from '../services/documentNavigation'
import { HTTPError, getDocument, getDocuments, type DocumentLocale, type DocumentSummary, type DocumentView } from '../services/http'
import { applyPageSEO } from '../services/seo'
import UiInlineState from '../ui/UiInlineState.vue'
import UiSkeletonRows from '../ui/UiSkeletonRows.vue'

const route = useRoute()
const router = useRouter()
const { locale, t } = useI18n()
const documents = ref<DocumentSummary[]>([])
const document = ref<DocumentView | null>(null)
const query = ref('')
const loading = ref(true)
const failed = ref(false)
const showingEnglishFallback = ref(false)
const hasEnglishFallbackItems = computed(() => docLocale.value !== 'en' && documents.value.some(item => item.locale === 'en'))

const docLocale = computed<DocumentLocale>(() => isDocumentLocale(route.params.docLocale) ? route.params.docLocale : 'en')
const docBase = computed(() => `/docs/${docLocale.value}`)

const requestedPath = computed(() => {
  if (route.meta.documentationProject === 'whale' || route.meta.documentationProject === 'stdlib') return route.meta.documentationProject
  const value = route.params.pathMatch
  return Array.isArray(value) ? value.join('/') : String(value ?? '')
})
const currentPath = computed(() => documentationPath(requestedPath.value))
const project = computed(() => documentationProject(requestedPath.value))
function projectLabel(value: string) {
  return value === 'stdlib' ? t('docs.standardLibrary') : value === 'whale' ? 'Whale' : 'Wave'
}
const projectName = computed(() => projectLabel(project.value))
const catalogBase = computed(() => documentationCatalog(docLocale.value, project.value))
const filtered = computed(() => {
  const needle = query.value.trim().toLocaleLowerCase(docLocale.value)
  if (!needle) return documents.value
  return documents.value.filter((item) => `${item.title} ${item.summary}`.toLocaleLowerCase(docLocale.value).includes(needle))
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
    practice: t('docs.practice'), stdlib: t('docs.standardLibrary'),
    reference: t('docs.reference'), toolchain: t('docs.tools'), whale: 'Whale',
  }
  return names[group] ?? group
}

let loadGeneration = 0
async function load() {
  const generation = ++loadGeneration
  const requestedLocale = docLocale.value
  const path = currentPath.value
  const selectedProject = project.value
  loading.value = true
  document.value = null
  failed.value = false
  showingEnglishFallback.value = false
  try {
    saveDocumentLocale(requestedLocale)
    const [translated, english] = await Promise.all([
      getDocuments(requestedLocale), requestedLocale === 'en' ? Promise.resolve([]) : getDocuments('en'),
    ])
    let loadedDocument: DocumentView | null = null
    let fallback = false
    if (path) {
      try {
        loadedDocument = await getDocument(path, requestedLocale)
      } catch (error) {
        if (requestedLocale === 'en' || !(error instanceof HTTPError) || error.status !== 404) throw error
        loadedDocument = await getDocument(path, 'en')
        fallback = true
      }
    }
    if (generation !== loadGeneration) return
    documents.value = projectDocuments(translated, english, selectedProject)
    document.value = loadedDocument
    showingEnglishFallback.value = fallback
  } catch {
    if (generation !== loadGeneration) return
    failed.value = true
    document.value = null
  } finally {
    if (generation === loadGeneration) loading.value = false
  }
}

watch([requestedPath, docLocale], () => { query.value = ''; void load() }, { immediate: true })

function changeDocumentLocale(event: Event) {
  const value = (event.target as HTMLSelectElement).value
  if (!isDocumentLocale(value)) return
  saveDocumentLocale(value)
  router.push(currentPath.value
    ? { name: 'document', params: { docLocale: value, pathMatch: currentPath.value.split('/') } }
    : { name: project.value === 'wave' ? 'docs-locale' : `docs-${project.value}`, params: { docLocale: value } })
}
watchEffect(() => {
  if (!document.value) {
    applyPageSEO({
      title: failed.value ? 'Page not found · Wave' : `${projectName.value} Documentation · Wave`,
      description: failed.value ? 'The requested document could not be loaded.' : project.value === 'stdlib' ? t('docs.stdlibLead') : project.value === 'whale' ? t('docs.whaleLead') : t('docs.lead'),
      locale: docLocale.value, path: route.path, noIndex: failed.value,
      schema: { '@type': 'CollectionPage' },
      alternates: !currentPath.value ? [...documentLocales.map((item) => ({ locale: item.id, path: documentationCatalog(item.id, project.value) })), { locale: 'x-default', path: documentationCatalog('en', project.value) }] : [],
    })
    return
  }
  applyPageSEO({
    title: `${document.value.title} · ${projectName.value} Documentation`,
    description: document.value.summary,
    locale: document.value.locale,
    alternates: [
      ...document.value.translations.map((item) => ({ locale: item.locale, path: `/docs/${item.locale}/${item.path}` })),
      ...(document.value.translations.some((item) => item.locale === 'en') ? [{ locale: 'x-default', path: `/docs/en/${currentPath.value}` }] : []),
    ],
    path: showingEnglishFallback.value ? `/docs/en/${currentPath.value}` : route.path,
    breadcrumbs: [
      { name: 'Home', path: '/' },
      { name: `${projectName.value} Documentation`, path: documentationCatalog(document.value.locale, project.value) },
      { name: document.value.title, path: showingEnglishFallback.value ? `/docs/en/${currentPath.value}` : route.path },
    ],
    schema: {
      '@type': 'TechArticle',
      headline: document.value.title,
      abstract: document.value.summary,
      dateModified: document.value.updatedAt,
      articleSection: 'Documentation',
      isAccessibleForFree: true,
      author: { '@type': 'Organization', name: 'Wave Foundation' },
    },
  })
})
</script>

<template>
  <main class="docs-service">
    <header class="docs-service-header">
      <div class="docs-width docs-service-header-inner">
        <div><RouterLink class="docs-service-title" :to="catalogBase">{{ t('docs.title') }}</RouterLink></div>
        <label class="docs-locale-select">
          <span>{{ t('docs.languageSelector') }}</span>
          <select :value="docLocale" @change="changeDocumentLocale">
            <option v-for="option in documentLocales" :key="option.id" :value="option.id">{{ option.label }}</option>
          </select>
        </label>
        <label class="docs-search"><Search :size="16" aria-hidden="true" /><input v-model="query" :aria-label="t('docs.search')" :placeholder="t('docs.search')" /></label>
      </div>
      <nav class="docs-width docs-project-tabs" :aria-label="t('docs.projectSelector')">
        <RouterLink v-for="item in (['wave', 'stdlib', 'whale'] as const)" :key="item" :to="documentationCatalog(docLocale, item)" :class="{ active: project === item }" :aria-current="project === item ? 'page' : undefined">{{ projectLabel(item) }}</RouterLink>
      </nav>
    </header>

    <div v-if="loading" class="docs-width docs-loading"><UiSkeletonRows :rows="8" /></div>
    <div v-else-if="failed" class="docs-width docs-loading"><UiInlineState :message="t('common.loadError')" :action="t('common.retry')" @action="load" /></div>

    <div v-else-if="document" class="docs-width docs-reader-layout">
      <aside class="docs-tree docs-tree-desktop" :aria-label="t('docs.navigation')">
        <section v-for="group in groups" :key="group.id">
          <strong>{{ group.title }}</strong>
          <RouterLink v-for="item in group.items" :key="item.path" :to="`${docBase}/${item.path}`" :class="{ active: item.path === document.path }">{{ item.title }}</RouterLink>
        </section>
      </aside>

      <details class="docs-tree-mobile">
        <summary>{{ t('docs.navigation') }}</summary>
        <div>
          <section v-for="group in groups" :key="group.id">
            <strong>{{ group.title }}</strong>
            <RouterLink v-for="item in group.items" :key="item.path" :to="`${docBase}/${item.path}`" :class="{ active: item.path === document.path }">{{ item.title }}</RouterLink>
          </section>
        </div>
      </details>

      <article class="document-page">
        <nav class="document-breadcrumb"><RouterLink :to="catalogBase">{{ projectName }} {{ t('docs.title') }}</RouterLink><span>/</span><span>{{ groupName(document.group) }}</span></nav>
        <p v-if="showingEnglishFallback" class="docs-translation-notice" role="status">{{ t('docs.englishFallback') }}</p>
        <header><h1>{{ document.title }}</h1><p>{{ document.summary }}</p></header>
        <div class="document-content"><MarkdownContent :source="document.markdown" runnable /></div>
        <nav class="document-pagination">
          <RouterLink v-if="previous" :to="`${docBase}/${previous.path}`"><small>{{ t('docs.previous') }}</small><span>← {{ previous.title }}</span></RouterLink><span v-else />
          <RouterLink v-if="next" :to="`${docBase}/${next.path}`"><small>{{ t('docs.next') }}</small><span>{{ next.title }} →</span></RouterLink>
        </nav>
      </article>

      <aside class="docs-toc" :aria-label="t('docs.contents')"><strong>{{ t('docs.contents') }}</strong><a v-for="heading in headings" :key="heading.anchor" :href="`#${heading.anchor}`" :class="`level-${heading.level}`">{{ heading.text }}</a></aside>
    </div>

    <div v-else class="docs-width docs-catalog-page">
      <header class="docs-titlebar"><h1>{{ projectName }} {{ t('docs.title') }}</h1><p>{{ project === 'stdlib' ? t('docs.stdlibLead') : project === 'whale' ? t('docs.whaleLead') : t('docs.lead') }}</p></header>
      <p v-if="hasEnglishFallbackItems" class="docs-translation-notice" role="status">{{ t('docs.englishFallbackCatalog') }}</p>
      <div class="docs-catalog">
        <section v-for="group in groups" :key="group.id" class="docs-catalog-group">
          <h2>{{ group.title }}</h2>
          <ul><li v-for="item in group.items" :key="item.path"><RouterLink :to="`${docBase}/${item.path}`"><strong>{{ item.title }}</strong><span>{{ item.summary }}</span></RouterLink></li></ul>
        </section>
      </div>
      <p v-if="groups.length === 0" class="docs-empty" role="status">{{ t(documents.length === 0 ? 'docs.emptyProject' : 'docs.noResults') }}</p>
    </div>
  </main>
</template>
