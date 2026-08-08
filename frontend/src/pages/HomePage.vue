<script setup lang="ts">
import { Check, Copy } from '@lucide/vue'
import { computed, onMounted, ref } from 'vue'

import { useI18n } from '../i18n'
import {
  getCommunityThreads,
  getCommunitySpaces,
  getQuestions,
  getReleases,
  getSourceRepositories,
	getSponsors,
  type CommunityThreadSummary,
  type QuestionSummary,
  type ReleaseSummary,
	type SponsorsView,
} from '../services/http'
import type { SourceRepository } from '../components/source/types'

const { locale, t } = useI18n()
const releases = ref<ReleaseSummary[]>([])
const discussions = ref<CommunityThreadSummary[]>([])
const unansweredQuestions = ref<QuestionSummary[]>([])
const repositories = ref<SourceRepository[]>([])
const sponsors = ref<SponsorsView | null>(null)
const copied = ref(false)
const installPlatform = ref<'unix' | 'windows'>('unix')
const installCommand = computed(() => installPlatform.value === 'windows'
  ? 'irm https://wave-lang.dev/install.ps1 -OutFile install.ps1; powershell -ExecutionPolicy Bypass -File .\\install.ps1 -Latest'
  : 'curl -fsSL https://wave-lang.dev/install.sh | bash -s -- latest')

const docs = computed(() => [
  { path: 'getting-started/install', label: t('docs.installation'), detail: t('docs.installation.detail') },
  { path: 'getting-started/overview', label: t('docs.firstProgram'), detail: t('docs.firstProgram.detail') },
  { path: 'reference/syntax-quick-reference', label: t('docs.language'), detail: t('docs.language.detail') },
])

function authorName(author: string) {
  const label = author.replace(/\s*<[^>]+>\s*$/, '').trim()
  return label.replace(/^"(.*)"$/, '$1') || author
}

function formatDate(value: string) {
  return new Intl.DateTimeFormat(locale.value === 'ko' ? 'ko-KR' : 'en-US', {
    year: 'numeric', month: 'short', day: 'numeric',
  }).format(new Date(value))
}

async function copyInstall() {
  await navigator.clipboard.writeText(installCommand.value)
  copied.value = true
  window.setTimeout(() => { copied.value = false }, 1600)
}

onMounted(async () => {
  const [releaseResult, discussionResult, spaceResult, questionResult, repositoryResult, sponsorResult] = await Promise.allSettled([
    getReleases(4),
    getCommunityThreads('', { sort: 'active', limit: 5 }),
    getCommunitySpaces(),
    getQuestions({ sort: 'unanswered', limit: 5 }),
    getSourceRepositories(),
	getSponsors(),
  ])
  if (releaseResult.status === 'fulfilled') releases.value = releaseResult.value
  if (discussionResult.status === 'fulfilled' && spaceResult.status === 'fulfilled') {
    const communitySpaceIDs = new Set(spaceResult.value.filter((space) => space.postingPolicy !== 'owner').map((space) => space.id))
    discussions.value = discussionResult.value.filter((thread) => communitySpaceIDs.has(thread.spaceId)).slice(0, 5)
  }
  if (questionResult.status === 'fulfilled') unansweredQuestions.value = questionResult.value.slice(0, 5)
  if (repositoryResult.status === 'fulfilled') {
    repositories.value = repositoryResult.value
      .filter((repository) => repository.headCommit)
      .sort((left, right) => (right.headCommit?.authoredAt ?? '').localeCompare(left.headCommit?.authoredAt ?? ''))
      .slice(0, 4)
  }
	if (sponsorResult.status === 'fulfilled') sponsors.value = sponsorResult.value
})
</script>

