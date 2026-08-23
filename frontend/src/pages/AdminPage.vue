<script setup lang="ts">
import coreuiStylesheet from '@coreui/coreui/dist/css/coreui.min.css?url'
import {
  CBadge, CButton, CCard, CCardBody, CCardHeader, CCol, CContainer, CHeader, CHeaderToggler,
  CModal, CModalBody, CModalFooter, CModalHeader, CModalTitle, CNavItem, CNavTitle, CRow,
  CSidebar, CSidebarBrand, CSidebarFooter, CSidebarHeader, CSidebarNav, CTable, CTableBody,
  CTableDataCell, CTableHead, CTableHeaderCell, CTableRow,
} from '@coreui/vue'
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  ArrowLeft, Boxes, Database, FileClock, FileText, Gauge, GitBranch, MailWarning, Menu, RefreshCw,
  Search, ShieldCheck, Users, Webhook,
} from '@lucide/vue'

import {
  getAdminSnapshot, getModules, getPlatformStatus,
  deleteAdminWebhook, getAdminBlogPost, getAdminBlogPosts, getAdminWebhooks, saveAdminBlogPost, saveAdminWebhook, testAdminWebhook,
  updateAdminAccountRole, updateAdminAccountStatus, updateAdminRFCMaintainer, updateAdminSourceMaintainer, updateLunaStevTimeZone,
  type AdminAccount, type AdminSnapshot, type BlogPostInput, type BlogPostSummary, type ModuleStatus, type PlatformStatus,
  type WebhookAdminView, type WebhookInput,
} from '../services/http'
import { useI18n } from '../i18n'
import type { Locale } from '../i18n'
import { useAuthStore } from '../stores/auth'
import ThemeSelector from '../components/platform/ThemeSelector.vue'
import MarkdownContent from '../components/MarkdownContent.vue'

type PendingAction = { kind: 'status'; account: AdminAccount; status: 'active' | 'suspended' }
  | { kind: 'role'; account: AdminAccount; administrator: boolean }
  | { kind: 'source-maintainer'; account: AdminAccount; enabled: boolean }
  | { kind: 'rfc-maintainer'; account: AdminAccount; enabled: boolean }

const router = useRouter()
const route = useRoute()
const auth = useAuthStore()
const { locale, setLocale, t } = useI18n()
const modules = ref<ModuleStatus[]>([])
const snapshot = ref<AdminSnapshot | null>(null)
const platform = ref<PlatformStatus | null>(null)
const error = ref('')
const actionError = ref('')
const loading = ref(true)
const actionLoading = ref(false)
const desktopLayout = ref(window.innerWidth >= 992)
const sidebarVisible = ref(desktopLayout.value)
const accountQuery = ref('')
const pendingAction = ref<PendingAction | null>(null)
const lunaStevTimeZone = ref('Asia/Seoul')
const blogPosts = ref<BlogPostSummary[]>([])
const blogSaving = ref(false)
const blogNotice = ref('')
const editingBlogSlug = ref('')
const blogForm = ref<BlogPostInput>({ slug: '', category: 'article', roadmapStatus: '', roadmapOrder: 0, targetDate: '', title: '', summary: '', content: '', status: 'draft' })
const webhookView = ref<WebhookAdminView>({ supportedEvents: [], endpoints: [], deliveries: [] })
const webhookForm = ref<WebhookInput>({ id: '', name: '', kind: 'generic', url: '', events: [], enabled: true, rotateSecret: false })
const webhookSecret = ref('')
const timeZones = ['Asia/Seoul', 'UTC', 'Asia/Tokyo', 'Asia/Singapore', 'Europe/London', 'Europe/Paris', 'America/New_York', 'America/Los_Angeles']
let stylesheet: HTMLLinkElement | null = null

const accounts = computed(() => {
  const needle = accountQuery.value.trim().toLowerCase()
  if (!needle) return snapshot.value?.accounts ?? []
  return (snapshot.value?.accounts ?? []).filter((account) =>
    `${account.username} ${account.displayName} ${account.email}`.toLowerCase().includes(needle))
})
const mailAttention = computed(() => (snapshot.value?.mail.queued ?? 0) + (snapshot.value?.mail.delivering ?? 0)
  + (snapshot.value?.mail.deferred ?? 0) + (snapshot.value?.mail.failed ?? 0))
const section = computed(() => String(route.params.section ?? route.meta.adminSection ?? 'overview'))
const sectionTitle = computed(() => t(({ overview: 'admin.overview', blog: 'admin.blog', accounts: 'admin.accounts', mailbox: 'admin.managementMailbox',
  'mail-queue': 'admin.mailQueue', 'git-mirrors': 'admin.gitMirrors', 'audit-log': 'admin.auditLog', security: 'admin.security',
  webhooks: 'admin.webhooks', modules: 'admin.modules', system: 'admin.system' } as Record<string, string>)[section.value] ?? 'admin.overview'))

onMounted(async () => {
  stylesheet = document.createElement('link')
  stylesheet.rel = 'stylesheet'
  stylesheet.href = coreuiStylesheet
  stylesheet.dataset.waveAdminStyles = 'coreui'
  document.head.append(stylesheet)
  document.body.classList.add('coreui-admin-active')
	window.addEventListener('resize', syncSidebar)
  await load()
})
watch(section, async () => {
	error.value = ''
	actionError.value = ''
	blogNotice.value = ''
	pendingAction.value = null
	if (window.innerWidth < 992) sidebarVisible.value = false
	try { await loadSection() }
	catch (reason) { error.value = reason instanceof Error ? reason.message : t('common.loadError') }
})

onBeforeUnmount(() => {
  stylesheet?.remove()
  document.body.classList.remove('coreui-admin-active')
	window.removeEventListener('resize', syncSidebar)
})

function syncSidebar() {
	const nextDesktopLayout = window.innerWidth >= 992
	if (nextDesktopLayout === desktopLayout.value) return
	desktopLayout.value = nextDesktopLayout
	sidebarVisible.value = nextDesktopLayout
}

async function load() {
  loading.value = true
  error.value = ''
  try {
    const [moduleResult, snapshotResult, platformResult] = await Promise.all([
      getModules(), getAdminSnapshot(), getPlatformStatus(),
    ])
    modules.value = moduleResult
    snapshot.value = snapshotResult
	lunaStevTimeZone.value = snapshotResult.lunaStevTimeZone
    platform.value = platformResult
	await loadSection()
  } catch (reason) {
    error.value = reason instanceof Error ? reason.message : t('common.loadError')
  } finally {
    loading.value = false
  }
}

async function loadSection() {
	if (section.value === 'blog') blogPosts.value = await getAdminBlogPosts()
	if (section.value === 'webhooks') webhookView.value = await getAdminWebhooks()
}

function newWebhook() {
	webhookSecret.value = ''
	webhookForm.value = { id: '', name: '', kind: 'generic', url: '', events: [], enabled: true, rotateSecret: false }
}

