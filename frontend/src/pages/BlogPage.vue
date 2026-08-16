<script setup lang="ts">
import { computed, onMounted, ref, watch, watchEffect } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import MarkdownContent from '../components/MarkdownContent.vue'
import { useI18n } from '../i18n'
import { getBlogPost, getBlogPosts, type BlogPost, type BlogPostSummary } from '../services/http'
import { applyPageSEO, plainTextDescription } from '../services/seo'
import UiInlineState from '../ui/UiInlineState.vue'

const route = useRoute()
const router = useRouter()
const { locale, t } = useI18n()
const posts = ref<BlogPostSummary[]>([])
const post = ref<BlogPost | null>(null)
const loading = ref(true)
const error = ref('')
const detail = computed(() => typeof route.params.slug === 'string' && route.params.slug !== '')
const category = computed<'article' | 'release' | 'roadmap' | ''>(() => {
	if (route.meta.blogCategory === 'release') return 'release'
	return ['release', 'article', 'roadmap'].includes(String(route.query.category)) ? route.query.category as 'article' | 'release' | 'roadmap' : ''
})
const releaseIndex = computed(() => !detail.value && category.value === 'release')
const releaseDetail = computed(() => detail.value && post.value?.category === 'release')

function formatDate(value: string) {
	if (!value) return ''
	const parsed = /^\d{4}-\d{2}-\d{2}$/.test(value) ? new Date(`${value}T00:00:00`) : new Date(value)
	return new Intl.DateTimeFormat(locale.value === 'ko' ? 'ko-KR' : 'en-US', {
		year: 'numeric', month: 'long', day: 'numeric',
	}).format(parsed)
}

async function load() {
	loading.value = true
	error.value = ''
	try {
		if (detail.value) {
			post.value = await getBlogPost(String(route.params.slug))
			if (route.name === 'release-detail' && post.value.category !== 'release') throw new Error(t('blog.release.notFound'))
			posts.value = []
		} else {
			posts.value = await getBlogPosts(category.value || undefined)
			post.value = null
		}
	} catch (reason) {
		error.value = reason instanceof Error ? reason.message : t('common.loadError')
	} finally {
		loading.value = false
	}
}

function selectCategory(value: 'article' | 'release' | 'roadmap' | '') {
	void router.push(value === 'release' ? { name: 'releases' } : { name: 'blog', query: value ? { category: value } : {} })
}

function roadmapStatusLabel(value: BlogPostSummary['roadmapStatus']) {
	return t(value === 'in-progress' ? 'blog.roadmap.inProgress' : value === 'released' ? 'blog.roadmap.released' : 'blog.roadmap.planned')
}

function releaseVersion(title: string) {
	return title.match(/\bv?\d+\.\d+\.\d+(?:-[a-z0-9]+(?:-[a-z0-9]+)*)?/i)?.[0] ?? title
}

onMounted(load)
watch(() => route.fullPath, load)
watchEffect(() => {
	if (releaseIndex.value) {
		applyPageSEO({
			title: `${t('blog.release.title')} · Wave`, description: t('blog.release.lead'), locale: locale.value,
			path: route.path, schema: { '@type': 'CollectionPage', name: t('blog.release.title') },
		})
		return
	}
	if (!post.value) return
	const release = post.value.category === 'release'
	applyPageSEO({
		title: `${post.value.title} · ${release ? 'Wave Releases' : 'Wave Blog'}`,
		description: plainTextDescription(post.value.summary || post.value.content, post.value.title),
		locale: locale.value,
		path: release ? `/releases/${encodeURIComponent(post.value.slug)}` : route.path,
		schema: { '@type': release ? 'TechArticle' : 'BlogPosting', headline: post.value.title, datePublished: post.value.publishedAt, dateModified: post.value.updatedAt },
	})
})
</script>

