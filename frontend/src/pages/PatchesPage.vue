<script setup lang="ts">
import { ArrowLeft, Check, Download, Mail, MessageSquarePlus } from '@lucide/vue'
import { computed, ref, watch, watchEffect } from 'vue'
import { RouterLink, useRoute } from 'vue-router'

import { useI18n } from '../i18n'
import {
	addPatchReviewComment,
	getPatch,
	getPatches,
	resolvePatchReviewComment,
	updatePatchReview,
	type PatchReviewComment,
	type PatchSummary,
} from '../services/http'
import { applyPageSEO } from '../services/seo'
import { useAuthStore } from '../stores/auth'
import UiInlineState from '../ui/UiInlineState.vue'
import UiSkeletonRows from '../ui/UiSkeletonRows.vue'

const route = useRoute()
const auth = useAuthStore()
const { locale, t } = useI18n()
const address = ref('patchs@wave-lang.dev')
const patches = ref<PatchSummary[]>([])
const patch = ref<PatchSummary | null>(null)
const search = ref('')
const loading = ref(true)
const error = ref('')
const reviewStatus = ref<PatchSummary['reviewStatus']>('received')
const targetRepository = ref('')
const reviewBusy = ref(false)
const reviewNotice = ref('')
const generalComment = ref('')
const selectedLine = ref(0)
const inlineComment = ref('')
const commentBusy = ref(false)
const patchID = computed(() => String(route.params.patch ?? ''))
const canMaintain = computed(() => Boolean(auth.account?.sourceMaintainer))
const generalComments = computed(() => patch.value?.reviewComments.filter((item) => item.line === 0) ?? [])
const patchLines = computed(() => {
	let path = ''
	return (patch.value?.body ?? '').split('\n').map((text, index) => {
		const file = text.match(/^diff --git a\/(.+?) b\/(.+)$/)
		if (file) path = file[2]
		return {
			number: index + 1,
			path,
			text,
			kind: text.startsWith('diff --git ') || text.startsWith('index ') || text.startsWith('@@') ? 'meta'
				: text.startsWith('+++') || text.startsWith('---') ? 'file'
					: text.startsWith('+') ? 'add' : text.startsWith('-') ? 'delete' : 'context',
		}
	})
})

function commentsForLine(line: number) {
	return patch.value?.reviewComments.filter((item) => item.line === line) ?? []
}

function formatDate(value: string) {
	return new Intl.DateTimeFormat(locale.value === 'ko' ? 'ko-KR' : 'en-US', { dateStyle: 'medium', timeStyle: 'short' }).format(new Date(value))
}

function syncReview(item: PatchSummary) {
	reviewStatus.value = item.reviewStatus
	targetRepository.value = item.targetRepository
}

async function load() {
	loading.value = true
	error.value = ''
	reviewNotice.value = ''
	try {
		if (patchID.value) {
			patch.value = await getPatch(patchID.value)
			syncReview(patch.value)
		} else {
			const result = await getPatches(search.value)
			address.value = result.address || address.value
			patches.value = result.patches
		}
	} catch (reason) {
		error.value = reason instanceof Error ? reason.message : t('common.loadError')
	} finally {
		loading.value = false
	}
}

async function saveReview() {
	if (!patch.value || reviewBusy.value) return
	reviewBusy.value = true
	reviewNotice.value = ''
	try {
		patch.value = await updatePatchReview(patch.value.id, reviewStatus.value, targetRepository.value)
		syncReview(patch.value)
		reviewNotice.value = t('patches.reviewSaved')
	} catch (reason) {
		reviewNotice.value = reason instanceof Error ? reason.message : t('common.loadError')
	} finally {
		reviewBusy.value = false
	}
}

async function addComment(line: number) {
	if (!patch.value || commentBusy.value) return
	const body = line === 0 ? generalComment.value : inlineComment.value
	if (!body.trim()) return
	commentBusy.value = true
	reviewNotice.value = ''
	try {
		await addPatchReviewComment(patch.value.id, line, body)
		patch.value = await getPatch(patch.value.id)
		if (line === 0) generalComment.value = ''
		else { inlineComment.value = ''; selectedLine.value = 0 }
		reviewNotice.value = t('patches.commentSaved')
	} catch (reason) {
		reviewNotice.value = reason instanceof Error ? reason.message : t('common.loadError')
	} finally {
		commentBusy.value = false
	}
}

async function toggleResolved(comment: PatchReviewComment) {
	if (!patch.value || commentBusy.value) return
	commentBusy.value = true
	reviewNotice.value = ''
	try {
		await resolvePatchReviewComment(patch.value.id, comment.id, !comment.resolved)
		patch.value = await getPatch(patch.value.id)
		reviewNotice.value = t(comment.resolved ? 'patches.commentReopened' : 'patches.commentResolved')
	} catch (reason) {
		reviewNotice.value = reason instanceof Error ? reason.message : t('common.loadError')
	} finally {
		commentBusy.value = false
	}
}