function editWebhook(id: string) {
	const item = webhookView.value.endpoints.find((endpoint) => endpoint.id === id)
	if (!item) return
	webhookSecret.value = ''
	webhookForm.value = { id: item.id, name: item.name, kind: item.kind, url: '', events: [...item.events], enabled: item.enabled, rotateSecret: false }
}

async function saveWebhook() {
	actionLoading.value = true
	actionError.value = ''
	try {
		const saved = await saveAdminWebhook(webhookForm.value)
		webhookSecret.value = saved.signingSecret
		webhookView.value = await getAdminWebhooks()
		editWebhook(saved.id)
		webhookSecret.value = saved.signingSecret
	} catch (reason) { actionError.value = reason instanceof Error ? reason.message : t('admin.actionFailed') }
	finally { actionLoading.value = false }
}

async function testWebhook(id: string) {
	actionLoading.value = true; actionError.value = ''
	try { await testAdminWebhook(id); webhookView.value = await getAdminWebhooks() }
	catch (reason) { actionError.value = reason instanceof Error ? reason.message : t('admin.actionFailed') }
	finally { actionLoading.value = false }
}

async function removeWebhook(id: string) {
	if (!window.confirm(t('admin.confirmDeleteWebhook'))) return
	actionLoading.value = true; actionError.value = ''
	try { await deleteAdminWebhook(id); newWebhook(); webhookView.value = await getAdminWebhooks() }
	catch (reason) { actionError.value = reason instanceof Error ? reason.message : t('admin.actionFailed') }
	finally { actionLoading.value = false }
}

function newBlogPost() {
	editingBlogSlug.value = ''
	blogNotice.value = ''
	actionError.value = ''
	blogForm.value = { slug: '', category: 'article', roadmapStatus: '', roadmapOrder: 0, targetDate: '', title: '', summary: '', content: '', status: 'draft' }
}

async function editBlogPost(slug: string) {
	actionError.value = ''
	blogNotice.value = ''
	try {
		const item = await getAdminBlogPost(slug)
		editingBlogSlug.value = item.slug
		blogForm.value = { slug: item.slug, category: item.category, roadmapStatus: item.roadmapStatus, roadmapOrder: item.roadmapOrder, targetDate: item.targetDate, title: item.title, summary: item.summary, content: item.content, status: item.status || 'draft' }
	} catch (reason) {
		actionError.value = reason instanceof Error ? reason.message : t('common.loadError')
	}
}

async function saveBlogPost() {
	blogSaving.value = true
	actionError.value = ''
	blogNotice.value = ''
	try {
		const input = { ...blogForm.value }
		if (input.category === 'roadmap' && !editingBlogSlug.value) input.slug = ''
		const saved = await saveAdminBlogPost(input)
		editingBlogSlug.value = saved.slug
		blogForm.value = { slug: saved.slug, category: saved.category, roadmapStatus: saved.roadmapStatus, roadmapOrder: saved.roadmapOrder, targetDate: saved.targetDate, title: saved.title, summary: saved.summary, content: saved.content, status: saved.status || 'draft' }
		blogPosts.value = await getAdminBlogPosts()
		blogNotice.value = t('admin.blogSaved')
	} catch (reason) {
		actionError.value = reason instanceof Error ? reason.message : t('admin.actionFailed')
	} finally {
		blogSaving.value = false
	}
}

async function saveLunaStevTimeZone() {
  actionError.value = ''
  try { await updateLunaStevTimeZone(lunaStevTimeZone.value); await load() }
  catch (reason) { actionError.value = reason instanceof Error ? reason.message : t('admin.actionFailed') }
}

function formatDate(value: string) {
  if (!value) return '—'
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? value : new Intl.DateTimeFormat(locale.value, {
    dateStyle: 'medium', timeStyle: 'short',
  }).format(date)
}

function formatBytes(value: number) {
  if (!Number.isFinite(value) || value <= 0) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  const unit = Math.min(Math.floor(Math.log(value) / Math.log(1024)), units.length - 1)
  return `${(value / 1024 ** unit).toFixed(unit === 0 ? 0 : 1)} ${units[unit]}`
}

function badgeColor(status: string) {
  if (['ready', 'active', 'delivered', 'success'].includes(status)) return 'success'
  if (['queued', 'pending', 'syncing', 'delivering'].includes(status)) return 'info'
  if (['deferred'].includes(status)) return 'warning'
  if (['failed', 'error', 'suspended', 'unavailable'].includes(status)) return 'danger'
  return 'secondary'
}

function ask(action: PendingAction) {
  actionError.value = ''
  pendingAction.value = action
}

async function confirmAction() {
  const action = pendingAction.value
  if (!action) return
  actionLoading.value = true
  actionError.value = ''
  try {
    if (action.kind === 'status') await updateAdminAccountStatus(action.account.id, action.status)
    else if (action.kind === 'role') await updateAdminAccountRole(action.account.id, action.administrator)
    else if (action.kind === 'source-maintainer') await updateAdminSourceMaintainer(action.account.id, action.enabled)
    else await updateAdminRFCMaintainer(action.account.id, action.enabled)
    pendingAction.value = null
    await load()
  } catch (reason) {
    actionError.value = reason instanceof Error ? reason.message : t('admin.actionFailed')
  } finally {
    actionLoading.value = false
  }
}

async function signOut() {
  await auth.signOut()
  await router.replace('/')
}

function changeLocale(event: Event) {
  setLocale((event.target as HTMLSelectElement).value as Locale)
}
</script>

