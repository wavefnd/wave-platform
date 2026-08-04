<script setup lang="ts">
import { onMounted, ref, watchEffect } from 'vue'
import { useRoute } from 'vue-router'

import MarkdownContent from '../components/MarkdownContent.vue'
import { useI18n } from '../i18n'
import { getRelease, type Release } from '../services/http'
import { applyPageSEO, plainTextDescription } from '../services/seo'

const route = useRoute()
const { locale, t } = useI18n()
const release = ref<Release | null>(null)
const error = ref('')

onMounted(async () => {
  try {
    release.value = await getRelease(String(route.params.slug))
  } catch {
    error.value = t('common.loadError')
  }
})

const formatDate = (value: string) => new Intl.DateTimeFormat(
  locale.value === 'ko' ? 'ko-KR' : 'en-US',
  { year: 'numeric', month: 'long', day: 'numeric' },
).format(new Date(value))

watchEffect(() => {
  if (!release.value) return
  applyPageSEO({
    title: `${release.value.title} · Wave`,
    description: plainTextDescription(release.value.summary || release.value.content, release.value.title),
    locale: locale.value,
    path: route.path,
    schema: {
      '@type': 'Article',
      headline: release.value.title,
      datePublished: release.value.publishedAt,
      articleSection: 'Release notes',
    },
  })
})
</script>

<template>
  <main class="service-page release-service">
    <div class="shell release-layout">
      <RouterLink class="release-back" to="/#releases">← {{ t('release.back') }}</RouterLink>
      <p v-if="error">{{ error }}</p>
      <p v-else-if="!release">{{ t('common.loading') }}</p>
      <article v-else class="release-article">
        <header>
          <span>{{ t('home.releaseLabel') }}</span>
          <h1>{{ release.title }}</h1>
          <time :datetime="release.publishedAt">{{ formatDate(release.publishedAt) }}</time>
        </header>
        <MarkdownContent :source="release.content" />
      </article>
    </div>
  </main>
</template>