<template>
	<main class="blog-page portal-width">
		<header v-if="!detail" class="blog-heading" :class="{ 'release-heading': releaseIndex }">
			<span>{{ t('blog.official') }}</span>
			<h1>{{ releaseIndex ? t('blog.release.title') : t('blog.title') }}</h1>
			<p>{{ releaseIndex ? t('blog.release.lead') : t('blog.lead') }}</p>
		</header>
		<nav v-if="!detail" class="blog-categories" :aria-label="t('blog.categories')">
			<button type="button" :class="{ active: category === '' }" @click="selectCategory('')">{{ t('blog.category.all') }}</button>
			<button type="button" :class="{ active: category === 'release' }" @click="selectCategory('release')">{{ t('blog.category.release') }}</button>
			<button type="button" :class="{ active: category === 'article' }" @click="selectCategory('article')">{{ t('blog.category.article') }}</button>
			<button type="button" :class="{ active: category === 'roadmap' }" @click="selectCategory('roadmap')">{{ t('blog.category.roadmap') }}</button>
		</nav>
		<UiInlineState v-if="loading" :message="t('common.loading')" />
		<UiInlineState v-else-if="error" :message="error" />
		<section v-else-if="releaseIndex" class="release-index" :aria-label="t('blog.release.archive')">
			<RouterLink v-for="item in posts" :key="item.slug" :to="`/releases/${encodeURIComponent(item.slug)}`" class="release-row">
				<div class="release-version"><small>{{ t('blog.release.version') }}</small><strong :title="item.title">{{ releaseVersion(item.title) }}</strong></div>
				<div class="release-description"><p>{{ item.summary }}</p><span>{{ t('blog.release.date') }} <time :datetime="item.publishedAt">{{ formatDate(item.publishedAt) }}</time></span></div>
				<span class="release-open" aria-hidden="true">→</span>
			</RouterLink>
			<p v-if="posts.length === 0" class="compact-empty">{{ t('blog.release.empty') }}</p>
		</section>
		<section v-else-if="!detail" class="blog-index" :class="{ 'roadmap-index': category === 'roadmap' }">
			<RouterLink v-for="item in posts" :key="item.slug" :to="item.category === 'release' ? `/releases/${encodeURIComponent(item.slug)}` : `/blog/${encodeURIComponent(item.slug)}`" class="blog-row" :class="{ 'roadmap-row': category === 'roadmap' }">
				<div v-if="category === 'roadmap'" class="roadmap-issue-state"><span :class="`is-${item.roadmapStatus}`">{{ roadmapStatusLabel(item.roadmapStatus) }}</span><small>#{{ item.roadmapOrder }}</small></div>
				<div class="blog-row-date"><small v-if="item.category === 'release'">{{ t('blog.release.date') }}</small><small v-else-if="item.category === 'roadmap'">{{ t('blog.roadmap.target') }}</small><time :datetime="item.category === 'roadmap' ? item.targetDate : item.publishedAt">{{ formatDate(item.category === 'roadmap' ? item.targetDate : item.publishedAt) }}</time></div>
				<div><small class="blog-category">{{ t(`blog.category.${item.category}`) }}</small><h2>{{ item.title }}</h2><p>{{ item.summary }}</p><small v-if="item.category !== 'roadmap'">{{ item.authorName }}</small></div>
			</RouterLink>
			<p v-if="posts.length === 0" class="compact-empty">{{ category === 'roadmap' ? t('blog.roadmap.empty') : t('blog.empty') }}</p>
		</section>
		<article v-else-if="post" class="blog-article" :class="{ 'release-article': releaseDetail }">
			<RouterLink class="blog-back" :to="releaseDetail ? '/releases' : '/blog'">← {{ releaseDetail ? t('blog.release.back') : t('blog.back') }}</RouterLink>
			<header><span>{{ t(`blog.category.${post.category}`) }}</span><h1>{{ post.title }}</h1><div><template v-if="post.category === 'roadmap'"><span class="roadmap-detail-status" :class="`is-${post.roadmapStatus}`">{{ roadmapStatusLabel(post.roadmapStatus) }}</span><time :datetime="post.targetDate">{{ t('blog.roadmap.target') }} {{ formatDate(post.targetDate) }}</time></template><template v-else-if="post.category === 'release'"><strong>{{ t('blog.release.date') }}</strong><time :datetime="post.publishedAt">{{ formatDate(post.publishedAt) }}</time></template><template v-else><time :datetime="post.publishedAt">{{ formatDate(post.publishedAt) }}</time><span>{{ post.authorName }}</span></template></div><p>{{ post.summary }}</p></header>
			<MarkdownContent :source="post.content" />
		</article>
	</main>
</template>
