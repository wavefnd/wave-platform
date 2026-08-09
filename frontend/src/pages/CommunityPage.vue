<script setup lang="ts">
import {
  ArrowBigDown, ArrowBigUp, Bell, BellOff, Code2, Eye, ImagePlus, Link, MessageSquare, Quote, Search,
} from '@lucide/vue'
import { computed, nextTick, ref, watch, watchEffect } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import MarkdownContent from '../components/MarkdownContent.vue'
import CommunityReplyEditor from '../components/community/CommunityReplyEditor.vue'
import { useI18n } from '../i18n'
import {
  createCommunityPost, createCommunityReply, getCommunitySpaces, getCommunityThread,
  getCommunityThreads, subscribeCommunity, uploadLunaStevImage, voteCommunity,
  type CommunityMessage, type CommunitySpace, type CommunityThread, type CommunityThreadSummary,
} from '../services/http'
import { useAuthStore } from '../stores/auth'
import { applyPageSEO, plainText, plainTextDescription } from '../services/seo'
import UiInlineState from '../ui/UiInlineState.vue'
import UiSkeletonRows from '../ui/UiSkeletonRows.vue'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const { locale, t } = useI18n()

const spaces = ref<CommunitySpace[]>([])
const threads = ref<CommunityThreadSummary[]>([])
const thread = ref<CommunityThread | null>(null)
const loading = ref(true)
const loadingMore = ref(false)
const failed = ref(false)
const hasMore = ref(false)
const submitting = ref(false)
const actionError = ref('')
const searchText = ref('')

const postSpace = ref('general')
const postTitle = ref('')
const postBody = ref('')
const postTags = ref('')
const postPreview = ref(false)
const postBodyInput = ref<HTMLTextAreaElement | null>(null)
const postImageInput = ref<HTMLInputElement | null>(null)
const uploadingImage = ref(false)

const replyBody = ref('')
const replyTo = ref('')
const replyPreview = ref(false)
const replyEditor = ref<InstanceType<typeof CommunityReplyEditor> | null>(null)

const activeSpace = computed(() => String(route.query.space ?? ''))
const threadID = computed(() => String(route.params.thread ?? ''))
const personalMode = computed(() => route.meta.personalSpace === true)
const isComposer = computed(() => route.name === 'community-new' || route.name === 'personal-space-new')
const rootRouteName = computed(() => personalMode.value ? 'personal-space' : 'community')
const composerRouteName = computed(() => personalMode.value ? 'personal-space-new' : 'community-new')
const threadRouteName = computed(() => personalMode.value ? 'personal-space-thread' : 'community-thread')
const rootPath = computed(() => personalMode.value ? '/lunastev' : '/community')
const activeSort = computed(() => ['latest', 'active', 'top'].includes(String(route.query.sort)) ? String(route.query.sort) : 'latest')
const activeQuery = computed(() => String(route.query.q ?? ''))
const postableSpaces = computed(() => spaces.value.filter((space) =>
  space.postingPolicy !== 'owner' || Boolean(auth.account?.owner)))
const personalSpaces = computed(() => {
  const order = new Map([['founder-notes', 0], ['development-log', 1]])
  return spaces.value.filter((space) => space.postingPolicy === 'owner')
    .sort((left, right) => (order.get(left.id) ?? 99) - (order.get(right.id) ?? 99))
})
const memberSpaces = computed(() => spaces.value.filter((space) => space.postingPolicy !== 'owner'))
const modeSpaces = computed(() => personalMode.value ? personalSpaces.value : memberSpaces.value)
const selectedSpaceID = computed(() => modeSpaces.value.some((space) => space.id === activeSpace.value) ? activeSpace.value : '')
const currentSpaceID = computed(() => selectedSpaceID.value || thread.value?.spaceId || '')
const activeSpacePolicy = computed(() => spaces.value.find((space) => space.id === currentSpaceID.value)?.postingPolicy ?? 'members')
const canOpenComposer = computed(() => {
  if (threadID.value && !thread.value) return false
  if (personalMode.value) return Boolean(auth.account?.owner)
  return activeSpacePolicy.value !== 'owner' || Boolean(auth.account?.owner)
})
const visibleThreads = computed(() => {
  const personalIDs = new Set(personalSpaces.value.map((space) => space.id))
  return threads.value.filter((item) => personalIDs.has(item.spaceId) === personalMode.value)
})
const composerSpaces = computed(() => postableSpaces.value.filter((space) =>
  (space.postingPolicy === 'owner') === personalMode.value))
