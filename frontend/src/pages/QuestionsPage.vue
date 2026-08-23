<script setup lang="ts">
import { ArrowDown, ArrowUp, Check, Search } from '@lucide/vue'
import { computed, onMounted, ref, watch, watchEffect } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { useI18n } from '../i18n'
import PlatformWaveEditor from '../components/editor/PlatformWaveEditor.vue'
import {
  acceptQuestionAnswer, createQuestion, createQuestionAnswer, getQuestion, getQuestions, voteQuestion,
  type QuestionMessage, type QuestionSummary, type QuestionView,
} from '../services/http'
import { useAuthStore } from '../stores/auth'
import { applyPageSEO, plainText, plainTextDescription } from '../services/seo'
import UiInlineState from '../ui/UiInlineState.vue'

const { locale, t } = useI18n()
const auth = useAuthStore()
const route = useRoute()
const router = useRouter()
const questions = ref<QuestionSummary[]>([])
const question = ref<QuestionView | null>(null)
const loading = ref(false)
const submitting = ref(false)
const error = ref('')
const actionError = ref('')
const filter = ref<'newest' | 'unanswered' | 'active'>('newest')
const query = ref('')
const title = ref('')
const body = ref('')
const tags = ref('')
const waveVersion = ref('')
const platform = ref('')
const answerBody = ref('')

const mode = computed(() => route.name === 'question-new' ? 'new' : route.name === 'question-detail' ? 'detail' : 'list')
const canAccept = computed(() => Boolean(question.value && auth.account &&
  (auth.account.administrator || auth.account.id === question.value.root.authorAccountId)))

function displayAuthor(value: string) {
  const label = value.replace(/\s*<[^>]+>\s*$/, '').trim()
  return label.replace(/^"(.*)"$/, '$1') || value
}

function authorEmail(value: string) {
  const bracketed = value.match(/<([^>]+)>/)?.[1]
  const address = (bracketed ?? value).trim().replace(/^"|"$/g, '')
  return address.includes('@') ? address : ''
}

function authorProfile(value: string, accountID = '') {
  if (accountID) return `/user/id/${encodeURIComponent(accountID)}`
  const localPart = authorEmail(value).split('@')[0]
  return localPart ? `/user/${encodeURIComponent(localPart)}` : ''
}

function formatDate(value: string) {
  if (!value) return ''
  return new Intl.DateTimeFormat(locale.value === 'ko' ? 'ko-KR' : 'en-US', {
    year: 'numeric', month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit',
		timeZone: auth.account?.timeZone || undefined,
  }).format(new Date(value))
}

function statusLabel(status: string) { return t(`questions.status.${status || 'open'}`) }

async function loadList() {
  loading.value = true
  error.value = ''
  try { questions.value = await getQuestions({ sort: filter.value, q: query.value.trim() }) }
  catch (reason) { error.value = reason instanceof Error ? reason.message : t('common.loadError') }
  finally { loading.value = false }
}

async function loadDetail() {
  const id = String(route.params.question ?? '')
  if (!id) return
  loading.value = true
  error.value = ''
  try { question.value = await getQuestion(id) }
  catch (reason) { error.value = reason instanceof Error ? reason.message : t('common.loadError') }
  finally { loading.value = false }
}

async function loadPage() {
  actionError.value = ''
  question.value = null
  if (mode.value === 'list') {
    const requestedFilter = String(route.query.sort ?? '')
    filter.value = requestedFilter === 'unanswered' || requestedFilter === 'active' ? requestedFilter : 'newest'
    query.value = String(route.query.q ?? '')
    await loadList()
  }
  else if (mode.value === 'detail') await loadDetail()
}

async function setFilter(value: 'newest' | 'unanswered' | 'active') {
  filter.value = value
  await loadList()
}

