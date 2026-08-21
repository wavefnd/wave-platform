<script setup lang="ts">
import { ArrowLeft, MessageSquare, Pencil, Plus } from '@lucide/vue'
import { computed, ref, watch, watchEffect } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'

import MarkdownContent from '../components/MarkdownContent.vue'
import { useI18n } from '../i18n'
import {
	addRFCComment,
	createRFC,
	getRFC,
	getRFCs,
	updateRFC,
	updateRFCStatus,
	type RFCProposal,
	type RFCStatus,
} from '../services/http'
import { applyPageSEO, plainTextDescription } from '../services/seo'
import { useAuthStore } from '../stores/auth'
import UiInlineState from '../ui/UiInlineState.vue'
import UiSkeletonRows from '../ui/UiSkeletonRows.vue'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const { locale, t } = useI18n()
const proposals = ref<RFCProposal[]>([])
const proposal = ref<RFCProposal | null>(null)
const loading = ref(true)
const busy = ref(false)
const error = ref('')
const notice = ref('')
const query = ref('')
const statusFilter = ref('')
const title = ref('')
const content = ref('')
const commentBody = ref('')
const selectedStatus = ref<RFCStatus>('draft')
const number = computed(() => Number(route.params.number) || 0)
const editing = computed(() => route.name === 'rfc-new' || route.name === 'rfc-edit')
const detail = computed(() => number.value > 0)
const canEdit = computed(() => Boolean(proposal.value && auth.account?.id === proposal.value.authorAccountId && proposal.value.status === 'draft'))
const canMaintain = computed(() => Boolean(auth.account?.rfcMaintainer))
const statuses: RFCStatus[] = ['draft', 'discussion', 'accepted', 'rejected', 'implementing', 'completed', 'withdrawn']

function rfcNumber(value: number) {
	return `RFC-${String(value).padStart(4, '0')}`
}

function formatDate(value: string) {
	return new Intl.DateTimeFormat(locale.value === 'ko' ? 'ko-KR' : 'en-US', { dateStyle: 'medium' }).format(new Date(value))
}

function defaultTemplate() {
	return locale.value === 'ko'
		? '## 배경\n\n해결하려는 문제와 현재 상황을 설명하세요.\n\n## 제안\n\n구체적인 설계를 설명하세요.\n\n## 호환성 영향\n\n기존 코드와 사용자에게 미치는 영향을 설명하세요.\n\n## 보안 고려사항\n\n보안 경계와 위험을 설명하세요.\n\n## 대안\n\n검토한 다른 방법을 설명하세요.'
		: '## Motivation\n\nDescribe the problem and current situation.\n\n## Proposal\n\nDescribe the concrete design.\n\n## Compatibility\n\nExplain the impact on existing code and users.\n\n## Security considerations\n\nDescribe security boundaries and risks.\n\n## Alternatives\n\nDescribe the alternatives considered.'
}

async function load() {
	loading.value = true
	error.value = ''
	notice.value = ''
	try {
		if (detail.value) {
			proposal.value = await getRFC(number.value)
			selectedStatus.value = proposal.value.status
			if (editing.value) {
				title.value = proposal.value.title
				content.value = proposal.value.content
			}
		} else if (route.name === 'rfc-new') {
			proposal.value = null
			title.value = ''
			content.value = defaultTemplate()
		} else {
			proposals.value = await getRFCs(query.value, statusFilter.value)
			proposal.value = null
		}
	} catch (reason) {
		error.value = reason instanceof Error ? reason.message : t('common.loadError')
	} finally {
		loading.value = false
	}
}

async function saveProposal() {
	if (busy.value) return
	busy.value = true
	error.value = ''
	try {
		const saved = number.value ? await updateRFC(number.value, title.value, content.value) : await createRFC(title.value, content.value)
		await router.push({ name: 'rfc-detail', params: { number: saved.number } })
	} catch (reason) {
		error.value = reason instanceof Error ? reason.message : t('rfc.saveFailed')
	} finally {
		busy.value = false
	}
}

