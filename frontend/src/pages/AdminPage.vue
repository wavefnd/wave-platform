<script setup lang="ts">
import coreuiStylesheet from '@coreui/coreui/dist/css/coreui.min.css?url'
import {
  CBadge, CCard, CCardBody, CCardHeader, CCol, CContainer, CHeader, CHeaderToggler,
  CNavItem, CNavTitle, CRow, CSidebar, CSidebarBrand, CSidebarFooter, CSidebarHeader,
  CSidebarNav, CTable, CTableBody, CTableDataCell, CTableHead, CTableHeaderCell, CTableRow,
} from '@coreui/vue'
import { onBeforeUnmount, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ArrowLeft, Cloud, Gauge, Menu, Settings } from '@lucide/vue'

import { getModules, getPlatformStats, getPlatformStatus, type ModuleStatus, type PlatformStats, type PlatformStatus } from '../services/http'
import { useI18n } from '../i18n'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const auth = useAuthStore()
const { t } = useI18n()
const modules = ref<ModuleStatus[]>([])
const stats = ref<PlatformStats | null>(null)
const platform = ref<PlatformStatus | null>(null)
const error = ref('')
const sidebarVisible = ref(true)
let stylesheet: HTMLLinkElement | null = null

onMounted(async () => {
  stylesheet = document.createElement('link')
  stylesheet.rel = 'stylesheet'
  stylesheet.href = coreuiStylesheet
  stylesheet.dataset.waveAdminStyles = 'coreui'
  document.head.append(stylesheet)
  document.body.classList.add('coreui-admin-active')

  try {
    const [moduleResult, statsResult, platformResult] = await Promise.all([getModules(), getPlatformStats(), getPlatformStatus()])
    modules.value = moduleResult
    stats.value = statsResult
    platform.value = platformResult
  } catch (reason) {
    error.value = reason instanceof Error ? reason.message : 'Administration data could not be loaded.'
  }
})

onBeforeUnmount(() => {
  stylesheet?.remove()
  document.body.classList.remove('coreui-admin-active')
})

async function signOut() {
  await auth.signOut()
  await router.replace('/')
}
</script>