const showTopics = computed(() => !personalMode.value && topicCounts.value.length > 0)
const topicCounts = computed(() => {
  const counts = new Map<string, number>()
  for (const item of visibleThreads.value) for (const tag of item.tags) counts.set(tag, (counts.get(tag) ?? 0) + 1)
  return Array.from(counts, ([name, count]) => ({ name, count }))
    .sort((left, right) => right.count - left.count || left.name.localeCompare(right.name)).slice(0, 10)
})
const orderedReplies = computed(() => {
  if (!thread.value) return []
  const replies = thread.value.replies
  const replyIDs = new Set(replies.map((reply) => reply.id))
  const children = new Map<string, CommunityMessage[]>()
  for (const reply of replies) {
    const parentID = replyIDs.has(reply.parentMessageId) ? reply.parentMessageId : thread.value.root.id
    const group = children.get(parentID) ?? []
    group.push(reply)
    children.set(parentID, group)
  }
  for (const group of children.values()) group.sort((left, right) =>
    new Date(left.createdAt).getTime() - new Date(right.createdAt).getTime())

  const ordered: Array<{ message: CommunityMessage; depth: number }> = []
  const visited = new Set<string>()
  function append(parentID: string, depth: number) {
    for (const message of children.get(parentID) ?? []) {
      if (visited.has(message.id)) continue
      visited.add(message.id)
      ordered.push({ message, depth })
      append(message.id, depth + 1)
    }
  }
  append(thread.value.root.id, 0)
  for (const message of replies) {
    if (visited.has(message.id)) continue
    ordered.push({ message, depth: 0 })
  }
  return ordered
})

function spaceName(space: CommunitySpace) {
  const localized = t(`community.space.${space.slug}`)
  return localized.startsWith('community.space.') ? space.name : localized
}

function selectedSpaceName(spaceID: string) {
  const space = spaces.value.find((item) => item.id === spaceID)
  return space ? spaceName(space) : spaceID
}

function authorName(author: string) {
  const label = author.replace(/\s*<[^>]+>\s*$/, '').trim()
  return label.replace(/^"(.*)"$/, '$1') || author
}

function authorInitial(author: string) {
  return Array.from(authorName(author))[0]?.toUpperCase() ?? 'W'
}

function formatDate(value: string) {
  if (!value) return ''
  return new Intl.DateTimeFormat(locale.value === 'ko' ? 'ko-KR' : 'en-US', {
    year: 'numeric', month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit',
  }).format(new Date(value))
}

function relativeDate(value: string) {
  if (!value) return ''
  const seconds = Math.round((new Date(value).getTime() - Date.now()) / 1000)
  const formatter = new Intl.RelativeTimeFormat(locale.value === 'ko' ? 'ko' : 'en', { numeric: 'auto' })
  const ranges: [Intl.RelativeTimeFormatUnit, number][] = [['year', 31536000], ['month', 2592000], ['day', 86400], ['hour', 3600], ['minute', 60]]
  for (const [unit, size] of ranges) if (Math.abs(seconds) >= size) return formatter.format(Math.round(seconds / size), unit)
  return formatter.format(seconds, 'second')
}

function replyIndent(depth: number) { return `${8 + Math.min(depth, 4) * 28}px` }

function routeQuery(changes: Record<string, string | undefined>) {
  const query = { ...route.query, ...changes }
  for (const [key, value] of Object.entries(query)) if (!value) delete query[key]
  return query
}

function selectSpace(spaceID: string) {
  router.push({ name: rootRouteName.value, query: routeQuery({ space: spaceID || undefined }) })
}

function selectSort(sort: string) {
  router.push({ name: rootRouteName.value, query: routeQuery({ sort: sort === 'latest' ? undefined : sort }) })
}