<template>
  <main class="coreui-admin-template">
    <CSidebar class="border-end admin-mobile-sidebar" color-scheme="dark" position="fixed" :overlaid="!desktopLayout" :visible="sidebarVisible" @visible-change="sidebarVisible = $event">
      <CSidebarHeader class="border-bottom">
        <CSidebarBrand as="div" class="admin-brand">Wave <span>{{ t('admin.title') }}</span></CSidebarBrand>
      </CSidebarHeader>
      <CSidebarNav>
        <CNavItem><RouterLink class="nav-link" :class="{ active: section === 'overview' }" to="/admin"><Gauge class="nav-icon" :size="20" />{{ t('admin.overview') }}</RouterLink></CNavItem>
        <CNavTitle>{{ t('admin.management') }}</CNavTitle>
		<CNavItem><RouterLink class="nav-link" :class="{ active: section === 'blog' }" to="/admin/blog"><FileText class="nav-icon" :size="20" />{{ t('admin.blog') }}</RouterLink></CNavItem>
		<CNavItem><RouterLink class="nav-link" :class="{ active: section === 'webhooks' }" to="/admin/webhooks"><Webhook class="nav-icon" :size="20" />{{ t('admin.webhooks') }}</RouterLink></CNavItem>
        <CNavItem><RouterLink class="nav-link" :class="{ active: section === 'accounts' }" to="/admin/accounts"><Users class="nav-icon" :size="20" />{{ t('admin.accounts') }}</RouterLink></CNavItem>
        <CNavItem><RouterLink class="nav-link" :class="{ active: section === 'mailbox' }" to="/admin/mailbox"><MailWarning class="nav-icon" :size="20" />{{ t('admin.managementMailbox') }}</RouterLink></CNavItem>
        <CNavItem><RouterLink class="nav-link" :class="{ active: section === 'mail-queue' }" to="/admin/mail-queue"><MailWarning class="nav-icon" :size="20" />{{ t('admin.mailQueue') }}</RouterLink></CNavItem>
        <CNavItem><RouterLink class="nav-link" :class="{ active: section === 'git-mirrors' }" to="/admin/git-mirrors"><GitBranch class="nav-icon" :size="20" />{{ t('admin.gitMirrors') }}</RouterLink></CNavItem>
        <CNavItem><RouterLink class="nav-link" :class="{ active: section === 'audit-log' }" to="/admin/audit-log"><FileClock class="nav-icon" :size="20" />{{ t('admin.auditLog') }}</RouterLink></CNavItem>
        <CNavTitle>{{ t('admin.platform') }}</CNavTitle>
        <CNavItem><RouterLink class="nav-link" :class="{ active: section === 'security' }" to="/admin/security"><ShieldCheck class="nav-icon" :size="20" />{{ t('admin.security') }}</RouterLink></CNavItem>
        <CNavItem><RouterLink class="nav-link" :class="{ active: section === 'modules' }" to="/admin/modules"><Boxes class="nav-icon" :size="20" />{{ t('admin.modules') }}</RouterLink></CNavItem>
        <CNavItem><RouterLink class="nav-link" :class="{ active: section === 'system' }" to="/admin/system"><Database class="nav-icon" :size="20" />{{ t('admin.system') }}</RouterLink></CNavItem>
      </CSidebarNav>
      <CSidebarFooter class="border-top">
        <RouterLink class="nav-link" to="/"><ArrowLeft class="nav-icon" :size="20" />{{ t('admin.backToPlatform') }}</RouterLink>
      </CSidebarFooter>
    </CSidebar>

    <div class="admin-wrapper wrapper d-flex flex-column min-vh-100">
      <CHeader position="sticky" class="mb-4 p-0 border-bottom">
        <CContainer class="px-4" fluid>
          <CHeaderToggler class="me-3" :aria-label="t('nav.menu')" @click="sidebarVisible = !sidebarVisible"><Menu :size="22" /></CHeaderToggler>
          <strong>Wave Platform</strong>
          <div class="ms-auto d-flex align-items-center gap-3">
            <ThemeSelector class="admin-theme-toggle" />
            <select class="form-select form-select-sm admin-preference-select" :value="locale" :aria-label="t('nav.language')" @change="changeLocale"><option value="en">English</option><option value="ko">한국어</option></select>
            <span class="text-body-secondary small d-none d-sm-inline">{{ auth.account?.displayName }}</span>
            <button class="btn btn-outline-secondary btn-sm" type="button" @click="signOut">{{ t('auth.signOut') }}</button>
          </div>
        </CContainer>
      </CHeader>

      <div class="body flex-grow-1">
        <CContainer class="px-4" fluid>
          <div id="overview" class="admin-heading mb-4">
            <div>
              <div class="text-body-secondary small mb-1">{{ t('admin.title') }} / {{ sectionTitle }}</div>
              <h1 class="h3 mb-1">{{ sectionTitle }}</h1>
              <p v-if="section === 'overview'" class="text-body-secondary mb-0">{{ t('admin.overviewLead') }}</p>
            </div>
            <CButton color="secondary" variant="outline" size="sm" :disabled="loading" @click="load"><RefreshCw :size="15" /> {{ t('admin.refresh') }}</CButton>
          </div>

          <div v-if="error" class="alert alert-danger" role="alert">{{ error }}</div>
          <div v-if="loading && !snapshot" class="admin-loading text-body-secondary">{{ t('common.loading') }}</div>

          <template v-if="snapshot">
            <CRow v-if="section === 'overview'" :xs="{ cols: 1, gutter: 3 }" :sm="{ cols: 2 }" :xl="{ cols: 4 }" class="mb-4">
              <CCol><CCard class="h-100"><CCardBody><div class="text-body-secondary small">{{ t('admin.accounts') }}</div><div class="fs-3 fw-semibold">{{ snapshot.security.activeAccounts + snapshot.security.suspendedAccounts }}</div><small class="text-body-secondary">{{ snapshot.security.suspendedAccounts }} {{ t('admin.suspended').toLowerCase() }}</small></CCardBody></CCard></CCol>
              <CCol><CCard class="h-100"><CCardBody><div class="text-body-secondary small">{{ t('admin.mailAttention') }}</div><div class="fs-3 fw-semibold">{{ mailAttention }}</div><small class="text-body-secondary">{{ snapshot.mail.failed }} {{ t('admin.failed').toLowerCase() }}</small></CCardBody></CCard></CCol>
              <CCol><CCard class="h-100"><CCardBody><div class="text-body-secondary small">{{ t('admin.gitMirrors') }}</div><div class="fs-3 fw-semibold">{{ snapshot.gitMirrors.filter((item) => item.status === 'ready').length }} / {{ snapshot.gitMirrors.length }}</div><small class="text-body-secondary">{{ t('admin.ready') }}</small></CCardBody></CCard></CCol>
              <CCol><CCard class="h-100"><CCardBody><div class="text-body-secondary small">{{ t('admin.storageUsed') }}</div><div class="fs-3 fw-semibold">{{ formatBytes(snapshot.storage.filesBytes) }}</div><small class="text-body-secondary">{{ snapshot.storage.health }}</small></CCardBody></CCard></CCol>
            </CRow>

			<CCard v-if="section === 'blog'" class="mb-4">
			  <CCardHeader class="admin-card-header">
				<div><strong>{{ t('admin.blogPosts') }}</strong><div class="text-body-secondary small">{{ t('admin.blogHelp') }}</div></div>
				<CButton color="primary" size="sm" @click="newBlogPost">{{ t('admin.newPost') }}</CButton>
			  </CCardHeader>
			  <CCardBody class="p-0">
				<div class="admin-blog-layout">
				  <aside class="admin-blog-list">
					<button v-for="post in blogPosts" :key="post.slug" type="button" :class="{ active: editingBlogSlug === post.slug }" @click="editBlogPost(post.slug)">
					  <span><strong>{{ post.title }}</strong><small>{{ t(`blog.category.${post.category}`) }}<template v-if="post.category === 'roadmap'"> #{{ post.roadmapOrder }}</template> · {{ post.status === 'published' ? t('admin.published') : t('admin.draft') }}</small></span>
					  <time :datetime="post.updatedAt">{{ formatDate(post.updatedAt) }}</time>
					</button>
					<p v-if="blogPosts.length === 0" class="text-body-secondary small p-3 mb-0">{{ t('admin.noBlogPosts') }}</p>
				  </aside>
				  <form class="admin-blog-editor" @submit.prevent="saveBlogPost">
					<h2 class="h5">{{ editingBlogSlug ? t('admin.editPost') : t('admin.newPost') }}</h2>
					<div class="admin-blog-fields">
					  <label v-if="blogForm.category !== 'roadmap'">{{ t('admin.blogSlug') }}<input v-model="blogForm.slug" class="form-control" required maxlength="120" pattern="[a-z0-9]+(?:[.-][a-z0-9]+)*" :readonly="Boolean(editingBlogSlug)" /></label>
					  <label>{{ t('admin.blogCategory') }}<select v-model="blogForm.category" class="form-select"><option value="article">{{ t('blog.category.article') }}</option><option value="release">{{ t('blog.category.release') }}</option><option value="roadmap">{{ t('blog.category.roadmap') }}</option></select></label>
					  <template v-if="blogForm.category === 'roadmap'">
						<label>{{ t('admin.roadmapStatus') }}<select v-model="blogForm.roadmapStatus" class="form-select" required><option value="" disabled>{{ t('admin.selectRoadmapStatus') }}</option><option value="planned">{{ t('blog.roadmap.planned') }}</option><option value="in-progress">{{ t('blog.roadmap.inProgress') }}</option><option value="released">{{ t('blog.roadmap.released') }}</option></select></label>
						<label>{{ t('admin.roadmapOrder') }}<input v-model.number="blogForm.roadmapOrder" class="form-control" type="number" min="1" max="1000000" required /><small class="text-body-secondary">{{ t('admin.roadmapOrderHelp') }}</small></label>
						<label>{{ t('admin.targetReleaseDate') }}<input v-model="blogForm.targetDate" class="form-control" type="date" required /></label>
					  </template>
					  <label class="wide">{{ t('admin.blogTitle') }}<input v-model="blogForm.title" class="form-control" required maxlength="160" /></label>
					  <label v-if="blogForm.category !== 'roadmap'" class="wide">{{ t('admin.blogSummary') }}<textarea v-model="blogForm.summary" class="form-control" required maxlength="500" rows="3" /></label>
					  <label class="wide">{{ t('admin.blogContent') }}<textarea v-model="blogForm.content" class="form-control admin-blog-content" required rows="18" /></label>
					  <label>{{ t('admin.blogStatus') }}<select v-model="blogForm.status" class="form-select"><option value="draft">{{ t('admin.draft') }}</option><option value="published">{{ t('admin.published') }}</option></select></label>
					</div>
					<div class="d-flex align-items-center gap-3"><CButton color="primary" type="submit" :disabled="blogSaving">{{ t('common.save') }}</CButton><span v-if="blogNotice" class="text-success small" role="status">{{ blogNotice }}</span></div>
					<div v-if="actionError" class="alert alert-danger mt-3 mb-0" role="alert">{{ actionError }}</div>
					<section v-if="blogForm.content" class="admin-blog-preview"><h3 class="h6">{{ t('admin.blogPreview') }}</h3><MarkdownContent :source="blogForm.content" /></section>
				  </form>
				</div>
			  </CCardBody>
			</CCard>

			<div v-if="section === 'webhooks'" class="admin-webhook-section">
			  <CCard class="mb-4">
				<CCardHeader class="admin-card-header"><div><strong>{{ t('admin.webhookConfiguration') }}</strong><div class="text-body-secondary small">{{ t('admin.webhookHelp') }}</div></div><CButton color="secondary" variant="outline" size="sm" @click="newWebhook">{{ t('admin.newWebhook') }}</CButton></CCardHeader>
				<CCardBody>
				  <form class="admin-webhook-form" @submit.prevent="saveWebhook">
					<label>{{ t('admin.webhookName') }}<input v-model="webhookForm.name" class="form-control" required maxlength="80" /></label>
					<label>{{ t('admin.webhookKind') }}<select v-model="webhookForm.kind" class="form-select"><option value="generic">Generic JSON</option><option value="discord">Discord</option></select></label>
					<label class="wide">{{ t('admin.webhookUrl') }}<input v-model="webhookForm.url" class="form-control" type="url" :required="!webhookForm.id" placeholder="https://…" /><small v-if="webhookForm.id" class="text-body-secondary">{{ t('admin.webhookUrlRetained') }}</small></label>
					<fieldset class="wide"><legend>{{ t('admin.webhookEvents') }}</legend><label v-for="event in webhookView.supportedEvents" :key="event" class="admin-webhook-event"><input v-model="webhookForm.events" type="checkbox" :value="event" /> <code>{{ event }}</code></label></fieldset>
					<label class="admin-webhook-event"><input v-model="webhookForm.enabled" type="checkbox" /> {{ t('admin.enabled') }}</label>
					<label v-if="webhookForm.id" class="admin-webhook-event"><input v-model="webhookForm.rotateSecret" type="checkbox" /> {{ t('admin.rotateWebhookSecret') }}</label>
					<div class="wide d-flex align-items-center gap-3"><CButton color="primary" type="submit" :disabled="actionLoading || webhookForm.events.length === 0">{{ t('common.save') }}</CButton><span v-if="webhookForm.id" class="text-body-secondary small">{{ webhookForm.id }}</span></div>
				  </form>
				  <div v-if="webhookSecret" class="alert alert-warning mt-3 mb-0"><strong>{{ t('admin.webhookSecretOnce') }}</strong><code class="d-block mt-2 text-break">{{ webhookSecret }}</code></div>
				  <div v-if="actionError" class="alert alert-danger mt-3 mb-0">{{ actionError }}</div>
				</CCardBody>
			  </CCard>
			  <CCard class="mb-4"><CCardHeader><strong>{{ t('admin.webhookEndpoints') }}</strong></CCardHeader><CCardBody class="p-0"><CTable align="middle" class="mb-0 admin-table" responsive><CTableHead><CTableRow><CTableHeaderCell class="ps-4">{{ t('admin.webhookName') }}</CTableHeaderCell><CTableHeaderCell>{{ t('admin.destination') }}</CTableHeaderCell><CTableHeaderCell>{{ t('admin.webhookEvents') }}</CTableHeaderCell><CTableHeaderCell>{{ t('admin.status') }}</CTableHeaderCell><CTableHeaderCell>{{ t('admin.actions') }}</CTableHeaderCell></CTableRow></CTableHead><CTableBody><CTableRow v-for="endpoint in webhookView.endpoints" :key="endpoint.id"><CTableDataCell class="ps-4"><strong>{{ endpoint.name }}</strong><div class="d-flex align-items-center gap-1 mt-1"><CBadge :color="endpoint.scope === 'platform' ? 'primary' : endpoint.scope === 'account' ? 'info' : 'warning'">{{ endpoint.scope || (locale === 'ko' ? '검토 필요' : 'review required') }}</CBadge><small class="text-body-secondary">{{ endpoint.kind }}</small></div></CTableDataCell><CTableDataCell><code>{{ endpoint.destination }}</code></CTableDataCell><CTableDataCell><div v-for="event in endpoint.events" :key="event"><code>{{ event }}</code></div></CTableDataCell><CTableDataCell><CBadge :color="endpoint.enabled && endpoint.scope ? 'success' : 'secondary'">{{ endpoint.enabled && endpoint.scope ? t('admin.enabled') : t('admin.disabled') }}</CBadge></CTableDataCell><CTableDataCell class="admin-actions"><CButton color="secondary" variant="outline" size="sm" @click="editWebhook(endpoint.id)">{{ t('common.edit') }}</CButton><CButton color="secondary" variant="outline" size="sm" :disabled="actionLoading || !endpoint.scope" @click="testWebhook(endpoint.id)">{{ t('admin.testWebhook') }}</CButton><CButton color="danger" variant="outline" size="sm" :disabled="actionLoading" @click="removeWebhook(endpoint.id)">{{ t('common.delete') }}</CButton></CTableDataCell></CTableRow><CTableRow v-if="webhookView.endpoints.length === 0"><CTableDataCell colspan="5" class="py-4 text-center text-body-secondary">{{ t('admin.noWebhooks') }}</CTableDataCell></CTableRow></CTableBody></CTable></CCardBody></CCard>
			  <CCard class="mb-4"><CCardHeader><strong>{{ t('admin.webhookDeliveries') }}</strong></CCardHeader><CCardBody class="p-0"><CTable align="middle" class="mb-0 admin-table" responsive><CTableHead><CTableRow><CTableHeaderCell class="ps-4">{{ t('admin.webhookEvents') }}</CTableHeaderCell><CTableHeaderCell>{{ t('admin.status') }}</CTableHeaderCell><CTableHeaderCell>HTTP</CTableHeaderCell><CTableHeaderCell>{{ t('admin.attempts') }}</CTableHeaderCell><CTableHeaderCell>{{ t('admin.updated') }}</CTableHeaderCell></CTableRow></CTableHead><CTableBody><CTableRow v-for="delivery in webhookView.deliveries" :key="delivery.id"><CTableDataCell class="ps-4"><code>{{ delivery.eventType }}</code><div class="text-body-secondary small">{{ delivery.title }}</div><div v-if="delivery.lastError" class="admin-error-detail">{{ delivery.lastError }}</div></CTableDataCell><CTableDataCell><CBadge :color="badgeColor(delivery.status)">{{ delivery.status }}</CBadge></CTableDataCell><CTableDataCell>{{ delivery.httpStatus || '—' }}</CTableDataCell><CTableDataCell>{{ delivery.attempts }}</CTableDataCell><CTableDataCell>{{ formatDate(delivery.lastAttemptAt || delivery.createdAt) }}</CTableDataCell></CTableRow><CTableRow v-if="webhookView.deliveries.length === 0"><CTableDataCell colspan="5" class="py-4 text-center text-body-secondary">{{ t('admin.noWebhookDeliveries') }}</CTableDataCell></CTableRow></CTableBody></CTable></CCardBody></CCard>
			</div>

            <CCard v-if="section === 'accounts'" class="mb-4">
              <CCardHeader class="admin-card-header">
                <div><strong>{{ t('admin.accountManagement') }}</strong><div class="text-body-secondary small">{{ t('admin.accountManagementHelp') }}</div></div>
                <label class="admin-search"><Search :size="16" /><span class="visually-hidden">{{ t('admin.searchAccounts') }}</span><input v-model="accountQuery" type="search" :placeholder="t('admin.searchAccounts')"></label>
              </CCardHeader>
              <CCardBody class="p-0">
                <CTable align="middle" class="mb-0 admin-table" hover responsive>
                  <CTableHead><CTableRow><CTableHeaderCell class="ps-4">{{ t('admin.account') }}</CTableHeaderCell><CTableHeaderCell>{{ t('admin.role') }}</CTableHeaderCell><CTableHeaderCell>{{ t('admin.security') }}</CTableHeaderCell><CTableHeaderCell>{{ t('admin.status') }}</CTableHeaderCell><CTableHeaderCell class="text-end pe-4">{{ t('admin.actions') }}</CTableHeaderCell></CTableRow></CTableHead>
                  <CTableBody>
                    <CTableRow v-for="account in accounts" :key="account.id">
                      <CTableDataCell class="ps-4"><strong>{{ account.displayName }}</strong><div class="text-body-secondary small">{{ account.email }} · @{{ account.username }}</div></CTableDataCell>
                      <CTableDataCell><div class="d-flex flex-wrap gap-1"><CBadge v-if="account.owner" color="primary">{{ t('admin.owner') }}</CBadge><CBadge v-if="!account.owner && account.administrator" color="info">{{ t('admin.administrator') }}</CBadge><CBadge v-if="account.sourceMaintainer" color="warning">{{ t('admin.sourceMaintainer') }}</CBadge><CBadge v-if="account.rfcMaintainer" color="dark">{{ t('admin.rfcMaintainer') }}</CBadge><span v-if="!account.owner && !account.administrator && !account.sourceMaintainer && !account.rfcMaintainer" class="text-body-secondary">{{ t('admin.member') }}</span></div></CTableDataCell>
                      <CTableDataCell><span :class="account.totpEnabled ? 'text-success' : 'text-danger'">TOTP</span><div class="text-body-secondary small">{{ account.recoveryVerified ? t('admin.recoveryVerified') : t('admin.recoveryUnverified') }}</div></CTableDataCell>
                      <CTableDataCell><CBadge :color="badgeColor(account.status)">{{ account.status }}</CBadge></CTableDataCell>
                      <CTableDataCell class="text-end pe-4 admin-actions">
                        <CButton v-if="auth.account?.owner && !account.owner && account.id !== auth.account.id" color="secondary" variant="ghost" size="sm" @click="ask({ kind: 'role', account, administrator: !account.administrator })">{{ account.administrator ? t('admin.removeAdmin') : t('admin.makeAdmin') }}</CButton>
                        <CButton v-if="auth.account?.owner && !account.owner" color="warning" variant="ghost" size="sm" @click="ask({ kind: 'source-maintainer', account, enabled: !account.sourceMaintainer })">{{ account.sourceMaintainer ? t('admin.removeSourceMaintainer') : t('admin.makeSourceMaintainer') }}</CButton>
                        <CButton v-if="auth.account?.owner && !account.owner" color="dark" variant="ghost" size="sm" @click="ask({ kind: 'rfc-maintainer', account, enabled: !account.rfcMaintainer })">{{ account.rfcMaintainer ? t('admin.removeRFCMaintainer') : t('admin.makeRFCMaintainer') }}</CButton>
                        <CButton v-if="account.id !== auth.account?.id && (auth.account?.owner || !account.owner)" :color="account.status === 'active' ? 'danger' : 'success'" variant="outline" size="sm" @click="ask({ kind: 'status', account, status: account.status === 'active' ? 'suspended' : 'active' })">{{ account.status === 'active' ? t('admin.suspend') : t('admin.restore') }}</CButton>
                      </CTableDataCell>
                    </CTableRow>
                    <CTableRow v-if="accounts.length === 0"><CTableDataCell colspan="5" class="py-4 text-center text-body-secondary">{{ t('admin.noAccounts') }}</CTableDataCell></CTableRow>
                  </CTableBody>
                </CTable>
              </CCardBody>
            </CCard>

            <CCard v-if="section === 'mailbox'" class="mb-4">
              <CCardHeader><strong>{{ t('admin.managementMailbox') }}</strong></CCardHeader>
              <CCardBody>
                <p class="mb-2">{{ locale === 'ko' ? '관리자 설정과 운영 메일 처리를 분리했습니다.' : 'Administrative settings and operational mail are now separated.' }}</p>
                <p class="text-body-secondary mb-3">{{ locale === 'ko' ? '통합 메일의 읽기, 답장, 발송은 Mail의 팀 메일에서 처리하세요. 이 관리자 화면은 계정 권한, 보안, 전달 상태 같은 플랫폼 설정에 집중합니다.' : 'Read, reply to, and send shared messages from Team mail in Mail. This admin area remains focused on platform settings such as account access, security, and delivery status.' }}</p>
                <div class="d-flex flex-wrap gap-2 mb-4" aria-label="Management mail addresses">
                  <CBadge color="secondary">admin@wave-lang.dev</CBadge>
                  <CBadge color="secondary">help@wave-lang.dev</CBadge>
                  <CBadge color="secondary">info@wave-lang.dev</CBadge>
                  <CBadge color="secondary">support@wave-lang.dev</CBadge>
                </div>
                <RouterLink class="btn btn-primary d-inline-flex align-items-center gap-2" to="/mail/team">
                  <MailWarning :size="17" />{{ locale === 'ko' ? '팀 메일 열기' : 'Open Team mail' }}
                </RouterLink>
              </CCardBody>
            </CCard>

            <CCard v-if="section === 'mail-queue'" class="mb-4">
              <CCardHeader><strong>{{ t('admin.mailQueue') }}</strong><span class="ms-2 text-body-secondary small">{{ t('admin.recentDeliveries') }}</span></CCardHeader>
              <CCardBody class="p-0">
                <div class="admin-status-strip"><span>{{ t('admin.queued') }} <strong>{{ snapshot.mail.queued }}</strong></span><span>{{ t('admin.delivering') }} <strong>{{ snapshot.mail.delivering }}</strong></span><span>{{ t('admin.deferred') }} <strong>{{ snapshot.mail.deferred }}</strong></span><span>{{ t('admin.failed') }} <strong>{{ snapshot.mail.failed }}</strong></span><span>{{ t('admin.delivered') }} <strong>{{ snapshot.mail.delivered }}</strong></span></div>
                <CTable align="middle" class="mb-0 admin-table" responsive>
                  <CTableHead><CTableRow><CTableHeaderCell class="ps-4">{{ t('admin.recipient') }}</CTableHeaderCell><CTableHeaderCell>{{ t('admin.destination') }}</CTableHeaderCell><CTableHeaderCell>{{ t('admin.status') }}</CTableHeaderCell><CTableHeaderCell>{{ t('admin.attempts') }}</CTableHeaderCell><CTableHeaderCell>{{ t('admin.updated') }}</CTableHeaderCell></CTableRow></CTableHead>
                  <CTableBody><CTableRow v-for="delivery in snapshot.deliveries" :key="delivery.id"><CTableDataCell class="ps-4"><strong>{{ delivery.recipient }}</strong><div v-if="delivery.lastError" class="admin-error-detail" :title="delivery.lastError">{{ delivery.lastError }}</div></CTableDataCell><CTableDataCell>{{ delivery.destination }}</CTableDataCell><CTableDataCell><CBadge :color="badgeColor(delivery.status)">{{ delivery.status }}</CBadge></CTableDataCell><CTableDataCell>{{ delivery.attempts }}</CTableDataCell><CTableDataCell>{{ formatDate(delivery.lastAttemptAt || delivery.createdAt) }}</CTableDataCell></CTableRow><CTableRow v-if="snapshot.deliveries.length === 0"><CTableDataCell colspan="5" class="py-4 text-center text-body-secondary">{{ t('admin.noDeliveries') }}</CTableDataCell></CTableRow></CTableBody>
                </CTable>
              </CCardBody>
            </CCard>

            <CCard v-if="section === 'git-mirrors'" class="mb-4">
              <CCardHeader><strong>{{ t('admin.gitMirrors') }}</strong><span class="ms-2 text-body-secondary small">{{ t('admin.syncEvery') }} {{ snapshot.gitSyncInterval }}</span></CCardHeader>
              <CCardBody class="p-0"><CTable align="middle" class="mb-0 admin-table" responsive><CTableHead><CTableRow><CTableHeaderCell class="ps-4">{{ t('admin.repository') }}</CTableHeaderCell><CTableHeaderCell>{{ t('admin.branch') }}</CTableHeaderCell><CTableHeaderCell>{{ t('admin.head') }}</CTableHeaderCell><CTableHeaderCell>{{ t('admin.lastFetched') }}</CTableHeaderCell><CTableHeaderCell>{{ t('admin.status') }}</CTableHeaderCell></CTableRow></CTableHead><CTableBody><CTableRow v-for="repository in snapshot.gitMirrors" :key="repository.id"><CTableDataCell class="ps-4"><strong>{{ repository.owner }}/{{ repository.name }}</strong></CTableDataCell><CTableDataCell>{{ repository.defaultBranch }}</CTableDataCell><CTableDataCell><code>{{ repository.headOid ? repository.headOid.slice(0, 10) : '—' }}</code></CTableDataCell><CTableDataCell>{{ formatDate(repository.headCommit?.authoredAt ?? '') }}</CTableDataCell><CTableDataCell><CBadge :color="badgeColor(repository.status)">{{ repository.status }}</CBadge></CTableDataCell></CTableRow></CTableBody></CTable></CCardBody>
            </CCard>

            <CRow v-if="section === 'security' || section === 'system'" :xs="{ cols: 1, gutter: 4 }" class="mb-4">
              <CCol v-if="section === 'security'"><CCard class="h-100"><CCardHeader><strong>{{ t('admin.security') }}</strong></CCardHeader><CCardBody><dl class="admin-definition-list"><div><dt>{{ t('admin.totpEnrollment') }}</dt><dd>{{ snapshot.security.totpAccounts }} / {{ snapshot.security.activeAccounts + snapshot.security.suspendedAccounts }}</dd></div><div><dt>{{ t('admin.verifiedRecovery') }}</dt><dd>{{ snapshot.security.verifiedRecoveries }}</dd></div><div><dt>{{ t('admin.registration') }}</dt><dd><CBadge :color="snapshot.security.registrationOpen ? 'success' : 'secondary'">{{ snapshot.security.registrationOpen ? t('admin.open') : t('admin.closed') }}</CBadge></dd></div><div><dt>Cloudflare Turnstile</dt><dd><CBadge :color="snapshot.security.turnstileEnabled ? 'success' : 'secondary'">{{ snapshot.security.turnstileEnabled ? t('admin.enabled') : t('admin.disabled') }}</CBadge></dd></div></dl></CCardBody></CCard></CCol>
              <CCol v-if="section === 'system'"><CCard class="h-100"><CCardHeader><strong>{{ t('admin.system') }}</strong><span class="ms-2 text-body-secondary small">{{ platform?.environment }} · {{ platform?.version }}</span></CCardHeader><CCardBody><dl class="admin-definition-list"><div><dt>{{ t('admin.databaseHealth') }}</dt><dd><CBadge :color="badgeColor(snapshot.storage.health)">{{ snapshot.storage.health }}</CBadge></dd></div><div><dt>{{ t('admin.lsmStorage') }}</dt><dd>{{ formatBytes(snapshot.storage.databaseBytes) }}</dd></div><div><dt>{{ t('admin.valueLog') }}</dt><dd>{{ formatBytes(snapshot.storage.valueLogBytes) }}</dd></div><div><dt>{{ t('admin.totalStorage') }}</dt><dd>{{ formatBytes(snapshot.storage.filesBytes) }}</dd></div></dl><form class="admin-timezone-setting" @submit.prevent="saveLunaStevTimeZone"><label>{{ t('admin.lunaStevTimeZone') }}<select v-model="lunaStevTimeZone" class="form-select"><option v-for="zone in timeZones" :key="zone" :value="zone">{{ zone }}</option></select></label><p class="text-body-secondary small">{{ t('admin.lunaStevTimeZoneHelp') }}</p><CButton color="primary" type="submit">{{ t('common.save') }}</CButton></form><div v-if="actionError" class="alert alert-danger mt-3">{{ actionError }}</div></CCardBody></CCard></CCol>
            </CRow>

            <CCard v-if="section === 'modules'" class="mb-4"><CCardHeader><strong>{{ t('admin.modules') }}</strong></CCardHeader><CCardBody class="p-0"><CTable align="middle" class="mb-0 admin-table" hover responsive><CTableHead><CTableRow><CTableHeaderCell class="ps-4">{{ t('admin.module') }}</CTableHeaderCell><CTableHeaderCell>{{ t('admin.status') }}</CTableHeaderCell><CTableHeaderCell>{{ t('admin.availability') }}</CTableHeaderCell></CTableRow></CTableHead><CTableBody><CTableRow v-for="module in modules" :key="module.name"><CTableDataCell class="ps-4"><strong>{{ module.name }}</strong></CTableDataCell><CTableDataCell>{{ module.status }}</CTableDataCell><CTableDataCell><CBadge :color="module.enabled ? 'success' : 'secondary'">{{ module.enabled ? t('admin.enabled') : t('admin.disabled') }}</CBadge></CTableDataCell></CTableRow></CTableBody></CTable></CCardBody></CCard>

            <CCard v-if="section === 'audit-log'" class="mb-4"><CCardHeader><strong>{{ t('admin.auditLog') }}</strong><span class="ms-2 text-body-secondary small">{{ t('admin.recentEvents') }}</span></CCardHeader><CCardBody class="p-0"><CTable align="middle" class="mb-0 admin-table" responsive><CTableHead><CTableRow><CTableHeaderCell class="ps-4">{{ t('admin.time') }}</CTableHeaderCell><CTableHeaderCell>{{ t('admin.actor') }}</CTableHeaderCell><CTableHeaderCell>{{ t('admin.action') }}</CTableHeaderCell><CTableHeaderCell>{{ t('admin.resource') }}</CTableHeaderCell><CTableHeaderCell>{{ t('admin.result') }}</CTableHeaderCell></CTableRow></CTableHead><CTableBody><CTableRow v-for="event in snapshot.auditLog" :key="event.id"><CTableDataCell class="ps-4">{{ formatDate(event.occurredAt) }}</CTableDataCell><CTableDataCell><code>{{ event.actorId }}</code></CTableDataCell><CTableDataCell>{{ event.action }}</CTableDataCell><CTableDataCell><code>{{ event.resourceId }}</code></CTableDataCell><CTableDataCell><CBadge :color="badgeColor(event.result)">{{ event.result }}</CBadge></CTableDataCell></CTableRow><CTableRow v-if="snapshot.auditLog.length === 0"><CTableDataCell colspan="5" class="py-4 text-center text-body-secondary">{{ t('admin.noAuditEvents') }}</CTableDataCell></CTableRow></CTableBody></CTable></CCardBody></CCard>
          </template>
        </CContainer>
      </div>
    </div>

    <CModal :visible="Boolean(pendingAction)" alignment="center" @close="pendingAction = null">
      <CModalHeader><CModalTitle>{{ t('admin.confirmAction') }}</CModalTitle></CModalHeader>
      <CModalBody v-if="pendingAction"><p>{{ pendingAction.kind === 'status' ? (pendingAction.status === 'suspended' ? t('admin.confirmSuspend') : t('admin.confirmRestore')) : pendingAction.kind === 'role' ? (pendingAction.administrator ? t('admin.confirmMakeAdmin') : t('admin.confirmRemoveAdmin')) : pendingAction.kind === 'source-maintainer' ? (pendingAction.enabled ? t('admin.confirmMakeSourceMaintainer') : t('admin.confirmRemoveSourceMaintainer')) : (pendingAction.enabled ? t('admin.confirmMakeRFCMaintainer') : t('admin.confirmRemoveRFCMaintainer')) }}</p><strong>{{ pendingAction.account.displayName }}</strong><div class="text-body-secondary small">{{ pendingAction.account.email }}</div><div v-if="actionError" class="alert alert-danger mt-3 mb-0" role="alert">{{ actionError }}</div></CModalBody>
      <CModalFooter><CButton color="secondary" variant="outline" :disabled="actionLoading" @click="pendingAction = null">{{ t('common.cancel') }}</CButton><CButton :color="pendingAction?.kind === 'status' && pendingAction.status === 'suspended' ? 'danger' : 'primary'" :disabled="actionLoading" @click="confirmAction">{{ t('admin.confirm') }}</CButton></CModalFooter>
    </CModal>

  </main>
