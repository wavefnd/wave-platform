<script setup lang="ts">
import {
	Archive, Inbox, MailOpen, PenLine, RefreshCw, Reply, Search, Send, Trash2, X,
} from '@lucide/vue'
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'

import GmailDeliveryWarning from '../GmailDeliveryWarning.vue'
import { useI18n } from '../../i18n'
import { containsGmailAddress } from '../../services/email-address'
import {
	getMailbox, getMailMessage, getManagementMailbox, getManagementMailMessage,
	sendMail, sendManagementMail, updateMailEntry, updateManagementMailEntry,
	type MailboxItem, type MailboxView, type MailMessageView,
} from '../../services/http'
import { useAuthStore } from '../../stores/auth'
import UiInlineState from '../../ui/UiInlineState.vue'

const props = defineProps<{ mode: 'personal' | 'team' }>()
const { locale, t } = useI18n()
const auth = useAuthStore()
const mailbox = ref<MailboxView | null>(null)
const selected = ref<MailMessageView | null>(null)
const loading = ref(false)
const readerLoading = ref(false)
const submitting = ref(false)
const error = ref('')
const actionError = ref('')
const folder = ref('Inbox')
const query = ref('')
const composing = ref(false)
const composeFrom = ref('')
const composeTo = ref('')
const composeSubject = ref('')
const composeBody = ref('')
const composeParentEntryID = ref('')
const teamMode = computed(() => props.mode === 'team')
const gmailRecipient = computed(() => containsGmailAddress(composeTo.value))
let refreshTimer: number | undefined

const folders = computed(() => [
	{ id: 'Inbox', label: t('mail.inbox'), icon: Inbox },
	{ id: 'Sent', label: t('mail.sent'), icon: Send },
	{ id: 'Archive', label: t('mail.archive'), icon: Archive },
	{ id: 'Trash', label: t('mail.trash'), icon: Trash2 },
])

function displayAddress(value: string) {
	const label = value.replace(/\s*<[^>]+>\s*$/, '').trim()
	return label.replace(/^"(.*)"$/, '$1') || value
}

function rawAddress(value: string) {
	return value.match(/<([^>]+)>/)?.[1] ?? value.trim()
}

function formatDate(value: string) {
	if (!value) return ''
	return new Intl.DateTimeFormat(locale.value === 'ko' ? 'ko-KR' : 'en-US', {
		month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit',
		timeZone: auth.account?.timeZone || undefined,
	}).format(new Date(value))
}

function deliveryLabel(status: string) {
	return status ? t(`mail.delivery.${status}`) : ''
}

async function fetchMailbox() {
	return teamMode.value ? getManagementMailbox(folder.value, query.value.trim()) : getMailbox(folder.value, query.value.trim())
}

async function fetchMessage(entryID: string) {
	return teamMode.value ? getManagementMailMessage(entryID) : getMailMessage(entryID)
}

async function mutateEntry(entryID: string, action: 'archive' | 'trash' | 'read' | 'unread') {
	return teamMode.value ? updateManagementMailEntry(entryID, action) : updateMailEntry(entryID, action)
}

async function loadMailbox() {
	if (!auth.account) return
	loading.value = true
	error.value = ''
	try {
		mailbox.value = await fetchMailbox()
		if (teamMode.value && !composeFrom.value) composeFrom.value = mailbox.value.addresses?.[0] || mailbox.value.address
	} catch (reason) {
		error.value = reason instanceof Error ? reason.message : t('common.loadError')
	} finally {
		loading.value = false
	}
}

async function refreshMailboxState() {
	if (!auth.account || composing.value || submitting.value || document.hidden) return
	try {
		mailbox.value = await fetchMailbox()
		if (selected.value) selected.value = await fetchMessage(selected.value.entryId)
	} catch {
		// Background refresh must not replace the current mailbox with an error.
	}
}

async function selectFolder(value: string) {
	folder.value = value
	selected.value = null
	composing.value = false
	query.value = ''
	await loadMailbox()
}

async function openMessage(item: MailboxItem) {
	composing.value = false
	readerLoading.value = true
	actionError.value = ''
	try {
		selected.value = await fetchMessage(item.id)
		if (item.flags.includes('unread')) {
			selected.value = await mutateEntry(item.id, 'read')
			item.flags = item.flags.filter((flag) => flag !== 'unread')
		}
	} catch (reason) {
		actionError.value = reason instanceof Error ? reason.message : t('common.loadError')
	} finally {
		readerLoading.value = false
	}
}

