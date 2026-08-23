<script setup lang="ts">
import { Bell, Code2, FileCode2, MailPlus, MessageSquareReply, RefreshCw, Search, Users } from '@lucide/vue'
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { useI18n } from '../../i18n'
import PlatformWaveEditor from '../editor/PlatformWaveEditor.vue'
import {
	getMailingLists, getMailingListThread, getMailingListThreads, postMailingListThread,
	replyMailingListThread, setMailingListSubscription,
	type MailingListSummary, type MailingListThread, type MailingListThreadSummary,
} from '../../services/http'
import { useAuthStore } from '../../stores/auth'
import UiInlineState from '../../ui/UiInlineState.vue'

const { locale, t } = useI18n()
const auth = useAuthStore()
const route = useRoute()
const router = useRouter()
const lists = ref<MailingListSummary[]>([])
const threads = ref<MailingListThreadSummary[]>([])
const thread = ref<MailingListThread | null>(null)
const query = ref('')
const loading = ref(false)
const threadLoading = ref(false)
const saving = ref(false)
const error = ref('')
const actionError = ref('')
const composing = ref(false)
const subject = ref('')
const body = ref('')
const replyBody = ref('')

const listID = computed(() => String(route.params.list ?? ''))
const threadID = computed(() => String(route.params.thread ?? ''))
const selectedList = computed(() => lists.value.find((item) => item.id === listID.value) ?? null)
const canPost = computed(() => Boolean(selectedList.value?.subscribed) && (
	selectedList.value?.postingPolicy !== 'staff' || Boolean(auth.account?.owner || auth.account?.administrator)
))

const listIcons: Record<string, typeof Bell> = { announce: Bell, development: Code2, patchs: FileCode2 }

function displayAddress(value: string) {
	const label = value.replace(/\s*<[^>]+>\s*$/, '').trim().replace(/^"(.*)"$/, '$1')
	return label || value
}

function formatDate(value: string) {
	if (!value) return ''
	return new Intl.DateTimeFormat(locale.value === 'ko' ? 'ko-KR' : 'en-US', {
		year: 'numeric', month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit',
		timeZone: auth.account?.timeZone || undefined,
	}).format(new Date(value))
}

async function loadLists() {
	loading.value = true
	error.value = ''
	try {
		lists.value = await getMailingLists()
		if (!listID.value && lists.value.length) {
			const first = lists.value.find((item) => item.subscribed) ?? lists.value.find((item) => item.id === 'development') ?? lists.value[0]
			await router.replace(`/mail/lists/${first.id}`)
		}
	} catch (reason) {
		error.value = reason instanceof Error ? reason.message : t('common.loadError')
	} finally {
		loading.value = false
	}
}

async function loadThreads() {
	if (!listID.value) return
	threadLoading.value = true
	actionError.value = ''
	try {
		threads.value = await getMailingListThreads(listID.value, query.value.trim())
		thread.value = threadID.value ? await getMailingListThread(listID.value, threadID.value) : null
	} catch (reason) {
		actionError.value = reason instanceof Error ? reason.message : t('common.loadError')
	} finally {
		threadLoading.value = false
	}
}

async function chooseList(id: string) {
	composing.value = false
	query.value = ''
	await router.push(`/mail/lists/${id}`)
}

async function chooseThread(id: string) {
	composing.value = false
	await router.push(`/mail/lists/${listID.value}/thread/${id}`)
}

async function toggleSubscription() {
	if (!selectedList.value) return
	saving.value = true
	actionError.value = ''
	try {
		await setMailingListSubscription(selectedList.value.id, !selectedList.value.subscribed)
		selectedList.value.subscribed = !selectedList.value.subscribed
	} catch (reason) {
		actionError.value = reason instanceof Error ? reason.message : t('mail.lists.subscriptionFailed')
	} finally {
		saving.value = false
	}
}

function openComposer() {
	thread.value = null
	composing.value = true
	actionError.value = ''
}

async function submitThread() {
	if (!selectedList.value) return
	saving.value = true
	actionError.value = ''
	try {
		const created = await postMailingListThread(selectedList.value.id, subject.value, body.value)
		subject.value = ''
		body.value = ''
		composing.value = false
		await router.push(`/mail/lists/${selectedList.value.id}/thread/${created.id}`)
	} catch (reason) {
		actionError.value = reason instanceof Error ? reason.message : t('mail.lists.postFailed')
	} finally {
		saving.value = false
	}
}

async function submitReply() {
	if (!selectedList.value || !thread.value) return
	saving.value = true
	actionError.value = ''
	try {
		thread.value = await replyMailingListThread(selectedList.value.id, thread.value.id, replyBody.value)
		replyBody.value = ''
		await loadThreads()
	} catch (reason) {
		actionError.value = reason instanceof Error ? reason.message : t('mail.lists.replyFailed')
	} finally {
		saving.value = false
	}
}

onMounted(async () => {
	await auth.initialize()
	if (!auth.account) return
	await loadLists()
	await loadThreads()
})

watch([listID, threadID], async ([currentList], [previousList]) => {
	if (!auth.account || !currentList) return
	if (currentList !== previousList) {
		query.value = ''
		composing.value = false
	}
	await loadThreads()
})
</script>

