<script setup lang="ts">
import { computed, onMounted, ref, watch, watchEffect } from 'vue'
import { useRoute } from 'vue-router'

import MarkdownContent from '../components/MarkdownContent.vue'
import { useI18n } from '../i18n'
import { getBlogPost, getBlogPosts, type BlogPost, type BlogPostSummary } from '../services/http'
import { applyPageSEO, firstMarkdownImage, plainTextDescription } from '../services/seo'
import { useAuthStore } from '../stores/auth'
import UiInlineState from '../ui/UiInlineState.vue'
import '../ui/blog.css'

const route = useRoute()
const { locale, t } = useI18n()
const auth = useAuthStore()
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
const editorialIndex = computed(() => !detail.value && category.value === '')
const canEdit = computed(() => Boolean(auth.account?.owner || auth.account?.administrator))

function publishedTime(item: BlogPostSummary) {
	const value = item.publishedAt || item.updatedAt
	const parsed = Date.parse(value)
	return Number.isFinite(parsed) ? parsed : 0
}

const publishedPosts = computed(() => posts.value
	.filter((item) => item.category !== 'roadmap')
	.slice()
	.sort((left, right) => publishedTime(right) - publishedTime(left)))
const articlePosts = computed(() => publishedPosts.value.filter((item) => item.category === 'article'))
const releasePosts = computed(() => publishedPosts.value.filter((item) => item.category === 'release'))
const roadmapPosts = computed(() => posts.value
	.filter((item) => item.category === 'roadmap')
	.map((item, index) => ({ item, index }))
	.sort((left, right) => left.item.roadmapOrder - right.item.roadmapOrder || left.index - right.index)
	.map(({ item }) => item))
const featuredPost = computed(() => publishedPosts.value[0] ?? roadmapPosts.value[0] ?? null)
const homeArticles = computed(() => articlePosts.value.slice(0, 4))
const homeReleases = computed(() => releasePosts.value.slice(0, 3))
const homeRoadmap = computed(() => roadmapPosts.value.slice(0, 4))
const latestRelease = computed(() => releasePosts.value[0] ?? null)
const seoPosts = computed(() => {
	if (releaseIndex.value) return releasePosts.value
	if (category.value === 'roadmap') return roadmapPosts.value
	if (category.value === 'article') return articlePosts.value
	return [...publishedPosts.value.filter((item) => item.category !== 'release'), ...roadmapPosts.value]
})
const releaseArchive = computed(() => releasePosts.value.slice(1).reduce<Array<{ year: string, items: BlogPostSummary[] }>>((groups, item) => {
	const year = releaseYear(item.publishedAt)
	const existing = groups.find((group) => group.year === year)
	if (existing) existing.items.push(item)
	else groups.push({ year, items: [item] })
	return groups
}, []))
const detailBack = computed(() => {
	if (releaseDetail.value) return '/releases'
	if (post.value?.category === 'roadmap') return { path: '/blog', query: { category: 'roadmap' } }
	return { path: '/blog', query: { category: 'article' } }
})

function formatDate(value: string) {
	if (!value) return ''
	const parsed = /^\d{4}-\d{2}-\d{2}$/.test(value) ? new Date(`${value}T00:00:00`) : new Date(value)
	if (Number.isNaN(parsed.getTime())) return ''
	return new Intl.DateTimeFormat(locale.value === 'ko' ? 'ko-KR' : 'en-US', {
		year: 'numeric', month: 'long', day: 'numeric',
	}).format(parsed)
}

function releaseYear(value: string) {
	const direct = value.match(/^(\d{4})/)?.[1]
	if (direct) return direct
	const parsed = new Date(value)
	return Number.isNaN(parsed.getTime()) ? '—' : String(parsed.getFullYear())
}

function postLink(item: BlogPostSummary) {
	return item.category === 'release' ? `/releases/${encodeURIComponent(item.slug)}` : `/blog/${encodeURIComponent(item.slug)}`
}

function postDate(item: BlogPostSummary) {
	return item.category === 'roadmap' ? item.targetDate : item.publishedAt
}

