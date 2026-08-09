<script setup lang="ts">
import {
  Archive, Inbox, Mail, MailOpen, PenLine, RefreshCw, Search, Send, Trash2, X,
} from '@lucide/vue'
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'

import { useI18n } from '../i18n'
import {
  getMailbox, getMailMessage, sendMail, updateMailEntry,
  type MailboxItem, type MailboxView, type MailMessageView,
} from '../services/http'
import { useAuthStore } from '../stores/auth'
import UiInlineState from '../ui/UiInlineState.vue'

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
const composeTo = ref('')
const composeSubject = ref('')
const composeBody = ref('')
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

function formatDate(value: string) {
  if (!value) return ''
  return new Intl.DateTimeFormat(locale.value === 'ko' ? 'ko-KR' : 'en-US', {
    month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit',
		timeZone: auth.account?.timeZone || undefined,
  }).format(new Date(value))
}

function deliveryLabel(status: string) {
  if (!status) return ''
  return t(`mail.delivery.${status}`)
}

async function loadMailbox() {
  if (!auth.account) return
  loading.value = true
  error.value = ''
  try { mailbox.value = await getMailbox(folder.value, query.value.trim()) }
  catch (reason) { error.value = reason instanceof Error ? reason.message : t('common.loadError') }
  finally { loading.value = false }
}

async function refreshMailboxState() {
  if (!auth.account || composing.value || submitting.value || document.hidden) return
  try {
    mailbox.value = await getMailbox(folder.value, query.value.trim())
    if (selected.value) selected.value = await getMailMessage(selected.value.entryId)
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
    selected.value = await getMailMessage(item.id)
    if (item.flags.includes('unread')) {
      selected.value = await updateMailEntry(item.id, 'read')
      item.flags = item.flags.filter((flag) => flag !== 'unread')
    }
  } catch (reason) {
    actionError.value = reason instanceof Error ? reason.message : t('common.loadError')
  } finally { readerLoading.value = false }
}

function openComposer() {
  selected.value = null
  composing.value = true
  actionError.value = ''
}

function closeReader() {
  selected.value = null
  composing.value = false
}

async function submitMail() {
  submitting.value = true
  actionError.value = ''
  try {
    const sent = await sendMail(composeTo.value, composeSubject.value, composeBody.value)
    composeTo.value = ''; composeSubject.value = ''; composeBody.value = ''
    folder.value = 'Sent'
    composing.value = false
    selected.value = sent
    await loadMailbox()
  } catch (reason) {
    actionError.value = reason instanceof Error ? reason.message : t('mail.sendFailed')
  } finally { submitting.value = false }
}

async function applyAction(action: 'archive' | 'trash' | 'read' | 'unread') {
  if (!selected.value) return
  actionError.value = ''
  try {
    const updated = await updateMailEntry(selected.value.entryId, action)
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
watch(() => auth.account?.id, async () => { mailbox.value = null; selected.value = null; await loadMailbox() })
</script>

<template>
  <main class="mail-service">
    <section v-if="auth.initialized && !auth.account" class="mail-signin" aria-labelledby="mail-title">
      <Mail :size="34" :stroke-width="1.5" aria-hidden="true" />
      <h1 id="mail-title">Wave Mail</h1>
      <p>{{ t('mail.signInRequired') }}</p>
      <RouterLink class="ui-button primary" to="/login">{{ t('mail.signIn') }}</RouterLink>
      <RouterLink class="mail-create-account" to="/register">{{ t('auth.signUp') }}</RouterLink>
    </section>

    <section v-else-if="auth.account" class="mail-workspace" :class="{ 'reader-open': composing || selected }" aria-label="Wave Mail">
      <aside class="mail-folders">
        <header><strong>{{ auth.account.email }}</strong></header>
        <button class="mail-compose-button" type="button" @click="openComposer"><PenLine :size="17" aria-hidden="true" />{{ t('mail.compose') }}</button>
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
          <button type="button" :disabled="!selected" :title="t('mail.archive')" @click="applyAction('archive')"><Archive :size="16" />{{ t('mail.archive') }}</button>
          <button type="button" :disabled="!selected" :title="t('mail.delete')" @click="applyAction('trash')"><Trash2 :size="16" />{{ t('mail.delete') }}</button>
          <button type="button" :disabled="!selected" :title="t('mail.markUnread')" @click="applyAction('unread')"><MailOpen :size="16" />{{ t('mail.markUnread') }}</button>
          <button v-if="composing || selected" class="mail-reader-close" type="button" :aria-label="t('common.cancel')" @click="closeReader"><X :size="17" /></button>
        </header>

        <UiInlineState v-if="readerLoading" :message="t('common.loading')" />
        <form v-else-if="composing" class="mail-composer" @submit.prevent="submitMail">
          <h1>{{ t('mail.newMessage') }}</h1>
          <label>{{ t('mail.to') }}<input v-model="composeTo" required type="text" inputmode="email" :placeholder="t('mail.toPlaceholder')" /></label>
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
  </main>
</template>