function search() {
  router.push({ name: rootRouteName.value, query: routeQuery({ q: searchText.value.trim() || undefined }) })
}

function openComposer() {
  if (!auth.account) {
    router.push({ name: 'login', query: { redirect: `${rootPath.value}/new` } })
    return
  }
  const selected = spaces.value.find((space) => space.id === currentSpaceID.value)
  const canUseSelected = selected && (selected.postingPolicy !== 'owner' || auth.account.owner)
  router.push({ name: composerRouteName.value, query: canUseSelected ? { space: selected.id } : {} })
}

async function load() {
  loading.value = true
  failed.value = false
  hasMore.value = false
  actionError.value = ''
  thread.value = null
  searchText.value = activeQuery.value
  try {
    await auth.initialize()
    spaces.value = await getCommunitySpaces()
    if (selectedSpaceID.value && composerSpaces.value.some((space) => space.id === selectedSpaceID.value)) postSpace.value = selectedSpaceID.value
    else if (!composerSpaces.value.some((space) => space.id === postSpace.value)) postSpace.value = composerSpaces.value[0]?.id ?? ''
    if (threadID.value) {
      const [selected, latest] = await Promise.all([getCommunityThread(threadID.value), getCommunityThreads('', { sort: 'active', limit: 30 })])
      const selectedSpace = spaces.value.find((space) => space.id === selected.spaceId)
      if (!selectedSpace || (selectedSpace.postingPolicy === 'owner') !== personalMode.value) throw new Error('thread belongs to another space')
      thread.value = selected
      threads.value = latest
    } else {
      threads.value = await getCommunityThreads(selectedSpaceID.value, { sort: activeSort.value, q: activeQuery.value, limit: 30 })
      hasMore.value = threads.value.length === 30
    }
  } catch {
    failed.value = true
  } finally {
    loading.value = false
  }
}

async function loadMore() {
  if (!hasMore.value || loading.value || loadingMore.value) return
  loadingMore.value = true
  try {
    const next = await getCommunityThreads(selectedSpaceID.value, {
      sort: activeSort.value, q: activeQuery.value, limit: 30, offset: threads.value.length,
    })
    threads.value.push(...next)
    hasMore.value = next.length === 30
  } finally { loadingMore.value = false }
}

async function submitPost() {
  if (!auth.account) { openComposer(); return }
  submitting.value = true
  actionError.value = ''
  try {
    const created = await createCommunityPost({
      spaceId: postSpace.value, title: postTitle.value, body: postBody.value,
      tags: postTags.value.split(',').map((tag) => tag.trim()).filter(Boolean),
    })
    postTitle.value = ''; postBody.value = ''; postTags.value = ''
    await router.replace({ name: threadRouteName.value, params: { thread: created.id } })
  } catch (reason) {
    actionError.value = reason instanceof Error ? reason.message : t('community.actionFailed')
  } finally { submitting.value = false }
}

async function submitReply() {
  if (!thread.value || !auth.account) {
    await router.push({ name: 'login', query: { redirect: route.fullPath } })
    return
  }
  submitting.value = true
  actionError.value = ''
  try {
    thread.value = await createCommunityReply(thread.value.id, replyBody.value, replyTo.value)
    replyBody.value = ''; replyTo.value = ''; replyPreview.value = false
  } catch (reason) {
    actionError.value = reason instanceof Error ? reason.message : t('community.actionFailed')
  } finally { submitting.value = false }
}

async function beginReply(message: CommunityMessage) {
  if (!auth.account) {
    router.push({ name: 'login', query: { redirect: route.fullPath } })
    return
  }
  replyTo.value = message.id
  replyPreview.value = false
  await nextTick()
  replyEditor.value?.focus()
}

async function voteSummary(item: CommunityThreadSummary, value: -1 | 1) {
  if (!auth.account) { await router.push({ name: 'login', query: { redirect: route.fullPath } }); return }
  const next = item.viewerVote === value ? 0 : value
  try {
    const result = await voteCommunity(item.id, 'thread', item.id, next as -1 | 0 | 1)
    item.score = result.score; item.viewerVote = result.viewerVote
  } catch (reason) { actionError.value = reason instanceof Error ? reason.message : t('community.actionFailed') }
}