async function load() {
	loading.value = true
	error.value = ''
	try {
		if (detail.value) {
			post.value = null
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

function roadmapStatusLabel(value: BlogPostSummary['roadmapStatus']) {
	return t(value === 'in-progress' ? 'blog.roadmap.inProgress' : value === 'released' ? 'blog.roadmap.released' : 'blog.roadmap.planned')
}

function releaseVersion(title: string) {
	return title.match(/\bv?\d+\.\d+\.\d+(?:-[a-z0-9]+(?:-[a-z0-9]+)*)?/i)?.[0] ?? title
}

onMounted(load)
watch(() => route.fullPath, load)
watchEffect(() => {
	if (!detail.value) {
		const pageTitle = releaseIndex.value
			? t('blog.release.title')
			: category.value === 'roadmap'
				? t('blog.category.roadmap')
				: category.value === 'article' ? t('blog.category.article') : t('blog.title')
		applyPageSEO({
			title: `${pageTitle} · Wave`,
			description: releaseIndex.value ? t('blog.release.lead') : t('blog.lead'),
			locale: locale.value,
			path: route.path,
			breadcrumbs: [
				{ name: 'Home', path: '/' },
				{ name: releaseIndex.value ? 'Releases' : 'Blog', path: releaseIndex.value ? '/releases' : '/blog' },
			],
			schema: {
				'@type': 'CollectionPage', name: pageTitle,
				mainEntity: {
					'@type': 'ItemList',
					itemListElement: seoPosts.value.slice(0, 20).map((item, index) => ({
						'@type': 'ListItem', position: index + 1, name: item.title, url: new URL(postLink(item), window.location.origin).toString(),
					})),
				},
			},
		})
		return
	}
	if (!post.value) {
		if (error.value) applyPageSEO({
			title: 'Page not found · Wave', description: 'The requested Wave page was not found.',
			locale: locale.value, path: route.path, noIndex: true, schema: { '@type': 'WebPage' },
		})
		return
	}
	const release = post.value.category === 'release'
	const canonicalPath = release ? `/releases/${encodeURIComponent(post.value.slug)}` : route.path
	const sectionName = release ? 'Releases' : post.value.category === 'roadmap' ? 'Roadmap' : 'Blog'
	const sectionPath = release ? '/releases' : '/blog'
	const image = firstMarkdownImage(post.value.content)
	applyPageSEO({
		title: `${post.value.title} · ${release ? 'Wave Releases' : 'Wave Blog'}`,
		description: plainTextDescription(post.value.summary || post.value.content, post.value.title),
		locale: locale.value,
		path: canonicalPath,
		breadcrumbs: [
			{ name: 'Home', path: '/' },
			{ name: sectionName, path: sectionPath },
			{ name: post.value.title, path: canonicalPath },
		],
		image: image?.url,
		imageAlt: image?.alt || post.value.title,
		schema: {
			'@type': release || post.value.category === 'roadmap' ? 'TechArticle' : 'BlogPosting',
			headline: post.value.title,
			...(post.value.publishedAt ? { datePublished: post.value.publishedAt } : {}),
			dateModified: post.value.updatedAt,
			articleSection: sectionName,
			isAccessibleForFree: true,
			author: {
				'@type': post.value.authorName === 'Wave Foundation' ? 'Organization' : 'Person',
				name: post.value.authorName || 'Wave Foundation',
				...(post.value.authorAccountId ? { url: new URL(`/user/id/${encodeURIComponent(post.value.authorAccountId)}`, window.location.origin).toString() } : {}),
			},
		},
	})
})
</script>

<template>
	<main class="blog-page portal-width">
		<header v-if="!detail" class="blog-heading" :class="{ 'release-heading': releaseIndex }">
			<div class="blog-heading-kicker"><span>{{ t('blog.official') }}</span><RouterLink v-if="canEdit" to="/blog/editor">WaveEditor →</RouterLink></div>
			<h1>{{ releaseIndex ? t('blog.release.title') : t('blog.title') }}</h1>
			<p>{{ releaseIndex ? t('blog.release.lead') : t('blog.lead') }}</p>
		</header>

		<nav v-if="!detail" class="blog-categories" :aria-label="t('blog.categories')">
			<RouterLink to="/blog" :class="{ active: category === '' }">{{ t('blog.category.all') }}</RouterLink>
			<RouterLink :to="{ path: '/blog', query: { category: 'article' } }" :class="{ active: category === 'article' }">{{ t('blog.category.article') }}</RouterLink>
			<RouterLink to="/releases" :class="{ active: category === 'release' }">{{ t('blog.category.release') }}</RouterLink>
			<RouterLink :to="{ path: '/blog', query: { category: 'roadmap' } }" :class="{ active: category === 'roadmap' }">{{ t('blog.category.roadmap') }}</RouterLink>
		</nav>

		<UiInlineState v-if="loading" :message="t('common.loading')" />
		<UiInlineState v-else-if="error" :message="error" />

		<section v-else-if="editorialIndex" class="blog-editorial" :aria-label="t('blog.title')">
			<RouterLink v-if="featuredPost" :to="postLink(featuredPost)" class="blog-feature">
				<div class="blog-feature-meta">
					<strong>{{ t(`blog.category.${featuredPost.category}`) }}</strong>
					<time :datetime="postDate(featuredPost)">{{ formatDate(postDate(featuredPost)) }}</time>
				</div>
				<h2>{{ featuredPost.title }}</h2>
				<p>{{ featuredPost.summary }}</p>
				<span class="blog-feature-open" aria-hidden="true">→</span>
			</RouterLink>

			<div class="blog-editorial-sections">
				<section class="blog-editorial-section">
					<RouterLink class="blog-section-heading" :to="{ path: '/blog', query: { category: 'article' } }">
						<h2>{{ t('blog.category.article') }}</h2><span aria-hidden="true">→</span>
					</RouterLink>
					<div v-if="homeArticles.length" class="blog-compact-list">
						<RouterLink v-for="item in homeArticles" :key="item.slug" :to="postLink(item)">
							<time :datetime="item.publishedAt">{{ formatDate(item.publishedAt) }}</time>
							<strong>{{ item.title }}</strong>
							<p>{{ item.summary }}</p>
						</RouterLink>
					</div>
					<p v-else-if="articlePosts.length === 0" class="compact-empty">{{ t('blog.empty') }}</p>
				</section>

				<section class="blog-editorial-section">
					<RouterLink class="blog-section-heading" to="/releases">
						<h2>{{ t('blog.category.release') }}</h2><span aria-hidden="true">→</span>
					</RouterLink>
					<div v-if="homeReleases.length" class="blog-compact-list release-compact-list">
						<RouterLink v-for="item in homeReleases" :key="item.slug" :to="postLink(item)">
							<strong>{{ releaseVersion(item.title) }}</strong>
							<time :datetime="item.publishedAt">{{ formatDate(item.publishedAt) }}</time>
						</RouterLink>
					</div>
					<p v-else-if="releasePosts.length === 0" class="compact-empty">{{ t('blog.release.empty') }}</p>
				</section>

				<section class="blog-editorial-section">
					<RouterLink class="blog-section-heading" :to="{ path: '/blog', query: { category: 'roadmap' } }">
						<h2>{{ t('blog.category.roadmap') }}</h2><span aria-hidden="true">→</span>
					</RouterLink>
					<div v-if="homeRoadmap.length" class="blog-compact-list roadmap-compact-list">
						<RouterLink v-for="item in homeRoadmap" :key="item.slug" :to="postLink(item)">
							<span class="roadmap-detail-status" :class="`is-${item.roadmapStatus}`">{{ roadmapStatusLabel(item.roadmapStatus) }}</span>
							<strong>{{ item.title }}</strong>
							<time :datetime="item.targetDate">{{ formatDate(item.targetDate) }}</time>
						</RouterLink>
					</div>
					<p v-else class="compact-empty">{{ t('blog.roadmap.empty') }}</p>
				</section>
			</div>
		</section>

		<template v-else-if="releaseIndex">
			<section v-if="latestRelease" class="release-latest" :aria-label="t('blog.release.title')">
				<div class="release-latest-version">
					<small>{{ t('blog.release.version') }}</small>
					<strong>{{ releaseVersion(latestRelease.title) }}</strong>
				</div>
				<RouterLink :to="postLink(latestRelease)" class="release-latest-body">
					<h2>{{ latestRelease.title }}</h2>
					<p>{{ latestRelease.summary }}</p>
					<span>{{ t('blog.release.date') }} <time :datetime="latestRelease.publishedAt">{{ formatDate(latestRelease.publishedAt) }}</time></span>
					<i aria-hidden="true">→</i>
				</RouterLink>
			</section>

			<section v-if="releaseArchive.length" class="release-archive" :aria-label="t('blog.release.archive')">
				<h2>{{ t('blog.release.archive') }}</h2>
				<section v-for="group in releaseArchive" :key="group.year" class="release-year">
					<h3>{{ group.year }}</h3>
					<div>
						<RouterLink v-for="item in group.items" :key="item.slug" :to="postLink(item)" class="release-archive-row">
							<strong>{{ releaseVersion(item.title) }}</strong>
							<span>{{ item.title }}</span>
							<time :datetime="item.publishedAt">{{ formatDate(item.publishedAt) }}</time>
							<i aria-hidden="true">→</i>
						</RouterLink>
					</div>
				</section>
			</section>
			<p v-if="!latestRelease" class="compact-empty release-empty">{{ t('blog.release.empty') }}</p>
		</template>

		<section v-else-if="!detail" class="blog-index" :class="{ 'roadmap-index': category === 'roadmap' }">
			<RouterLink v-for="item in category === 'roadmap' ? roadmapPosts : articlePosts" :key="item.slug" :to="postLink(item)" class="blog-row" :class="{ 'roadmap-row': category === 'roadmap' }">
				<div v-if="category === 'roadmap'" class="roadmap-issue-state">
					<span :class="`is-${item.roadmapStatus}`">{{ roadmapStatusLabel(item.roadmapStatus) }}</span>
					<small>#{{ item.roadmapOrder }}</small>
				</div>
				<div class="blog-row-date">
					<small>{{ item.category === 'roadmap' ? t('blog.roadmap.target') : t('blog.category.article') }}</small>
					<time :datetime="postDate(item)">{{ formatDate(postDate(item)) }}</time>
				</div>
				<div class="blog-row-copy">
					<h2>{{ item.title }}</h2>
					<p>{{ item.summary }}</p>
					<small v-if="item.category !== 'roadmap'">{{ item.authorName }}</small>
				</div>
				<span class="blog-row-open" aria-hidden="true">→</span>
			</RouterLink>
			<p v-if="(category === 'roadmap' ? roadmapPosts : articlePosts).length === 0" class="compact-empty">{{ category === 'roadmap' ? t('blog.roadmap.empty') : t('blog.empty') }}</p>
		</section>

		<article v-else-if="post" class="blog-article" :class="{ 'release-article': releaseDetail, 'roadmap-article': post.category === 'roadmap' }">
			<div class="blog-detail-actions"><RouterLink class="blog-back" :to="detailBack">← {{ releaseDetail ? t('blog.release.back') : t('blog.back') }}</RouterLink><RouterLink v-if="canEdit" :to="`/blog/editor/${encodeURIComponent(post.slug)}`">{{ t('admin.editPost') }}</RouterLink></div>
			<header>
				<span>{{ t(`blog.category.${post.category}`) }}</span>
				<h1>{{ post.title }}</h1>
				<div class="blog-article-meta">
					<template v-if="post.category === 'roadmap'">
						<span class="roadmap-detail-status" :class="`is-${post.roadmapStatus}`">{{ roadmapStatusLabel(post.roadmapStatus) }}</span>
						<time :datetime="post.targetDate">{{ t('blog.roadmap.target') }} {{ formatDate(post.targetDate) }}</time>
					</template>
					<template v-else-if="post.category === 'release'">
						<strong>{{ t('blog.release.date') }}</strong>
						<time :datetime="post.publishedAt">{{ formatDate(post.publishedAt) }}</time>
					</template>
					<template v-else>
						<time :datetime="post.publishedAt">{{ formatDate(post.publishedAt) }}</time>
						<span>{{ post.authorName }}</span>
					</template>
				</div>
				<p v-if="post.summary" class="blog-article-summary">{{ post.summary }}</p>
			</header>
			<MarkdownContent :source="post.content" />
		</article>
	</main>
</template>
