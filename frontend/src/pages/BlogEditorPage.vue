<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import MarkdownContent from '../components/MarkdownContent.vue'
import PlatformWaveEditor from '../components/editor/PlatformWaveEditor.vue'
import { useI18n } from '../i18n'
import { getBlogEditorPost, getBlogEditorPosts, saveBlogEditorPost, type BlogPostInput, type BlogPostSummary } from '../services/http'
import { applyPageSEO } from '../services/seo'
import '../ui/blog-editor.css'

const route = useRoute()
const router = useRouter()
const { locale, t } = useI18n()
const posts = ref<BlogPostSummary[]>([])
const form = ref<BlogPostInput>(blankPost())
const originalSlug = ref('')
const loading = ref(true)
const saving = ref(false)
const error = ref('')
const notice = ref('')
const preview = ref(true)

const editing = computed(() => Boolean(originalSlug.value))
const generatedSlug = computed(() => slugFromTitle(form.value.title))

function blankPost(): BlogPostInput {
  return { slug: '', category: 'article', roadmapStatus: '', roadmapOrder: 1, targetDate: '', title: '', summary: '', content: '', status: 'draft' }
}

function slugFromTitle(title: string) {
  return title.trim().toLowerCase().replace(/[^a-z0-9.]+/g, '-').replace(/^[.-]+|[.-]+$/g, '')
}

async function load() {
  loading.value = true
  error.value = ''
  try {
    posts.value = await getBlogEditorPosts()
    const slug = typeof route.params.slug === 'string' ? route.params.slug : ''
    if (slug) {
      const item = await getBlogEditorPost(slug)
      originalSlug.value = item.slug
      form.value = { slug: item.slug, category: item.category, roadmapStatus: item.roadmapStatus, roadmapOrder: item.roadmapOrder,
        targetDate: item.targetDate, title: item.title, summary: item.summary, content: item.content, status: item.status || 'draft' }
    } else {
      originalSlug.value = ''
      form.value = blankPost()
    }
  } catch (reason) {
    error.value = reason instanceof Error ? reason.message : t('common.loadError')
  } finally { loading.value = false }
}

async function save() {
  if (saving.value) return
  saving.value = true
  error.value = ''
  notice.value = ''
  try {
    const saved = await saveBlogEditorPost({ ...form.value, slug: originalSlug.value || generatedSlug.value })
    originalSlug.value = saved.slug
    form.value = { slug: saved.slug, category: saved.category, roadmapStatus: saved.roadmapStatus, roadmapOrder: saved.roadmapOrder,
      targetDate: saved.targetDate, title: saved.title, summary: saved.summary, content: saved.content, status: saved.status || 'draft' }
    posts.value = await getBlogEditorPosts()
    notice.value = t('blog.editor.saved')
    if (route.params.slug !== saved.slug) await router.replace({ name: 'blog-editor-post', params: { slug: saved.slug } })
  } catch (reason) {
    error.value = reason instanceof Error ? reason.message : t('admin.actionFailed')
  } finally { saving.value = false }
}

function publicLink(item: BlogPostInput) {
  return item.category === 'release' ? `/releases/${encodeURIComponent(item.slug)}` : `/blog/${encodeURIComponent(item.slug)}`
}

onMounted(() => {
  applyPageSEO({ title: `${t('blog.editor.title')} · Wave`, description: t('blog.editor.lead'), locale: locale.value, path: '/blog/editor', noIndex: true })
  void load()
})
watch(() => route.params.slug, load)
</script>