async function saveStatus() {
	if (!proposal.value || busy.value || selectedStatus.value === proposal.value.status) return
	if (!window.confirm(t('rfc.confirmStatus', { status: t(`rfc.status.${selectedStatus.value}`) }))) return
	busy.value = true
	error.value = ''
	try {
		proposal.value = await updateRFCStatus(proposal.value.number, selectedStatus.value)
		notice.value = t('rfc.statusSaved')
	} catch (reason) {
		error.value = reason instanceof Error ? reason.message : t('rfc.saveFailed')
	} finally {
		busy.value = false
	}
}

async function addComment() {
	if (!proposal.value || busy.value || !commentBody.value.trim()) return
	busy.value = true
	error.value = ''
	try {
		await addRFCComment(proposal.value.number, commentBody.value)
		commentBody.value = ''
		proposal.value = await getRFC(proposal.value.number)
		notice.value = t('rfc.commentSaved')
	} catch (reason) {
		error.value = reason instanceof Error ? reason.message : t('rfc.commentFailed')
	} finally {
		busy.value = false
	}
}

let searchTimer = 0
watch(() => route.fullPath, load, { immediate: true })
watch([query, statusFilter], () => {
	if (detail.value || editing.value) return
	window.clearTimeout(searchTimer)
	searchTimer = window.setTimeout(load, 250)
})
watchEffect(() => applyPageSEO({
	title: proposal.value ? `${rfcNumber(proposal.value.number)}: ${proposal.value.title} · Wave RFCs` : `${t('rfc.title')} · Wave`,
	description: proposal.value ? plainTextDescription(proposal.value.summary, proposal.value.title) : t('rfc.lead'),
	locale: locale.value, path: route.path,
}))
</script>