function downloadURL(series = false) {
	return `/api/v1/patches/${encodeURIComponent(patch.value?.id ?? '')}/download${series ? '?series=1' : ''}`
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
          <div class="patch-list-meta">
            <span class="patch-status">{{ t(`patches.status.${item.reviewStatus}`) }}</span>
            <span v-if="item.reviewCommentCount">{{ item.reviewCommentCount }} {{ t('patches.comments') }}</span>
            <span v-if="item.version > 1">v{{ item.version }}</span><span v-if="item.total">{{ item.part }}/{{ item.total }}</span><small>{{ item.files.length }} {{ t('patches.files') }}</small>
          </div>
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

        <section class="patch-review" :aria-labelledby="'patch-review-title'">
          <header>
            <div><h2 id="patch-review-title">{{ t('patches.review') }}</h2><p>{{ t('patches.reviewHelp') }}</p></div>
            <span class="patch-review-state">{{ t(`patches.status.${patch.reviewStatus}`) }}</span>
          </header>
          <dl class="patch-review-facts">
            <div v-if="patch.targetRepository"><dt>{{ t('patches.targetRepository') }}</dt><dd>{{ patch.targetRepository }}</dd></div>
            <div v-if="patch.assigneeName"><dt>{{ t('patches.assignee') }}</dt><dd>{{ patch.assigneeName }}</dd></div>
            <div v-if="patch.sha256"><dt>{{ t('patches.checksum') }}</dt><dd><code>{{ patch.sha256 }}</code></dd></div>
          </dl>
          <div v-if="canMaintain" class="patch-review-actions">
            <a class="ui-button" :href="downloadURL()"><Download :size="15" aria-hidden="true" />{{ t('patches.downloadMbox') }}</a>
            <a v-if="patch.seriesCount > 1" class="ui-button" :href="downloadURL(true)"><Download :size="15" aria-hidden="true" />{{ t('patches.downloadSeries') }}</a>
          </div>
          <form v-if="canMaintain" class="patch-review-form" @submit.prevent="saveReview">
            <label><span>{{ t('patches.reviewStatus') }}</span><select v-model="reviewStatus"><option v-for="status in ['received', 'reviewing', 'accepted', 'rejected', 'applied']" :key="status" :value="status">{{ t(`patches.status.${status}`) }}</option></select></label>
            <label><span>{{ t('patches.targetRepository') }}</span><input v-model.trim="targetRepository" type="text" maxlength="101" pattern="[A-Za-z0-9][A-Za-z0-9._-]{0,49}/[A-Za-z0-9][A-Za-z0-9._-]{0,49}" :placeholder="t('patches.targetRepositoryPlaceholder')" /></label>
            <button class="ui-button primary" type="submit" :disabled="reviewBusy">{{ t('patches.saveReview') }}</button>
          </form>
          <p v-if="reviewNotice" class="patch-review-notice" aria-live="polite">{{ reviewNotice }}</p>

          <div class="patch-comments-general">
            <h3>{{ t('patches.generalComments') }}</h3>
            <p v-if="generalComments.length === 0" class="patch-comments-empty">{{ t('patches.noComments') }}</p>
            <article v-for="comment in generalComments" :key="comment.id" class="patch-comment" :class="{ resolved: comment.resolved }">
              <header><strong>{{ comment.authorName }}</strong><time :datetime="comment.createdAt">{{ formatDate(comment.createdAt) }}</time><span v-if="comment.resolved"><Check :size="13" />{{ t('patches.resolved') }}</span></header>
              <p>{{ comment.body }}</p>
              <button v-if="canMaintain" type="button" :disabled="commentBusy" @click="toggleResolved(comment)">{{ t(comment.resolved ? 'patches.reopen' : 'patches.resolve') }}</button>
            </article>
            <form v-if="canMaintain" class="patch-comment-form" @submit.prevent="addComment(0)"><label><span>{{ t('patches.addGeneralComment') }}</span><textarea v-model="generalComment" required maxlength="4000" rows="3" /></label><button class="ui-button" type="submit" :disabled="commentBusy">{{ t('patches.addComment') }}</button></form>
          </div>
        </section>

        <div v-if="patch.files.length" class="patch-files"><strong>{{ t('patches.changedFiles') }}</strong><code v-for="file in patch.files" :key="file">{{ file }}</code></div>
        <div class="patch-body" tabindex="0" :aria-label="t('patches.diff')">
          <div v-for="line in patchLines" :key="line.number" class="patch-code-group">
            <div class="patch-code-line" :class="`patch-line-${line.kind}`">
              <span class="patch-line-number" aria-hidden="true">{{ line.number }}</span>
              <button v-if="canMaintain && line.path" class="patch-line-comment" type="button" :aria-label="t('patches.commentOnLine', { line: line.number })" @click="selectedLine = selectedLine === line.number ? 0 : line.number"><MessageSquarePlus :size="13" aria-hidden="true" /></button>
              <span v-else class="patch-line-comment-placeholder" />
              <code>{{ line.text || ' ' }}</code>
            </div>
            <article v-for="comment in commentsForLine(line.number)" :key="comment.id" class="patch-comment inline" :class="{ resolved: comment.resolved }">
              <header><code>{{ comment.path }}:{{ comment.line }}</code><strong>{{ comment.authorName }}</strong><time :datetime="comment.createdAt">{{ formatDate(comment.createdAt) }}</time><span v-if="comment.resolved"><Check :size="13" />{{ t('patches.resolved') }}</span></header>
              <p>{{ comment.body }}</p>
              <button v-if="canMaintain" type="button" :disabled="commentBusy" @click="toggleResolved(comment)">{{ t(comment.resolved ? 'patches.reopen' : 'patches.resolve') }}</button>
            </article>
            <form v-if="canMaintain && selectedLine === line.number" class="patch-comment-form inline" @submit.prevent="addComment(line.number)">
              <label><span>{{ t('patches.commentingOn', { path: line.path, line: line.number }) }}</span><textarea v-model="inlineComment" required maxlength="4000" rows="3" autofocus /></label>
              <div><button type="button" class="ui-button" @click="selectedLine = 0; inlineComment = ''">{{ t('common.cancel') }}</button><button class="ui-button primary" type="submit" :disabled="commentBusy">{{ t('patches.addComment') }}</button></div>
            </form>
          </div>
        </div>
      </template>
    </section>
  </main>
</template>