</template>

<style>
body.coreui-admin-active { background-color: var(--cui-tertiary-bg); }
.coreui-admin-template { min-height: 100vh; font-family: var(--cui-body-font-family); }
.coreui-admin-template .wrapper { width: 100%; padding-inline: var(--cui-sidebar-occupy-start, 0) var(--cui-sidebar-occupy-end, 0); transition: padding .15s; }
.coreui-admin-template .header > .container-fluid, .coreui-admin-template .sidebar-header { min-height: calc(4rem + 1px); }
.coreui-admin-template .header > .container-fluid { display: flex; }
.coreui-admin-template .admin-brand { display: flex; align-items: baseline; gap: .45rem; font-size: 1.1rem; font-weight: 700; }
.coreui-admin-template .admin-brand span { color: var(--cui-secondary-color); font-size: .72rem; font-weight: 500; }
.coreui-admin-template .sidebar-footer .nav-link { display: flex; align-items: center; color: var(--cui-sidebar-nav-link-color); text-decoration: none; }
.coreui-admin-template .body { padding-bottom: 2rem; }
.admin-heading, .admin-card-header { display: flex; align-items: center; justify-content: space-between; gap: 20px; }
.admin-preference-select { width: auto; max-width: 150px; }
.admin-theme-toggle { display: grid; width: 32px; height: 31px; flex: 0 0 auto; place-items: center; padding: 0; border: 1px solid var(--cui-border-color); border-radius: var(--cui-border-radius); background: var(--cui-body-bg); color: var(--cui-body-color); }
.admin-theme-toggle:hover { background: var(--cui-tertiary-bg); }
.admin-heading .btn { display: inline-flex; align-items: center; gap: 6px; white-space: nowrap; }
.admin-loading { padding: 48px 0; text-align: center; }
.admin-search { display: flex; width: min(100%, 320px); height: 34px; align-items: center; gap: 7px; padding: 0 9px; border: 1px solid var(--cui-border-color); border-radius: var(--cui-border-radius); background: var(--cui-body-bg); color: var(--cui-secondary-color); }
.admin-search input { min-width: 0; flex: 1; border: 0; outline: 0; background: transparent; color: var(--cui-body-color); }
.admin-table td { font-size: .84rem; }
.admin-actions { white-space: nowrap; }
.admin-actions .btn + .btn { margin-left: 5px; }
.admin-error-detail { max-width: 360px; overflow: hidden; margin-top: 2px; color: var(--cui-danger-text-emphasis); font-size: .75rem; text-overflow: ellipsis; white-space: nowrap; }
.admin-status-strip { display: flex; flex-wrap: wrap; gap: 20px; padding: 12px 24px; border-bottom: 1px solid var(--cui-border-color); color: var(--cui-secondary-color); font-size: .8rem; }
.admin-status-strip strong { margin-left: 4px; color: var(--cui-body-color); }
.admin-definition-list { margin: 0; }
.admin-definition-list > div { display: flex; align-items: center; justify-content: space-between; gap: 20px; padding: 10px 0; border-bottom: 1px solid var(--cui-border-color); }
.admin-definition-list > div:last-child { border-bottom: 0; }
.admin-definition-list dt { color: var(--cui-secondary-color); font-weight: 500; }
.admin-definition-list dd { margin: 0; font-weight: 650; }
.admin-timezone-setting { max-width: 520px; margin-top: 24px; padding-top: 20px; border-top: 1px solid var(--cui-border-color); }
.admin-timezone-setting label { display: grid; gap: 7px; font-weight: 600; }
.admin-mail-body { max-height: 52vh; overflow: auto; margin: 20px 0 0; padding: 16px; border: 1px solid var(--cui-border-color); background: var(--cui-tertiary-bg); color: var(--cui-body-color); font: 13px/1.6 ui-monospace, monospace; white-space: pre-wrap; }
.admin-blog-layout { display: grid; grid-template-columns: minmax(240px, 320px) minmax(0, 1fr); min-height: 620px; }
.admin-blog-list { border-right: 1px solid var(--cui-border-color); }
.admin-blog-list > button { display: grid; width: 100%; grid-template-columns: minmax(0, 1fr) auto; gap: 12px; padding: 13px 16px; border: 0; border-bottom: 1px solid var(--cui-border-color); background: transparent; color: var(--cui-body-color); text-align: left; }
.admin-blog-list > button:hover, .admin-blog-list > button.active { background: var(--cui-tertiary-bg); }
.admin-blog-list span, .admin-blog-list strong, .admin-blog-list small { display: block; min-width: 0; }
.admin-blog-list strong { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.admin-blog-list small, .admin-blog-list time { margin-top: 3px; color: var(--cui-secondary-color); font-size: .72rem; }
.admin-blog-editor { min-width: 0; padding: 22px 24px 32px; }
.admin-blog-fields { display: grid; grid-template-columns: minmax(0, 2fr) minmax(150px, 1fr); gap: 16px; margin: 18px 0; }
.admin-blog-fields label { display: grid; align-content: start; gap: 6px; font-size: .82rem; font-weight: 600; }
.admin-blog-fields .wide { grid-column: 1 / -1; }
.admin-blog-content { min-height: 340px; font: 13px/1.55 ui-monospace, SFMono-Regular, Consolas, monospace; resize: vertical; }
.admin-blog-preview { margin-top: 28px; padding-top: 20px; border-top: 1px solid var(--cui-border-color); }
.admin-blog-preview .markdown-content { max-width: 820px; }
.admin-webhook-form { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 14px 18px; }
.admin-webhook-form > label { display: grid; gap: 6px; font-size: .84rem; font-weight: 600; }
.admin-webhook-form .wide { grid-column: 1 / -1; }
.admin-webhook-form fieldset { display: flex; flex-wrap: wrap; gap: 8px 18px; margin: 0; padding: 12px 0 0; border: 0; border-top: 1px solid var(--cui-border-color); }
.admin-webhook-form legend { width: 100%; margin: 0; font-size: .84rem; font-weight: 600; }
.admin-webhook-event { display: inline-flex !important; align-items: center; gap: 6px !important; font-size: .82rem !important; font-weight: 500 !important; }
@media (max-width: 991.98px) { .coreui-admin-template .wrapper { padding-inline-start: 0; }.coreui-admin-template .admin-mobile-sidebar { --cui-is-mobile: 1; } }
@media (max-width: 767.98px) { .admin-card-header { align-items: stretch; flex-direction: column; }.admin-search { width: 100%; }.admin-heading { align-items: flex-start; }.coreui-admin-template .container-fluid { padding-right: 16px !important; padding-left: 16px !important; }.admin-preference-select { max-width: 96px; }.coreui-admin-template .header .gap-3 { gap: .4rem !important; }.admin-blog-layout { grid-template-columns: 1fr; }.admin-blog-list { max-height: 260px; overflow-y: auto; border-right: 0; border-bottom: 1px solid var(--cui-border-color); }.admin-blog-fields, .admin-webhook-form { grid-template-columns: 1fr; }.admin-blog-fields .wide, .admin-webhook-form .wide { grid-column: auto; }.admin-blog-editor { padding: 18px 16px 26px; } }
@media (max-width: 480px) { .coreui-admin-template .header strong { display: none; }.admin-preference-select { max-width: 88px; } }
</style>