<template>
  <main class="rfc-page portal-width">
    <UiSkeletonRows v-if="loading" :rows="8" />
    <UiInlineState v-else-if="error && !editing" :message="error" :action="t('common.retry')" @action="load" />

    <template v-else-if="editing">
      <RouterLink class="rfc-back" :to="number ? `/rfcs/${number}` : '/rfcs'"><ArrowLeft :size="15" />{{ t('rfc.back') }}</RouterLink>
      <header class="rfc-page-heading"><div><span>{{ number ? rfcNumber(number) : t('rfc.new') }}</span><h1>{{ number ? t('rfc.edit') : t('rfc.create') }}</h1><p>{{ t('rfc.editorHelp') }}</p></div></header>
      <form class="rfc-editor" @submit.prevent="saveProposal">
        <label><span>{{ t('rfc.proposalTitle') }}</span><input v-model.trim="title" required minlength="5" maxlength="180" /></label>
        <label><span>{{ t('rfc.content') }}</span><textarea v-model="content" required minlength="20" maxlength="200000" rows="24" /></label>
        <p v-if="error" class="rfc-error" role="alert">{{ error }}</p>
        <footer><RouterLink class="ui-button" :to="number ? `/rfcs/${number}` : '/rfcs'">{{ t('common.cancel') }}</RouterLink><button class="ui-button primary" type="submit" :disabled="busy">{{ t('common.save') }}</button></footer>
      </form>
    </template>

    <template v-else-if="proposal">
      <RouterLink class="rfc-back" to="/rfcs"><ArrowLeft :size="15" />{{ t('rfc.all') }}</RouterLink>
      <article class="rfc-detail">
        <header>
          <div class="rfc-detail-meta"><strong>{{ rfcNumber(proposal.number) }}</strong><span :class="`rfc-status status-${proposal.status}`">{{ t(`rfc.status.${proposal.status}`) }}</span></div>
          <h1>{{ proposal.title }}</h1>
          <p>{{ t('rfc.by', { author: proposal.authorName }) }} · {{ formatDate(proposal.createdAt) }}</p>
          <RouterLink v-if="canEdit" class="ui-button" :to="`/rfcs/${proposal.number}/edit`"><Pencil :size="15" />{{ t('common.edit') }}</RouterLink>
        </header>

        <div v-if="canMaintain" class="rfc-maintainer-control">
          <label><span>{{ t('rfc.officialStatus') }}</span><select v-model="selectedStatus"><option v-for="item in statuses" :key="item" :value="item">{{ t(`rfc.status.${item}`) }}</option></select></label>
          <button class="ui-button" type="button" :disabled="busy || selectedStatus === proposal.status" @click="saveStatus">{{ t('rfc.updateStatus') }}</button>
        </div>
        <p v-if="notice" class="rfc-notice" aria-live="polite">{{ notice }}</p>
        <p v-if="error" class="rfc-error" role="alert">{{ error }}</p>
        <MarkdownContent class="rfc-content" :source="proposal.content" />
      </article>

      <section class="rfc-discussion" aria-labelledby="rfc-discussion-title">
        <header><div><h2 id="rfc-discussion-title">{{ t('rfc.discussion') }}</h2><p>{{ t('rfc.discussionHelp') }}</p></div><span><MessageSquare :size="15" />{{ proposal.commentCount }}</span></header>
        <p v-if="proposal.comments.length === 0" class="rfc-empty">{{ t('rfc.noComments') }}</p>
        <article v-for="comment in proposal.comments" :key="comment.id" class="rfc-comment">
          <header><strong>{{ comment.authorName }}</strong><time :datetime="comment.createdAt">{{ formatDate(comment.createdAt) }}</time></header>
          <p>{{ comment.body }}</p>
        </article>
        <form v-if="auth.account" class="rfc-comment-form" @submit.prevent="addComment"><label><span>{{ t('rfc.yourComment') }}</span><textarea v-model="commentBody" required maxlength="10000" rows="5" /></label><button class="ui-button primary" type="submit" :disabled="busy">{{ t('rfc.addComment') }}</button></form>
        <p v-else class="rfc-signin"><RouterLink :to="{ name: 'login', query: { redirect: route.fullPath } }">{{ t('auth.signIn') }}</RouterLink> {{ t('rfc.toComment') }}</p>
      </section>
    </template>

    <template v-else>
      <header class="rfc-page-heading">
        <div><span>Wave RFCs</span><h1>{{ t('rfc.title') }}</h1><p>{{ t('rfc.lead') }}</p></div>
        <RouterLink v-if="auth.account" class="ui-button primary" to="/rfcs/new"><Plus :size="16" />{{ t('rfc.new') }}</RouterLink>
      </header>
      <div class="rfc-tools">
        <label><span class="sr-only">{{ t('rfc.search') }}</span><input v-model="query" type="search" :placeholder="t('rfc.search')" /></label>
        <label><span class="sr-only">{{ t('rfc.filterStatus') }}</span><select v-model="statusFilter"><option value="">{{ t('rfc.allStatuses') }}</option><option v-for="item in statuses" :key="item" :value="item">{{ t(`rfc.status.${item}`) }}</option></select></label>
      </div>
      <div class="rfc-list">
        <article v-for="item in proposals" :key="item.number">
          <div class="rfc-number">{{ rfcNumber(item.number) }}</div>
          <div><RouterLink :to="`/rfcs/${item.number}`">{{ item.title }}</RouterLink><p>{{ item.summary }}</p><small>{{ item.authorName }} · {{ formatDate(item.updatedAt) }}</small></div>
          <div class="rfc-list-state"><span :class="`rfc-status status-${item.status}`">{{ t(`rfc.status.${item.status}`) }}</span><small><MessageSquare :size="13" />{{ item.commentCount }}</small></div>
        </article>
        <p v-if="proposals.length === 0" class="rfc-empty">{{ t('rfc.empty') }}</p>
      </div>
    </template>
  </main>
</template>
