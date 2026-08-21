<script setup lang="ts">
import { ArrowLeft, Mail } from '@lucide/vue'
import { computed, ref, watch, watchEffect } from 'vue'
import { RouterLink, useRoute } from 'vue-router'

import { useI18n } from '../i18n'
import { getPatch, getPatches, type PatchSummary } from '../services/http'
import { applyPageSEO } from '../services/seo'
import UiInlineState from '../ui/UiInlineState.vue'
import UiSkeletonRows from '../ui/UiSkeletonRows.vue'

const route = useRoute()
const { locale, t } = useI18n()
const address = ref('patchs@wave-lang.dev')
const patches = ref<PatchSummary[]>([])
const patch = ref<PatchSummary | null>(null)
const search = ref('')
const loading = ref(true)
const error = ref('')
const patchID = computed(() => String(route.params.patch ?? ''))
const patchLines = computed(() => (patch.value?.body ?? '').split('\n').map((text) => ({
	text,
	kind: text.startsWith('diff --git ') || text.startsWith('index ') || text.startsWith('@@') ? 'meta'
		: text.startsWith('+++') || text.startsWith('---') ? 'file'
			: text.startsWith('+') ? 'add' : text.startsWith('-') ? 'delete' : 'context',
})))

function formatDate(value: string) {
	return new Intl.DateTimeFormat(locale.value === 'ko' ? 'ko-KR' : 'en-US', { dateStyle: 'medium', timeStyle: 'short' }).format(new Date(value))
}

async function load() {
	loading.value = true; error.value = ''
	try {
		if (patchID.value) patch.value = await getPatch(patchID.value)
		else { const result = await getPatches(search.value); address.value = result.address || address.value; patches.value = result.patches }
	} catch (reason) { error.value = reason instanceof Error ? reason.message : t('common.loadError') }
	finally { loading.value = false }
}

let searchTimer = 0
watch(() => route.fullPath, load, { immediate: true })
watch(search, () => { window.clearTimeout(searchTimer); searchTimer = window.setTimeout(load, 250) })
watchEffect(() => applyPageSEO({
	title: patch.value ? `${patch.value.subject} · Wave Patches` : 'Wave Patches',
	description: patch.value?.preview || t('patches.lead'), locale: locale.value, path: route.path,
}))
</script>

<template>
  <main class="source-forge patch-archive">
    <header class="source-org-header">
      <div class="source-width source-org-row">
        <RouterLink class="source-org-name" to="/source">Wave Source</RouterLink>
        <nav :aria-label="t('source.sections')">
          <RouterLink to="/source">{{ t('source.repositories') }}</RouterLink>
          <RouterLink to="/patches" class="active">{{ t('patches.title') }}</RouterLink>
        </nav>
      </div>
    </header>

    <section v-if="!patchID" class="source-width patch-index">
      <header class="patch-heading">
        <div><h1>{{ t('patches.title') }}</h1><p>{{ t('patches.lead') }}</p></div>
        <a class="patch-address" :href="`mailto:${address}`"><Mail :size="16" aria-hidden="true" />{{ address }}</a>
      </header>
      <div class="patch-submit-help"><strong>{{ t('patches.submit') }}</strong><code>git send-email --to={{ address }} *.patch</code><p>{{ t('patches.submitHelp') }}</p></div>
      <label class="source-repository-search"><span class="sr-only">{{ t('patches.search') }}</span><input v-model="search" type="search" :placeholder="t('patches.search')" /></label>
      <div class="patch-list">
        <UiSkeletonRows v-if="loading" :rows="6" />
        <UiInlineState v-else-if="error" :message="t('common.loadError')" :action="t('common.retry')" @action="load" />
        <article v-for="item in patches" v-else :key="item.id">
          <div class="patch-list-main">
            <RouterLink :to="`/patches/${item.id}`">{{ item.subject }}</RouterLink>
            <p v-if="item.preview">{{ item.preview }}</p>
            <small>{{ item.authorName }} &lt;{{ item.authorEmail }}&gt; · {{ formatDate(item.receivedAt) }}</small>
          </div>
          <div class="patch-list-meta"><span v-if="item.version > 1">v{{ item.version }}</span><span v-if="item.total">{{ item.part }}/{{ item.total }}</span><small>{{ item.files.length }} {{ t('patches.files') }}</small></div>
        </article>
        <p v-if="!loading && !error && patches.length === 0" class="portal-empty-state">{{ t('patches.empty') }}</p>
      </div>
    </section>

    <section v-else class="source-width patch-detail">
      <UiSkeletonRows v-if="loading" :rows="8" />
      <UiInlineState v-else-if="error" :message="t('patches.notFound')" :action="t('common.retry')" @action="load" />
      <template v-else-if="patch">
        <RouterLink class="patch-back" to="/patches"><ArrowLeft :size="15" aria-hidden="true" />{{ t('patches.back') }}</RouterLink>
        <header><h1>{{ patch.subject }}</h1><p>{{ patch.authorName }} &lt;<a :href="`mailto:${patch.authorEmail}`">{{ patch.authorEmail }}</a>&gt;</p><time :datetime="patch.receivedAt">{{ formatDate(patch.receivedAt) }}</time></header>
        <div v-if="patch.files.length" class="patch-files"><strong>{{ t('patches.changedFiles') }}</strong><code v-for="file in patch.files" :key="file">{{ file }}</code></div>
        <pre class="patch-body" tabindex="0"><code><span v-for="(line, index) in patchLines" :key="index" :class="`patch-line-${line.kind}`">{{ line.text }}
</span></code></pre>
      </template>
    </section>
  </main>
</template>