async function submitQuestion() {
  if (!auth.account) { await router.push('/login'); return }
  submitting.value = true
  actionError.value = ''
  try {
    const created = await createQuestion({
      title: title.value, body: body.value,
      tags: tags.value.split(',').map((tag) => tag.trim()).filter(Boolean),
      waveVersion: waveVersion.value, platform: platform.value,
    })
    await router.push(`/questions/${created.id}`)
  } catch (reason) { actionError.value = reason instanceof Error ? reason.message : t('questions.createFailed') }
  finally { submitting.value = false }
}

async function submitAnswer() {
  if (!question.value) return
  if (!auth.account) { await router.push('/login'); return }
  submitting.value = true
  actionError.value = ''
  try {
    question.value = await createQuestionAnswer(question.value.id, answerBody.value)
    answerBody.value = ''
  } catch (reason) { actionError.value = reason instanceof Error ? reason.message : t('questions.answerFailed') }
  finally { submitting.value = false }
}

async function castVote(targetType: 'question' | 'answer', target: QuestionMessage) {
  if (!question.value) return
  if (!auth.account) { await router.push('/login'); return }
  actionError.value = ''
  try {
    const current = targetType === 'question' ? question.value.viewerVote : target.viewerVote
    const desired = (current === 1 ? 0 : 1) as 0 | 1
    const result = await voteQuestion(question.value.id, targetType, targetType === 'question' ? question.value.id : target.id, desired)
    target.score = result.score
    target.viewerVote = result.viewerVote
    if (targetType === 'question') { question.value.score = result.score; question.value.viewerVote = result.viewerVote }
  } catch (reason) { actionError.value = reason instanceof Error ? reason.message : t('questions.voteFailed') }
}

async function downVote(targetType: 'question' | 'answer', target: QuestionMessage) {
  if (!question.value) return
  if (!auth.account) { await router.push('/login'); return }
  actionError.value = ''
  try {
    const current = targetType === 'question' ? question.value.viewerVote : target.viewerVote
    const desired = (current === -1 ? 0 : -1) as -1 | 0
    const result = await voteQuestion(question.value.id, targetType, targetType === 'question' ? question.value.id : target.id, desired)
    target.score = result.score
    target.viewerVote = result.viewerVote
    if (targetType === 'question') { question.value.score = result.score; question.value.viewerVote = result.viewerVote }
  } catch (reason) { actionError.value = reason instanceof Error ? reason.message : t('questions.voteFailed') }
}

async function toggleAccepted(answer: QuestionMessage) {
  if (!question.value || !canAccept.value) return
  actionError.value = ''
  try { question.value = await acceptQuestionAnswer(question.value.id, answer.accepted ? '' : answer.id) }
  catch (reason) { actionError.value = reason instanceof Error ? reason.message : t('questions.acceptFailed') }
}

onMounted(async () => { await auth.initialize(); await loadPage() })
watch(() => route.fullPath, loadPage)
watch(() => auth.account?.id, async () => { if (mode.value === 'detail') await loadDetail() })
watchEffect(() => {
  if (!question.value) return
  const answers = question.value.answers.map((answer) => ({
    '@type': 'Answer',
    text: plainText(answer.body),
    dateCreated: answer.createdAt,
    upvoteCount: answer.score,
    author: { '@type': 'Person', name: displayAuthor(answer.author) },
  }))
  const accepted = question.value.answers.findIndex((answer) => answer.accepted)
  const answerProperty = accepted >= 0
    ? { acceptedAnswer: answers[accepted], suggestedAnswer: answers.filter((_, index) => index !== accepted) }
    : answers.length ? { suggestedAnswer: answers } : {}
  applyPageSEO({
    title: `${question.value.title} · Wave Questions`,
    description: plainTextDescription(question.value.root.body, question.value.title),
    locale: locale.value,
    path: route.path,
    schema: {
      '@type': 'QAPage',
      mainEntity: {
        '@type': 'Question',
        name: question.value.title,
        text: plainText(question.value.root.body),
        dateCreated: question.value.root.createdAt,
        answerCount: answers.length,
        upvoteCount: Math.max(0, question.value.score),
        author: { '@type': 'Person', name: displayAuthor(question.value.root.author) },
        ...answerProperty,
      },
    },
  })
})
</script>