<template>
  <main class="coreui-admin-template">
    <CSidebar class="border-end" color-scheme="dark" position="fixed" :visible="sidebarVisible" @visible-change="sidebarVisible = $event">
      <CSidebarHeader class="border-bottom">
        <CSidebarBrand as="div" class="admin-brand">Wave <span>{{ t('admin.title') }}</span></CSidebarBrand>
      </CSidebarHeader>
      <CSidebarNav>
        <CNavItem>
          <RouterLink class="nav-link active" to="/admin"><Gauge class="nav-icon" :size="20" />{{ t('admin.overview') }}</RouterLink>
        </CNavItem>
        <CNavTitle>{{ t('admin.platform') }}</CNavTitle>
        <CNavItem><a class="nav-link" href="#modules"><Settings class="nav-icon" :size="20" />{{ t('admin.modules') }}</a></CNavItem>
      </CSidebarNav>
      <CSidebarFooter class="border-top">
        <RouterLink class="nav-link" to="/"><ArrowLeft class="nav-icon" :size="20" />{{ t('admin.backToPlatform') }}</RouterLink>
      </CSidebarFooter>
    </CSidebar>

    <div class="admin-wrapper wrapper d-flex flex-column min-vh-100">
      <CHeader position="sticky" class="mb-4 p-0 border-bottom">
        <CContainer class="px-4" fluid>
          <CHeaderToggler class="me-3" @click="sidebarVisible = !sidebarVisible">
            <Menu :size="22" />
          </CHeaderToggler>
          <strong>Wave Platform</strong>
          <div class="ms-auto d-flex align-items-center gap-3">
            <span class="text-body-secondary small">{{ auth.account?.displayName }}</span>
            <button class="btn btn-outline-secondary btn-sm" type="button" @click="signOut">{{ t('auth.signOut') }}</button>
          </div>
        </CContainer>
      </CHeader>

      <div class="body flex-grow-1">
        <CContainer class="px-4" lg>
          <div class="mb-4">
            <div class="text-body-secondary small mb-1">{{ t('admin.title') }} / {{ t('admin.overview') }}</div>
            <h1 class="h3 mb-1">{{ t('admin.platformOverview') }}</h1>
            <p class="text-body-secondary mb-0">{{ t('admin.lead') }}</p>
          </div>

          <div v-if="error" class="alert alert-danger" role="alert">{{ error }}</div>

          <CRow v-else :xs="{ cols: 1, gutter: 4 }" :md="{ cols: 3 }" class="mb-4">
            <CCol>
              <CCard class="h-100 border-start border-start-4 border-start-primary">
                <CCardBody><div class="text-body-secondary small">{{ t('admin.accounts') }}</div><div class="fs-3 fw-semibold">{{ stats?.accounts ?? '—' }}</div></CCardBody>
              </CCard>
            </CCol>
            <CCol>
              <CCard class="h-100 border-start border-start-4 border-start-info">
                <CCardBody><div class="text-body-secondary small">{{ t('admin.messagesToday') }}</div><div class="fs-3 fw-semibold">{{ stats?.messagesToday ?? '—' }}</div></CCardBody>
              </CCard>
            </CCol>
            <CCol>
              <CCard class="h-100 border-start border-start-4 border-start-success">
                <CCardBody><div class="text-body-secondary small">{{ t('admin.gitMirrors') }}</div><div class="fs-3 fw-semibold">{{ stats?.gitMirrors ?? '—' }}</div></CCardBody>
              </CCard>
            </CCol>
          </CRow>

          <CCard id="modules" class="mb-4">
            <CCardHeader class="d-flex align-items-center justify-content-between">
              <strong>{{ t('admin.modules') }}</strong>
              <span class="text-body-secondary small">{{ platform?.environment }} · {{ platform?.version }}</span>
            </CCardHeader>
            <CCardBody class="p-0">
              <CTable align="middle" class="mb-0" hover responsive>
                <CTableHead><CTableRow><CTableHeaderCell class="ps-4">{{ t('admin.module') }}</CTableHeaderCell><CTableHeaderCell>{{ t('admin.status') }}</CTableHeaderCell><CTableHeaderCell>{{ t('admin.availability') }}</CTableHeaderCell></CTableRow></CTableHead>
                <CTableBody>
                  <CTableRow v-for="module in modules" :key="module.name">
                    <CTableDataCell class="ps-4"><Cloud class="me-2 text-body-secondary" :size="16" /><strong>{{ module.name }}</strong></CTableDataCell>
                    <CTableDataCell>{{ module.status }}</CTableDataCell>
                    <CTableDataCell><CBadge :color="module.enabled ? 'success' : 'secondary'">{{ module.enabled ? t('admin.enabled') : t('admin.disabled') }}</CBadge></CTableDataCell>
                  </CTableRow>
                </CTableBody>
              </CTable>
            </CCardBody>
          </CCard>
        </CContainer>
      </div>
    </div>
  </main>
</template>

<style>
body.coreui-admin-active { background-color: var(--cui-tertiary-bg); }
.coreui-admin-template { min-height: 100vh; font-family: var(--cui-body-font-family); }
.coreui-admin-template .wrapper { width: 100%; padding-inline: var(--cui-sidebar-occupy-start, 0) var(--cui-sidebar-occupy-end, 0); transition: padding .15s; }
.coreui-admin-template .header > .container-fluid,
.coreui-admin-template .sidebar-header { min-height: calc(4rem + 1px); }
.coreui-admin-template .header > .container-fluid { display: flex; }
.coreui-admin-template .admin-brand { display: flex; align-items: baseline; gap: .45rem; font-size: 1.1rem; font-weight: 700; }
.coreui-admin-template .admin-brand span { color: var(--cui-secondary-color); font-size: .72rem; font-weight: 500; }
.coreui-admin-template .sidebar-footer .nav-link { display: flex; align-items: center; color: var(--cui-sidebar-nav-link-color); text-decoration: none; }
.coreui-admin-template .body { padding-bottom: 2rem; }
@media (max-width: 991.98px) { .coreui-admin-template .wrapper { padding-inline-start: 0; } }
</style>