function openComposer(parentEntryID = '') {
	selected.value = null
	composing.value = true
	actionError.value = ''
	composeParentEntryID.value = parentEntryID
	if (teamMode.value && !composeFrom.value) composeFrom.value = mailbox.value?.addresses?.[0] || mailbox.value?.address || ''
}

function replyToSelected() {
	if (!selected.value) return
	const parentEntryID = selected.value.entryId
	const replyingFromSent = folder.value === 'Sent'
	composeTo.value = replyingFromSent ? selected.value.to.map(rawAddress).join(', ') : rawAddress(selected.value.from)
	if (teamMode.value) {
		const candidate = replyingFromSent
			? rawAddress(selected.value.from)
			: selected.value.to.map(rawAddress).find((address) => mailbox.value?.addresses?.some((alias) => alias.toLowerCase() === address.toLowerCase()))
		const allowedAlias = mailbox.value?.addresses?.find((alias) => alias.toLowerCase() === candidate?.toLowerCase())
		if (allowedAlias) composeFrom.value = allowedAlias
	}
	composeSubject.value = /^re:/i.test(selected.value.subject) ? selected.value.subject : `Re: ${selected.value.subject}`
	composeBody.value = ''
	openComposer(parentEntryID)
}

function closeReader() {
	selected.value = null
	composing.value = false
	composeTo.value = ''
	composeSubject.value = ''
	composeBody.value = ''
	composeParentEntryID.value = ''
}

async function submitMail() {
	submitting.value = true
	actionError.value = ''
	try {
		const sent = teamMode.value
			? await sendManagementMail(composeFrom.value, composeTo.value, composeSubject.value, composeBody.value, composeParentEntryID.value)
			: await sendMail(composeTo.value, composeSubject.value, composeBody.value, composeParentEntryID.value)
		composeTo.value = ''; composeSubject.value = ''; composeBody.value = ''
		composeParentEntryID.value = ''
		folder.value = 'Sent'
		composing.value = false
		selected.value = sent
		await loadMailbox()
	} catch (reason) {
		actionError.value = reason instanceof Error ? reason.message : t('mail.sendFailed')
	} finally {
		submitting.value = false
	}
}

async function applyAction(action: 'archive' | 'trash' | 'read' | 'unread') {
	if (!selected.value) return
	actionError.value = ''
	try {
		const updated = await mutateEntry(selected.value.entryId, action)
		if (action === 'read' || action === 'unread') selected.value = updated
		else selected.value = null
		await loadMailbox()
	} catch (reason) {
		actionError.value = reason instanceof Error ? reason.message : t('mail.actionFailed')
	}
}

onMounted(async () => {
	await auth.initialize()
	await loadMailbox()
	refreshTimer = window.setInterval(refreshMailboxState, 15_000)
})
onUnmounted(() => { if (refreshTimer !== undefined) window.clearInterval(refreshTimer) })
watch(() => props.mode, async () => {
	mailbox.value = null; selected.value = null; composing.value = false; folder.value = 'Inbox'; query.value = ''; composeFrom.value = ''
	await loadMailbox()
})
watch(() => auth.account?.id, async () => { mailbox.value = null; selected.value = null; await loadMailbox() })
</script>

