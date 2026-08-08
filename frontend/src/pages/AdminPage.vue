<script setup lang="ts">
import coreuiStylesheet from '@coreui/coreui/dist/css/coreui.min.css?url'
import {
  CBadge, CButton, CCard, CCardBody, CCardHeader, CCol, CContainer, CHeader, CHeaderToggler,
  CModal, CModalBody, CModalFooter, CModalHeader, CModalTitle, CNavItem, CNavTitle, CRow,
  CSidebar, CSidebarBrand, CSidebarFooter, CSidebarHeader, CSidebarNav, CTable, CTableBody,
  CTableDataCell, CTableHead, CTableHeaderCell, CTableRow,
} from '@coreui/vue'
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import {
  ArrowLeft, Boxes, Database, FileClock, Gauge, GitBranch, MailWarning, Menu, RefreshCw,
  Search, ShieldCheck, Users,
} from '@lucide/vue'

import {
  getAdminSnapshot, getModules, getPlatformStatus, updateAdminAccountRole, updateAdminAccountStatus,
  type AdminAccount, type AdminSnapshot, type ModuleStatus, type PlatformStatus,
} from '../services/http'
import { useI18n } from '../i18n'
import type { Locale } from '../i18n'
import { useAuthStore } from '../stores/auth'
import ThemeSelector from '../components/platform/ThemeSelector.vue'

type PendingAction = { kind: 'status'; account: AdminAccount; status: 'active' | 'suspended' }
  | { kind: 'role'; account: AdminAccount; administrator: boolean }

const router = useRouter()
const auth = useAuthStore()
const { locale, setLocale, t } = useI18n()
const modules = ref<ModuleStatus[]>([])
const snapshot = ref<AdminSnapshot | null>(null)
const platform = ref<PlatformStatus | null>(null)
const error = ref('')
const actionError = ref('')
const loading = ref(true)
const actionLoading = ref(false)
const sidebarVisible = ref(window.innerWidth >= 992)
const accountQuery = ref('')
const pendingAction = ref<PendingAction | null>(null)
let stylesheet: HTMLLinkElement | null = null

const accounts = computed(() => {
  const needle = accountQuery.value.trim().toLowerCase()
  if (!needle) return snapshot.value?.accounts ?? []
  return (snapshot.value?.accounts ?? []).filter((account) =>
    `${account.username} ${account.displayName} ${account.email}`.toLowerCase().includes(needle))
})
const mailAttention = computed(() => (snapshot.value?.mail.queued ?? 0) + (snapshot.value?.mail.delivering ?? 0)
  + (snapshot.value?.mail.deferred ?? 0) + (snapshot.value?.mail.failed ?? 0))

onMounted(async () => {
  stylesheet = document.createElement('link')
  stylesheet.rel = 'stylesheet'
  stylesheet.href = coreuiStylesheet
  stylesheet.dataset.waveAdminStyles = 'coreui'
  document.head.append(stylesheet)
  document.body.classList.add('coreui-admin-active')
  await load()
})

onBeforeUnmount(() => {
  stylesheet?.remove()
  document.body.classList.remove('coreui-admin-active')
})