<template>
  <main class="questions-service questions-width">
    <template v-if="mode === 'list'">
      <header class="questions-titlebar">
        <div><h1>{{ t('questions.title') }}</h1></div>
        <RouterLink class="ui-button primary" :to="auth.account ? '/questions/new' : '/login'">{{ t('questions.ask') }}</RouterLink>
      </header>
      <div class="questions-list-tools">
        <form role="search" @submit.prevent="loadList"><input v-model="query" type="search" :placeholder="t('questions.search')" /><button type="submit"><Search :size="15" /><span>{{ t('questions.search') }}</span></button></form>
        <nav class="ui-tabs question-filters" :aria-label="t('questions.filters')">
          <button type="button" :class="{ active: filter === 'newest' }" @click="setFilter('newest')">{{ t('questions.newest') }}</button>
          <button type="button" :class="{ active: filter === 'unanswered' }" @click="setFilter('unanswered')">{{ t('questions.unanswered') }}</button>
          <button type="button" :class="{ active: filter === 'active' }" @click="setFilter('active')">{{ t('questions.active') }}</button>
        </nav>
      </div>
      <UiInlineState v-if="loading" :message="t('common.loading')" />
      <UiInlineState v-else-if="error" :message="error" />
      <section v-else class="question-list" aria-live="polite">
        <p v-if="questions.length === 0" class="compact-empty">{{ t('questions.empty') }}</p>
        <article v-for="item in questions" v-else :key="item.id" class="question-row">
          <div class="question-row-stats"><span><strong>{{ item.score }}</strong>{{ t('questions.votes') }}</span><span :class="{ accepted: item.accepted }"><strong>{{ item.answerCount }}</strong>{{ t('questions.answers') }}</span><span><strong>{{ item.viewCount }}</strong>{{ t('questions.views') }}</span></div>
          <div class="question-row-copy">
            <RouterLink :to="`/questions/${item.id}`"><h2>{{ item.title }}</h2></RouterLink>
            <p>{{ item.excerpt }}</p>
            <div class="question-tags"><span v-for="tag in item.tags" :key="tag">{{ tag }}</span><small v-if="item.waveVersion">Wave {{ item.waveVersion }}</small></div>
            <footer><RouterLink v-if="authorProfile(item.author, item.authorAccountId)" class="author-profile-link" :to="authorProfile(item.author, item.authorAccountId)" :title="authorEmail(item.author)">{{ displayAuthor(item.author) }}</RouterLink><span v-else>{{ displayAuthor(item.author) }}</span><time :datetime="item.lastActivityAt">{{ formatDate(item.lastActivityAt) }}</time><em>{{ statusLabel(item.status) }}</em></footer>
          </div>
        </article>
      </section>
    </template>

    <form v-else-if="mode === 'new'" class="question-editor" @submit.prevent="submitQuestion">
      <header><h1>{{ t('questions.ask') }}</h1><RouterLink to="/questions">{{ t('common.cancel') }}</RouterLink></header>
      <label><span>{{ t('questions.questionTitle') }}</span><input v-model="title" required minlength="10" maxlength="180" /></label>
		<PlatformWaveEditor v-model="body" :label="t('questions.details')" required :min-length="20" :max-length="30000" :rows="15" />
      <div class="question-editor-meta">
        <label><span>{{ t('questions.tagsLabel') }}</span><input v-model="tags" required :placeholder="t('questions.tagsPlaceholder')" /></label>
        <label><span>{{ t('questions.waveVersion') }}</span><input v-model="waveVersion" maxlength="40" placeholder="0.2.0-pre-beta" /></label>
        <label><span>{{ t('questions.platform') }}</span><input v-model="platform" maxlength="80" placeholder="Linux x86_64" /></label>
      </div>
      <p v-if="actionError" class="question-action-error" role="alert">{{ actionError }}</p>
      <footer><button class="ui-button primary" type="submit" :disabled="submitting">{{ t('questions.publish') }}</button></footer>
    </form>

    <template v-else>
      <UiInlineState v-if="loading" :message="t('common.loading')" />
      <UiInlineState v-else-if="error" :message="error" />
      <article v-else-if="question" class="question-detail">
        <header class="question-detail-heading">
          <RouterLink to="/questions">{{ t('questions.title') }}</RouterLink>
          <h1>{{ question.title }}</h1>
          <div><span>{{ formatDate(question.root.createdAt) }}</span><span>{{ question.viewCount }} {{ t('questions.views') }}</span><span v-if="question.waveVersion">Wave {{ question.waveVersion }}</span><span v-if="question.platform">{{ question.platform }}</span><em>{{ statusLabel(question.status) }}</em></div>
        </header>
        <section class="question-post">
          <aside class="question-vote" :aria-label="t('questions.vote')"><button type="button" :class="{ active: question.viewerVote === 1 }" @click="castVote('question', question.root)"><ArrowUp :size="22" /></button><strong>{{ question.score }}</strong><button type="button" :class="{ active: question.viewerVote === -1 }" @click="downVote('question', question.root)"><ArrowDown :size="22" /></button></aside>
          <div class="question-post-content"><pre>{{ question.root.body }}</pre><div class="question-tags"><span v-for="tag in question.tags" :key="tag">{{ tag }}</span></div><footer><RouterLink v-if="authorProfile(question.root.author, question.root.authorAccountId)" class="author-profile-link" :to="authorProfile(question.root.author, question.root.authorAccountId)" :title="authorEmail(question.root.author)">{{ displayAuthor(question.root.author) }}</RouterLink><span v-else>{{ displayAuthor(question.root.author) }}</span><span>· {{ formatDate(question.root.createdAt) }}</span></footer></div>
        </section>
        <h2 class="question-answer-count">{{ question.answers.length }} {{ t('questions.answers') }}</h2>
        <section v-for="answer in question.answers" :key="answer.id" class="question-post question-answer" :class="{ accepted: answer.accepted }">
          <aside class="question-vote"><button type="button" :class="{ active: answer.viewerVote === 1 }" @click="castVote('answer', answer)"><ArrowUp :size="22" /></button><strong>{{ answer.score }}</strong><button type="button" :class="{ active: answer.viewerVote === -1 }" @click="downVote('answer', answer)"><ArrowDown :size="22" /></button><button v-if="answer.accepted || canAccept" class="question-accept" :class="{ active: answer.accepted }" type="button" :disabled="!canAccept" :title="answer.accepted ? t('questions.accepted') : t('questions.accept')" @click="toggleAccepted(answer)"><Check :size="22" /></button></aside>
          <div class="question-post-content"><pre>{{ answer.body }}</pre><footer><RouterLink v-if="authorProfile(answer.author, answer.authorAccountId)" class="author-profile-link" :to="authorProfile(answer.author, answer.authorAccountId)" :title="authorEmail(answer.author)">{{ displayAuthor(answer.author) }}</RouterLink><span v-else>{{ displayAuthor(answer.author) }}</span><span>· {{ formatDate(answer.createdAt) }}</span><strong v-if="answer.accepted">{{ t('questions.accepted') }}</strong></footer></div>
        </section>
        <p v-if="actionError" class="question-action-error" role="alert">{{ actionError }}</p>
		<form v-if="auth.account && question.status !== 'closed'" class="question-answer-editor" @submit.prevent="submitAnswer"><h2>{{ t('questions.yourAnswer') }}</h2><PlatformWaveEditor v-model="answerBody" required :max-length="20000" :rows="10" /><button class="ui-button primary" type="submit" :disabled="submitting">{{ t('questions.postAnswer') }}</button></form>
        <p v-else-if="!auth.account" class="question-login-note"><RouterLink to="/login">{{ t('mail.signIn') }}</RouterLink> {{ t('questions.toAnswer') }}</p>
      </article>
    </template>
  </main>
</template>