<template>
	<section class="mail-workspace" :class="{ 'reader-open': composing || selected }" :aria-label="teamMode ? t('mail.team') : 'Wave Mail'">
		<aside class="mail-folders">
			<header><strong>{{ teamMode ? t('mail.team') : auth.account?.email }}</strong><small v-if="teamMode">{{ mailbox?.addresses?.join(' · ') }}</small></header>
			<button class="mail-compose-button" type="button" @click="openComposer()"><PenLine :size="17" aria-hidden="true" />{{ t('mail.compose') }}</button>
			<nav :aria-label="t('mail.folders')">
				<button v-for="item in folders" :key="item.id" type="button" :class="{ active: folder === item.id }" @click="selectFolder(item.id)">
					<component :is="item.icon" :size="17" aria-hidden="true" />{{ item.label }}
				</button>
			</nav>
		</aside>

		<section class="mail-list" :aria-label="t('mail.messages')">
			<header class="mail-list-header">
				<div><h1>{{ folders.find((item) => item.id === folder)?.label }}</h1><button type="button" :title="t('mail.refresh')" :aria-label="t('mail.refresh')" @click="loadMailbox"><RefreshCw :size="16" /></button></div>
				<form role="search" @submit.prevent="loadMailbox"><Search :size="15" aria-hidden="true" /><input v-model="query" type="search" :placeholder="t('mail.search')" :aria-label="t('mail.search')" /></form>
			</header>
			<UiInlineState v-if="loading" :message="t('common.loading')" />
			<UiInlineState v-else-if="error" :message="error" />
			<div v-else-if="mailbox?.items.length === 0" class="mail-empty"><strong>{{ query ? t('mail.noSearchResults') : t('mail.emptyTitle') }}</strong><span>{{ query ? t('mail.changeSearch') : t('mail.emptyDetail') }}</span></div>
			<div v-else class="mail-rows">
				<button v-for="item in mailbox?.items ?? []" :key="item.id" class="mail-row" :class="{ selected: selected?.entryId === item.id, unread: item.flags.includes('unread') }" type="button" @click="openMessage(item)">
					<span class="mail-row-author">{{ folder === 'Sent' ? item.to.join(', ') : displayAddress(item.from) }}</span>
					<span class="mail-row-copy"><strong>{{ item.subject }}</strong><small>{{ item.preview }}</small></span>
					<span class="mail-row-end"><small v-if="folder === 'Sent' && item.deliveryStatus" :class="`delivery-${item.deliveryStatus}`">{{ deliveryLabel(item.deliveryStatus) }}</small><time :datetime="item.receivedAt">{{ formatDate(item.receivedAt) }}</time></span>
				</button>
			</div>
		</section>

		<section class="mail-reader" :aria-label="t('mail.reader')">
			<header class="mail-reader-toolbar">
				<button type="button" :disabled="!selected" :title="t('mail.reply')" @click="replyToSelected"><Reply :size="16" />{{ t('mail.reply') }}</button>
				<button type="button" :disabled="!selected" :title="t('mail.archive')" @click="applyAction('archive')"><Archive :size="16" />{{ t('mail.archive') }}</button>
				<button type="button" :disabled="!selected" :title="t('mail.delete')" @click="applyAction('trash')"><Trash2 :size="16" />{{ t('mail.delete') }}</button>
				<button type="button" :disabled="!selected" :title="t('mail.markUnread')" @click="applyAction('unread')"><MailOpen :size="16" />{{ t('mail.markUnread') }}</button>
				<button v-if="composing || selected" class="mail-reader-close" type="button" :aria-label="t('common.cancel')" @click="closeReader"><X :size="17" /></button>
			</header>

			<UiInlineState v-if="readerLoading" :message="t('common.loading')" />
			<form v-else-if="composing" class="mail-composer" @submit.prevent="submitMail">
				<h1>{{ teamMode ? t('mail.newTeamMessage') : t('mail.newMessage') }}</h1>
				<label v-if="teamMode">{{ t('mail.from') }}<select v-model="composeFrom" required><option v-for="address in mailbox?.addresses ?? []" :key="address" :value="address">{{ address }}</option></select></label>
				<label>{{ t('mail.to') }}<input v-model="composeTo" required type="text" inputmode="email" :placeholder="t('mail.toPlaceholder')" /></label>
				<GmailDeliveryWarning v-if="gmailRecipient" />
				<label>{{ t('mail.subject') }}<input v-model="composeSubject" required maxlength="180" /></label>
				<label class="mail-compose-body"><span class="sr-only">{{ t('mail.body') }}</span><textarea v-model="composeBody" required maxlength="50000" rows="16" /></label>
				<p v-if="actionError" class="mail-action-error" role="alert">{{ actionError }}</p>
				<footer><button class="ui-button primary" type="submit" :disabled="submitting">{{ t('mail.sendMessage') }}</button></footer>
			</form>

			<article v-else-if="selected" class="mail-message-view">
				<header><h1>{{ selected.subject }}</h1><div><strong>{{ displayAddress(selected.from) }}</strong><span v-if="folder === 'Sent' && selected.deliveryStatus" class="mail-delivery-status" :class="`delivery-${selected.deliveryStatus}`">{{ deliveryLabel(selected.deliveryStatus) }}</span><time :datetime="selected.date">{{ formatDate(selected.date) }}</time></div><p>{{ t('mail.to') }} {{ selected.to.join(', ') }}</p></header>
				<pre>{{ selected.body }}</pre>
				<p v-if="actionError" class="mail-action-error" role="alert">{{ actionError }}</p>
			</article>

			<div v-else class="mail-reader-empty"><strong>{{ t('mail.selectTitle') }}</strong><span>{{ t('mail.selectDetail') }}</span></div>
		</section>
	</section>
</template>