async function voteDetail(targetType: 'thread' | 'message', targetID: string, value: -1 | 1, message?: CommunityMessage) {
  if (!thread.value) return
  if (!auth.account) { await router.push({ name: 'login', query: { redirect: route.fullPath } }); return }
  const current = message ? message.viewerVote : thread.value.viewerVote
  const next = current === value ? 0 : value
  try {
    const result = await voteCommunity(thread.value.id, targetType, targetID, next as -1 | 0 | 1)
    if (message) { message.score = result.score; message.viewerVote = result.viewerVote }
    else { thread.value.score = result.score; thread.value.viewerVote = result.viewerVote; thread.value.root.score = result.score; thread.value.root.viewerVote = result.viewerVote }
  } catch (reason) { actionError.value = reason instanceof Error ? reason.message : t('community.actionFailed') }
}

async function toggleSubscription() {
  if (!thread.value) return
  if (!auth.account) { await router.push({ name: 'login', query: { redirect: route.fullPath } }); return }
  try { await subscribeCommunity(thread.value.id, !thread.value.subscribed); thread.value.subscribed = !thread.value.subscribed }
  catch (reason) { actionError.value = reason instanceof Error ? reason.message : t('community.actionFailed') }
}

async function insertMarkup(prefix: string, suffix = '') {
  const element = postBodyInput.value
  if (!element) return
  const start = element.selectionStart
  const end = element.selectionEnd
  const selected = postBody.value.slice(start, end)
  postBody.value = postBody.value.slice(0, start) + prefix + selected + suffix + postBody.value.slice(end)
  await nextTick()
  element.focus(); element.setSelectionRange(start + prefix.length, end + prefix.length)
}

async function uploadPostImage(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  input.value = ''
  if (!file || !personalMode.value || !auth.account?.owner) return
  uploadingImage.value = true
  actionError.value = ''
  try {
    const uploaded = await uploadLunaStevImage(file)
    postPreview.value = false
    await nextTick()
    const element = postBodyInput.value
    const offset = element?.selectionStart ?? postBody.value.length
    const alternate = file.name.replace(/\.[^.]+$/, '').replace(/[\[\]]/g, '').trim() || 'Image'
    const markdown = `\n![${alternate}](${uploaded.url})\n`
    postBody.value = postBody.value.slice(0, offset) + markdown + postBody.value.slice(offset)
    await nextTick()
    element?.focus()
    element?.setSelectionRange(offset + markdown.length, offset + markdown.length)
  } catch (reason) {
    actionError.value = reason instanceof Error ? reason.message : t('community.imageUploadFailed')
  } finally {
    uploadingImage.value = false
  }
}

watch(() => route.fullPath, load, { immediate: true })
watchEffect(() => {
  if (!thread.value) return
  const root = thread.value.root
  const author = { '@type': 'Person', name: authorName(root.author) }
  const comments = thread.value.replies.map((reply) => ({
    '@type': 'Comment',
    text: plainText(reply.body),
    datePublished: reply.createdAt,
    author: { '@type': 'Person', name: authorName(reply.author) },
  }))
  applyPageSEO({
    title: `${thread.value.title} · ${personalMode.value ? 'LunaStev' : 'Wave Community'}`,
    description: plainTextDescription(root.body, thread.value.title),
    locale: locale.value,
    path: route.path,
    schema: {
      '@type': personalMode.value ? 'BlogPosting' : 'DiscussionForumPosting',
      headline: thread.value.title,
      text: plainText(root.body),
      datePublished: root.createdAt,
      author,
      commentCount: comments.length,
      ...(comments.length ? { comment: comments } : {}),
      interactionStatistic: [
        { '@type': 'InteractionCounter', interactionType: 'https://schema.org/ViewAction', userInteractionCount: thread.value.viewCount },
        { '@type': 'InteractionCounter', interactionType: 'https://schema.org/ReplyAction', userInteractionCount: comments.length },
        { '@type': 'InteractionCounter', interactionType: 'https://schema.org/LikeAction', userInteractionCount: Math.max(0, thread.value.score) },
      ],
    },
  })
})
</script>

