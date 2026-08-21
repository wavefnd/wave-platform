<script setup lang="ts">
import { computed, ref, watch, watchEffect } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'

import CommitDiff from '../components/source/CommitDiff.vue'
import LanguageBreakdown from '../components/source/LanguageBreakdown.vue'
import HighlightedSource from '../components/source/HighlightedSource.vue'
import MarkdownContent from '../components/MarkdownContent.vue'
import type { SourceBlob, SourceCommit, SourceCommitDetail, SourceRefs, SourceRepository, SourceTree } from '../components/source/types'
import { useI18n } from '../i18n'
import { getSourceBlob, getSourceCommitDetail, getSourceCommits, getSourceRefs, getSourceRepositories, getSourceTree } from '../services/http'
import { applyPageSEO } from '../services/seo'
import UiInlineState from '../ui/UiInlineState.vue'
import UiSkeletonRows from '../ui/UiSkeletonRows.vue'

const route = useRoute()
const router = useRouter()
const { locale, t } = useI18n()

const repositories = ref<SourceRepository[]>([])
const tree = ref<SourceTree | null>(null)
const blob = ref<SourceBlob | null>(null)
const commits = ref<SourceCommit[]>([])
const commitDetail = ref<SourceCommitDetail | null>(null)
const refs = ref<SourceRefs>({ branches: [], tags: [] })
const search = ref('')
const loading = ref(true)
const error = ref('')