<template>
	<section class="mailing-list-workspace" :class="{ 'reader-open': composing || thread }" :aria-label="t('mail.lists.title')">
		<aside class="mailing-list-directory">
			<header><div><Users :size="17" aria-hidden="true" /><strong>{{ t('mail.lists.title') }}</strong></div><small>{{ t('mail.lists.internalOnly') }}</small></header>
			<UiInlineState v-if="loading" :message="t('common.loading')" />
			<nav v-else :aria-label="t('mail.lists.directory')">
				<button v-for="item in lists" :key="item.id" type="button" :class="{ active: item.id === listID }" @click="chooseList(item.id)">
					<component :is="listIcons[item.id] ?? Users" :size="17" aria-hidden="true" />
					<span><strong>{{ item.name }}</strong><small>{{ item.address }}</small></span>
					<i v-if="item.subscribed" :title="t('mail.lists.subscribed')" />
				</button>
			</nav>
		</aside>

		<section class="mailing-thread-list" :aria-label="t('mail.lists.threads')">
			<header v-if="selectedList" class="mailing-thread-header">
				<div><div><h1>{{ selectedList.name }}</h1><p>{{ selectedList.description }}</p></div><button type="button" :disabled="saving" @click="toggleSubscription">{{ selectedList.subscribed ? t('mail.lists.leave') : t('mail.lists.join') }}</button></div>
				<div class="mailing-thread-actions">
					<button v-if="canPost" class="ui-button primary" type="button" @click="openComposer"><MailPlus :size="15" />{{ t('mail.lists.newThread') }}</button>
					<RouterLink v-if="selectedList.id === 'patchs'" class="ui-button" to="/mail/lists/patchs/reviews"><FileCode2 :size="15" />{{ t('mail.lists.patchReview') }}</RouterLink>
				</div>
				<form role="search" @submit.prevent="loadThreads"><Search :size="15" aria-hidden="true" /><input v-model="query" type="search" maxlength="200" :placeholder="t('mail.lists.search')" /><button type="submit" :title="t('mail.refresh')"><RefreshCw :size="15" /></button></form>
			</header>
			<UiInlineState v-if="threadLoading && !threads.length" :message="t('common.loading')" />
			<UiInlineState v-else-if="error" :message="error" />
			<div v-else-if="threads.length" class="mailing-thread-rows">
				<button v-for="item in threads" :key="item.id" type="button" :class="{ active: item.id === threadID }" @click="chooseThread(item.id)">
					<strong>{{ item.subject }}</strong><span>{{ item.preview }}</span>
					<footer><span>{{ displayAddress(item.author) }}</span><span>{{ item.messageCount }} {{ t('mail.lists.messages') }}</span><time :datetime="item.lastActivityAt">{{ formatDate(item.lastActivityAt) }}</time></footer>
				</button>
			</div>
			<div v-else class="mail-empty"><strong>{{ t('mail.lists.emptyTitle') }}</strong><span>{{ selectedList?.subscribed ? t('mail.lists.emptyDetail') : t('mail.lists.joinToPost') }}</span></div>
		</section>

		<section class="mailing-thread-reader" :aria-label="t('mail.lists.threadReader')">
			<form v-if="composing" class="mailing-list-composer" @submit.prevent="submitThread">
				<header><button type="button" @click="composing = false">{{ t('common.cancel') }}</button><strong>{{ t('mail.lists.newThread') }}</strong></header>
				<p v-if="selectedList?.id === 'patchs'" class="mailing-patch-help"><FileCode2 :size="17" aria-hidden="true" />{{ t('patches.submitHelp') }}</p>
				<label>{{ t('mail.subject') }}<input v-model="subject" required maxlength="180" :placeholder="selectedList?.id === 'patchs' ? t('mail.lists.patchSubjectPlaceholder') : ''" /></label>
				<PlatformWaveEditor v-model="body" :label="t('mail.body')" mode="plain" required :max-length="20000" :rows="16" />
				<p v-if="actionError" class="mail-action-error" role="alert">{{ actionError }}</p>
				<footer><button class="ui-button primary" type="submit" :disabled="saving">{{ t('mail.lists.publish') }}</button></footer>
			</form>

			<article v-else-if="thread" class="mailing-thread-view">
				<header><button class="mailing-mobile-back" type="button" @click="router.push(`/mail/lists/${listID}`)">← {{ t('mail.lists.threads') }}</button><h1>{{ thread.subject }}</h1><p>{{ thread.address }}</p></header>
				<section v-for="message in thread.messages" :key="message.id" class="mailing-thread-message">
					<header><RouterLink v-if="message.authorAccountId" :to="`/user/id/${encodeURIComponent(message.authorAccountId)}`">{{ displayAddress(message.from) }}</RouterLink><strong v-else>{{ displayAddress(message.from) }}</strong><time :datetime="message.createdAt">{{ formatDate(message.createdAt) }}</time></header>
					<pre>{{ message.body }}</pre>
				</section>
				<form v-if="canPost" class="mailing-reply-form" @submit.prevent="submitReply">
					<div class="mailing-reply-editor"><MessageSquareReply :size="17" aria-hidden="true" /><PlatformWaveEditor v-model="replyBody" :label="t('mail.lists.reply')" mode="plain" required :max-length="20000" :rows="5" /></div>
					<p v-if="actionError" class="mail-action-error" role="alert">{{ actionError }}</p>
					<footer><button class="ui-button primary" type="submit" :disabled="saving">{{ t('mail.lists.sendReply') }}</button></footer>
				</form>
			</article>

			<div v-else class="mail-reader-empty"><strong>{{ t('mail.lists.selectTitle') }}</strong><span>{{ t('mail.lists.selectDetail') }}</span></div>
		</section>
	</section>
</template>