<template>
  <main class="community-hub">
    <header class="community-bar">
      <div class="community-width community-bar-inner">
        <h1>{{ personalMode ? t('community.personalTitle') : t('community.title') }}</h1>
        <form class="community-search" role="search" @submit.prevent="search">
          <Search :size="16" aria-hidden="true" />
          <input v-model="searchText" :placeholder="t('community.search')" :aria-label="t('community.search')" />
        </form>
        <button v-if="canOpenComposer" class="community-write" type="button" @click="openComposer">{{ t('community.write') }}</button>
      </div>
    </header>

    <div class="community-width community-workspace" :class="{ 'has-topics': showTopics, 'personal-section': personalMode }">
      <aside v-if="!personalMode" class="community-spaces" :aria-label="t('community.spaces')">
        <strong>{{ t('community.spaces') }}</strong>
        <button type="button" :class="{ active: !activeSpace && !threadID }" @click="selectSpace('')">{{ t('community.all') }}</button>
        <button v-for="space in memberSpaces" :key="space.id" type="button"
          :class="{ active: activeSpace === space.id || thread?.spaceId === space.id }" @click="selectSpace(space.id)">
          {{ spaceName(space) }}
        </button>
      </aside>

      <section class="community-feed" aria-live="polite">
        <nav v-if="personalMode && !threadID && !isComposer" class="personal-category-tabs" :aria-label="t('community.personalCategories')">
          <button type="button" :class="{ active: !selectedSpaceID }" @click="selectSpace('')">{{ t('community.personalAll') }}</button>
          <button v-for="space in personalSpaces" :key="space.id" type="button"
            :class="{ active: selectedSpaceID === space.id }" @click="selectSpace(space.id)">
            {{ spaceName(space) }}
          </button>
        </nav>
        <UiSkeletonRows v-if="loading" :rows="6" />
        <UiInlineState v-else-if="failed" :message="t('common.loadError')" :action="t('common.retry')" @action="load" />

        <form v-else-if="isComposer" class="community-composer" @submit.prevent="submitPost">
          <header>
            <RouterLink :to="rootPath">← {{ personalMode ? t('community.personalTitle') : t('community.back') }}</RouterLink>
            <h2>{{ t('community.createPost') }}</h2>
          </header>
          <label for="post-space">{{ t('community.category') }}</label>
          <select id="post-space" v-model="postSpace" required>
            <option v-for="space in composerSpaces" :key="space.id" :value="space.id">{{ spaceName(space) }}</option>
          </select>
          <label for="post-title">{{ t('community.postTitle') }}</label>
          <input id="post-title" v-model="postTitle" required minlength="5" maxlength="180" />
          <label for="post-body">{{ t('community.body') }}</label>
          <div class="community-editor-toolbar" :aria-label="t('community.formatting')">
            <button type="button" :title="t('community.code')" @click="insertMarkup('\n```wave\n', '\n```\n')"><Code2 :size="16" /></button>
            <button type="button" :title="t('community.link')" @click="insertMarkup('[', '](https://)')"><Link :size="16" /></button>
            <button type="button" :title="t('community.quote')" @click="insertMarkup('> ')"><Quote :size="16" /></button>
            <button v-if="personalMode && auth.account?.owner" type="button" :title="t('community.imageUpload')"
              :aria-label="t('community.imageUpload')" :disabled="uploadingImage" @click="postImageInput?.click()">
              <ImagePlus :size="16" />
            </button>
            <input v-if="personalMode && auth.account?.owner" ref="postImageInput" hidden type="file"
              accept="image/jpeg,image/png,image/webp" @change="uploadPostImage" />
            <button type="button" class="preview-toggle" @click="postPreview = !postPreview">{{ postPreview ? t('community.edit') : t('community.preview') }}</button>
          </div>
          <small v-if="personalMode && auth.account?.owner" class="community-image-help">{{ uploadingImage ? t('community.imageUploading') : t('community.imageHelp') }}</small>
          <MarkdownContent v-if="postPreview" class="community-editor-preview" :source="postBody" />
          <textarea v-else id="post-body" ref="postBodyInput" v-model="postBody" required maxlength="20000" rows="12" />
          <label for="post-tags">{{ t('community.tags') }}</label>
          <input id="post-tags" v-model="postTags" :placeholder="t('community.tagsPlaceholder')" />
          <p v-if="actionError" class="community-action-error" role="alert">{{ actionError }}</p>
          <footer><RouterLink :to="rootPath">{{ t('common.cancel') }}</RouterLink><button class="ui-button primary" type="submit" :disabled="submitting">{{ t('community.publish') }}</button></footer>
        </form>

        <template v-else-if="thread">
          <header class="thread-heading">
            <RouterLink class="thread-back" :to="rootPath">← {{ personalMode ? t('community.personalTitle') : t('community.back') }}</RouterLink>
            <div class="thread-heading-meta"><span>{{ selectedSpaceName(thread.spaceId) }}</span><time :datetime="thread.root.createdAt">{{ formatDate(thread.root.createdAt) }}</time></div>
            <h2>{{ thread.title }}</h2>
            <div class="thread-tags"><span v-for="tag in thread.tags" :key="tag">{{ tag }}</span><span v-if="thread.locked">{{ t('community.locked') }}</span></div>
          </header>

          <article class="forum-message root-message">
            <div class="community-vote-rail">
              <button type="button" :class="{ active: thread.viewerVote === 1 }" :aria-label="t('community.upvote')" @click="voteDetail('thread', thread.id, 1)"><ArrowBigUp :size="22" /></button>
              <strong>{{ thread.score }}</strong>
              <button type="button" :class="{ active: thread.viewerVote === -1 }" :aria-label="t('community.downvote')" @click="voteDetail('thread', thread.id, -1)"><ArrowBigDown :size="22" /></button>
            </div>
            <div class="forum-message-content">
              <header><span class="author-avatar">{{ authorInitial(thread.root.author) }}</span><strong>{{ authorName(thread.root.author) }}</strong><time :datetime="thread.root.createdAt">{{ relativeDate(thread.root.createdAt) }}</time></header>
              <MarkdownContent :source="thread.root.body" />
              <footer>
                <span><MessageSquare :size="14" />{{ thread.replies.length }}</span><span><Eye :size="14" />{{ thread.viewCount }}</span>
                <button type="button" @click="beginReply(thread.root)">{{ t('community.reply') }}</button>
                <button type="button" @click="toggleSubscription"><BellOff v-if="thread.subscribed" :size="14" /><Bell v-else :size="14" />{{ thread.subscribed ? t('community.unsubscribe') : t('community.subscribe') }}</button>
              </footer>
            </div>
          </article>
          <CommunityReplyEditor v-if="replyTo === thread.root.id" ref="replyEditor"
            v-model="replyBody" v-model:preview="replyPreview" :authenticated="Boolean(auth.account)"
            :submitting="submitting" :error="actionError" :replying="true" :login-target="route.fullPath"
            @submit="submitReply" @cancel="replyTo = ''" />

          <section class="thread-replies" :aria-label="t('community.comments')">
            <header><h3>{{ t('community.comments') }}</h3><span>{{ thread.replies.length }}</span></header>
            <template v-for="item in orderedReplies" :key="item.message.id">
              <article class="forum-message reply-message" :class="{ 'is-nested': item.depth > 0 }"
                :style="{ marginLeft: replyIndent(item.depth) }" :data-depth="item.depth">
                <div class="community-vote-rail compact">
                  <button type="button" :class="{ active: item.message.viewerVote === 1 }" :aria-label="t('community.upvote')" @click="voteDetail('message', item.message.id, 1, item.message)"><ArrowBigUp :size="18" /></button>
                  <strong>{{ item.message.score }}</strong>
                  <button type="button" :class="{ active: item.message.viewerVote === -1 }" :aria-label="t('community.downvote')" @click="voteDetail('message', item.message.id, -1, item.message)"><ArrowBigDown :size="18" /></button>
                </div>
                <div class="forum-message-content">
                  <header><span class="author-avatar">{{ authorInitial(item.message.author) }}</span><strong>{{ authorName(item.message.author) }}</strong><time :datetime="item.message.createdAt">{{ relativeDate(item.message.createdAt) }}</time></header>
                  <MarkdownContent :source="item.message.body" />
                  <footer><button v-if="!thread.locked" type="button" @click="beginReply(item.message)">{{ t('community.reply') }}</button></footer>
                </div>
              </article>
              <CommunityReplyEditor v-if="replyTo === item.message.id" ref="replyEditor"
                v-model="replyBody" v-model:preview="replyPreview" :authenticated="Boolean(auth.account)"
                :submitting="submitting" :error="actionError" :replying="true" :login-target="route.fullPath"
                :style="{ marginLeft: replyIndent(item.depth + 1) }"
                @submit="submitReply" @cancel="replyTo = ''" />
            </template>
            <p v-if="thread.replies.length === 0" class="community-state compact">{{ t('community.noComments') }}</p>
          </section>

          <CommunityReplyEditor v-if="!thread.locked && !replyTo" ref="replyEditor"
            v-model="replyBody" v-model:preview="replyPreview" :authenticated="Boolean(auth.account)"
            :submitting="submitting" :error="actionError" :replying="false" :login-target="route.fullPath"
            @submit="submitReply" @cancel="replyTo = ''" />
        </template>

        <template v-else>
          <nav class="community-feed-tabs" :aria-label="t('community.sort')">
            <button v-for="sort in ['latest', 'active', 'top']" :key="sort" type="button" :class="{ active: activeSort === sort }" @click="selectSort(sort)">{{ t(`community.sort.${sort}`) }}</button>
            <span>{{ visibleThreads.length }} {{ t('community.posts') }}</span>
          </nav>
          <div v-if="visibleThreads.length" class="thread-list">
            <article v-for="item in visibleThreads" :key="item.id" class="forum-post-row">
              <div class="community-vote-rail compact">
                <button type="button" :class="{ active: item.viewerVote === 1 }" :aria-label="t('community.upvote')" @click="voteSummary(item, 1)"><ArrowBigUp :size="19" /></button>
                <strong>{{ item.score }}</strong>
                <button type="button" :class="{ active: item.viewerVote === -1 }" :aria-label="t('community.downvote')" @click="voteSummary(item, -1)"><ArrowBigDown :size="19" /></button>
              </div>
              <div class="thread-summary">
                <div class="thread-labels"><span v-if="item.pinned" class="pinned-label">{{ t('community.pinned') }}</span><span>{{ selectedSpaceName(item.spaceId) }}</span></div>
                <RouterLink :to="{ name: threadRouteName, params: { thread: item.id } }">{{ item.title }}</RouterLink>
                <p>{{ item.excerpt }}</p>
                <footer>
                  <strong>{{ authorName(item.author) }}</strong><time :datetime="item.createdAt">{{ relativeDate(item.createdAt) }}</time>
                  <span><MessageSquare :size="13" />{{ item.replyCount }}</span><span><Eye :size="13" />{{ item.viewCount }}</span>
                  <span>{{ t('community.activity') }} {{ relativeDate(item.lastActivityAt) }}</span>
                  <span v-for="tag in item.tags" :key="tag" class="thread-tag">{{ tag }}</span>
                </footer>
              </div>
            </article>
            <button v-if="hasMore" class="community-load-more" type="button" :disabled="loadingMore" @click="loadMore">
              {{ loadingMore ? t('common.loading') : t('common.more') }}
            </button>
          </div>
          <p v-else class="community-state">{{ activeQuery ? t('community.noSearchResults') : t('community.empty') }}</p>
        </template>
      </section>

      <aside v-if="showTopics" class="community-topics">
        <section>
          <strong>{{ t('community.popularTags') }}</strong>
          <button v-for="topic in topicCounts" :key="topic.name" type="button" @click="searchText = topic.name; search()"><span># {{ topic.name }}</span><small>{{ topic.count }}</small></button>
        </section>
      </aside>
    </div>
  </main>
</template>