const repositoryID = computed(() => String(route.params.repository ?? ''))
const activeTab = computed(() => String(route.query.tab ?? 'code'))
const activePath = computed(() => String(route.query.path ?? ''))
const activeRef = computed(() => String(route.query.ref ?? ''))
const currentRepository = computed(() => repositories.value.find((repository) => repository.id === repositoryID.value) ?? tree.value?.repository)
const currentRef = computed(() => activeRef.value || `refs/heads/${currentRepository.value?.defaultBranch ?? 'master'}`)
const currentRefLabel = computed(() => currentRef.value.replace(/^refs\/(?:heads|tags)\//, ''))
const currentRefKind = computed(() => currentRef.value.startsWith('refs/tags/') ? t('source.tag') : currentRef.value.startsWith('refs/heads/') ? t('source.branch') : t('source.commit'))
const knownRef = computed(() => [...refs.value.branches.map((item) => `refs/heads/${item.name}`), ...refs.value.tags.map((item) => `refs/tags/${item.name}`)].includes(currentRef.value))
const filteredRepositories = computed(() => {
  const needle = search.value.trim().toLowerCase()
  if (!needle) return repositories.value
  return repositories.value.filter((repository) => `${repository.owner}/${repository.name} ${repository.description}`.toLowerCase().includes(needle))
})
const breadcrumbs = computed(() => activePath.value.split('/').filter(Boolean))

function formatDate(value?: string) {
  if (!value) return ''
  return new Intl.DateTimeFormat(locale.value === 'ko' ? 'ko-KR' : 'en-US', { year: 'numeric', month: 'short', day: 'numeric' }).format(new Date(value))
}

function pathThrough(index: number) {
  return breadcrumbs.value.slice(0, index + 1).join('/')
}

function codeLocation(path = '', kind: 'tree' | 'blob' = 'tree', reference = activeRef.value) {
  return {
    name: 'source-repository',
    params: { repository: repositoryID.value },
    query: { ...(path ? { path } : {}), ...(kind === 'blob' ? { kind } : {}), ...(reference ? { ref: reference } : {}) },
  }
}

function tabLocation(tab: string, path = '') {
  return { query: { tab, ...(path ? { path } : {}), ...(activeRef.value ? { ref: activeRef.value } : {}) } }
}

function commitLocation(oid: string) {
  return { query: { tab: 'commit', oid, ...(activeRef.value ? { ref: activeRef.value } : {}) } }
}

function switchReference(event: Event) {
  const reference = (event.target as HTMLSelectElement).value
  void router.push(codeLocation('', 'tree', reference))
}

function rawURL(path: string, reference = currentRef.value) {
  const query = new URLSearchParams({ path })
  if (reference) query.set('ref', reference)
  return `/api/v1/source/repositories/${encodeURIComponent(repositoryID.value)}/raw?${query.toString()}`
}

const binaryPreview = computed(() => {
  const extension = blob.value?.path.split('.').at(-1)?.toLowerCase() ?? ''
  if (['avif', 'gif', 'jpeg', 'jpg', 'png', 'svg', 'webp'].includes(extension)) return 'image'
  if (extension === 'pdf') return 'pdf'
  return 'file'
})

async function load() {
  loading.value = true
  error.value = ''
  tree.value = null
  blob.value = null
  commits.value = []
  commitDetail.value = null
  try {
    repositories.value = await getSourceRepositories()
    if (!repositoryID.value) return
    refs.value = await getSourceRefs(repositoryID.value)

    if (activeTab.value === 'commit') {
      commitDetail.value = await getSourceCommitDetail(repositoryID.value, String(route.query.oid ?? ''))
    } else if (activeTab.value === 'commits') {
      commits.value = await getSourceCommits(repositoryID.value, activePath.value, activeRef.value)
    } else if (activeTab.value === 'branches' || activeTab.value === 'tags') {
      return
    } else if (activePath.value && route.query.kind === 'blob') {
      blob.value = await getSourceBlob(repositoryID.value, activePath.value, activeRef.value)
    } else if (activePath.value) {
      try {
        tree.value = await getSourceTree(repositoryID.value, activePath.value, activeRef.value)
      } catch {
        blob.value = await getSourceBlob(repositoryID.value, activePath.value, activeRef.value)
      }
    } else {
      tree.value = await getSourceTree(repositoryID.value, '', activeRef.value)
    }
  } catch (cause) {
    error.value = cause instanceof Error ? cause.message : t('common.loadError')
  } finally {
    loading.value = false
  }
}

watch(() => route.fullPath, load, { immediate: true })
watchEffect(() => {
  if (!currentRepository.value) return
  const repository = currentRepository.value
  const pageName = activePath.value ? `${activePath.value} · ${repository.owner}/${repository.name}` : `${repository.owner}/${repository.name}`
  applyPageSEO({
    title: `${pageName} · Wave Source`,
    description: repository.description || `Read-only Git mirror of ${repository.owner}/${repository.name}.`,
    locale: locale.value,
    path: route.path,
    schema: {
      '@type': 'SoftwareSourceCode',
      name: `${repository.owner}/${repository.name}`,
      codeSampleType: 'full solution',
      ...(tree.value?.languages.length ? { programmingLanguage: tree.value.languages.map((language) => language.name) } : {}),
      version: currentRefLabel.value,
    },
  })
})
</script>

<template>
  <main class="source-forge">
    <header class="source-org-header">
      <div class="source-width source-org-row">
        <RouterLink class="source-org-name" to="/source">Wave Source</RouterLink>
        <nav :aria-label="t('source.sections')">
          <RouterLink to="/source" class="active">{{ t('source.repositories') }} <span>{{ repositories.length }}</span></RouterLink>
          <RouterLink to="/patches">{{ t('patches.title') }}</RouterLink>
        </nav>
      </div>
    </header>

    <section v-if="!repositoryID" class="source-width source-index">
      <header class="source-index-head">
        <div>
          <h1>{{ t('source.repositories') }}</h1>
          <p>{{ t('source.lead') }}</p>
        </div>
        <span>{{ repositories.length }}</span>
      </header>

      <label class="source-repository-search">
        <span class="sr-only">{{ t('source.searchRepositories') }}</span>
        <input v-model="search" type="search" :placeholder="t('source.searchRepositories')" />
      </label>

      <div class="source-repository-list">
        <UiSkeletonRows v-if="loading" :rows="5" />
        <UiInlineState v-else-if="error" :message="t('common.loadError')" :action="t('common.retry')" @action="load" />
        <article v-for="repository in filteredRepositories" v-else :key="repository.id">
          <div class="source-repository-copy">
            <div class="source-repository-title">
              <RouterLink :to="`/source/${repository.id}`">{{ repository.owner }} / <strong>{{ repository.name }}</strong></RouterLink>
              <span>{{ t('source.public') }}</span>
              <span>{{ t('source.mirror') }}</span>
            </div>
            <p>{{ repository.description }}</p>
            <small v-if="repository.headCommit">
              {{ repository.headCommit.author }} · {{ repository.headCommit.subject }} · {{ formatDate(repository.headCommit.authoredAt) }}
            </small>
          </div>
          <code v-if="repository.headCommit">{{ repository.headCommit.shortOid }}</code>
        </article>
      </div>
    </section>

    <section v-else class="source-repository-page">
      <header class="source-repo-heading source-width">
        <div>
          <RouterLink to="/source">{{ currentRepository?.owner ?? 'wavefnd' }}</RouterLink>
          <span>/</span>
          <h1>{{ currentRepository?.name ?? repositoryID }}</h1>
          <em>{{ t('source.public') }}</em>
          <em>{{ t('source.mirror') }}</em>
        </div>
      </header>

      <nav class="source-repo-tabs">
        <div class="source-width">
          <RouterLink :class="{ active: activeTab === 'code' }" :to="codeLocation()">{{ t('source.code') }}</RouterLink>
          <RouterLink :class="{ active: activeTab === 'commits' || activeTab === 'commit' }" :to="tabLocation('commits')">{{ t('source.commits') }}</RouterLink>
          <RouterLink :class="{ active: activeTab === 'branches' }" :to="tabLocation('branches')">{{ t('source.branches') }}</RouterLink>
          <RouterLink :class="{ active: activeTab === 'tags' }" :to="tabLocation('tags')">{{ t('source.tags') }}</RouterLink>
        </div>
      </nav>

      <div class="source-width source-repo-body">
        <UiSkeletonRows v-if="loading" :rows="6" />
        <UiInlineState v-else-if="error" :message="t('common.loadError')" :action="t('common.retry')" @action="load" />

        <template v-else-if="activeTab === 'code'">
          <div class="source-code-toolbar">
            <label class="source-ref-select">
              <span class="sr-only">{{ t('source.reference') }}</span>
              <select :value="currentRef" @change="switchReference">
                <option v-if="!knownRef" :value="currentRef">{{ currentRefLabel }}</option>
                <optgroup :label="t('source.branches')">
                  <option v-for="item in refs.branches" :key="`branch-${item.name}`" :value="`refs/heads/${item.name}`">{{ item.name }}</option>
                </optgroup>
                <optgroup :label="t('source.tags')">
                  <option v-for="item in refs.tags" :key="`tag-${item.name}`" :value="`refs/tags/${item.name}`">{{ item.name }}</option>
                </optgroup>
              </select>
            </label>
            <span>{{ currentRefKind }}</span>
          </div>

          <nav v-if="activePath" class="source-breadcrumbs" :aria-label="t('source.path')">
            <RouterLink :to="codeLocation()">{{ currentRepository?.name }}</RouterLink>
            <template v-for="(part, index) in breadcrumbs" :key="pathThrough(index)">
              <span>/</span>
              <RouterLink :to="codeLocation(pathThrough(index))">{{ part }}</RouterLink>
            </template>
          </nav>

          <div v-if="tree" class="source-code-layout">
            <div class="source-code-main">
              <section class="source-tree-panel">
                <header class="source-latest-commit">
                  <strong>{{ tree.commit.author }}</strong>
                  <span>{{ tree.commit.subject }}</span>
                  <code>{{ tree.commit.shortOid }}</code>
                </header>
                <RouterLink v-if="tree.path" class="source-tree-row source-tree-parent" :to="codeLocation(tree.path.split('/').slice(0, -1).join('/'))">
                  <span>..</span><small></small><time></time>
                </RouterLink>
                <RouterLink v-for="entry in tree.entries" :key="entry.oid + entry.path" class="source-tree-row" :to="codeLocation(entry.path, entry.type)">
                  <span><i :class="entry.type"></i>{{ entry.name }}</span>
                  <small>{{ entry.lastCommit?.subject }}</small>
                  <time>{{ formatDate(entry.lastCommit?.authoredAt) }}</time>
                </RouterLink>
              </section>

              <section v-if="tree.readme" class="source-readme-panel">
                <header>{{ tree.readme.path }}</header>
                <MarkdownContent
                  :source="tree.readme.content"
                  :repository="repositoryID"
                  :path="tree.readme.path"
                  :reference="tree.ref"
                />
              </section>
            </div>

            <aside v-if="!tree.path" class="source-about">
              <section>
                <h2>{{ t('source.about') }}</h2>
                <p>{{ tree.repository.description }}</p>
              </section>
              <section v-if="tree.languages.length">
                <LanguageBreakdown :languages="tree.languages" :title="t('source.languages')" subtitle="" primary-label="" />
              </section>
            </aside>
          </div>

          <section v-else-if="blob" class="source-blob-panel">
            <header>
              <strong>{{ blob.path.split('/').at(-1) }}</strong>
              <div>
                <RouterLink :to="tabLocation('commits', blob.path)">{{ t('source.history') }}</RouterLink>
                <a :href="rawURL(blob.path)" target="_blank" rel="noreferrer">Raw</a>
                <span>{{ blob.size.toLocaleString() }} bytes</span>
              </div>
            </header>
            <div v-if="blob.binary" class="source-binary-preview">
              <img v-if="binaryPreview === 'image'" :src="rawURL(blob.path)" :alt="blob.path.split('/').at(-1)" />
              <iframe v-else-if="binaryPreview === 'pdf'" :src="rawURL(blob.path)" :title="blob.path.split('/').at(-1)" />
              <a v-else :href="rawURL(blob.path)" target="_blank" rel="noreferrer">{{ t('source.openFile') }}</a>
            </div>
            <HighlightedSource v-else :content="blob.content" :path="blob.path" :tokens="blob.waveHighlight?.tokens" />
          </section>
        </template>

        <section v-else-if="activeTab === 'commit' && commitDetail" class="source-commit-detail">
          <RouterLink class="source-history-back" :to="tabLocation('commits', activePath)">← {{ t('source.commits') }}</RouterLink>
          <header>
            <h2>{{ commitDetail.commit.subject }}</h2>
            <p v-if="commitDetail.body">{{ commitDetail.body }}</p>
            <div>
              <span>{{ commitDetail.commit.author }} · {{ formatDate(commitDetail.commit.authoredAt) }}</span>
              <code>{{ commitDetail.commit.oid }}</code>
            </div>
          </header>
          <section class="source-changed-files">
            <h3>{{ commitDetail.files.length }} {{ t('source.changedFiles') }}</h3>
            <div v-for="file in commitDetail.files" :key="`${file.status}-${file.path}`">
              <span :class="`status-${file.status.toLowerCase()}`">{{ file.status }}</span>
              <RouterLink v-if="file.status !== 'D'" :to="codeLocation(file.path, 'blob', commitDetail.commit.oid)">{{ file.path }}</RouterLink>
              <span v-else>{{ file.path }}</span>
              <small v-if="file.oldPath">{{ file.oldPath }}</small>
            </div>
          </section>
          <CommitDiff :patch="commitDetail.patch" />
          <p v-if="commitDetail.patchTruncated" class="source-state">{{ t('source.diffTruncated') }}</p>
        </section>

        <section v-else-if="activeTab === 'commits'" class="source-history-panel">
          <header><h2>{{ t('source.commits') }}</h2></header>
          <article v-for="commit in commits" :key="commit.oid">
            <div><RouterLink :to="commitLocation(commit.oid)"><strong>{{ commit.subject }}</strong></RouterLink><small>{{ commit.author }} · {{ formatDate(commit.authoredAt) }}</small></div>
            <RouterLink :to="commitLocation(commit.oid)"><code>{{ commit.shortOid }}</code></RouterLink>
          </article>
        </section>

        <section v-else class="source-history-panel source-refs-panel">
          <header><h2>{{ activeTab === 'branches' ? t('source.branches') : t('source.tags') }}</h2></header>
          <article v-for="item in activeTab === 'branches' ? refs.branches : refs.tags" :key="item.name">
            <div><RouterLink :to="codeLocation('', 'tree', `${activeTab === 'branches' ? 'refs/heads' : 'refs/tags'}/${item.name}`)"><strong>{{ item.name }}</strong></RouterLink><small>{{ formatDate(item.updatedAt) }}</small></div>
            <RouterLink :to="commitLocation(item.oid)"><code>{{ item.oid.slice(0, 7) }}</code></RouterLink>
          </article>
        </section>
      </div>
    </section>
  </main>
</template>