<template>
  <main class="portal-home portal-width">
    <div class="portal-modules">
      <section v-if="releases.length" id="releases" class="portal-module">
        <header>
          <h1>{{ t('home.latestReleases') }}</h1>
        </header>
        <ul class="portal-data-list">
          <li v-for="(release, index) in releases" :key="release.slug" :class="{ featured: index === 0 }">
            <RouterLink :to="`/releases/${release.slug}`" :title="release.title">{{ release.title }}</RouterLink>
            <time :datetime="release.publishedAt">{{ formatDate(release.publishedAt) }}</time>
          </li>
        </ul>
      </section>

      <section v-if="discussions.length" class="portal-module">
        <header>
          <h2>{{ t('home.communityLatest') }}</h2>
          <RouterLink to="/community">{{ t('common.more') }}</RouterLink>
        </header>
        <ul class="portal-data-list portal-discussions">
          <li v-for="discussion in discussions" :key="discussion.id">
            <RouterLink :to="{ name: 'community-thread', params: { thread: discussion.id } }">{{ discussion.title }}</RouterLink>
            <small>{{ authorName(discussion.author) }} · {{ discussion.replyCount }} {{ t('community.replies') }}</small>
          </li>
        </ul>
      </section>

      <section v-if="unansweredQuestions.length" class="portal-module">
        <header>
          <h2>{{ t('home.unansweredQuestions') }}</h2>
          <RouterLink to="/questions?sort=unanswered">{{ t('common.more') }}</RouterLink>
        </header>
        <ul class="portal-data-list portal-discussions">
          <li v-for="question in unansweredQuestions" :key="question.id">
            <RouterLink :to="`/questions/${question.id}`">{{ question.title }}</RouterLink>
            <small>{{ authorName(question.author) }} · {{ question.answerCount }} {{ t('questions.answers') }}</small>
          </li>
        </ul>
      </section>

      <section class="portal-module portal-docs-module">
        <header>
          <h2>{{ t('home.docsStart') }}</h2>
          <RouterLink to="/docs">{{ t('common.more') }}</RouterLink>
        </header>
        <ul class="portal-data-list">
          <li v-for="item in docs" :key="item.label">
            <RouterLink :to="`/docs/${item.path}`">{{ item.label }}</RouterLink>
            <small>{{ item.detail }}</small>
          </li>
        </ul>
      </section>

      <section class="portal-module portal-install-module">
        <header>
          <h2>{{ t('home.installWave') }}</h2>
          <div class="install-platform-tabs" role="group" :aria-label="t('home.installPlatform')">
            <button type="button" :class="{ active: installPlatform === 'unix' }" :aria-pressed="installPlatform === 'unix'" @click="installPlatform = 'unix'">Linux / macOS</button>
            <button type="button" :class="{ active: installPlatform === 'windows' }" :aria-pressed="installPlatform === 'windows'" @click="installPlatform = 'windows'">Windows</button>
          </div>
        </header>
        <div class="terminal-command" :class="{ 'is-windows': installPlatform === 'windows' }"><code><span>{{ installPlatform === 'windows' ? 'PS>' : '$' }}</span> {{ installCommand }}</code><button type="button" :title="t('home.copyInstall')" :aria-label="t('home.copyInstall')" @click="copyInstall"><Check v-if="copied" :size="16" /><Copy v-else :size="16" /></button></div>
        <RouterLink to="/docs/getting-started/install">{{ t('home.start.docs') }} →</RouterLink>
      </section>

      <section v-if="repositories.length" class="portal-module portal-source-module">
        <header>
          <h2>{{ t('home.recentSource') }}</h2>
          <RouterLink to="/source">{{ t('common.more') }}</RouterLink>
        </header>
        <ul class="portal-data-list portal-discussions">
          <li v-for="repository in repositories" :key="repository.id">
            <RouterLink :to="{ name: 'source-repository', params: { repository: repository.id } }">{{ repository.owner }}/{{ repository.name }}</RouterLink>
            <small>{{ repository.headCommit?.subject }} · {{ formatDate(repository.headCommit?.authoredAt ?? '') }}</small>
          </li>
        </ul>
      </section>

	  <section v-if="sponsors" class="portal-module portal-sponsors-module">
		<header><h2>{{ t('home.sponsors') }}</h2><a :href="sponsors.url" target="_blank" rel="noopener noreferrer">{{ t('home.sponsorProject') }}</a></header>
		<div v-if="sponsors.tiers.some((tier) => tier.members.length)" class="sponsor-tiers">
		  <section v-for="tier in sponsors.tiers.filter((item) => item.members.length)" :key="tier.slug" class="sponsor-tier">
			<h3>{{ tier.name }}</h3>
			<ul><li v-for="member in tier.members" :key="member.profile"><a :href="member.profile" target="_blank" rel="noopener noreferrer">{{ member.name }}</a></li></ul>
		  </section>
		</div>
		<p v-else class="sponsor-empty"><a :href="sponsors.url" target="_blank" rel="noopener noreferrer">{{ t('home.noSponsors') }}</a></p>
	  </section>
    </div>
  </main>
</template>