async function load() {
  loading.value = true
  error.value = ''
  try {
    const [moduleResult, snapshotResult, platformResult] = await Promise.all([
      getModules(), getAdminSnapshot(), getPlatformStatus(),
    ])
    modules.value = moduleResult
    snapshot.value = snapshotResult
    platform.value = platformResult
  } catch (reason) {
    error.value = reason instanceof Error ? reason.message : t('common.loadError')
  } finally {
    loading.value = false
  }
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
    else await updateAdminAccountRole(action.account.id, action.administrator)
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
    <CSidebar class="border-end" color-scheme="dark" position="fixed" :visible="sidebarVisible" @visible-change="sidebarVisible = $event">
      <CSidebarHeader class="border-bottom">
        <CSidebarBrand as="div" class="admin-brand">Wave <span>{{ t('admin.title') }}</span></CSidebarBrand>
      </CSidebarHeader>
      <CSidebarNav>
        <CNavItem><a class="nav-link active" href="#overview"><Gauge class="nav-icon" :size="20" />{{ t('admin.overview') }}</a></CNavItem>
        <CNavTitle>{{ t('admin.management') }}</CNavTitle>
        <CNavItem><a class="nav-link" href="#accounts"><Users class="nav-icon" :size="20" />{{ t('admin.accounts') }}</a></CNavItem>
        <CNavItem><a class="nav-link" href="#mail-queue"><MailWarning class="nav-icon" :size="20" />{{ t('admin.mailQueue') }}</a></CNavItem>
        <CNavItem><a class="nav-link" href="#git-mirrors"><GitBranch class="nav-icon" :size="20" />{{ t('admin.gitMirrors') }}</a></CNavItem>
        <CNavItem><a class="nav-link" href="#audit-log"><FileClock class="nav-icon" :size="20" />{{ t('admin.auditLog') }}</a></CNavItem>
        <CNavTitle>{{ t('admin.platform') }}</CNavTitle>
        <CNavItem><a class="nav-link" href="#security"><ShieldCheck class="nav-icon" :size="20" />{{ t('admin.security') }}</a></CNavItem>
        <CNavItem><a class="nav-link" href="#modules"><Boxes class="nav-icon" :size="20" />{{ t('admin.modules') }}</a></CNavItem>
        <CNavItem><a class="nav-link" href="#system"><Database class="nav-icon" :size="20" />{{ t('admin.system') }}</a></CNavItem>
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
            <ThemeSelector class="form-select form-select-sm admin-preference-select" />
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
              <div class="text-body-secondary small mb-1">{{ t('admin.title') }} / {{ t('admin.overview') }}</div>
              <h1 class="h3 mb-1">{{ t('admin.platformOverview') }}</h1>
              <p class="text-body-secondary mb-0">{{ t('admin.lead') }}</p>
            </div>
            <CButton color="secondary" variant="outline" size="sm" :disabled="loading" @click="load"><RefreshCw :size="15" /> {{ t('admin.refresh') }}</CButton>
          </div>

          <div v-if="error" class="alert alert-danger" role="alert">{{ error }}</div>
          <div v-if="loading && !snapshot" class="admin-loading text-body-secondary">{{ t('common.loading') }}</div>

          <template v-if="snapshot">
            <CRow :xs="{ cols: 1, gutter: 3 }" :sm="{ cols: 2 }" :xl="{ cols: 4 }" class="mb-4">
              <CCol><CCard class="h-100"><CCardBody><div class="text-body-secondary small">{{ t('admin.accounts') }}</div><div class="fs-3 fw-semibold">{{ snapshot.security.activeAccounts + snapshot.security.suspendedAccounts }}</div><small class="text-body-secondary">{{ snapshot.security.suspendedAccounts }} {{ t('admin.suspended').toLowerCase() }}</small></CCardBody></CCard></CCol>
              <CCol><CCard class="h-100"><CCardBody><div class="text-body-secondary small">{{ t('admin.mailAttention') }}</div><div class="fs-3 fw-semibold">{{ mailAttention }}</div><small class="text-body-secondary">{{ snapshot.mail.failed }} {{ t('admin.failed').toLowerCase() }}</small></CCardBody></CCard></CCol>
              <CCol><CCard class="h-100"><CCardBody><div class="text-body-secondary small">{{ t('admin.gitMirrors') }}</div><div class="fs-3 fw-semibold">{{ snapshot.gitMirrors.filter((item) => item.status === 'ready').length }} / {{ snapshot.gitMirrors.length }}</div><small class="text-body-secondary">{{ t('admin.ready') }}</small></CCardBody></CCard></CCol>
              <CCol><CCard class="h-100"><CCardBody><div class="text-body-secondary small">{{ t('admin.storageUsed') }}</div><div class="fs-3 fw-semibold">{{ formatBytes(snapshot.storage.filesBytes) }}</div><small class="text-body-secondary">{{ snapshot.storage.health }}</small></CCardBody></CCard></CCol>
            </CRow>

            <CCard id="accounts" class="mb-4">
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
                      <CTableDataCell><CBadge v-if="account.owner" color="primary">{{ t('admin.owner') }}</CBadge><CBadge v-else-if="account.administrator" color="info">{{ t('admin.administrator') }}</CBadge><span v-else class="text-body-secondary">{{ t('admin.member') }}</span></CTableDataCell>
                      <CTableDataCell><span :class="account.totpEnabled ? 'text-success' : 'text-danger'">TOTP</span><div class="text-body-secondary small">{{ account.recoveryVerified ? t('admin.recoveryVerified') : t('admin.recoveryUnverified') }}</div></CTableDataCell>
                      <CTableDataCell><CBadge :color="badgeColor(account.status)">{{ account.status }}</CBadge></CTableDataCell>
                      <CTableDataCell class="text-end pe-4 admin-actions">
                        <CButton v-if="auth.account?.owner && !account.owner && account.id !== auth.account.id" color="secondary" variant="ghost" size="sm" @click="ask({ kind: 'role', account, administrator: !account.administrator })">{{ account.administrator ? t('admin.removeAdmin') : t('admin.makeAdmin') }}</CButton>
                        <CButton v-if="account.id !== auth.account?.id && (auth.account?.owner || !account.owner)" :color="account.status === 'active' ? 'danger' : 'success'" variant="outline" size="sm" @click="ask({ kind: 'status', account, status: account.status === 'active' ? 'suspended' : 'active' })">{{ account.status === 'active' ? t('admin.suspend') : t('admin.restore') }}</CButton>
                      </CTableDataCell>
                    </CTableRow>
                    <CTableRow v-if="accounts.length === 0"><CTableDataCell colspan="5" class="py-4 text-center text-body-secondary">{{ t('admin.noAccounts') }}</CTableDataCell></CTableRow>
                  </CTableBody>
                </CTable>
              </CCardBody>
            </CCard>

            <CCard id="mail-queue" class="mb-4">
              <CCardHeader><strong>{{ t('admin.mailQueue') }}</strong><span class="ms-2 text-body-secondary small">{{ t('admin.recentDeliveries') }}</span></CCardHeader>
              <CCardBody class="p-0">
                <div class="admin-status-strip"><span>{{ t('admin.queued') }} <strong>{{ snapshot.mail.queued }}</strong></span><span>{{ t('admin.delivering') }} <strong>{{ snapshot.mail.delivering }}</strong></span><span>{{ t('admin.deferred') }} <strong>{{ snapshot.mail.deferred }}</strong></span><span>{{ t('admin.failed') }} <strong>{{ snapshot.mail.failed }}</strong></span><span>{{ t('admin.delivered') }} <strong>{{ snapshot.mail.delivered }}</strong></span></div>
                <CTable align="middle" class="mb-0 admin-table" responsive>
                  <CTableHead><CTableRow><CTableHeaderCell class="ps-4">{{ t('admin.recipient') }}</CTableHeaderCell><CTableHeaderCell>{{ t('admin.destination') }}</CTableHeaderCell><CTableHeaderCell>{{ t('admin.status') }}</CTableHeaderCell><CTableHeaderCell>{{ t('admin.attempts') }}</CTableHeaderCell><CTableHeaderCell>{{ t('admin.updated') }}</CTableHeaderCell></CTableRow></CTableHead>
                  <CTableBody><CTableRow v-for="delivery in snapshot.deliveries" :key="delivery.id"><CTableDataCell class="ps-4"><strong>{{ delivery.recipient }}</strong><div v-if="delivery.lastError" class="admin-error-detail" :title="delivery.lastError">{{ delivery.lastError }}</div></CTableDataCell><CTableDataCell>{{ delivery.destination }}</CTableDataCell><CTableDataCell><CBadge :color="badgeColor(delivery.status)">{{ delivery.status }}</CBadge></CTableDataCell><CTableDataCell>{{ delivery.attempts }}</CTableDataCell><CTableDataCell>{{ formatDate(delivery.lastAttemptAt || delivery.createdAt) }}</CTableDataCell></CTableRow><CTableRow v-if="snapshot.deliveries.length === 0"><CTableDataCell colspan="5" class="py-4 text-center text-body-secondary">{{ t('admin.noDeliveries') }}</CTableDataCell></CTableRow></CTableBody>
                </CTable>
              </CCardBody>
            </CCard>

            <CCard id="git-mirrors" class="mb-4">
              <CCardHeader><strong>{{ t('admin.gitMirrors') }}</strong><span class="ms-2 text-body-secondary small">{{ t('admin.syncEvery') }} {{ snapshot.gitSyncInterval }}</span></CCardHeader>
              <CCardBody class="p-0"><CTable align="middle" class="mb-0 admin-table" responsive><CTableHead><CTableRow><CTableHeaderCell class="ps-4">{{ t('admin.repository') }}</CTableHeaderCell><CTableHeaderCell>{{ t('admin.branch') }}</CTableHeaderCell><CTableHeaderCell>{{ t('admin.head') }}</CTableHeaderCell><CTableHeaderCell>{{ t('admin.lastFetched') }}</CTableHeaderCell><CTableHeaderCell>{{ t('admin.status') }}</CTableHeaderCell></CTableRow></CTableHead><CTableBody><CTableRow v-for="repository in snapshot.gitMirrors" :key="repository.id"><CTableDataCell class="ps-4"><strong>{{ repository.owner }}/{{ repository.name }}</strong></CTableDataCell><CTableDataCell>{{ repository.defaultBranch }}</CTableDataCell><CTableDataCell><code>{{ repository.headOid ? repository.headOid.slice(0, 10) : '—' }}</code></CTableDataCell><CTableDataCell>{{ formatDate(repository.headCommit?.authoredAt ?? '') }}</CTableDataCell><CTableDataCell><CBadge :color="badgeColor(repository.status)">{{ repository.status }}</CBadge></CTableDataCell></CTableRow></CTableBody></CTable></CCardBody>
            </CCard>

            <CRow :xs="{ cols: 1, gutter: 4 }" :xl="{ cols: 2 }" class="mb-4">
              <CCol><CCard id="security" class="h-100"><CCardHeader><strong>{{ t('admin.security') }}</strong></CCardHeader><CCardBody><dl class="admin-definition-list"><div><dt>{{ t('admin.totpEnrollment') }}</dt><dd>{{ snapshot.security.totpAccounts }} / {{ snapshot.security.activeAccounts + snapshot.security.suspendedAccounts }}</dd></div><div><dt>{{ t('admin.verifiedRecovery') }}</dt><dd>{{ snapshot.security.verifiedRecoveries }}</dd></div><div><dt>{{ t('admin.registration') }}</dt><dd><CBadge :color="snapshot.security.registrationOpen ? 'success' : 'secondary'">{{ snapshot.security.registrationOpen ? t('admin.open') : t('admin.closed') }}</CBadge></dd></div><div><dt>Cloudflare Turnstile</dt><dd><CBadge :color="snapshot.security.turnstileEnabled ? 'success' : 'secondary'">{{ snapshot.security.turnstileEnabled ? t('admin.enabled') : t('admin.disabled') }}</CBadge></dd></div></dl></CCardBody></CCard></CCol>
              <CCol><CCard id="system" class="h-100"><CCardHeader><strong>{{ t('admin.system') }}</strong><span class="ms-2 text-body-secondary small">{{ platform?.environment }} · {{ platform?.version }}</span></CCardHeader><CCardBody><dl class="admin-definition-list"><div><dt>{{ t('admin.databaseHealth') }}</dt><dd><CBadge :color="badgeColor(snapshot.storage.health)">{{ snapshot.storage.health }}</CBadge></dd></div><div><dt>{{ t('admin.lsmStorage') }}</dt><dd>{{ formatBytes(snapshot.storage.databaseBytes) }}</dd></div><div><dt>{{ t('admin.valueLog') }}</dt><dd>{{ formatBytes(snapshot.storage.valueLogBytes) }}</dd></div><div><dt>{{ t('admin.totalStorage') }}</dt><dd>{{ formatBytes(snapshot.storage.filesBytes) }}</dd></div></dl></CCardBody></CCard></CCol>
            </CRow>

            <CCard id="modules" class="mb-4"><CCardHeader><strong>{{ t('admin.modules') }}</strong></CCardHeader><CCardBody class="p-0"><CTable align="middle" class="mb-0 admin-table" hover responsive><CTableHead><CTableRow><CTableHeaderCell class="ps-4">{{ t('admin.module') }}</CTableHeaderCell><CTableHeaderCell>{{ t('admin.status') }}</CTableHeaderCell><CTableHeaderCell>{{ t('admin.availability') }}</CTableHeaderCell></CTableRow></CTableHead><CTableBody><CTableRow v-for="module in modules" :key="module.name"><CTableDataCell class="ps-4"><strong>{{ module.name }}</strong></CTableDataCell><CTableDataCell>{{ module.status }}</CTableDataCell><CTableDataCell><CBadge :color="module.enabled ? 'success' : 'secondary'">{{ module.enabled ? t('admin.enabled') : t('admin.disabled') }}</CBadge></CTableDataCell></CTableRow></CTableBody></CTable></CCardBody></CCard>

            <CCard id="audit-log" class="mb-4"><CCardHeader><strong>{{ t('admin.auditLog') }}</strong><span class="ms-2 text-body-secondary small">{{ t('admin.recentEvents') }}</span></CCardHeader><CCardBody class="p-0"><CTable align="middle" class="mb-0 admin-table" responsive><CTableHead><CTableRow><CTableHeaderCell class="ps-4">{{ t('admin.time') }}</CTableHeaderCell><CTableHeaderCell>{{ t('admin.actor') }}</CTableHeaderCell><CTableHeaderCell>{{ t('admin.action') }}</CTableHeaderCell><CTableHeaderCell>{{ t('admin.resource') }}</CTableHeaderCell><CTableHeaderCell>{{ t('admin.result') }}</CTableHeaderCell></CTableRow></CTableHead><CTableBody><CTableRow v-for="event in snapshot.auditLog" :key="event.id"><CTableDataCell class="ps-4">{{ formatDate(event.occurredAt) }}</CTableDataCell><CTableDataCell><code>{{ event.actorId }}</code></CTableDataCell><CTableDataCell>{{ event.action }}</CTableDataCell><CTableDataCell><code>{{ event.resourceId }}</code></CTableDataCell><CTableDataCell><CBadge :color="badgeColor(event.result)">{{ event.result }}</CBadge></CTableDataCell></CTableRow><CTableRow v-if="snapshot.auditLog.length === 0"><CTableDataCell colspan="5" class="py-4 text-center text-body-secondary">{{ t('admin.noAuditEvents') }}</CTableDataCell></CTableRow></CTableBody></CTable></CCardBody></CCard>
          </template>
        </CContainer>
      </div>
    </div>

    <CModal :visible="Boolean(pendingAction)" alignment="center" @close="pendingAction = null">
      <CModalHeader><CModalTitle>{{ t('admin.confirmAction') }}</CModalTitle></CModalHeader>
      <CModalBody v-if="pendingAction"><p>{{ pendingAction.kind === 'status' ? (pendingAction.status === 'suspended' ? t('admin.confirmSuspend') : t('admin.confirmRestore')) : (pendingAction.administrator ? t('admin.confirmMakeAdmin') : t('admin.confirmRemoveAdmin')) }}</p><strong>{{ pendingAction.account.displayName }}</strong><div class="text-body-secondary small">{{ pendingAction.account.email }}</div><div v-if="actionError" class="alert alert-danger mt-3 mb-0" role="alert">{{ actionError }}</div></CModalBody>
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
@media (max-width: 991.98px) { .coreui-admin-template .wrapper { padding-inline-start: 0; } }
@media (max-width: 767.98px) { .admin-card-header { align-items: stretch; flex-direction: column; }.admin-search { width: 100%; }.admin-heading { align-items: flex-start; }.coreui-admin-template .container-fluid { padding-right: 16px !important; padding-left: 16px !important; }.admin-preference-select { max-width: 96px; }.coreui-admin-template .header .gap-3 { gap: .4rem !important; } }
@media (max-width: 480px) { .coreui-admin-template .header strong { display: none; }.admin-preference-select { max-width: 88px; } }
</style>