<template>
  <main class="blog-editor-page">
    <header class="blog-editor-heading">
      <div><span>WaveEditor</span><h1>{{ t('blog.editor.title') }}</h1><p>{{ t('blog.editor.lead') }}</p></div>
      <div class="blog-editor-heading-actions">
        <RouterLink class="ui-button" to="/blog">{{ t('blog.editor.viewBlog') }}</RouterLink>
        <RouterLink class="ui-button primary" to="/blog/editor">{{ t('admin.newPost') }}</RouterLink>
      </div>
    </header>

    <p v-if="error" class="ui-alert danger" role="alert">{{ error }}</p>
    <p v-if="loading" class="blog-editor-loading">{{ t('common.loading') }}</p>

    <div v-else class="blog-editor-workspace">
      <aside class="blog-editor-library" :aria-label="t('admin.blogPosts')">
        <h2>{{ t('admin.blogPosts') }} <small>{{ posts.length }}</small></h2>
        <nav>
          <RouterLink v-for="post in posts" :key="post.slug" :to="`/blog/editor/${encodeURIComponent(post.slug)}`" :class="{ active: originalSlug === post.slug }">
            <strong>{{ post.title }}</strong><span>{{ t(`blog.category.${post.category}`) }} · {{ post.status === 'published' ? t('admin.published') : t('admin.draft') }}</span>
          </RouterLink>
        </nav>
        <p v-if="posts.length === 0">{{ t('admin.noBlogPosts') }}</p>
      </aside>

      <form class="blog-editor-form" @submit.prevent="save">
        <div class="blog-editor-fields">
          <label>{{ t('admin.blogCategory') }}<select v-model="form.category"><option value="article">{{ t('blog.category.article') }}</option><option value="release">{{ t('blog.category.release') }}</option><option value="roadmap">{{ t('blog.category.roadmap') }}</option></select></label>
          <label>{{ t('admin.blogStatus') }}<select v-model="form.status"><option value="draft">{{ t('admin.draft') }}</option><option value="published">{{ t('admin.published') }}</option></select></label>
          <label class="wide">{{ t('admin.blogTitle') }}<input v-model="form.title" required maxlength="160" :placeholder="form.category === 'release' || form.category === 'roadmap' ? 'v0.3.0' : ''" /></label>
          <label class="wide slug-field">{{ t('admin.blogSlug') }}<input :value="originalSlug || generatedSlug" readonly /><small>{{ t('blog.editor.slugHelp') }}</small></label>
          <template v-if="form.category === 'roadmap'">
            <label>{{ t('admin.roadmapStatus') }}<select v-model="form.roadmapStatus" required><option value="" disabled>{{ t('admin.selectRoadmapStatus') }}</option><option value="planned">{{ t('blog.roadmap.planned') }}</option><option value="in-progress">{{ t('blog.roadmap.inProgress') }}</option><option value="released">{{ t('blog.roadmap.released') }}</option></select></label>
            <label>{{ t('admin.roadmapOrder') }}<input v-model.number="form.roadmapOrder" type="number" min="1" max="1000000" required /><small>{{ t('admin.roadmapOrderHelp') }}</small></label>
            <label>{{ t('admin.targetReleaseDate') }}<input v-model="form.targetDate" type="date" required /></label>
          </template>
          <label v-if="form.category !== 'roadmap'" class="wide">{{ t('admin.blogSummary') }}<input v-model="form.summary" required maxlength="500" /></label>
        </div>

        <div class="blog-editor-document-heading"><strong>{{ t('admin.blogContent') }}</strong><button type="button" @click="preview = !preview">{{ preview ? t('blog.editor.hidePreview') : t('blog.editor.showPreview') }}</button></div>
        <div class="blog-editor-document" :class="{ 'without-preview': !preview }">
          <PlatformWaveEditor v-model="form.content" :label="t('blog.editor.document')" required :rows="24" @save="save" />
          <section v-if="preview" class="blog-editor-preview">
            <span>{{ t('admin.blogPreview') }}</span><MarkdownContent v-if="form.content" :source="form.content" /><p v-else>{{ t('blog.editor.previewEmpty') }}</p>
          </section>
        </div>

        <footer class="blog-editor-actions">
          <button class="ui-button primary" type="submit" :disabled="saving || (!originalSlug && !generatedSlug)">{{ saving ? t('common.loading') : t('common.save') }}</button>
          <a v-if="editing && form.status === 'published'" :href="publicLink(form)">{{ t('blog.editor.openPublished') }}</a>
          <span v-if="notice" role="status">{{ notice }}</span><small>Ctrl/⌘ + S</small>
        </footer>
      </form>
    </div>
  </main>
</template>
