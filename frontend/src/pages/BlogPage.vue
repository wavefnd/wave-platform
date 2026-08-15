<script setup lang="ts">
import { computed, onMounted, ref, watch, watchEffect } from 'vue'
import { useRoute } from 'vue-router'

import MarkdownContent from '../components/MarkdownContent.vue'
import { useI18n } from '../i18n'
import { getBlogPost, getBlogPosts, type BlogPost, type BlogPostSummary } from '../services/http'
import { applyPageSEO, plainTextDescription } from '../services/seo'
import UiInlineState from '../ui/UiInlineState.vue'

const route = useRoute()
const { locale, t } = useI18n()
const posts = ref<BlogPostSummary[]>([])
const post = ref<BlogPost | null>(null)
const loading = ref(true)
const error = ref('')
const detail = computed(() => typeof route.params.slug === 'string' && route.params.slug !== '')

function formatDate(value: string) {
	if (!value) return ''
	return new Intl.DateTimeFormat(locale.value === 'ko' ? 'ko-KR' : 'en-US', {
		year: 'numeric', month: 'long', day: 'numeric',
	}).format(new Date(value))
}

async function load() {
	loading.value = true
	error.value = ''
	try {
		if (detail.value) {
			post.value = await getBlogPost(String(route.params.slug))
			posts.value = []
		} else {
			posts.value = await getBlogPosts(locale.value)
			post.value = null
		}
	} catch (reason) {
		error.value = reason instanceof Error ? reason.message : t('common.loadError')
	} finally {
		loading.value = false
	}
}

onMounted(load)
watch([() => route.fullPath, locale], load)
watchEffect(() => {
	if (!post.value) return
	applyPageSEO({
		title: `${post.value.title} · Wave Blog`,
		description: plainTextDescription(post.value.summary || post.value.content, post.value.title),
		locale: post.value.locale,
		path: route.path,
		schema: { '@type': 'BlogPosting', headline: post.value.title, datePublished: post.value.publishedAt, dateModified: post.value.updatedAt },
	})
})
</script>

<template>
	<main class="blog-page portal-width">
		<header v-if="!detail" class="blog-heading">
			<span>{{ t('blog.official') }}</span>
			<h1>{{ t('blog.title') }}</h1>
			<p>{{ t('blog.lead') }}</p>
		</header>
		<UiInlineState v-if="loading" :message="t('common.loading')" />
		<UiInlineState v-else-if="error" :message="error" />
		<section v-else-if="!detail" class="blog-index">
			<RouterLink v-for="item in posts" :key="item.slug" :to="`/blog/${encodeURIComponent(item.slug)}`" class="blog-row">
				<time :datetime="item.publishedAt">{{ formatDate(item.publishedAt) }}</time>
				<div><h2>{{ item.title }}</h2><p>{{ item.summary }}</p><small>{{ item.authorName }}</small></div>
			</RouterLink>
			<p v-if="posts.length === 0" class="compact-empty">{{ t('blog.empty') }}</p>
		</section>
		<article v-else-if="post" class="blog-article">
			<RouterLink class="blog-back" to="/blog">← {{ t('blog.back') }}</RouterLink>
			<header><span>Wave Blog</span><h1>{{ post.title }}</h1><div><time :datetime="post.publishedAt">{{ formatDate(post.publishedAt) }}</time><span>{{ post.authorName }}</span></div><p>{{ post.summary }}</p></header>
			<MarkdownContent :source="post.content" />
		</article>
	</main>
</template>
